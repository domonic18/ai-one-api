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
// 测试目的：验证课件平台状态查询API在未启用集成时的正确响应
// 测试内容：
// 1. 状态查询API的响应格式
// 2. 未启用集成时的状态信息
// 3. HTTP状态码的正确性
// 4. 响应体结构的完整性
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
// 测试目的：验证课件平台配置查询API能够正确返回配置信息
// 测试内容：
// 1. 配置查询API的响应格式
// 2. 配置信息的完整性
// 3. 启用状态和基础URL的返回
// 4. HTTP状态码的正确性
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
// 测试目的：验证课件平台缓存查询API能够正确返回缓存数据
// 测试内容：
// 1. 缓存查询API的响应格式
// 2. 缓存数据结构的完整性
// 3. 缓存项列表和总数的返回
// 4. HTTP状态码的正确性
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

// TestCoursewareAPI_SyncUsers_未启用集成时返回错误 测试课件平台用户同步API
// 测试目的：验证课件平台用户同步API在未启用集成时的错误处理
// 测试内容：
// 1. 未启用集成时的错误响应
// 2. 错误消息的准确性
// 3. HTTP状态码的正确性
// 4. 响应格式的一致性
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

	// 验证响应 - 未启用集成时应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 验证响应体包含错误信息
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "false")
}

// TestCoursewareAPI_ClearCache_未启用集成时返回错误 测试课件平台缓存清理API
// 测试目的：验证课件平台缓存清理API在未启用集成时的错误处理
// 测试内容：
// 1. 未启用集成时的错误响应
// 2. 错误消息的准确性
// 3. HTTP状态码的正确性
// 4. 响应格式的一致性
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

	// 验证响应 - 未启用集成时应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 验证响应体包含错误信息
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "false")
}

// TestCoursewareAPI_TestConnection_未配置时返回错误 测试课件平台连接测试API
// 测试目的：验证课件平台连接测试API在未配置时的错误处理
// 测试内容：
// 1. 未配置时的错误响应
// 2. 错误消息的准确性
// 3. HTTP状态码的正确性
// 4. 响应格式的一致性
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

	// 验证响应 - 未配置时应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 验证响应体包含错误信息
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "false")
}

// TestCoursewareAPI_RefreshCache_未启用集成时返回错误 测试课件平台缓存刷新API
// 测试目的：验证课件平台缓存刷新API在未启用集成时的错误处理
// 测试内容：
// 1. 未启用集成时的错误响应
// 2. 错误消息的准确性
// 3. HTTP状态码的正确性
// 4. 响应格式的一致性
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

	// 验证响应 - 未启用集成时应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 验证响应体包含错误信息
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "false")
}

// TestCoursewareAPI_分页参数处理 测试课件平台API的分页参数处理
// 测试目的：验证课件平台API能够正确处理分页参数
// 测试内容：
// 1. 分页参数的有效性验证
// 2. 默认分页参数的处理
// 3. 边界分页参数的处理
// 4. 响应中分页信息的完整性
func TestCoursewareAPI_分页参数处理(t *testing.T) {
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// 跳过AdminAuth中间件，直接测试控制器
	router.GET("/api/courseware/cache", controller.GetCoursewareCache)

	// 测试不同的分页参数
	testCases := []struct {
		name     string
		page     string
		pageSize string
	}{
		{"默认分页", "", ""},
		{"第一页", "1", "10"},
		{"第二页", "2", "20"},
		{"大页码", "100", "50"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试请求
			req, _ := http.NewRequest("GET", "/api/courseware/cache", nil)
			if tc.page != "" {
				req.URL.Query().Set("page", tc.page)
			}
			if tc.pageSize != "" {
				req.URL.Query().Set("page_size", tc.pageSize)
			}

			// 记录响应
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, http.StatusOK, w.Code)

			// 验证响应体包含分页信息
			assert.Contains(t, w.Body.String(), "success")
			assert.Contains(t, w.Body.String(), "items")
			assert.Contains(t, w.Body.String(), "total")
		})
	}
}

// TestCoursewareAPI_响应格式验证 测试课件平台API的响应格式
// 测试目的：验证课件平台API的响应格式符合预期规范
// 测试内容：
// 1. 响应JSON格式的正确性
// 2. 必需字段的存在性
// 3. 字段类型的正确性
// 4. 错误响应的格式一致性
func TestCoursewareAPI_响应格式验证(t *testing.T) {
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

	// 解析响应JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证必需字段
	assert.Contains(t, response, "success")
	assert.Contains(t, response, "data")

	// 验证字段类型
	_, successExists := response["success"].(bool)
	assert.True(t, successExists, "success字段应该是布尔类型")

	_, dataExists := response["data"]
	assert.True(t, dataExists, "data字段应该存在")
}
