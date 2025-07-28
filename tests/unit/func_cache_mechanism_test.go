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

// TestCacheManager 测试缓存管理器功能
func TestCacheManager(t *testing.T) {
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

	t.Run("缓存管理器基本功能测试", func(t *testing.T) {
		// 测试缓存未启用时的行为
		testKey := "test_key"
		testValue := "test_value"

		// 设置缓存应该失败
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

	t.Run("缓存统计信息测试", func(t *testing.T) {
		stats := cache.Mgr.GetStats()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.MissCount, int64(0))
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})

	t.Run("批量操作测试", func(t *testing.T) {
		// 准备测试数据
		items := []cache.Item{
			{Key: "batch_key1", Value: "value1", Expiration: time.Hour},
			{Key: "batch_key2", Value: "value2", Expiration: time.Hour},
		}

		// 批量设置应该失败（Redis未启用）
		result := cache.Mgr.BatchSet(ctx, items)
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, result.FailedCount)
		assert.Len(t, result.FailedKeys, 2)

		// 批量删除应该失败（Redis未启用）
		keys := []string{"batch_key1", "batch_key2"}
		result = cache.Mgr.BatchDelete(ctx, keys)
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, result.FailedCount)
		assert.Len(t, result.FailedKeys, 2)
	})

	t.Run("缓存状态检查测试", func(t *testing.T) {
		// 检查缓存是否启用
		assert.False(t, cache.Mgr.IsEnabled())

		// 检查缓存项是否存在
		exists := cache.Mgr.Exists(ctx, "non_existent_key")
		assert.False(t, exists)

		// 批量获取缓存项
		keys := []string{"key1", "key2", "key3"}
		results := cache.Mgr.GetMultiple(ctx, keys)
		assert.Empty(t, results)
	})
}

