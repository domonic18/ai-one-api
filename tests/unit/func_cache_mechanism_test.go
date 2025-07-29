package unit

import (
	"context"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/cache"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimpleCacheManager 测试简化的缓存管理器
func TestSimpleCacheManager(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 初始化缓存管理器
	cache.InitManager()
	require.NotNil(t, cache.Mgr)

	ctx := context.Background()

	t.Run("基本缓存操作测试", func(t *testing.T) {
		testKey := "test_key"
		testValue := "test_value"

		// 设置缓存应该失败（Redis未启用）
		err := cache.Mgr.Set(ctx, testKey, testValue, time.Hour)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "缓存未启用")

		// 获取缓存应该失败
		var result string
		err = cache.Mgr.Get(ctx, testKey, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "缓存未启用")

		// 删除缓存应该失败
		err = cache.Mgr.Delete(ctx, testKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "缓存未启用")
	})

	t.Run("GetWithFallback测试", func(t *testing.T) {
		testKey := "fallback_key"
		expectedValue := "fallback_value"

		var result string
		err := cache.Mgr.GetWithFallback(ctx, testKey, func() (interface{}, error) {
			return expectedValue, nil
		}, time.Hour, &result)

		// Redis未启用时，应该直接调用fallback函数
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, result)
	})

	t.Run("统计信息测试", func(t *testing.T) {
		stats := cache.Mgr.GetStats()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.MissCount, int64(0))
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})

	t.Run("批量删除测试", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		result := cache.Mgr.BatchDelete(ctx, keys)
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 3, result.FailedCount)
		assert.Len(t, result.FailedKeys, 3)
	})

	t.Run("缓存状态检查测试", func(t *testing.T) {
		assert.False(t, cache.Mgr.IsEnabled())
		assert.False(t, cache.Mgr.Exists(ctx, "non_existent_key"))

		results := cache.Mgr.GetMultiple(ctx, []string{"key1", "key2"})
		assert.Empty(t, results)
	})
}

// TestSimpleInvalidationManager 测试简化的缓存失效管理器
func TestSimpleInvalidationManager(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 初始化缓存失效管理器
	cache.InitInvalidationManager()
	require.NotNil(t, cache.Invalidator)

	ctx := context.Background()

	t.Run("按键失效测试", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		err := cache.Invalidator.InvalidateByKeys(ctx, keys)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("用户缓存失效测试", func(t *testing.T) {
		err := cache.Invalidator.InvalidateUserCache(ctx, "user123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空用户ID测试
		err = cache.Invalidator.InvalidateUserCache(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userId不能为空")
	})

	t.Run("学科组缓存失效测试", func(t *testing.T) {
		err := cache.Invalidator.InvalidateSubjectCache(ctx, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 无效学科组ID测试
		err = cache.Invalidator.InvalidateSubjectCache(ctx, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "subjectId必须大于0")
	})

	t.Run("缓存刷新测试", func(t *testing.T) {
		err := cache.Invalidator.RefreshCache(ctx, "test_key", "test_value", time.Hour)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})
}

// TestSimpleMonitor 测试简化的缓存监控器
func TestSimpleMonitor(t *testing.T) {
	// 初始化缓存监控器
	cache.InitMonitor()
	require.NotNil(t, cache.MonitorInstance)

	t.Run("监控器基本功能测试", func(t *testing.T) {
		// 检查初始状态
		assert.True(t, cache.MonitorInstance.IsMonitoring())

		// 获取健康状态
		health := cache.MonitorInstance.GetHealthStatus()
		assert.NotNil(t, health)
		assert.Contains(t, health, "is_healthy")
		assert.Contains(t, health, "redis_enabled")

		// 获取性能指标
		metrics := cache.MonitorInstance.GetPerformanceMetrics()
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "total_requests")
		assert.Contains(t, metrics, "hit_rate")
	})
}

// TestUserModelConfigCache 测试用户模型配置缓存（简化版）
func TestUserModelConfigCache(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("缓存常量测试", func(t *testing.T) {
		assert.Equal(t, "user_config:", smart.UserConfigCachePrefix)
		assert.Equal(t, 24*time.Hour, smart.UserConfigCacheTTL)
	})

	t.Run("获取用户模型配置测试", func(t *testing.T) {
		userId := "test_user_123"

		config, err := smart.GetUserConfigWithCache(ctx, userId)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效用户模型配置缓存测试", func(t *testing.T) {
		userId := "test_user_123"

		err := smart.InvalidateUserConfigCache(ctx, userId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("批量失效测试", func(t *testing.T) {
		userIds := []string{"user1", "user2", "user3"}

		err := smart.BatchInvalidateUserConfigCache(ctx, userIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表测试
		err = smart.BatchInvalidateUserConfigCache(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("预加载测试", func(t *testing.T) {
		userIds := []string{"user1", "user2"}

		err := smart.PreloadUserConfigs(ctx, userIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表测试
		err = smart.PreloadUserConfigs(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("缓存刷新测试", func(t *testing.T) {
		userId := "test_user_123"

		err := smart.RefreshUserConfigCache(ctx, userId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("获取缓存统计信息测试", func(t *testing.T) {
		stats := smart.GetUserConfigCacheStats()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})
}

// TestSubjectModelCache 测试学科组模型缓存（简化版）
func TestSubjectModelCache(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("获取老师信息测试", func(t *testing.T) {
		teacherId := "teacher_123"

		info, err := smart.GetTeacherInfoWithCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("失效老师信息缓存测试", func(t *testing.T) {
		teacherId := "teacher_123"

		err := smart.InvalidateTeacherInfoCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("批量失效老师信息缓存测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		err := smart.BatchInvalidateTeacherInfoCache(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表测试
		err = smart.BatchInvalidateTeacherInfoCache(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("预加载老师信息测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		err := smart.PreloadTeacherInfos(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表测试
		err = smart.PreloadTeacherInfos(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("获取缓存统计测试", func(t *testing.T) {
		stats := smart.GetSubjectInfoCacheStats()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})
}

// TestCacheIntegration 测试缓存集成功能（简化版）
func TestCacheIntegration(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	t.Run("缓存组件初始化测试", func(t *testing.T) {
		cache.InitManager()
		assert.NotNil(t, cache.Mgr)

		cache.InitInvalidationManager()
		assert.NotNil(t, cache.Invalidator)

		cache.InitMonitor()
		assert.NotNil(t, cache.MonitorInstance)
	})

	t.Run("缓存管理器状态测试", func(t *testing.T) {
		assert.False(t, cache.Mgr.IsEnabled())
		assert.True(t, cache.MonitorInstance.IsMonitoring())
	})
}
