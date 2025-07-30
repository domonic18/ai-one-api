package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/middleware/identity"
	"github.com/songquanpeng/one-api/tests/mocks"
)

// TestDefaultIdentityResolver 测试默认身份解析器
func TestDefaultIdentityResolver(t *testing.T) {
	resolver := &identity.DefaultIdentityResolver{}

	t.Run("ResolveGroup返回空字符串", func(t *testing.T) {
		group := resolver.ResolveGroup(context.Background(), "teacher_001")
		assert.Equal(t, "", group)
	})

	t.Run("ResolveModel返回原始模型", func(t *testing.T) {
		originalModel := "gpt-3.5-turbo"
		model := resolver.ResolveModel(context.Background(), "teacher_001", originalModel)
		assert.Equal(t, originalModel, model)
	})
}

// TestCoursewareIdentityResolver_ResolveGroup 测试课件平台身份解析器的用户组解析
func TestCoursewareIdentityResolver_ResolveGroup(t *testing.T) {
	ctx := context.Background()
	mockAPI := &mocks.MockCoursewareClient{}
	mockRedis := mocks.NewMockRedisClient()
	config := &identity.CoursewareConfig{
		DefaultGroup: "default",
		Timeout:      5 * time.Second,
	}

	cache := identity.NewCoursewareCache(mockRedis, config)
	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	t.Run("缓存命中", func(t *testing.T) {
		// 准备缓存数据
		userInfo := &identity.UserInfo{
			TeacherId:      "teacher_001",
			TeacherName:    "张老师",
			GroupName:      "beijing_math_group",
			PreferredModel: "gpt-4",
			UpdatedAt:      time.Now().Unix(),
		}

		userInfoBytes, _ := json.Marshal(userInfo)
		userInfoStr := string(userInfoBytes)

		// 设置mock期望 - 缓存命中
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(userInfoStr, nil))

		// 执行测试
		group := resolver.ResolveGroup(ctx, "teacher_001")

		// 验证结果
		assert.Equal(t, "beijing_math_group", group)
		mockRedis.AssertExpectations(t)
	})

	t.Run("缓存未命中_API成功", func(t *testing.T) {
		// 准备API响应
		apiUserInfo := &client.TeacherInfo{
			TeacherId:      "teacher_002",
			TeacherName:    "李老师",
			GroupName:      "beijing_chinese_group",
			PreferredModel: "gemini-pro",
		}

		// 设置mock期望
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_002").Return(redis.NewStringResult("", redis.Nil))
		mockAPI.On("GetTeacherInfo", mock.Anything, "teacher_002").Return(apiUserInfo, nil)
		mockRedis.On("Set", mock.Anything, "courseware:teacher:teacher_002", mock.Anything, config.CacheTTL).Return(redis.NewStatusResult("OK", nil))

		// 执行测试
		group := resolver.ResolveGroup(ctx, "teacher_002")

		// 验证结果
		assert.Equal(t, "beijing_chinese_group", group)
		mockRedis.AssertExpectations(t)
		mockAPI.AssertExpectations(t)
	})

	t.Run("API失败_使用默认分组", func(t *testing.T) {
		// 设置mock期望
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_003").Return(redis.NewStringResult("", redis.Nil))
		mockAPI.On("GetTeacherInfo", mock.Anything, "teacher_003").Return(nil, fmt.Errorf("API调用失败"))

		// 执行测试
		group := resolver.ResolveGroup(ctx, "teacher_003")

		// 验证结果
		assert.Equal(t, "default", group)
		mockRedis.AssertExpectations(t)
		mockAPI.AssertExpectations(t)
	})
}

