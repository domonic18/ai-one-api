package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// UserModelConfig 用户模型配置，用于智能模型选择
type UserModelConfig struct {
	UserId    string `json:"user_id"`    // 用户ID（老师ID）
	ModelName string `json:"model_name"` // 模型名称
	UpdatedAt int64  `json:"updated_at"` // 更新时间
	Source    string `json:"source"`     // 配置来源：user_preference, subject_default, school_default
}

// 缓存相关常量
const (
	// 用户模型配置缓存前缀
	UserModelConfigCachePrefix = "user_model_config:"
	// 用户模型配置缓存时间（24小时）
	UserModelConfigCacheTTL = 24 * time.Hour
)

// GetUserModelConfigWithCache 获取用户模型配置（带缓存）
func GetUserModelConfigWithCache(ctx context.Context, userId string) (*UserModelConfig, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 1. 尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("%s%s", UserModelConfigCachePrefix, userId)
	configStr, err := common.RedisGet(cacheKey)

	if err == nil && configStr != "" {
		// 缓存命中，解析并返回
		var config UserModelConfig
		err = json.Unmarshal([]byte(configStr), &config)
		if err == nil {
			logger.Debugf(ctx, "用户模型配置缓存命中: userId=%s, model=%s", userId, config.ModelName)
			return &config, nil
		}
		logger.Warnf(ctx, "用户模型配置缓存解析失败: userId=%s, error=%v", userId, err)
	}

	// 2. 缓存未命中，从API获取
	config, err := GetUserModelConfigFromAPI(ctx, userId)
	if err != nil {
		logger.Warnf(ctx, "获取用户模型配置失败: userId=%s, error=%v", userId, err)
		return nil, err
	}

	// 3. 更新缓存
	if config != nil {
		configJson, _ := json.Marshal(config)
		common.RedisSet(cacheKey, string(configJson), UserModelConfigCacheTTL)
		logger.Debugf(ctx, "用户模型配置已缓存: userId=%s, model=%s", userId, config.ModelName)
	}

	return config, nil
}

// GetUserModelConfigFromAPI 从课件平台API获取用户模型配置
var GetUserModelConfigFromAPI = func(ctx context.Context, userId string) (*UserModelConfig, error) {
	// TODO: 实现从课件平台API获取用户模型配置
	// 这里需要根据实际API调用实现
	// 当前返回空实现，后续集成实际API调用
	logger.Warnf(ctx, "GetUserModelConfigFromAPI未实现实际API调用: userId=%s", userId)

	// 临时返回nil，表示没有找到配置
	return nil, nil
}

// InvalidateUserModelConfigCache 使指定用户的模型配置缓存失效
func InvalidateUserModelConfigCache(ctx context.Context, userId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%s", UserModelConfigCachePrefix, userId)
	err := common.RedisDel(cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除用户模型配置缓存失败: userId=%s, error=%v", userId, err)
		return err
	}
	logger.Debugf(ctx, "用户模型配置缓存已删除: userId=%s", userId)
	return nil
}

// BatchInvalidateUserModelConfigCache 批量使用户模型配置缓存失效
func BatchInvalidateUserModelConfigCache(ctx context.Context, userIds []string) error {
	// 先检查是否为空列表
	if len(userIds) == 0 {
		return nil
	}

	// 再检查Redis是否启用
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	for _, userId := range userIds {
		cacheKey := fmt.Sprintf("%s%s", UserModelConfigCachePrefix, userId)
		err := common.RedisDel(cacheKey)
		if err != nil {
			logger.Warnf(ctx, "删除用户模型配置缓存失败: userId=%s, error=%v", userId, err)
			// 继续处理其他用户，不中断
		}
	}

	logger.Debugf(ctx, "批量删除用户模型配置缓存成功: count=%d", len(userIds))
	return nil
}
