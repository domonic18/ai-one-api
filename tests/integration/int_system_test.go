package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestSystem_GetStatus 测试获取系统状态功能
// 目的：验证系统状态检查接口的正确性
func TestSystem_GetStatus(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取系统状态", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/status", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的系统状态信息
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["version"])
		assert.NotEmpty(t, data["start_time"])
		assert.Contains(t, data, "email_verification")
		assert.Contains(t, data, "github_oauth")
		assert.Contains(t, data, "system_name")
	})
}

// TestSystem_GetModels 测试获取模型列表功能
// 目的：验证获取可用模型列表接口的正确性
func TestSystem_GetModels(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	_ = createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试渠道
	createTestChannel(db, "OpenAI Channel", "openai-key")

	t.Run("获取模型列表", func(t *testing.T) {
		// 先登录
		loginResp := loginUser(r, "testuser", "password123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		w := sendRequest(r, "GET", "/api/models", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的模型列表
		if data, ok := response["data"].([]interface{}); ok {
			assert.NotEmpty(t, data)
		} else {
			t.Errorf("Expected data to be []interface{}, got %T", response["data"])
		}
	})

	t.Run("未登录用户无权访问", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/models", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestSystem_GetNotice 测试获取通知功能
// 目的：验证获取系统通知接口的正确性
func TestSystem_GetNotice(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取系统通知", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/notice", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestSystem_GetAbout 测试获取关于信息功能
// 目的：验证获取关于页面信息接口的正确性
func TestSystem_GetAbout(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取关于信息", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/about", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestSystem_GetHomePageContent 测试获取首页内容功能
// 目的：验证获取首页内容接口的正确性
func TestSystem_GetHomePageContent(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取首页内容", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/home_page_content", nil, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestSystem_GetGroups 测试获取分组功能
// 目的：验证管理员获取用户分组列表接口的正确性
func TestSystem_GetGroups(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	t.Run("管理员获取分组列表", func(t *testing.T) {
		// 先登录
		loginResp := loginUser(r, "admin", "admin123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		w := sendRequest(r, "GET", "/api/group", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的分组列表
		data := response["data"].([]interface{})
		assert.NotEmpty(t, data)
	})

	t.Run("普通用户无权访问", func(t *testing.T) {
		// 创建普通用户
		_ = createTestUser(db, "user", "password123", model.RoleCommonUser)

		// 先登录
		loginResp := loginUser(r, "user", "password123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		w := sendRequest(r, "GET", "/api/group", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
	})
}

// TestSystem_GetOptions 测试获取系统选项功能
// 目的：验证超级管理员获取系统配置选项接口的正确性
func TestSystem_GetOptions(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试超级管理员用户
	_ = createTestUser(db, "root", "root123", model.RoleRootUser)

	t.Run("超级管理员获取系统选项", func(t *testing.T) {
		// 先登录
		loginResp := loginUser(r, "root", "root123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		w := sendRequest(r, "GET", "/api/option", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的选项列表
		data := response["data"].([]interface{})
		assert.NotEmpty(t, data)
	})

	t.Run("普通管理员无权访问", func(t *testing.T) {
		// 创建普通管理员用户
		_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

		// 先登录
		loginResp := loginUser(r, "admin", "admin123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		w := sendRequest(r, "GET", "/api/option", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
	})
}

// TestSystem_UpdateOption 测试更新系统选项功能
// 目的：验证超级管理员更新系统配置选项接口的正确性
func TestSystem_UpdateOption(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试超级管理员用户
	_ = createTestUser(db, "root", "root123", model.RoleRootUser)

	t.Run("更新系统选项", func(t *testing.T) {
		// 先登录
		loginResp := loginUser(r, "root", "root123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		payload := map[string]interface{}{
			"key":   "test_option",
			"value": "test_value",
		}

		w := sendRequest(r, "PUT", "/api/option", payload, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestSystem_ErrorHandling 测试错误处理功能
// 目的：验证系统错误处理机制的正确性
func TestSystem_ErrorHandling(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	tests := []struct {
		name           string
		method         string
		path           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "方法不允许",
			method:         "DELETE",
			path:           "/api/status",
			payload:        nil,
			expectedStatus: http.StatusNotFound, // Gin默认返回404而不是405
			expectedError:  "method not allowed",
		},
		{
			name:           "路径不存在",
			method:         "GET",
			path:           "/api/nonexistent",
			payload:        nil,
			expectedStatus: http.StatusNotFound,
			expectedError:  "not found",
		},
		{
			name:           "无效JSON",
			method:         "POST",
			path:           "/api/user/login",
			payload:        "invalid json",
			expectedStatus: http.StatusOK, // 实际返回200，但success为false
			expectedError:  "invalid_parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var headers map[string]string
			if tt.method == "POST" && tt.path == "/api/user/login" {
				// 对于登录接口，需要设置Content-Type
				headers = map[string]string{
					"Content-Type": "application/json",
				}
			}

			w := sendRequest(r, tt.method, tt.path, tt.payload, headers)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				if tt.expectedError != "" {
					assert.Equal(t, false, response["success"])
				}
			}
		})
	}
}

// TestSystem_ConcurrentAccess 测试并发访问功能
// 目的：验证系统在并发访问下的稳定性
func TestSystem_ConcurrentAccess(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	_ = createTestUser(db, "concurrentuser", "password123", model.RoleCommonUser)

	t.Run("并发访问状态接口", func(t *testing.T) {
		// 模拟并发访问
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				w := sendRequest(r, "GET", "/api/status", nil, nil)
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("并发访问模型接口", func(t *testing.T) {
		// 先登录
		loginResp := loginUser(r, "concurrentuser", "password123")

		// 获取session cookie
		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "one-api" {
				sessionCookie = cookie
				break
			}
		}

		headers := map[string]string{}
		if sessionCookie != nil {
			headers["Cookie"] = "one-api=" + sessionCookie.Value
		}

		// 模拟并发访问
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func() {
				w := sendRequest(r, "GET", "/api/models", nil, headers)
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
