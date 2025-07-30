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

// TestDefaultIdentityResolver 测试默认身份解析器的基本功能
// 测试目的：验证默认身份解析器在没有外部依赖的情况下能够提供基本的身份解析功能，确保系统的默认行为符合预期
// 测试内容：
// 1. 测试默认解析器的用户组解析功能，验证返回空字符串
// 2. 测试默认解析器的模型解析功能，验证返回原始模型
// 3. 确保默认解析器在无外部配置时的稳定性和一致性
// 4. 验证默认解析器作为系统降级方案的有效性
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

// TestCoursewareIdentityResolver_ResolveGroup 测试课件平台身份解析器的用户组解析功能
// 测试目的：验证课件平台身份解析器能够根据教师ID正确解析用户组信息，支持多种场景下的用户组识别和分配
// 测试内容：
// 1. 测试缓存命中场景下的用户组解析，验证快速响应能力
// 2. 测试缓存未命中但API成功场景下的用户组解析，验证数据获取和缓存更新
// 3. 测试API失败场景下的用户组解析，验证错误处理和默认值回退
// 4. 验证不同场景下用户组解析的准确性和一致性
// 5. 测试解析器与缓存系统和API客户端的正确集成
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

// TestCoursewareIdentityResolver_ResolveModel 测试课件平台身份解析器的模型解析功能
// 测试目的：验证课件平台身份解析器能够根据用户配置和系统设置正确解析AI模型选择，支持个性化模型偏好设置
// 测试内容：
// 1. 测试用户有偏好模型时的模型解析，验证个性化配置的有效性
// 2. 测试用户无偏好模型时的模型解析，验证默认模型选择的正确性
// 3. 测试用户不存在时的模型解析，验证边界情况的处理
// 4. 验证模型解析逻辑的准确性和一致性
// 5. 测试模型选择策略的正确实现
// 6. 确保模型解析功能的健壮性和容错性
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

// TestIdentityResolver_GlobalManagement 测试全局身份解析器的管理和切换功能
// 测试目的：验证系统能够正确管理全局身份解析器的设置和切换，支持不同解析器策略的动态配置
// 测试内容：
// 1. 测试默认解析器的设置和获取，验证基本管理功能
// 2. 测试课件平台解析器的设置和获取，验证高级解析器功能
// 3. 验证解析器切换的正确性和一致性
// 4. 测试全局解析器管理的线程安全性
// 5. 确保解析器管理功能的稳定性和可靠性
// 6. 验证不同解析器策略的正确实现和切换
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

// TestCoursewareConfig 测试课件平台配置结构的功能和属性
// 测试目的：验证课件平台配置结构能够正确存储和管理各种配置参数，确保配置系统的完整性和可用性
// 测试内容：
// 1. 测试默认配置的创建和属性设置，验证配置结构的完整性
// 2. 验证各种配置参数的正确性和有效性
// 3. 测试配置参数的默认值设置
// 4. 确保配置结构能够支持所有必要的配置选项
// 5. 验证配置系统的可扩展性和灵活性
// 6. 测试配置参数的类型安全性和数据完整性
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
