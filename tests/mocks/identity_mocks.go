package mocks

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/middleware/identity"
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
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return redis.NewStringResult("", redis.Nil)
	}
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	_ = m.Called(ctx, key, value, expiration)
	if str, ok := value.(string); ok {
		m.data[key] = str
	} else if bytes, ok := value.([]byte); ok {
		m.data[key] = string(bytes)
	}
	return redis.NewStatusResult("OK", nil)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	_ = m.Called(ctx, keys)
	for _, key := range keys {
		delete(m.data, key)
	}
	return redis.NewIntResult(int64(len(keys)), nil)
}

func (m *MockRedisClient) Pipeline() redis.Pipeliner {
	_ = m.Called()
	return nil
}

// 确保MockRedisClient实现了RedisClient接口
var _ identity.RedisClient = (*MockRedisClient)(nil)
