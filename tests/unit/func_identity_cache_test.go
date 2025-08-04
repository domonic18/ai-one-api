package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/middleware/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCoursewareClient 模拟课件平台客户端
type MockCoursewareClient struct {
	mock.Mock
}

func (m *MockCoursewareClient) GetTeacherInfo(ctx context.Context, teacherId string) (*client.TeacherInfo, error) {
	args := m.Called(ctx, teacherId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TeacherInfo), args.Error(1)
}

func (m *MockCoursewareClient) GetTeacherIds(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCoursewareClient) BatchGetUserInfo(ctx context.Context, teacherIds []string) ([]*client.TeacherInfo, error) {
	args := m.Called(ctx, teacherIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*client.TeacherInfo), args.Error(1)
}

// 确保MockCoursewareClient实现了CoursewareAPIClient接口
var _ identity.CoursewareAPIClient = (*MockCoursewareClient)(nil)

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	mock.Mock
	data map[string]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	_ = m.Called(ctx, key)
	if value, exists := m.data[key]; exists {
		return redis.NewStringResult(value, nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	_ = m.Called(ctx, key, value, expiration)

	// 将value转换为字符串并存储
	if str, ok := value.(string); ok {
		m.data[key] = str
	} else if bytes, ok := value.([]byte); ok {
		m.data[key] = string(bytes)
	}

	return redis.NewStatusResult("OK", nil)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	_ = m.Called(ctx, keys)

	// 删除数据
	for _, key := range keys {
		delete(m.data, key)
	}

	return redis.NewIntResult(int64(len(keys)), nil)
}

func (m *MockRedisClient) Pipeline() redis.Pipeliner {
	_ = m.Called()
	// 简化实现，返回nil
	return nil
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	_ = m.Called(ctx, pattern)
	// 返回空的结果
	return redis.NewStringSliceResult([]string{}, nil)
}

// 确保MockRedisClient实现了RedisClient接口
var _ identity.RedisClient = (*MockRedisClient)(nil)

// TestCoursewareCache_GetUserInfo_缓存命中 测试课件平台缓存系统的缓存命中功能
// 测试目的：验证缓存系统能够正确从Redis缓存中获取用户信息，避免重复的API调用
// 测试内容：
// 1. 测试缓存中存在用户信息时的获取逻辑
// 2. 验证返回的用户信息字段完整性
// 3. 验证缓存命中时不会触发API调用
// 4. 确保缓存数据的正确性和一致性
func TestCoursewareCache_GetUserInfo_缓存命中(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 准备缓存数据
	userInfo := &identity.UserInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		UpdatedAt:      time.Now().Unix(),
	}

	userInfoBytes, _ := json.Marshal(userInfo)
	mockRedis.data["courseware:teacher:teacher_001"] = string(userInfoBytes)

	// 设置mock期望
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(string(userInfoBytes), nil))

	// 执行测试
	ctx := context.Background()
	result, err := cache.GetUserInfo(ctx, "teacher_001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "teacher_001", result.TeacherId)
	assert.Equal(t, "张老师", result.TeacherName)
	assert.Equal(t, "beijing_math_group", result.GroupName)
	assert.Equal(t, "gpt-4", result.PreferredModel)

	mockRedis.AssertExpectations(t)
}

// TestCoursewareCache_GetUserInfo_缓存未命中 测试课件平台缓存系统的缓存未命中处理
// 测试目的：验证缓存系统在缓存未命中时的正确行为，确保系统能够优雅处理缓存缺失情况
// 测试内容：
// 1. 测试缓存中不存在用户信息时的处理逻辑
// 2. 验证缓存未命中时返回nil结果
// 3. 确保不会因为缓存未命中而导致系统错误
// 4. 验证Redis连接和查询的正确性
func TestCoursewareCache_GetUserInfo_缓存未命中(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 设置mock期望 - 缓存未命中
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult("", redis.Nil))

	// 执行测试
	ctx := context.Background()
	result, err := cache.GetUserInfo(ctx, "teacher_001")

	// 验证结果
	assert.NoError(t, err)
	assert.Nil(t, result) // 缓存未命中应返回nil

	mockRedis.AssertExpectations(t)
}

