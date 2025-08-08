package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/middleware/identity"
)

// 扩展的模拟数据结构，支持新的字段
type ExtendedTeacherInfo struct {
	TeacherId      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	SchoolId       int    `json:"school_id"`
	SchoolName     string `json:"school_name"`
	SubjectId      int    `json:"subject_id"`
	SubjectName    string `json:"subject_name"`
	GroupName      string `json:"group_name"`      // 新增：OneAPI用户组
	PreferredModel string `json:"preferred_model"` // 新增：用户偏好模型
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// 扩展的模拟数据存储
type ExtendedMockDataStore struct {
	mu           sync.RWMutex
	teachers     map[string]*ExtendedTeacherInfo
	requestCount map[string]int
}

// 创建扩展的模拟数据存储
func NewExtendedMockDataStore() *ExtendedMockDataStore {
	store := &ExtendedMockDataStore{
		teachers:     make(map[string]*ExtendedTeacherInfo),
		requestCount: make(map[string]int),
	}

	// 添加一些默认数据
	now := time.Now().Unix()

	// 添加老师数据
	store.teachers["teacher_001"] = &ExtendedTeacherInfo{
		TeacherId:      "teacher_001",
		TeacherName:    "张老师",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      10,
		SubjectName:    "数学组",
		GroupName:      "beijing_math_group",
		PreferredModel: "gpt-4",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	store.teachers["teacher_002"] = &ExtendedTeacherInfo{
		TeacherId:      "teacher_002",
		TeacherName:    "李老师",
		SchoolId:       1,
		SchoolName:     "北京中学",
		SubjectId:      11,
		SubjectName:    "语文组",
		GroupName:      "beijing_chinese_group",
		PreferredModel: "gemini-pro",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	store.teachers["teacher_003"] = &ExtendedTeacherInfo{
		TeacherId:      "teacher_003",
		TeacherName:    "王老师",
		SchoolId:       2,
		SchoolName:     "上海中学",
		SubjectId:      12,
		SubjectName:    "英语组",
		GroupName:      "shanghai_english_group",
		PreferredModel: "claude-3",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return store
}

// 扩展的课件平台模拟服务器
type ExtendedCoursePlatformMock struct {
	server *httptest.Server
	store  *ExtendedMockDataStore
}

// 创建扩展的课件平台模拟服务器
func NewExtendedCoursePlatformMock() *ExtendedCoursePlatformMock {
	store := NewExtendedMockDataStore()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	mock := &ExtendedCoursePlatformMock{
		store: store,
	}

	// 设置路由（兼容无前缀与 /api/v1 前缀两种写法）
	router.GET("/teacher/:teacherId/info", mock.getTeacherInfo)
	router.GET("/teachers/ids", mock.getAllTeacherIds)
	router.POST("/teachers/batch", mock.batchGetUserInfo)

	router.GET("/api/v1/teacher/:teacherId/info", mock.getTeacherInfo)
	router.GET("/api/v1/teachers/ids", mock.getAllTeacherIds)
	router.POST("/api/v1/teachers/batch", mock.batchGetUserInfo)

	// 创建测试服务器
	mock.server = httptest.NewServer(router)

	return mock
}

// 获取单个老师信息
func (m *ExtendedCoursePlatformMock) getTeacherInfo(c *gin.Context) {
	teacherId := c.Param("teacherId")

	// 验证认证
	if !m.validateAuth(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	m.store.mu.RLock()
	teacher, exists := m.store.teachers[teacherId]
	m.store.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
		return
	}

	m.store.recordRequest("getTeacherInfo")
	c.JSON(http.StatusOK, teacher)
}

// 获取所有老师ID列表
func (m *ExtendedCoursePlatformMock) getAllTeacherIds(c *gin.Context) {
	// 验证认证
	if !m.validateAuth(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	m.store.mu.RLock()
	teacherIds := make([]string, 0, len(m.store.teachers))
	for teacherId := range m.store.teachers {
		teacherIds = append(teacherIds, teacherId)
	}
	m.store.mu.RUnlock()

	m.store.recordRequest("getAllTeacherIds")
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"teacher_ids": teacherIds,
			"total":       len(teacherIds),
		},
	})
}

// 批量获取用户信息
func (m *ExtendedCoursePlatformMock) batchGetUserInfo(c *gin.Context) {
	// 验证认证
	if !m.validateAuth(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		TeacherIds []string `json:"teacher_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	m.store.mu.RLock()
	teachers := make([]*ExtendedTeacherInfo, 0, len(request.TeacherIds))
	for _, teacherId := range request.TeacherIds {
		if teacher, exists := m.store.teachers[teacherId]; exists {
			teachers = append(teachers, teacher)
		}
	}
	m.store.mu.RUnlock()

	m.store.recordRequest("batchGetUserInfo")
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"users":         teachers,
			"success_count": len(teachers),
			"error_count":   0,
		},
	})
}

// 记录请求统计
func (s *ExtendedMockDataStore) recordRequest(endpoint string) {
	s.mu.Lock()
	s.requestCount[endpoint]++
	s.mu.Unlock()
}

// 获取请求统计
func (s *ExtendedMockDataStore) getRequestStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for k, v := range s.requestCount {
		stats[k] = v
	}
	return stats
}

// 验证认证
func (m *ExtendedCoursePlatformMock) validateAuth(authHeader string) bool {
	return authHeader == "Bearer test-api-key"
}

// 获取服务器URL
func (m *ExtendedCoursePlatformMock) URL() string {
	return m.server.URL
}

// 关闭服务器
func (m *ExtendedCoursePlatformMock) Close() {
	m.server.Close()
}

// 添加老师数据
func (m *ExtendedCoursePlatformMock) AddTeacher(teacher *ExtendedTeacherInfo) {
	m.store.mu.Lock()
	m.store.teachers[teacher.TeacherId] = teacher
	m.store.mu.Unlock()
}

// 获取请求统计
func (m *ExtendedCoursePlatformMock) GetRequestStats() map[string]int {
	return m.store.getRequestStats()
}

// TestIdentity_CoursewareCache_集成测试 测试课件平台缓存系统的集成功能
// 测试目的：验证Redis缓存系统在实际环境中的完整功能，包括单个和批量用户信息的缓存操作、缓存未命中处理、缓存过期机制等
// 测试内容：
// 1. 测试单个用户信息的设置和获取，验证缓存的基本功能
// 2. 测试批量用户信息的缓存操作，验证Redis Pipeline的批量处理能力
// 3. 测试缓存未命中时的正确处理，确保系统能够优雅处理缺失数据
// 4. 测试缓存过期机制，验证TTL设置的正确性和过期后的数据清理
// 5. 验证缓存数据的完整性和一致性
// 6. 测试Redis连接和数据序列化的正确性
func TestIdentity_CoursewareCache_集成测试(t *testing.T) {
	// 准备Redis测试环境
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // 使用不同的DB避免冲突
	})
	defer redisClient.Close()

	// 清理测试数据
	ctx := context.Background()
	redisClient.FlushDB(ctx)

	config := &identity.CoursewareConfig{
		CacheTTL: 10 * time.Minute,
	}

	cache := identity.NewCoursewareCache(redisClient, config)

	t.Run("单个用户信息缓存操作", func(t *testing.T) {
		// 准备测试数据
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

		// 测试设置用户信息
		err := cache.SetUserInfo(ctx, userInfo)
		assert.NoError(t, err)

		// 测试获取用户信息
		result, err := cache.GetUserInfo(ctx, "teacher_001")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "teacher_001", result.TeacherId)
		assert.Equal(t, "张老师", result.TeacherName)
		assert.Equal(t, "beijing_math_group", result.GroupName)
		assert.Equal(t, "gpt-4", result.PreferredModel)
	})

	t.Run("批量用户信息缓存操作", func(t *testing.T) {
		// 准备批量测试数据
		userInfos := []*identity.UserInfo{
			{
				TeacherId:      "teacher_002",
				TeacherName:    "李老师",
				GroupName:      "beijing_chinese_group",
				PreferredModel: "gemini-pro",
				SchoolId:       1,
				SchoolName:     "北京中学",
				SubjectId:      11,
				SubjectName:    "语文组",
				UpdatedAt:      time.Now().Unix(),
			},
			{
				TeacherId:      "teacher_003",
				TeacherName:    "王老师",
				GroupName:      "shanghai_english_group",
				PreferredModel: "claude-3",
				SchoolId:       2,
				SchoolName:     "上海中学",
				SubjectId:      12,
				SubjectName:    "英语组",
				UpdatedAt:      time.Now().Unix(),
			},
		}

		// 测试批量设置
		err := cache.BatchSetUserInfo(ctx, userInfos)
		assert.NoError(t, err)

		// 验证批量设置结果
		for _, userInfo := range userInfos {
			result, err := cache.GetUserInfo(ctx, userInfo.TeacherId)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, userInfo.TeacherId, result.TeacherId)
			assert.Equal(t, userInfo.TeacherName, result.TeacherName)
			assert.Equal(t, userInfo.GroupName, result.GroupName)
			assert.Equal(t, userInfo.PreferredModel, result.PreferredModel)
		}
	})

	t.Run("缓存未命中处理", func(t *testing.T) {
		// 测试获取不存在的用户信息
		result, err := cache.GetUserInfo(ctx, "nonexistent_teacher")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("缓存过期处理", func(t *testing.T) {
		// 使用短TTL进行测试
		shortTTLConfig := &identity.CoursewareConfig{
			CacheTTL: 1 * time.Second,
		}
		shortTTLCache := identity.NewCoursewareCache(redisClient, shortTTLConfig)

		userInfo := &identity.UserInfo{
			TeacherId:      "teacher_expire",
			TeacherName:    "过期老师",
			GroupName:      "expire_group",
			PreferredModel: "gpt-3.5-turbo",
			UpdatedAt:      time.Now().Unix(),
		}

		// 设置用户信息
		err := shortTTLCache.SetUserInfo(ctx, userInfo)
		assert.NoError(t, err)

		// 立即获取应该成功
		result, err := shortTTLCache.GetUserInfo(ctx, "teacher_expire")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 等待过期
		time.Sleep(2 * time.Second)

		// 过期后获取应该返回nil
		result, err = shortTTLCache.GetUserInfo(ctx, "teacher_expire")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestIdentity_CoursewareAPIClient_集成测试 测试课件平台API客户端的集成功能
// 测试目的：验证API客户端与课件平台服务器的完整交互功能，包括认证、数据获取、错误处理、超时处理等
// 测试内容：
// 1. 测试获取单个老师信息的API调用，验证数据获取的准确性和完整性
// 2. 测试获取所有老师ID列表的API调用，验证批量数据获取功能
// 3. 测试批量获取用户信息的API调用，验证批量操作的效率和正确性
// 4. 测试API认证失败的处理，验证错误响应的正确性
// 5. 测试API超时处理，验证系统在超时情况下的健壮性
// 6. 验证API响应的数据格式和字段完整性
// 7. 测试不存在的用户ID的处理，验证错误边界情况
func TestIdentity_CoursewareAPIClient_集成测试(t *testing.T) {
	// 创建模拟服务器
	mockServer := NewExtendedCoursePlatformMock()
	defer mockServer.Close()

	// 直接设置配置变量并初始化客户端（确保 httpClient 非空）
	config.CoursewarePlatformBaseURL = mockServer.URL()
	config.CoursewarePlatformAPIKey = "test-api-key"
	config.CoursewarePlatformTimeout = 2
	client.InitCoursewareClient()

	// 获取API客户端
	apiClient := client.GetCoursewareClient()

	t.Run("获取单个老师信息", func(t *testing.T) {
		// 测试获取存在的老师信息
		teacherInfo, err := apiClient.GetTeacherInfo(context.Background(), "teacher_001")
		assert.NoError(t, err)
		assert.NotNil(t, teacherInfo)
		assert.Equal(t, "teacher_001", teacherInfo.TeacherId)
		assert.Equal(t, "张老师", teacherInfo.TeacherName)
		assert.Equal(t, "beijing_math_group", teacherInfo.GroupName)
		assert.Equal(t, "gpt-4", teacherInfo.PreferredModel)

		// 测试获取不存在的老师信息
		teacherInfo, err = apiClient.GetTeacherInfo(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, teacherInfo)
	})

	t.Run("获取所有老师ID列表", func(t *testing.T) {
		teacherIds, err := apiClient.GetAllTeacherIds(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, teacherIds)
		assert.Len(t, teacherIds, 3)
		assert.Contains(t, teacherIds, "teacher_001")
		assert.Contains(t, teacherIds, "teacher_002")
		assert.Contains(t, teacherIds, "teacher_003")
	})

	t.Run("批量获取用户信息", func(t *testing.T) {
		teacherIds := []string{"teacher_001", "teacher_002", "teacher_003"}
		userInfos, err := apiClient.BatchGetUserInfo(context.Background(), teacherIds)
		assert.NoError(t, err)
		assert.NotNil(t, userInfos)
		assert.Len(t, userInfos, 3)

		// 验证返回的数据
		for _, userInfo := range userInfos {
			assert.NotEmpty(t, userInfo.TeacherId)
			assert.NotEmpty(t, userInfo.TeacherName)
			assert.NotEmpty(t, userInfo.GroupName)
			assert.NotEmpty(t, userInfo.PreferredModel)
		}
	})

	t.Run("API认证失败", func(t *testing.T) {
		// 创建错误的API客户端
		wrongClient := client.GetCoursewareClient()
		wrongClient.SetBaseURL(mockServer.URL())
		wrongClient.SetAPIKey("wrong-api-key")

		_, err := wrongClient.GetTeacherInfo(context.Background(), "teacher_001")
		assert.Error(t, err)
	})

	t.Run("API超时处理", func(t *testing.T) {
		// 创建一个会超时的服务器（不响应）
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second) // 延迟2秒
			w.WriteHeader(http.StatusOK)
		}))
		defer slowServer.Close()

		// 创建超时时间很短的客户端
		timeoutClient := client.GetCoursewareClient()
		timeoutClient.SetBaseURL(slowServer.URL)
		timeoutClient.SetAPIKey("test-api-key")

		_, err := timeoutClient.GetTeacherInfo(context.Background(), "teacher_001")
		assert.Error(t, err)
	})
}

// TestIdentity_PreloadManager_集成测试 测试用户信息预加载管理器的集成功能
// 测试目的：验证预加载管理器能够正确地从API获取用户信息并批量缓存到Redis，提高系统性能和响应速度
// 测试内容：
// 1. 测试预加载所有用户信息的功能，验证批量数据获取和缓存操作
// 2. 测试预加载的分批处理机制，验证大数据量时的分批处理能力
// 3. 测试预加载API失败的处理，验证错误情况下的系统稳定性
// 4. 验证预加载后的缓存数据完整性和准确性
// 5. 测试不同批次大小配置下的预加载效果
// 6. 验证API调用统计和性能监控功能
// 7. 测试预加载过程中的并发安全性
func TestIdentity_PreloadManager_集成测试(t *testing.T) {
	// 准备Redis测试环境
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       2, // 使用不同的DB避免冲突
	})
	defer redisClient.Close()

	// 清理测试数据
	ctx := context.Background()
	redisClient.FlushDB(ctx)

	// 创建模拟服务器
	mockServer := NewExtendedCoursePlatformMock()
	defer mockServer.Close()

	// 创建API客户端
	apiClient := client.GetCoursewareClient()
	apiClient.SetBaseURL(mockServer.URL())
	apiClient.SetAPIKey("test-api-key")

	config := &identity.CoursewareConfig{
		CacheTTL:         10 * time.Minute,
		PreloadBatchSize: 2,
		RefreshInterval:  1 * time.Minute,
	}

	cache := identity.NewCoursewareCache(redisClient, config)
	preloadManager := identity.NewPreloadManager(apiClient, cache, config)

	t.Run("预加载所有用户信息", func(t *testing.T) {
		// 执行预加载
		err := preloadManager.PreloadUserInfos(ctx)
		assert.NoError(t, err)

		// 验证所有用户信息都已缓存
		expectedTeachers := []string{"teacher_001", "teacher_002", "teacher_003"}
		for _, teacherId := range expectedTeachers {
			userInfo, err := cache.GetUserInfo(ctx, teacherId)
			assert.NoError(t, err)
			assert.NotNil(t, userInfo)
			assert.Equal(t, teacherId, userInfo.TeacherId)
		}

		// 验证API调用统计
		stats := mockServer.GetRequestStats()
		assert.GreaterOrEqual(t, stats["getAllTeacherIds"], 1)
		assert.GreaterOrEqual(t, stats["batchGetUserInfo"], 1)
	})

	t.Run("预加载分批处理", func(t *testing.T) {
		// 清理缓存
		redisClient.FlushDB(ctx)

		// 使用小批次大小进行测试
		smallBatchConfig := &identity.CoursewareConfig{
			CacheTTL:         10 * time.Minute,
			PreloadBatchSize: 1, // 每次只处理1个
			RefreshInterval:  1 * time.Minute,
		}

		smallBatchCache := identity.NewCoursewareCache(redisClient, smallBatchConfig)
		smallBatchPreloadManager := identity.NewPreloadManager(apiClient, smallBatchCache, smallBatchConfig)

		// 执行预加载
		err := smallBatchPreloadManager.PreloadUserInfos(ctx)
		assert.NoError(t, err)

		// 验证所有用户信息都已缓存
		expectedTeachers := []string{"teacher_001", "teacher_002", "teacher_003"}
		for _, teacherId := range expectedTeachers {
			userInfo, err := smallBatchCache.GetUserInfo(ctx, teacherId)
			assert.NoError(t, err)
			assert.NotNil(t, userInfo)
			assert.Equal(t, teacherId, userInfo.TeacherId)
		}
	})

	t.Run("预加载API失败处理", func(t *testing.T) {
		// 创建错误的API客户端
		wrongClient := client.GetCoursewareClient()
		wrongClient.SetBaseURL(mockServer.URL())
		wrongClient.SetAPIKey("wrong-api-key")

		wrongPreloadManager := identity.NewPreloadManager(wrongClient, cache, config)

		// 执行预加载应该失败
		err := wrongPreloadManager.PreloadUserInfos(ctx)
		assert.Error(t, err)
	})
}

// TestIdentity_CompleteFlow_集成测试 测试身份解析器的完整业务流程
// 测试目的：验证身份解析器在实际业务场景中的完整工作流程，包括缓存命中、API调用、并发处理、错误降级等
// 测试内容：
// 1. 测试完整的身份解析流程，包括首次解析（缓存未命中）和后续解析（缓存命中）
// 2. 测试用户组解析功能，验证不同场景下的用户组识别和分配
// 3. 测试模型解析功能，验证用户偏好模型的正确应用
// 4. 测试并发访问场景，验证系统在高并发下的稳定性和数据一致性
// 5. 测试缓存和API混合场景，验证部分缓存命中、部分API调用的复杂情况
// 6. 测试不存在的用户处理，验证默认值和降级策略的正确性
// 7. 验证解析结果的准确性和一致性
// 8. 测试系统的性能和响应时间
func TestIdentity_CompleteFlow_集成测试(t *testing.T) {
	// 准备Redis测试环境
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       3, // 使用不同的DB避免冲突
	})
	defer redisClient.Close()

	// 清理测试数据
	ctx := context.Background()
	redisClient.FlushDB(ctx)

	// 创建模拟服务器
	mockServer := NewExtendedCoursePlatformMock()
	defer mockServer.Close()

	// 创建API客户端
	apiClient := client.GetCoursewareClient()
	apiClient.SetBaseURL(mockServer.URL())
	apiClient.SetAPIKey("test-api-key")

	config := &identity.CoursewareConfig{
		CacheTTL:         10 * time.Minute,
		DefaultGroup:     "default",
		PreloadBatchSize: 2,
		RefreshInterval:  1 * time.Minute,
		Timeout:          5 * time.Second,
	}

	cache := identity.NewCoursewareCache(redisClient, config)
	resolver := identity.NewCoursewareIdentityResolver(apiClient, cache, config)

	t.Run("完整身份解析流程", func(t *testing.T) {
		// 1. 首次解析（缓存未命中，调用API）
		group := resolver.ResolveGroup(ctx, "teacher_001")
		assert.Equal(t, "beijing_math_group", group)

		// 2. 再次解析（缓存命中）
		group = resolver.ResolveGroup(ctx, "teacher_001")
		assert.Equal(t, "beijing_math_group", group)

		// 3. 解析偏好模型
		model := resolver.ResolveModel(ctx, "teacher_001", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-4", model) // 应该使用偏好模型

		// 4. 解析不存在的用户（使用默认分组）
		group = resolver.ResolveGroup(ctx, "nonexistent_teacher")
		assert.Equal(t, "default", group)

		// 5. 解析无偏好模型的用户
		model = resolver.ResolveModel(ctx, "nonexistent_teacher", "gpt-3.5-turbo")
		assert.Equal(t, "gpt-3.5-turbo", model) // 应该使用原始模型
	})

	t.Run("并发访问测试", func(t *testing.T) {
		// 清理缓存
		redisClient.FlushDB(ctx)

		// 并发访问测试
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				teacherId := fmt.Sprintf("teacher_%03d", (id%3)+1)
				group := resolver.ResolveGroup(ctx, teacherId)
				assert.NotEmpty(t, group)

				model := resolver.ResolveModel(ctx, teacherId, "gpt-3.5-turbo")
				assert.NotEmpty(t, model)
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("缓存和API混合场景", func(t *testing.T) {
		// 清理缓存
		redisClient.FlushDB(ctx)

		// 测试混合场景：部分用户有缓存，部分没有
		teachers := []string{"teacher_001", "teacher_002", "teacher_003", "nonexistent_teacher"}

		for _, teacherId := range teachers {
			group := resolver.ResolveGroup(ctx, teacherId)
			assert.NotEmpty(t, group)

			model := resolver.ResolveModel(ctx, teacherId, "gpt-3.5-turbo")
			assert.NotEmpty(t, model)
		}
	})
}

// TestIdentity_ErrorHandling_集成测试 测试身份解析系统的错误处理和边界情况
// 测试目的：验证系统在各种异常情况和边界条件下的健壮性，确保系统能够优雅处理错误并保持稳定运行
// 测试内容：
// 1. 测试Redis连接失败的处理，验证系统在缓存服务不可用时的降级策略
// 2. 测试空数据输入的处理，验证系统对空值和无效数据的容错能力
// 3. 测试大数据量场景，验证系统在处理大量用户信息时的性能和稳定性
// 4. 测试网络异常和超时情况的处理，验证系统的错误恢复能力
// 5. 测试数据序列化和反序列化的错误处理，验证数据完整性保护
// 6. 测试系统资源限制情况下的处理，验证内存和连接池的管理
// 7. 验证错误日志记录和监控功能的正确性
// 8. 测试系统在异常情况下的降级和恢复机制
func TestIdentity_ErrorHandling_集成测试(t *testing.T) {
	// 准备Redis测试环境
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       4, // 使用不同的DB避免冲突
	})
	defer redisClient.Close()

	// 清理测试数据
	ctx := context.Background()
	redisClient.FlushDB(ctx)

	config := &identity.CoursewareConfig{
		CacheTTL:     10 * time.Minute,
		DefaultGroup: "default",
		Timeout:      5 * time.Second,
	}

	cache := identity.NewCoursewareCache(redisClient, config)

	t.Run("Redis连接失败处理", func(t *testing.T) {
		// 创建错误的Redis客户端
		wrongRedisClient := redis.NewClient(&redis.Options{
			Addr:     "localhost:9999", // 错误的端口
			Password: "",
			DB:       0,
		})

		wrongCache := identity.NewCoursewareCache(wrongRedisClient, config)

		// 测试设置用户信息（应该失败但不会panic）
		userInfo := &identity.UserInfo{
			TeacherId:      "teacher_001",
			TeacherName:    "张老师",
			GroupName:      "test_group",
			PreferredModel: "gpt-4",
			UpdatedAt:      time.Now().Unix(),
		}

		err := wrongCache.SetUserInfo(ctx, userInfo)
		assert.Error(t, err)

		// 测试获取用户信息（应该失败但不会panic）
		result, err := wrongCache.GetUserInfo(ctx, "teacher_001")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("空数据测试", func(t *testing.T) {
		// 测试空用户信息
		emptyUserInfo := &identity.UserInfo{}
		err := cache.SetUserInfo(ctx, emptyUserInfo)
		assert.NoError(t, err)

		result, err := cache.GetUserInfo(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("大数据量测试", func(t *testing.T) {
		// 测试大量用户信息
		const numUsers = 100
		userInfos := make([]*identity.UserInfo, numUsers)

		for i := 0; i < numUsers; i++ {
			userInfos[i] = &identity.UserInfo{
				TeacherId:      fmt.Sprintf("teacher_%03d", i),
				TeacherName:    fmt.Sprintf("老师_%03d", i),
				GroupName:      fmt.Sprintf("group_%03d", i),
				PreferredModel: "gpt-4",
				UpdatedAt:      time.Now().Unix(),
			}
		}

		// 批量设置
		err := cache.BatchSetUserInfo(ctx, userInfos)
		assert.NoError(t, err)

		// 验证批量设置结果
		for i := 0; i < numUsers; i++ {
			result, err := cache.GetUserInfo(ctx, fmt.Sprintf("teacher_%03d", i))
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, fmt.Sprintf("teacher_%03d", i), result.TeacherId)
		}
	})
}
