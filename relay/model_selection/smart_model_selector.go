package model_selection

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/constant"
)

// UserModelConfig 用户模型配置
type UserModelConfig struct {
	UserID     string                 `json:"user_id"`
	ModelName  string                 `json:"model_name"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SmartModelSelector 智能模型选择器
type SmartModelSelector struct{}

// NewSmartModelSelector 创建一个新的智能模型选择器
func NewSmartModelSelector() *SmartModelSelector {
	return &SmartModelSelector{}
}

// GetUserModelConfig 从Redis获取用户配置的模型信息
func (s *SmartModelSelector) GetUserModelConfig(userID string) (*UserModelConfig, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	key := constant.UserModelConfigKeyPrefix + userID
	configStr, err := common.RedisGet(key)
	if err != nil {
		return nil, fmt.Errorf("从Redis获取用户模型配置失败: %w", err)
	}

	var config UserModelConfig
	err = json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return nil, fmt.Errorf("解析用户模型配置失败: %w", err)
	}

	return &config, nil
}

// SetUserModelConfig 将用户配置的模型信息存储到Redis
func (s *SmartModelSelector) SetUserModelConfig(userID string, config *UserModelConfig) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化用户模型配置失败: %w", err)
	}

	key := constant.UserModelConfigKeyPrefix + userID
	// 设置过期时间为30天
	err = common.RedisSet(key, string(configBytes), 30*24*time.Hour)
	if err != nil {
		return fmt.Errorf("向Redis存储用户模型配置失败: %w", err)
	}

	logger.SysLog(fmt.Sprintf("成功为用户 %s 设置智能模型选择配置: %s", userID, config.ModelName))
	return nil
}

// DeleteUserModelConfig 从Redis删除用户配置的模型信息
func (s *SmartModelSelector) DeleteUserModelConfig(userID string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	key := constant.UserModelConfigKeyPrefix + userID
	err := common.RedisDel(key)
	if err != nil {
		return fmt.Errorf("从Redis删除用户模型配置失败: %w", err)
	}

	logger.SysLog(fmt.Sprintf("成功删除用户 %s 的智能模型选择配置", userID))
	return nil
}

// IsSmartSelectModel 检查是否为智能选择模型
func IsSmartSelectModel(modelName string) bool {
	return modelName == constant.SmartSelect
}
