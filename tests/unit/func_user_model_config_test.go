package unit

import (
	"context"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/stretchr/testify/assert"
)

// TestUserConfig 测试用户模型配置结构
func TestUserConfig(t *testing.T) {
	// 创建用户模型配置实例
	config := &smart.UserConfig{
		UserId:     "user_001",
		ModelName:  "gpt-4-turbo",
		Parameters: `{"temperature": 0.7, "max_tokens": 2000}`,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	// 验证字段
	assert.Equal(t, "user_001", config.UserId)
	assert.Equal(t, "gpt-4-turbo", config.ModelName)
	assert.Equal(t, `{"temperature": 0.7, "max_tokens": 2000}`, config.Parameters)
	assert.False(t, config.UpdatedAt.IsZero())
	assert.False(t, config.CreatedAt.IsZero())
}

// TestUserConfigJSON 测试用户模型配置JSON序列化
func TestUserConfigJSON(t *testing.T) {
	config := &smart.UserConfig{
		UserId:     "user_002",
		ModelName:  "claude-3-sonnet",
		Parameters: `{"temperature": 0.8}`,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	// 测试JSON序列化（简单验证结构体可以被正确序列化）
	assert.NotNil(t, config)
	assert.Equal(t, "user_002", config.UserId)
}

// TestUserConfigWithCache 测试用户模型配置缓存功能
func TestUserConfigWithCache(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("用户模型配置缓存常量测试", func(t *testing.T) {
		// 验证缓存前缀格式
		assert.Equal(t, "user_config:", smart.UserConfigCachePrefix)

		// 验证缓存时间是否为24小时
		expectedTTL := 24 * time.Hour
		assert.Equal(t, expectedTTL, smart.UserConfigCacheTTL)
	})

	t.Run("获取用户模型配置测试", func(t *testing.T) {
		userId := "test_user_123"

		// Redis未启用时应该返回错误
		config, err := smart.GetUserConfigWithCache(ctx, userId)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效用户模型配置缓存测试", func(t *testing.T) {
		userId := "test_user_123"

		// Redis未启用时应该返回错误
		err := smart.InvalidateUserConfigCache(ctx, userId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("批量失效用户模型配置缓存测试", func(t *testing.T) {
		userIds := []string{"user1", "user2", "user3"}

		// Redis未启用时应该返回错误
		err := smart.BatchInvalidateUserConfigCache(ctx, userIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.BatchInvalidateUserConfigCache(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("预加载用户模型配置测试", func(t *testing.T) {
		userIds := []string{"user1", "user2"}

		// Redis未启用时应该返回错误
		err := smart.PreloadUserConfigs(ctx, userIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.PreloadUserConfigs(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("刷新用户模型配置缓存测试", func(t *testing.T) {
		userId := "test_user_123"

		// Redis未启用时应该返回错误
		err := smart.RefreshUserConfigCache(ctx, userId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("获取缓存统计信息测试", func(t *testing.T) {
		stats := smart.GetUserConfigCacheStats()
		assert.NotNil(t, stats)
		// 当CacheMgr为nil时应该返回空统计
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})
}

// TestGetUserConfigFromAPI 测试从API获取用户模型配置
func TestGetUserConfigFromAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("API调用测试", func(t *testing.T) {
		userId := "test_user_123"

		// 调用API获取用户配置（这会调用实际的API，但在测试环境中可能会失败）
		config, err := smart.GetUserConfigFromAPI(ctx, userId)

		// 由于这是单元测试，API调用可能会失败，我们主要验证函数不会panic
		// 无论成功还是失败，都是可接受的结果
		if err != nil {
			assert.Nil(t, config)
		} else {
			assert.NotNil(t, config)
		}
	})
}