// TestCoursewareIdentityResolver_ResolveModel 测试课件平台身份解析器的模型解析
func TestCoursewareIdentityResolver_ResolveModel(t *testing.T) {
	ctx := context.Background()
	mockAPI := &mocks.MockCoursewareClient{}
	mockRedis := mocks.NewMockRedisClient()
	config := &identity.CoursewareConfig{
		DefaultGroup: "default",
		Timeout:      5 * time.Second,
	}

	cache := identity.NewCoursewareCache(mockRedis, config)
	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	t.Run("有偏好模型", func(t *testing.T) {
		// 准备缓存数据
		userInfo := &identity.UserInfo{
			TeacherId:      "teacher_001",
			TeacherName:    "张老师",
			GroupName:      "beijing_math_group",
			PreferredModel: "gpt-4",
			UpdatedAt:      time.Now().Unix(),
		}

		userInfoBytes, _ := json.Marshal(userInfo)
		userInfoStr := string(userInfoBytes)

		// 设置mock期望
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(userInfoStr, nil))

		// 执行测试
		originalModel := "gpt-3.5-turbo"
		model := resolver.ResolveModel(ctx, "teacher_001", originalModel)

		// 验证结果
		assert.Equal(t, "gpt-4", model)
		mockRedis.AssertExpectations(t)
	})

	t.Run("无偏好模型", func(t *testing.T) {
		// 准备缓存数据
		userInfo := &identity.UserInfo{
			TeacherId:      "teacher_002",
			TeacherName:    "李老师",
			GroupName:      "beijing_chinese_group",
			PreferredModel: "", // 无偏好模型
			UpdatedAt:      time.Now().Unix(),
		}

		userInfoBytes, _ := json.Marshal(userInfo)
		userInfoStr := string(userInfoBytes)

		// 设置mock期望
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_002").Return(redis.NewStringResult(userInfoStr, nil))

		// 执行测试
		originalModel := "gpt-3.5-turbo"
		model := resolver.ResolveModel(ctx, "teacher_002", originalModel)

		// 验证结果
		assert.Equal(t, originalModel, model)
		mockRedis.AssertExpectations(t)
	})

	t.Run("用户不存在", func(t *testing.T) {
		// 设置mock期望
		mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_003").Return(redis.NewStringResult("", redis.Nil))

		// 执行测试
		originalModel := "gpt-3.5-turbo"
		model := resolver.ResolveModel(ctx, "teacher_003", originalModel)

		// 验证结果
		assert.Equal(t, originalModel, model)
		mockRedis.AssertExpectations(t)
	})
}

// TestIdentityResolver_GlobalManagement 测试全局解析器管理
func TestIdentityResolver_GlobalManagement(t *testing.T) {
	t.Run("默认解析器", func(t *testing.T) {
		// 重置为默认解析器
		identity.SetIdentityResolver(&identity.DefaultIdentityResolver{})

		resolver := identity.GetIdentityResolver()
		assert.IsType(t, &identity.DefaultIdentityResolver{}, resolver)

		// 测试默认行为
		group := resolver.ResolveGroup(context.Background(), "teacher_001")
		assert.Equal(t, "", group)

		model := resolver.ResolveModel(context.Background(), "teacher_001", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-3.5-turbo", model)
	})

	t.Run("课件平台解析器", func(t *testing.T) {
		mockAPI := &mocks.MockCoursewareClient{}
		mockRedis := mocks.NewMockRedisClient()
		config := &identity.CoursewareConfig{
			DefaultGroup: "default",
			Timeout:      5 * time.Second,
		}

		cache := identity.NewCoursewareCache(mockRedis, config)
		coursewareResolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

		// 设置为课件平台解析器
		identity.SetIdentityResolver(coursewareResolver)

		resolver := identity.GetIdentityResolver()
		assert.IsType(t, &identity.CoursewareIdentityResolver{}, resolver)
	})
}

// TestCoursewareConfig 测试课件平台配置
func TestCoursewareConfig(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		config := &identity.CoursewareConfig{
			Enabled:          true,
			BaseURL:          "https://api.example.com",
			APIKey:           "test-key",
			Timeout:          5 * time.Second,
			CacheTTL:         10 * time.Minute,
			DefaultGroup:     "default",
			PreloadBatchSize: 100,
			RefreshInterval:  1 * time.Hour,
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, "https://api.example.com", config.BaseURL)
		assert.Equal(t, "test-key", config.APIKey)
		assert.Equal(t, 5*time.Second, config.Timeout)
		assert.Equal(t, 10*time.Minute, config.CacheTTL)
		assert.Equal(t, "default", config.DefaultGroup)
		assert.Equal(t, 100, config.PreloadBatchSize)
		assert.Equal(t, 1*time.Hour, config.RefreshInterval)
	})
}
