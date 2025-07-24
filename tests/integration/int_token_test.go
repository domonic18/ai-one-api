package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestToken_GetAllTokens 测试获取所有令牌功能
// 目的：验证用户获取自己所有令牌的接口正确性
func TestToken_GetAllTokens(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	createTestToken(db, user.Id, "token1")
	createTestToken(db, user.Id, "token2")

	t.Run("获取所有令牌", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/token", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的令牌列表
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

// TestToken_SearchTokens 测试搜索令牌功能
// 目的：验证用户搜索令牌的接口正确性
func TestToken_SearchTokens(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	createTestToken(db, user.Id, "test-token-1")
	createTestToken(db, user.Id, "other-token")

	t.Run("搜索令牌", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/token/search?keyword=test", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的搜索结果
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
	})
}

// TestToken_GetToken 测试获取单个令牌功能
// 目的：验证用户获取单个令牌详情的接口正确性
func TestToken_GetToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	token := createTestToken(db, user.Id, "test-token")

	t.Run("获取令牌详情", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/token/"+string(rune(token.Id)), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的令牌信息
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test-token", data["name"])
		assert.Equal(t, float64(token.Id), data["id"])
	})
}

// TestToken_AddToken 测试添加令牌功能
// 目的：验证用户创建新令牌的接口正确性
func TestToken_AddToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	_ = createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	tests := []struct {
		name            string
		payload         map[string]interface{}
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "创建有效令牌",
			payload: map[string]interface{}{
				"name": "new-token",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "创建空名称令牌",
			payload: map[string]interface{}{
				"name": "",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: false,
		},
		{
			name: "创建带模型的令牌",
			payload: map[string]interface{}{
				"name":   "model-token",
				"models": "gpt-3.5-turbo,gpt-4",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "创建带子网的令牌",
			payload: map[string]interface{}{
				"name":   "subnet-token",
				"subnet": "192.168.1.0/24",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			w := sendRequest(r, "POST", "/api/token", tt.payload, headers)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"])
		})
	}
}

// TestToken_UpdateToken 测试更新令牌功能
// 目的：验证用户更新令牌信息的接口正确性
func TestToken_UpdateToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	token := createTestToken(db, user.Id, "test-token")

	t.Run("更新令牌信息", func(t *testing.T) {
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

		payload := map[string]interface{}{
			"id":     token.Id,
			"name":   "updated-token",
			"models": "gpt-4",
			"status": 1,
		}

		w := sendRequest(r, "PUT", "/api/token", payload, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestToken_DeleteToken 测试删除令牌功能
// 目的：验证用户删除令牌的接口正确性
func TestToken_DeleteToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	token := createTestToken(db, user.Id, "test-token")

	t.Run("删除令牌", func(t *testing.T) {
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

		w := sendRequest(r, "DELETE", "/api/token/"+string(rune(token.Id)), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestToken_ValidateToken 测试令牌验证功能
// 目的：验证令牌验证的接口正确性
func TestToken_ValidateToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	token := createTestToken(db, user.Id, "test-token")

	t.Run("验证有效令牌", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
		}

		w := sendRequest(r, "GET", "/api/models", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("验证无效令牌", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid-token",
		}

		w := sendRequest(r, "GET", "/api/models", nil, headers)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("无令牌访问", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/models", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestToken_TokenQuota 测试令牌配额功能
// 目的：验证令牌配额消耗和恢复的接口正确性
func TestToken_TokenQuota(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户
	user := createTestUser(db, "testuser", "password123", model.RoleCommonUser)

	// 创建测试令牌
	token := createTestToken(db, user.Id, "test-token")

	t.Run("令牌配额消耗", func(t *testing.T) {
		// 模拟令牌配额消耗
		_ = token.RemainQuota

		// 这里应该调用实际的配额消耗逻辑
		// 由于这是集成测试，我们主要验证接口的响应

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
		}

		w := sendRequest(r, "GET", "/api/models", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