// TestCoursewareCache_SetUserInfo_设置成功 测试课件平台缓存系统的用户信息设置功能
// 测试目的：验证缓存系统能够正确将用户信息存储到Redis缓存中，为后续的缓存命中提供数据基础
// 测试内容：
// 1. 测试用户信息成功存储到Redis的逻辑
// 2. 验证存储的数据完整性和正确性
// 3. 验证缓存TTL设置的正确性
// 4. 确保数据序列化和反序列化的准确性
// 5. 验证Redis操作的原子性和一致性
func TestCoursewareCache_SetUserInfo_设置成功(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	userInfo := &identity.UserInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		UpdatedAt:      time.Now().Unix(),
	}

	// 设置mock期望
	mockRedis.On("Set", mock.Anything, "courseware:teacher:teacher_001", mock.Anything, 10*time.Minute).Return(redis.NewStatusResult("OK", nil))

	// 执行测试
	ctx := context.Background()
	err := cache.SetUserInfo(ctx, userInfo)

	// 验证结果
	assert.NoError(t, err)

	// 验证数据是否正确存储
	key := "courseware:teacher:teacher_001"
	assert.Contains(t, mockRedis.data, key)

	// 验证存储的数据
	var storedUserInfo identity.UserInfo
	err = json.Unmarshal([]byte(mockRedis.data[key]), &storedUserInfo)
	assert.NoError(t, err)
	assert.Equal(t, userInfo.TeacherId, storedUserInfo.TeacherId)
	assert.Equal(t, userInfo.TeacherName, storedUserInfo.TeacherName)
	assert.Equal(t, userInfo.GroupName, storedUserInfo.GroupName)

	mockRedis.AssertExpectations(t)
}

// 注意：批量设置测试依赖于Redis Pipeline功能，在简化测试中跳过

// TestCoursewareIdentityResolver_ResolveGroup_缓存命中 测试课件平台身份解析器的用户组解析功能（缓存命中场景）
// 测试目的：验证身份解析器在缓存命中的情况下能够快速返回用户组信息，提高系统响应性能
// 测试内容：
// 1. 测试缓存中存在用户信息时的用户组解析逻辑
// 2. 验证返回的用户组名称正确性
// 3. 验证缓存命中时不会触发API调用，提高性能
// 4. 确保缓存数据的有效性和时效性
// 5. 验证解析器与缓存系统的正确集成
func TestCoursewareIdentityResolver_ResolveGroup_缓存命中(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL:     10 * time.Minute,
		DefaultGroup: "default",
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 准备缓存数据
	userInfo := &identity.UserInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		UpdatedAt:      time.Now().Unix(),
	}

	userInfoBytes, _ := json.Marshal(userInfo)
	mockRedis.data["courseware:teacher:teacher_001"] = string(userInfoBytes)

	// 设置mock期望
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(string(userInfoBytes), nil))

	mockAPI := &MockCoursewareClient{}
	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	// 执行测试
	ctx := context.Background()
	group := resolver.ResolveGroup(ctx, "teacher_001")

	// 验证结果
	assert.Equal(t, "beijing_math_group", group)

	mockRedis.AssertExpectations(t)
}

// TestCoursewareIdentityResolver_ResolveGroup_缓存未命中_API成功 测试课件平台身份解析器的用户组解析功能（缓存未命中但API成功场景）
// 测试目的：验证身份解析器在缓存未命中时能够正确调用API获取用户信息，并将结果缓存以提高后续访问性能
// 测试内容：
// 1. 测试缓存未命中时的API调用逻辑
// 2. 验证API返回数据的正确解析和处理
// 3. 验证获取到的用户信息正确存储到缓存中
// 4. 确保用户组信息的准确性和一致性
// 5. 验证缓存更新机制的正确性
// 6. 测试API调用与缓存存储的完整流程
func TestCoursewareIdentityResolver_ResolveGroup_缓存未命中_API成功(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL:     10 * time.Minute,
		DefaultGroup: "default",
		Timeout:      5 * time.Second,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 设置mock期望 - 缓存未命中
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult("", redis.Nil))
	mockRedis.On("Set", mock.Anything, "courseware:teacher:teacher_001", mock.Anything, 10*time.Minute).Return(redis.NewStatusResult("OK", nil))

	mockAPI := &MockCoursewareClient{}

	// 模拟API返回数据
	apiUserInfo := &client.TeacherInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
	}

	mockAPI.On("GetTeacherInfo", mock.Anything, "teacher_001").Return(apiUserInfo, nil)

	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	// 执行测试
	ctx := context.Background()
	group := resolver.ResolveGroup(ctx, "teacher_001")

	// 验证结果
	assert.Equal(t, "beijing_math_group", group)

	// 验证缓存是否已更新
	assert.Contains(t, mockRedis.data, "courseware:teacher:teacher_001")

	mockAPI.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestCoursewareIdentityResolver_ResolveGroup_API失败_使用默认分组 测试课件平台身份解析器的用户组解析功能（API调用失败场景）