// TestCacheInvalidationManager 测试缓存失效管理器
func TestCacheInvalidationManager(t *testing.T) {
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

	t.Run("失效模式管理测试", func(t *testing.T) {
		// 添加失效模式
		pattern := cache.InvalidationPattern{
			Pattern:     "test_pattern:*",
			TTL:         time.Hour,
			Description: "测试模式",
		}
		cache.Invalidator.AddInvalidationPattern(pattern)

		// 获取失效统计信息
		stats := cache.Invalidator.GetInvalidationStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "total_invalidations")
		assert.Contains(t, stats, "patterns")

		// 移除失效模式
		cache.Invalidator.RemoveInvalidationPattern("test_pattern:*")
	})

	t.Run("事件驱动失效测试", func(t *testing.T) {
		// 创建失效事件
		event := cache.InvalidationEvent{
			EventType:  "test_event",
			ObjectType: "teacher",
			ObjectID:   "teacher_123",
			ChangeType: "update",
			Timestamp:  time.Now().Unix(),
		}

		// 执行事件失效（Redis未启用时应该失败）
		err := cache.Invalidator.InvalidateByEvent(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("相关缓存失效测试", func(t *testing.T) {
		// 失效相关缓存（Redis未启用时应该失败）
		err := cache.Invalidator.InvalidateRelatedCaches(ctx, "teacher", "teacher_123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("最近失效检查测试", func(t *testing.T) {
		// 检查是否最近被失效过
		isRecent := cache.Invalidator.IsRecentlyInvalidated("test_key", time.Minute)
		assert.False(t, isRecent)
	})
}

// TestCacheMonitor 测试缓存监控器
func TestCacheMonitor(t *testing.T) {
	// 初始化缓存监控器
	cache.InitMonitor()
	require.NotNil(t, cache.MonitorInstance)

	ctx := context.Background()

	t.Run("监控器基本功能测试", func(t *testing.T) {
		// 检查初始状态
		assert.False(t, cache.MonitorInstance.IsMonitoring())

		// 获取性能指标
		metrics := cache.MonitorInstance.GetPerformanceMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.TotalRequests, int64(0))

		// 获取健康状态
		health := cache.MonitorInstance.GetHealthStatus()
		assert.NotNil(t, health)
		assert.GreaterOrEqual(t, health.OverallScore, 0.0)

		// 获取告警阈值
		thresholds := cache.MonitorInstance.GetAlertThresholds()
		assert.NotNil(t, thresholds)
		assert.Greater(t, thresholds.MinHitRate, 0.0)
	})

	t.Run("告警阈值设置测试", func(t *testing.T) {
		// 设置新的告警阈值
		newThresholds := &cache.AlertThresholds{
			MinHitRate:          90.0,
			MaxResponseTime:     200.0,
			MaxErrorRate:        5.0,
			HealthCheckInterval: 30,
		}

		cache.MonitorInstance.SetAlertThresholds(newThresholds)

		// 验证设置是否生效
		thresholds := cache.MonitorInstance.GetAlertThresholds()
		assert.Equal(t, 90.0, thresholds.MinHitRate)
		assert.Equal(t, 200.0, thresholds.MaxResponseTime)
		assert.Equal(t, 5.0, thresholds.MaxErrorRate)
		assert.Equal(t, int64(30), thresholds.HealthCheckInterval)
	})

	t.Run("性能测试功能测试", func(t *testing.T) {
		// 准备测试配置
		testConfig := &cache.TestConfig{
			Operations:  10,
			Concurrency: 1,
		}

		// 运行性能测试
		result := cache.MonitorInstance.RunPerformanceTest(ctx, testConfig)
		assert.NotNil(t, result)
		assert.Equal(t, testConfig.Operations, len(result.Operations))
		assert.GreaterOrEqual(t, result.SuccessRate, 0.0)
		assert.LessOrEqual(t, result.SuccessRate, 100.0)
	})
}

// TestUserModelConfigCache 测试用户模型配置缓存
func TestUserModelConfigCache(t *testing.T) {
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

// TestSubjectModelCache 测试学科组模型缓存
func TestSubjectModelCache(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	ctx := context.Background()

	t.Run("学科组信息缓存常量测试", func(t *testing.T) {
		// 验证缓存前缀格式
		assert.Equal(t, "subject_info:", smart.SubjectInfoCachePrefix)
		assert.Equal(t, "teacher_info:", smart.TeacherInfoCachePrefix)

		// 验证缓存时间是否为24小时
		expectedTTL := 24 * time.Hour
		assert.Equal(t, expectedTTL, smart.DefaultCacheTTL)
	})

	t.Run("获取学科组信息测试", func(t *testing.T) {
		subjectId := 123

		// Redis未启用时应该返回错误
		info, err := smart.GetSubjectInfoWithCache(ctx, subjectId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("获取老师信息测试", func(t *testing.T) {
		teacherId := "teacher_123"

		// Redis未启用时应该返回错误
		info, err := smart.GetTeacherInfoWithCache(ctx, teacherId)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("根据老师ID获取模型测试", func(t *testing.T) {
		// 空用户ID测试
		modelName, err := smart.GetModelByTeacherId(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacherId不能为空")
		assert.Equal(t, "", modelName)

		// 有效用户ID但Redis未启用
		modelName, err = smart.GetModelByTeacherId(ctx, "teacher_123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
		assert.Equal(t, "", modelName)
	})

	t.Run("失效缓存测试", func(t *testing.T) {
		// 失效学科组信息缓存
		err := smart.InvalidateSubjectInfoCache(ctx, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 失效老师信息缓存
		err = smart.InvalidateTeacherInfoCache(ctx, "teacher_123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")
	})

	t.Run("批量失效老师信息缓存测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		// Redis未启用时应该返回错误
		err := smart.BatchInvalidateTeacherInfoCache(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.BatchInvalidateTeacherInfoCache(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("预加载老师信息测试", func(t *testing.T) {
		teacherIds := []string{"teacher1", "teacher2"}

		// Redis未启用时应该返回错误
		err := smart.PreloadTeacherInfos(ctx, teacherIds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis未启用")

		// 空列表应该直接返回nil
		err = smart.PreloadTeacherInfos(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("获取学科组信息缓存统计测试", func(t *testing.T) {
		stats := smart.GetSubjectInfoCacheStats()
		assert.NotNil(t, stats)
		// 当CacheMgr为nil时应该返回空统计
		assert.GreaterOrEqual(t, stats.TotalCount, int64(0))
	})
}

// TestCacheIntegration 测试缓存集成功能
func TestCacheIntegration(t *testing.T) {
	// 禁用Redis以进行单元测试
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	t.Run("缓存组件初始化测试", func(t *testing.T) {
		// 测试缓存管理器初始化
		cache.InitManager()
		assert.NotNil(t, cache.Mgr)

		// 测试缓存失效管理器初始化
		cache.InitInvalidationManager()
		assert.NotNil(t, cache.Invalidator)

		// 测试缓存监控器初始化
		cache.InitMonitor()
		assert.NotNil(t, cache.MonitorInstance)
	})

	t.Run("缓存管理器状态测试", func(t *testing.T) {
		// 验证缓存管理器状态
		assert.False(t, cache.Mgr.IsEnabled())

		// 验证监控器状态
		assert.False(t, cache.MonitorInstance.IsMonitoring())
	})

	t.Run("缓存性能基准测试", func(t *testing.T) {
		ctx := context.Background()

		// 运行基准测试
		testConfig := &cache.TestConfig{
			Operations:  5,
			Concurrency: 1,
		}

		result := cache.MonitorInstance.RunPerformanceTest(ctx, testConfig)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Operations))

		// 验证测试结果结构
		assert.GreaterOrEqual(t, result.EndTime, result.StartTime)
		assert.GreaterOrEqual(t, result.SuccessRate, 0.0)
		assert.LessOrEqual(t, result.SuccessRate, 100.0)
	})
}
