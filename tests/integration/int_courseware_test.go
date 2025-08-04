package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/controller"
	"github.com/stretchr/testify/assert"
)

// TestCoursewareAPI_GetStatus_未启用集成 测试课件平台状态API
func TestCoursewareAPI_GetStatus_未启用集成(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.GET("/api/courseware/status", controller.GetCoursewareStatus)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/courseware/status", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证响应体包含success字段
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "data")
}

// TestCoursewareAPI_GetConfig_返回配置信息 测试课件平台配置API
func TestCoursewareAPI_GetConfig_返回配置信息(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.GET("/api/courseware/config", controller.GetCoursewareConfig)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/courseware/config", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证响应体包含配置信息
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "enabled")
	assert.Contains(t, w.Body.String(), "base_url")
}

// TestCoursewareAPI_GetCache_返回缓存数据 测试课件平台缓存API
func TestCoursewareAPI_GetCache_返回缓存数据(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.GET("/api/courseware/cache", controller.GetCoursewareCache)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/courseware/cache", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证响应体包含缓存数据结构
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "items")
	assert.Contains(t, w.Body.String(), "total")
}

// TestCoursewareAPI_SyncUsers_未启用集成时返回错误 测试课件平台同步API
func TestCoursewareAPI_SyncUsers_未启用集成时返回错误(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.POST("/api/courseware/sync", controller.SyncCoursewareUsers)

	// 创建测试请求
	req, _ := http.NewRequest("POST", "/api/courseware/sync", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 由于未启用集成，应该返回错误
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "课件平台集成未启用")
}

// TestCoursewareAPI_ClearCache_未启用集成时返回错误 测试课件平台清理缓存API
func TestCoursewareAPI_ClearCache_未启用集成时返回错误(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.DELETE("/api/courseware/cache", controller.ClearCoursewareCache)

	// 创建测试请求
	req, _ := http.NewRequest("DELETE", "/api/courseware/cache", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 由于未启用集成，应该返回错误
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "课件平台集成未启用")
}

// TestCoursewareAPI_TestConnection_未配置时返回错误 测试课件平台连接测试API
func TestCoursewareAPI_TestConnection_未配置时返回错误(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.POST("/api/courseware/test", controller.TestCoursewareConnection)

	// 创建测试请求
	req, _ := http.NewRequest("POST", "/api/courseware/test", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 由于未配置API客户端，应该返回错误
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "课件平台API客户端未正确配置")
}

// TestCoursewareAPI_RefreshCache_未启用集成时返回错误 测试课件平台刷新缓存API
func TestCoursewareAPI_RefreshCache_未启用集成时返回错误(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.POST("/api/courseware/cache/refresh", controller.RefreshCoursewareCache)

	// 创建测试请求
	req, _ := http.NewRequest("POST", "/api/courseware/cache/refresh", nil)

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 由于未启用集成，应该返回错误
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "课件平台集成未启用")
}

// TestCoursewareAPI_分页参数处理 测试缓存API的分页功能
func TestCoursewareAPI_分页参数处理(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	router.GET("/api/courseware/cache", controller.GetCoursewareCache)

	// 测试不同的分页参数
	testCases := []struct {
		name     string
		query    string
		expected int
	}{
		{"默认分页", "", http.StatusOK},
		{"自定义页码", "?page=2", http.StatusOK},
		{"自定义大小", "?size=10", http.StatusOK},
		{"搜索参数", "?search=test", http.StatusOK},
		{"无效页码", "?page=0", http.StatusOK},
		{"无效大小", "?size=0", http.StatusOK},
		{"超大大小", "?size=1000", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/courseware/cache"+tc.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
			assert.Contains(t, w.Body.String(), "success")
		})
	}
}

// TestCoursewareAPI_响应格式验证 测试API响应格式
func TestCoursewareAPI_响应格式验证(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	router.GET("/api/courseware/status", controller.GetCoursewareStatus)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/api/courseware/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应格式
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应结构
	assert.Contains(t, response, "success")
	assert.Contains(t, response, "data")

	// 验证数据类型
	assert.IsType(t, bool(true), response["success"])
	assert.IsType(t, map[string]interface{}{}, response["data"])
}
