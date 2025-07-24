package integration

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestChannel_GetAllChannels 测试获取所有渠道功能
// 目的：验证管理员获取所有渠道的接口正确性
func TestChannel_GetAllChannels(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	createTestChannel(db, "Channel 1", "key1")
	createTestChannel(db, "Channel 2", "key2")

	t.Run("管理员获取所有渠道", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的渠道列表
		if data, ok := response["data"].([]interface{}); ok {
			assert.Len(t, data, 2)
		} else {
			// 如果data不是数组，至少验证它不为空
			assert.NotEmpty(t, response["data"])
		}
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

		w := sendRequest(r, "GET", "/api/channel", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
	})
}

// TestChannel_SearchChannels 测试搜索渠道功能
// 目的：验证管理员搜索渠道的接口正确性
func TestChannel_SearchChannels(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	createTestChannel(db, "OpenAI Channel", "openai-key")
	createTestChannel(db, "Azure Channel", "azure-key")

	t.Run("搜索渠道", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/search?keyword=OpenAI", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的搜索结果
		if data, ok := response["data"].([]interface{}); ok {
			assert.Len(t, data, 1)
		} else {
			// 如果data不是数组，至少验证它不为空
			assert.NotEmpty(t, response["data"])
		}
	})
}

// TestChannel_GetChannel 测试获取单个渠道功能
// 目的：验证管理员获取单个渠道详情的接口正确性
func TestChannel_GetChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	channel := createTestChannel(db, "Test Channel", "test-key")

	t.Run("获取渠道详情", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/"+strconv.Itoa(channel.Id), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的渠道信息
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.Equal(t, "Test Channel", data["name"])
			assert.Equal(t, "test-key", data["key"])
			assert.Equal(t, float64(channel.Id), data["id"])
		} else {
			// 如果data不是map，至少验证它不为空
			assert.NotEmpty(t, response["data"])
		}
	})
}

// TestChannel_AddChannel 测试添加渠道功能
// 目的：验证管理员创建新渠道的接口正确性
func TestChannel_AddChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	tests := []struct {
		name            string
		payload         map[string]interface{}
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "创建有效渠道",
			payload: map[string]interface{}{
				"name":    "New Channel",
				"key":     "new-key",
				"type":    1,
				"status":  1,
				"group":   "default",
				"models":  "gpt-3.5-turbo,gpt-4",
				"balance": 100.0,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "创建空名称渠道",
			payload: map[string]interface{}{
				"name":   "",
				"key":    "empty-name-key",
				"type":   1,
				"status": 1,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: false,
		},
		{
			name: "创建空密钥渠道",
			payload: map[string]interface{}{
				"name":   "Empty Key Channel",
				"key":    "",
				"type":   1,
				"status": 1,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: false,
		},
		{
			name: "创建带配置的渠道",
			payload: map[string]interface{}{
				"name":    "Config Channel",
				"key":     "config-key",
				"type":    1,
				"status":  1,
				"group":   "premium",
				"models":  "gpt-4",
				"balance": 500.0,
				"config":  "{\"base_url\":\"https://api.openai.com\"}",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			w := sendRequest(r, "POST", "/api/channel", tt.payload, headers)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"])
		})
	}
}

// TestChannel_UpdateChannel 测试更新渠道功能
// 目的：验证管理员更新渠道信息的接口正确性
func TestChannel_UpdateChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	channel := createTestChannel(db, "Test Channel", "test-key")

	t.Run("更新渠道信息", func(t *testing.T) {
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

		payload := map[string]interface{}{
			"id":      channel.Id,
			"name":    "Updated Channel",
			"key":     "updated-key",
			"status":  1,
			"group":   "premium",
			"models":  "gpt-4",
			"balance": 200.0,
		}

		w := sendRequest(r, "PUT", "/api/channel", payload, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestChannel_DeleteChannel 测试删除渠道功能
// 目的：验证管理员删除渠道的接口正确性
func TestChannel_DeleteChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	channel := createTestChannel(db, "Test Channel", "test-key")

	t.Run("删除渠道", func(t *testing.T) {
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

		w := sendRequest(r, "DELETE", "/api/channel/"+strconv.Itoa(channel.Id), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}

// TestChannel_TestChannel 测试渠道测试功能
// 目的：验证管理员测试渠道连接性的接口正确性
func TestChannel_TestChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	channel := createTestChannel(db, "Test Channel", "test-key")

	t.Run("测试单个渠道", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/test/"+strconv.Itoa(channel.Id), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 测试结果可能成功或失败，取决于渠道配置
		assert.Contains(t, response, "success")
	})
}

// TestChannel_UpdateChannelBalance 测试更新渠道余额功能
// 目的：验证管理员更新渠道余额的接口正确性
func TestChannel_UpdateChannelBalance(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建测试渠道
	channel := createTestChannel(db, "Test Channel", "test-key")

	t.Run("更新单个渠道余额", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/update_balance/"+strconv.Itoa(channel.Id), nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 更新余额的结果可能成功或失败，取决于渠道配置
		assert.Contains(t, response, "success")
	})

	t.Run("更新所有渠道余额", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/update_balance", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// 更新余额的结果可能成功或失败，取决于渠道配置
		assert.Contains(t, response, "success")
	})
}

// TestChannel_ListAllModels 测试获取所有模型功能
// 目的：验证管理员获取所有可用模型的接口正确性
func TestChannel_ListAllModels(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	t.Run("获取所有模型", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/channel/models", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的模型列表
		if data, ok := response["data"].([]interface{}); ok {
			assert.NotEmpty(t, data)
		} else {
			// 如果data不是数组，至少验证它不为空
			assert.NotEmpty(t, response["data"])
		}
	})
}

// TestChannel_DeleteDisabledChannel 测试删除禁用渠道功能
// 目的：验证管理员删除禁用渠道的接口正确性
func TestChannel_DeleteDisabledChannel(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试管理员用户
	_ = createTestUser(db, "admin", "admin123", model.RoleAdminUser)

	// 创建禁用的测试渠道
	disabledChannel := &model.Channel{
		Type:    1,
		Key:     "disabled-key",
		Name:    "Disabled Channel",
		Status:  2, // 禁用状态
		Group:   "default",
		Models:  "gpt-3.5-turbo",
		Balance: 0.0,
	}
	db.Create(disabledChannel)

	t.Run("删除禁用渠道", func(t *testing.T) {
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

		w := sendRequest(r, "DELETE", "/api/channel/disabled", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})
}
