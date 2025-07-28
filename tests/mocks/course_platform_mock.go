package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 模拟数据结构
type TeacherInfo struct {
	TeacherId   string `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	SchoolId    int    `json:"school_id"`
	SchoolName  string `json:"school_name"`
	SubjectId   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type SubjectInfo struct {
	SubjectId    int    `json:"subject_id"`
	SubjectName  string `json:"subject_name"`
	SchoolId     int    `json:"school_id"`
	SchoolName   string `json:"school_name"`
	DefaultModel string `json:"default_model"`
	GroupName    string `json:"group_name"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type UserConfig struct {
	UserId     string `json:"user_id"`
	ModelName  string `json:"model_name"`
	Parameters string `json:"parameters,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// 模拟数据存储
type MockDataStore struct {
	mu           sync.RWMutex
	teachers     map[string]*TeacherInfo
	subjects     map[int]*SubjectInfo
	userConfigs  map[string]*UserConfig
	requestCount map[string]int
}

// 创建新的数据存储
func NewMockDataStore() *MockDataStore {
	store := &MockDataStore{
		teachers:     make(map[string]*TeacherInfo),
		subjects:     make(map[int]*SubjectInfo),
		userConfigs:  make(map[string]*UserConfig),
		requestCount: make(map[string]int),
	}

	// 添加一些默认数据
	now := time.Now().Unix()

	// 添加学科组
	store.subjects[1] = &SubjectInfo{
		SubjectId:    1,
		SubjectName:  "语文组",
		SchoolId:     1,
		SchoolName:   "示例学校",
		DefaultModel: "gpt-3.5-turbo",
		GroupName:    "chinese",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	store.subjects[2] = &SubjectInfo{
		SubjectId:    2,
		SubjectName:  "数学组",
		SchoolId:     1,
		SchoolName:   "示例学校",
		DefaultModel: "gpt-4",
		GroupName:    "math",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 添加老师
	store.teachers["teacher1"] = &TeacherInfo{
		TeacherId:   "teacher1",
		TeacherName: "张老师",
		SchoolId:    1,
		SchoolName:  "示例学校",
		SubjectId:   1,
		SubjectName: "语文组",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	store.teachers["teacher2"] = &TeacherInfo{
		TeacherId:   "teacher2",
		TeacherName: "李老师",
		SchoolId:    1,
		SchoolName:  "示例学校",
		SubjectId:   2,
		SubjectName: "数学组",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 添加用户模型配置
	store.userConfigs["teacher1"] = &UserConfig{
		UserId:    "teacher1",
		ModelName: "gpt-4",
		CreatedAt: now,
		UpdatedAt: now,
	}

	return store
}

// 记录API请求
func (s *MockDataStore) recordRequest(endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestCount[endpoint]++
}

// 获取API请求统计
func (s *MockDataStore) getRequestStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := make(map[string]int)
	for k, v := range s.requestCount {
		stats[k] = v
	}
	return stats
}

// 课件平台模拟服务器
type CoursePlatformMock struct {
	server *httptest.Server
	store  *MockDataStore
}

// 创建新的模拟服务器
func NewCoursePlatformMock() *CoursePlatformMock {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	store := NewMockDataStore()
	mock := &CoursePlatformMock{
		store: store,
	}

	// 设置API路由
	router.GET("/teacher/:id/info", mock.getTeacherInfo)
	router.GET("/subject/:id", mock.getSubjectInfo)
	router.GET("/teacher/:id/model", mock.getUserConfig)
	router.GET("/stats", mock.getStats)

	// 创建测试服务器
	mock.server = httptest.NewServer(router)
	return mock
}

// 获取服务器URL
func (m *CoursePlatformMock) URL() string {
	return m.server.URL
}

// 关闭服务器
func (m *CoursePlatformMock) Close() {
	m.server.Close()
}

// 获取老师信息API
func (m *CoursePlatformMock) getTeacherInfo(c *gin.Context) {
	m.store.recordRequest("getTeacherInfo")

	// 检查授权
	authHeader := c.GetHeader("Authorization")
	if !m.validateAuth(authHeader) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	teacherId := c.Param("id")
	m.store.mu.RLock()
	teacher, exists := m.store.teachers[teacherId]
	m.store.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到老师信息"})
		return
	}

	c.JSON(http.StatusOK, teacher)
}

// 获取学科组信息API
func (m *CoursePlatformMock) getSubjectInfo(c *gin.Context) {
	m.store.recordRequest("getSubjectInfo")

	// 检查授权
	authHeader := c.GetHeader("Authorization")
	if !m.validateAuth(authHeader) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	subjectId := c.Param("id")
	id := 0
	_, err := fmt.Sscanf(subjectId, "%d", &id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的学科组ID"})
		return
	}

	m.store.mu.RLock()
	subject, exists := m.store.subjects[id]
	m.store.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到学科组信息"})
		return
	}

	c.JSON(http.StatusOK, subject)
}

// 获取用户模型配置API
func (m *CoursePlatformMock) getUserConfig(c *gin.Context) {
	m.store.recordRequest("getUserConfig")

	// 检查授权
	authHeader := c.GetHeader("Authorization")
	if !m.validateAuth(authHeader) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	userId := c.Param("id")
	m.store.mu.RLock()
	config, exists := m.store.userConfigs[userId]
	m.store.mu.RUnlock()

	if !exists {
		// 对于不存在的用户，返回404而不是默认配置
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到用户配置"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 获取请求统计API
func (m *CoursePlatformMock) getStats(c *gin.Context) {
	stats := m.store.getRequestStats()
	c.JSON(http.StatusOK, stats)
}

// 验证授权
func (m *CoursePlatformMock) validateAuth(authHeader string) bool {
	// 简单验证，检查是否有Bearer前缀
	return len(authHeader) > 7 && authHeader[:7] == "Bearer "
}

// 添加老师信息
func (m *CoursePlatformMock) AddTeacher(teacher *TeacherInfo) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	m.store.teachers[teacher.TeacherId] = teacher
}

// 添加学科组信息
func (m *CoursePlatformMock) AddSubject(subject *SubjectInfo) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	m.store.subjects[subject.SubjectId] = subject
}

// 添加用户模型配置
func (m *CoursePlatformMock) AddUserConfig(config *UserConfig) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	m.store.userConfigs[config.UserId] = config
}

// 序列化为JSON
func (m *CoursePlatformMock) ToJSON() string {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()

	data := map[string]interface{}{
		"teachers":     m.store.teachers,
		"subjects":     m.store.subjects,
		"userConfigs":  m.store.userConfigs,
		"requestCount": m.store.requestCount,
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}
