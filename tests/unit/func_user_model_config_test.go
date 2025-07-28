package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

func TestUserModelConfig(t *testing.T) {
	t.Run("用户模型配置结构体测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "teacher_001",
			ModelName: "gpt-4-turbo",
			UpdatedAt: time.Now().Unix(),
			Source:    "user_preference",
		}

		assert.Equal(t, "teacher_001", config.UserId)
		assert.Equal(t, "gpt-4-turbo", config.ModelName)
		assert.Equal(t, "user_preference", config.Source)
		assert.True(t, config.UpdatedAt > 0)
	})

	t.Run("用户模型配置JSON序列化测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "teacher_002",
			ModelName: "claude-3-sonnet",
			UpdatedAt: time.Now().Unix(),
			Source:    "admin_setting",
		}

		jsonData, err := json.Marshal(config)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var decodedConfig model.UserModelConfig
		err = json.Unmarshal(jsonData, &decodedConfig)
		assert.NoError(t, err)
		assert.Equal(t, config.UserId, decodedConfig.UserId)
		assert.Equal(t, config.ModelName, decodedConfig.ModelName)
		assert.Equal(t, config.UpdatedAt, decodedConfig.UpdatedAt)
		assert.Equal(t, config.Source, decodedConfig.Source)
	})

	t.Run("用户模型配置默认值测试", func(t *testing.T) {
		config := &model.UserModelConfig{}

		assert.Equal(t, "", config.UserId)
		assert.Equal(t, "", config.ModelName)
		assert.Equal(t, int64(0), config.UpdatedAt)
		assert.Equal(t, "", config.Source)
	})
}

func TestUserModelConfigCacheConstants(t *testing.T) {
	t.Run("缓存前缀和时间常量测试", func(t *testing.T) {
		// 验证缓存前缀格式
		assert.Equal(t, "user_model_config:", model.UserModelConfigCachePrefix)

		// 验证缓存时间是否为24小时
		expectedTTL := 24 * time.Hour
		assert.Equal(t, expectedTTL, model.UserModelConfigCacheTTL)
	})
}

func TestGetUserModelConfigFromAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("API调用未实现测试", func(t *testing.T) {
		// 当前API调用函数返回nil，这是预期的行为
		config, err := model.GetUserModelConfigFromAPI(ctx, "teacher_001")

		// 应该返回nil，表示没有找到配置
		assert.Nil(t, config)
		assert.NoError(t, err)
	})
}

func TestGetUserModelConfigWithCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		userId := "teacher_001"
		config, err := model.GetUserModelConfigWithCache(ctx, userId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

func TestInvalidateUserModelConfigCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Redis未启用测试", func(t *testing.T) {
		// 保存原始值
		originalRedisEnabled := common.RedisEnabled
		// 测试结束后恢复
		defer func() { common.RedisEnabled = originalRedisEnabled }()

		// 设置Redis为未启用
		common.RedisEnabled = false

		userId := "teacher_001"
		err := model.InvalidateUserModelConfigCache(ctx, userId)

		// 应该返回错误，表示Redis未启用
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

func TestUserModelConfigValidation(t *testing.T) {
	t.Run("有效配置测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "teacher_001",
			ModelName: "gpt-4-turbo",
			UpdatedAt: time.Now().Unix(),
			Source:    "user_preference",
		}

		// 验证配置有效性
		assert.NotEmpty(t, config.UserId)
		assert.NotEmpty(t, config.ModelName)
		assert.True(t, config.UpdatedAt > 0)
		assert.NotEmpty(t, config.Source)
	})

	t.Run("无效配置测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "",
			ModelName: "",
			UpdatedAt: 0,
			Source:    "",
		}

		// 验证配置无效性
		assert.Empty(t, config.UserId)
		assert.Empty(t, config.ModelName)
		assert.Equal(t, int64(0), config.UpdatedAt)
		assert.Empty(t, config.Source)
	})
}

func TestUserModelConfigEdgeCases(t *testing.T) {
	t.Run("特殊字符用户ID测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "teacher@school.edu",
			ModelName: "gpt-3.5-turbo",
			UpdatedAt: time.Now().Unix(),
			Source:    "api_setting",
		}

		jsonData, err := json.Marshal(config)
		assert.NoError(t, err)

		var decodedConfig model.UserModelConfig
		err = json.Unmarshal(jsonData, &decodedConfig)
		assert.NoError(t, err)
		assert.Equal(t, config.UserId, decodedConfig.UserId)
	})

	t.Run("长模型名称测试", func(t *testing.T) {
		longModelName := "very-long-model-name-that-might-cause-issues-in-some-systems"
		config := &model.UserModelConfig{
			UserId:    "teacher_001",
			ModelName: longModelName,
			UpdatedAt: time.Now().Unix(),
			Source:    "admin_override",
		}

		jsonData, err := json.Marshal(config)
		assert.NoError(t, err)

		var decodedConfig model.UserModelConfig
		err = json.Unmarshal(jsonData, &decodedConfig)
		assert.NoError(t, err)
		assert.Equal(t, longModelName, decodedConfig.ModelName)
	})

	t.Run("时间戳边界值测试", func(t *testing.T) {
		config := &model.UserModelConfig{
			UserId:    "teacher_001",
			ModelName: "gpt-4",
			UpdatedAt: 0, // 最小时间戳
			Source:    "system_default",
		}

		jsonData, err := json.Marshal(config)
		assert.NoError(t, err)

		var decodedConfig model.UserModelConfig
		err = json.Unmarshal(jsonData, &decodedConfig)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), decodedConfig.UpdatedAt)
	})
}
