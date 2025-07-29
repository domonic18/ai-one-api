package smart

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/cache"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/logger"
)

// UserConfig 用户模型配置结构
type UserConfig struct {
	UserId     string    `json:"user_id"`
	ModelName  string    `json:"model_name"`
	Parameters string    `json:"parameters,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// 缓存相关常量
const (
	UserConfigCachePrefix = "user_config:"
	UserConfigCacheTTL    = 24 * time.Hour
)

// GetUserConfigWithCache 获取用户模型配置（使用简化的缓存接口）
func GetUserConfigWithCache(ctx context.Context, userId string) (*UserConfig, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%s", UserConfigCachePrefix, userId)
	var config UserConfig

	// 使用简化的GetWithFallback方法
	err := cache.Mgr.GetWithFallback(ctx, cacheKey, func() (interface{}, error) {
		return GetUserConfigFromAPI(ctx, userId)
	}, UserConfigCacheTTL, &config)

	if err != nil {
		logger.Warnf(ctx, "获取用户模型配置失败: userId=%s, error=%v", userId, err)
		return nil, err
	}

	logger.Debugf(ctx, "用户模型配置获取成功: userId=%s, model=%s", userId, config.ModelName)
	return &config, nil
}

// GetUserConfigFromAPI 从课件平台API获取用户模型配置
func GetUserConfigFromAPI(ctx context.Context, userId string) (*UserConfig, error) {
	// 使用课件平台API客户端获取用户模型配置
	coursewareClient := client.GetCoursewareClient()
	if coursewareClient == nil {
		logger.Warnf(ctx, "课件平台API客户端未初始化: userId=%s", userId)
		return nil, fmt.Errorf("课件平台API客户端未初始化")
	}

	// 调用API获取用户模型配置
	apiConfig, err := coursewareClient.GetUserConfig(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "从API获取用户模型配置失败: userId=%s, error=%v", userId, err)
		return nil, err
	}

	// 如果API返回空，则返回默认配置
	if apiConfig == nil {
		return &UserConfig{
			UserId:    userId,
			ModelName: "gpt-3.5-turbo", // 默认模型
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	// 转换为内部UserConfig结构
	userConfig := &UserConfig{
		UserId:     apiConfig.UserId,
		ModelName:  apiConfig.ModelName,
		Parameters: apiConfig.Parameters,
		CreatedAt:  time.Unix(apiConfig.CreatedAt, 0),
		UpdatedAt:  time.Unix(apiConfig.UpdatedAt, 0),
	}

	logger.Debugf(ctx, "从API获取用户模型配置成功: userId=%s, model=%s", userId, userConfig.ModelName)
	return userConfig, nil
}

// InvalidateUserConfigCache 使指定用户的模型配置缓存失效
func InvalidateUserConfigCache(ctx context.Context, userId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 使用简化的失效管理器
	return cache.Invalidator.InvalidateUserCache(ctx, userId)
}

// BatchInvalidateUserConfigCache 批量使用户模型配置缓存失效
func BatchInvalidateUserConfigCache(ctx context.Context, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 构建缓存键列表
	cacheKeys := make([]string, len(userIds))
	for i, userId := range userIds {
		cacheKeys[i] = fmt.Sprintf("%s%s", UserConfigCachePrefix, userId)
	}

	return cache.Invalidator.InvalidateByKeys(ctx, cacheKeys)
}

// PreloadUserConfigs 预加载用户模型配置到缓存
func PreloadUserConfigs(ctx context.Context, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	successCount := 0
	for _, userId := range userIds {
		_, err := GetUserConfigWithCache(ctx, userId)
		if err != nil {
			logger.Warnf(ctx, "预加载用户模型配置失败: userId=%s, error=%v", userId, err)
		} else {
			successCount++
		}
	}

	logger.Debugf(ctx, "预加载用户模型配置完成: 请求=%d, 成功=%d", len(userIds), successCount)
	return nil
}

// RefreshUserConfigCache 刷新用户模型配置缓存
func RefreshUserConfigCache(ctx context.Context, userId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 从API获取最新配置
	config, err := GetUserConfigFromAPI(ctx, userId)
	if err != nil {
		return err
	}

	// 使用简化的刷新方法
	cacheKey := fmt.Sprintf("%s%s", UserConfigCachePrefix, userId)
	return cache.Invalidator.RefreshCache(ctx, cacheKey, config, UserConfigCacheTTL)
}

// GetUserConfigCacheStats 获取用户配置缓存统计信息
func GetUserConfigCacheStats() *cache.SimpleStats {
	if cache.Mgr == nil {
		return &cache.SimpleStats{}
	}
	return cache.Mgr.GetStats()
}