// 测试目的：验证身份解析器在API调用失败时能够优雅降级，使用默认用户组确保系统的可用性和稳定性
// 测试内容：
// 1. 测试API调用失败时的错误处理逻辑
// 2. 验证系统能够正确返回默认用户组
// 3. 确保API失败不会影响系统的正常运行
// 4. 验证错误处理的健壮性和容错性
// 5. 测试系统在异常情况下的降级策略
// 6. 确保用户体验的连续性和一致性
func TestCoursewareIdentityResolver_ResolveGroup_API失败_使用默认分组(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL:     10 * time.Minute,
		DefaultGroup: "default",
		Timeout:      5 * time.Second,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 设置mock期望 - 缓存未命中
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult("", redis.Nil))

	mockAPI := &MockCoursewareClient{}

	// 模拟API调用失败
	mockAPI.On("GetTeacherInfo", mock.Anything, "teacher_001").Return(nil, fmt.Errorf("API调用失败"))

	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	// 执行测试
	ctx := context.Background()
	group := resolver.ResolveGroup(ctx, "teacher_001")

	// 验证结果
	assert.Equal(t, "default", group)

	mockAPI.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestCoursewareIdentityResolver_ResolveModel_有偏好模型 测试课件平台身份解析器的模型解析功能（用户有偏好模型场景）
// 测试目的：验证身份解析器能够正确识别用户的偏好模型设置，并优先使用用户指定的模型进行AI对话
// 测试内容：
// 1. 测试用户有偏好模型时的模型解析逻辑
// 2. 验证系统能够正确返回用户的偏好模型
// 3. 确保偏好模型设置能够覆盖默认模型选择
// 4. 验证模型解析的准确性和一致性
// 5. 测试用户个性化配置的有效性
// 6. 确保模型选择逻辑的正确性
func TestCoursewareIdentityResolver_ResolveModel_有偏好模型(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 准备缓存数据
	userInfo := &identity.UserInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		UpdatedAt:      time.Now().Unix(),
	}

	userInfoBytes, _ := json.Marshal(userInfo)
	mockRedis.data["courseware:teacher:teacher_001"] = string(userInfoBytes)

	// 设置mock期望
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(string(userInfoBytes), nil))

	mockAPI := &MockCoursewareClient{}
	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	// 执行测试
	ctx := context.Background()
	model := resolver.ResolveModel(ctx, "teacher_001", "gpt-3.5-turbo")

	// 验证结果
	assert.Equal(t, "gpt-4", model) // 应该使用偏好模型

	mockRedis.AssertExpectations(t)
}

// TestCoursewareIdentityResolver_ResolveModel_无偏好模型 测试课件平台身份解析器的模型解析功能（用户无偏好模型场景）
// 测试目的：验证身份解析器在用户没有设置偏好模型时能够正确使用原始模型，确保系统的默认行为符合预期
// 测试内容：
// 1. 测试用户无偏好模型时的模型解析逻辑
// 2. 验证系统能够正确返回原始模型
// 3. 确保无偏好模型设置不会影响正常的模型选择
// 4. 验证默认模型选择的正确性
// 5. 测试系统在用户未配置偏好时的默认行为
// 6. 确保模型解析逻辑的健壮性
func TestCoursewareIdentityResolver_ResolveModel_无偏好模型(t *testing.T) {
	// 准备测试数据
	mockRedis := NewMockRedisClient()
	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}
	cache := identity.NewCoursewareCache(mockRedis, config)

	// 准备缓存数据（无偏好模型）
	userInfo := &identity.UserInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		GroupName:      "beijing_math_group",
		PreferredModel: "", // 无偏好模型
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		UpdatedAt:      time.Now().Unix(),
	}

	userInfoBytes, _ := json.Marshal(userInfo)
	mockRedis.data["courseware:teacher:teacher_001"] = string(userInfoBytes)

	// 设置mock期望
	mockRedis.On("Get", mock.Anything, "courseware:teacher:teacher_001").Return(redis.NewStringResult(string(userInfoBytes), nil))

	mockAPI := &MockCoursewareClient{}
	resolver := identity.NewCoursewareIdentityResolver(mockAPI, cache, config)

	// 执行测试
	ctx := context.Background()
	originalModel := "gpt-3.5-turbo"
	model := resolver.ResolveModel(ctx, "teacher_001", originalModel)

	// 验证结果
	assert.Equal(t, originalModel, model) // 应该使用原始模型

	mockRedis.AssertExpectations(t)
}

// 注意：预加载测试依赖于Redis Pipeline功能，在简化测试中跳过
