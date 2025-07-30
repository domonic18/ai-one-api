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

func (m *MockCoursewareClient) GetAllTeacherIds(ctx context.Context) ([]string, error) {
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

// 确保MockRedisClient实现了RedisClient接口
var _ identity.RedisClient = (*MockRedisClient)(nil)

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
