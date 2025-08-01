package identity

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/logger"
)

// InitializeIdentityResolver 初始化身份解析器
// OneAPI系统启动时的初始化
func InitializeIdentityResolver() {
	// 1. 读取配置
	config := loadCoursewareConfig() // 可选配置

	if config != nil && config.Enabled {
		logger.SysLog("课件平台集成已启用，初始化身份解析器")

		// 2. 初始化Redis客户端
		redisClient := initRedisClient()
		if redisClient == nil {
			logger.SysLog("Redis客户端初始化失败，回退到默认模式")
			SetIdentityResolver(&DefaultIdentityResolver{})
			return
		}

		// 3. 创建缓存管理器
		cache := NewCoursewareCache(redisClient, config)

		// 4. 创建 API 客户端
		apiClient := client.GetCoursewareClient()

		// 5. 验证API客户端是否有效
		if apiClient == nil || !apiClient.IsValid() {
			logger.SysLog("课件平台API客户端初始化失败，回退到默认模式")
			SetIdentityResolver(&DefaultIdentityResolver{})
			return
		}

		// 6. 创建身份解析器
		resolver := NewCoursewareIdentityResolver(apiClient, cache, config)

		// 7. 设置为全局解析器
		SetIdentityResolver(resolver)

		// 8. 创建预加载管理器
		preloadManager := NewPreloadManager(apiClient, cache, config)

		// 9. 执行初始预加载（在goroutine中执行，避免阻塞启动）
		ctx := context.Background()
		go func() {
			logger.SysLog("开始执行初始预加载")
			if err := preloadManager.PreloadUserInfos(ctx); err != nil {
				logger.SysLog("初始预加载失败: " + err.Error())
			} else {
				logger.SysLog("初始预加载完成")
			}
		}()

		// 10. 启动定期刷新任务
		preloadManager.StartPeriodicRefresh(ctx)

		logger.SysLog("课件平台身份解析器初始化完成")
	} else {
		// 11. 课件平台集成未启用，使用默认解析器
		SetIdentityResolver(&DefaultIdentityResolver{})

		logger.SysLog("课件平台集成未启用，OneAPI将保持原有功能")
	}
}

// initRedisClient 初始化Redis客户端
func initRedisClient() RedisClient {
	// 使用OneAPI现有的Redis配置
	if !common.RedisEnabled {
		logger.SysLog("Redis未启用，无法使用课件平台集成")
		return nil
	}

	// 复用OneAPI的Redis连接
	if common.RDB == nil {
		logger.SysLog("Redis客户端未初始化，无法使用课件平台集成")
		return nil
	}

	// 将common.RDB转换为RedisClient接口
	if client, ok := common.RDB.(RedisClient); ok {
		return client
	}

	logger.SysLog("Redis客户端类型不支持，无法使用课件平台集成")
	return nil
}

// loadCoursewareConfig 加载课件平台配置
func loadCoursewareConfig() *CoursewareConfig {
	// 从环境变量加载配置
	enabled := os.Getenv("COURSEWARE_ENABLED") == "true"
	if !enabled {
		logger.SysLog("课件平台集成未启用 (COURSEWARE_ENABLED != true)")
		return nil
	}

	baseURL := os.Getenv("COURSEWARE_PLATFORM_BASE_URL")
	apiKey := os.Getenv("COURSEWARE_PLATFORM_API_KEY")

	// 添加调试日志
	logger.SysLogf("调试: 从环境变量读取 baseURL=%s, apiKey=%s", baseURL, apiKey)

	if baseURL == "" || apiKey == "" {
		logger.SysLog("课件平台配置不完整，禁用集成 (baseURL或apiKey为空)")
		return nil
	}

	// 解析超时配置
	timeoutStr := os.Getenv("COURSEWARE_PLATFORM_TIMEOUT")
	timeout := 5 * time.Second // 默认超时
	if timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			timeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	// 解析缓存TTL
	cacheTTLStr := os.Getenv("COURSEWARE_CACHE_TTL")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		cacheTTL = 10 * time.Minute // 默认缓存时间
	}

	// 解析刷新间隔
	refreshIntervalStr := os.Getenv("COURSEWARE_REFRESH_INTERVAL")
	refreshInterval, err := time.ParseDuration(refreshIntervalStr)
	if err != nil {
		refreshInterval = 1 * time.Hour // 默认1小时刷新一次
	}

	// 解析预加载批次大小
	preloadBatchSize := 100 // 默认批次大小
	if batchSizeStr := os.Getenv("COURSEWARE_PRELOAD_BATCH_SIZE"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil && batchSize > 0 {
			preloadBatchSize = batchSize
		}
	}

	// 解析默认分组
	defaultGroup := os.Getenv("COURSEWARE_DEFAULT_GROUP")
	if defaultGroup == "" {
		defaultGroup = "default"
	}

	logger.SysLogf("课件平台配置加载成功: baseURL=%s, timeout=%v, cacheTTL=%v", baseURL, timeout, cacheTTL)

	return &CoursewareConfig{
		Enabled:          true,
		BaseURL:          baseURL,
		APIKey:           apiKey,
		Timeout:          timeout,
		CacheTTL:         cacheTTL,
		DefaultGroup:     defaultGroup,
		PreloadBatchSize: preloadBatchSize,
		RefreshInterval:  refreshInterval,
	}
}
