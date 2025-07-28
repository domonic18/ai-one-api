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
	// 用户模型配置缓存前缀
	UserConfigCachePrefix = "user_config:"
	// 用户模型配置缓存时间（24小时）
	UserConfigCacheTTL = 24 * time.Hour
)

// GetUserConfigWithCache 获取用户模型配置（带缓存）
func GetUserConfigWithCache(ctx context.Context, userId string) (*UserConfig, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 1. 尝试从缓存管理器获取
	cacheKey := fmt.Sprintf("%s%s", UserConfigCachePrefix, userId)
	var config UserConfig

	err := cache.Mgr.Get(ctx, cacheKey, &config)
	if err == nil {
		logger.Debugf(ctx, "用户模型配置缓存命中: userId=%s, model=%s", userId, config.ModelName)
		return &config, nil
	}

	// 2. 缓存未命中，从API获取
	apiConfig, err := GetUserConfigFromAPI(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "从API获取用户模型配置失败: userId=%s, error=%v", userId, err)
		return nil, err
	}

	// 3. 使用缓存管理器更新缓存
	err = cache.Mgr.Set(ctx, cacheKey, apiConfig, UserConfigCacheTTL)
	if err != nil {
		logger.Warnf(ctx, "设置用户模型配置缓存失败: userId=%s, error=%v", userId, err)
	} else {
		logger.Debugf(ctx, "用户模型配置已缓存: userId=%s, model=%s", userId, apiConfig.ModelName)
	}

	return apiConfig, nil
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

	logger.Debugf(ctx, "从API获取用户模型配置成功: userId=%s, model=%s",
		userId, userConfig.ModelName)

	return userConfig, nil
}

// InvalidateUserConfigCache 使指定用户的模型配置缓存失效
func InvalidateUserConfigCache(ctx context.Context, userId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%s", UserConfigCachePrefix, userId)
	err := cache.Mgr.Delete(ctx, cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除用户模型配置缓存失败: userId=%s, error=%v", userId, err)
		return err
	}

	logger.Debugf(ctx, "用户模型配置缓存已删除: userId=%s", userId)
	return nil
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

	// 使用缓存管理器批量删除
	result := cache.Mgr.BatchDelete(ctx, cacheKeys)

	if result.FailedCount > 0 {
		logger.Warnf(ctx, "批量删除用户模型配置缓存部分失败: 成功=%d, 失败=%d",
			result.SuccessCount, result.FailedCount)
	}

	logger.Debugf(ctx, "批量删除用户模型配置缓存完成: 成功=%d, 失败=%d",
		result.SuccessCount, result.FailedCount)

	return nil
}

// PreloadUserConfigs 预加载用户模型配置到缓存
func PreloadUserConfigs(ctx context.Context, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 构建缓存项列表
	cacheItems := make([]cache.Item, 0, len(userIds))

	for _, userId := range userIds {
		// 从API获取配置
		config, err := GetUserConfigFromAPI(ctx, userId)
		if err != nil {
			logger.Warnf(ctx, "预加载用户模型配置失败: userId=%s, error=%v", userId, err)
			continue
		}

		cacheItem := cache.Item{
			Key:        fmt.Sprintf("%s%s", UserConfigCachePrefix, userId),
			Value:      config,
			Expiration: UserConfigCacheTTL,
			CreatedAt:  time.Now(),
		}
		cacheItems = append(cacheItems, cacheItem)
	}

	// 批量设置缓存
	result := cache.Mgr.BatchSet(ctx, cacheItems)

	logger.Debugf(ctx, "预加载用户模型配置完成: 请求=%d, 成功=%d, 失败=%d",
		len(userIds), result.SuccessCount, result.FailedCount)

	return nil
}

// RefreshUserConfigCache 刷新用户模型配置缓存
func RefreshUserConfigCache(ctx context.Context, userId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 先删除旧缓存
	err := InvalidateUserConfigCache(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "删除旧缓存失败: userId=%s, error=%v", userId, err)
	}

	// 重新获取并缓存
	_, err = GetUserConfigWithCache(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "刷新用户模型配置缓存失败: userId=%s, error=%v", userId, err)
		return err
	}

	logger.Debugf(ctx, "用户模型配置缓存已刷新: userId=%s", userId)
	return nil
}

// GetUserConfigCacheStats 获取用户模型配置缓存统计信息
func GetUserConfigCacheStats() *cache.Stats {
	if cache.Mgr == nil {
		return &cache.Stats{}
	}

	return cache.Mgr.GetStats()
}
