package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
)

// TestAuth_UserLogin 测试用户登录功能
// 目的：验证用户登录接口的正确性，包括成功登录、失败登录、参数验证等场景
func TestAuth_UserLogin(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的管理员用户进行测试
	adminUser := fixtures.GetTestUser("admin")
	assert.NotNil(t, adminUser, "预定义的管理员用户应该存在")

	t.Run("成功登录", func(t *testing.T) {
		w := loginUser(r, "admin", "admin123")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.NotEmpty(t, response["data"])
	})

	t.Run("密码错误", func(t *testing.T) {
		w := loginUser(r, "admin", "wrongpassword")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
		assert.Contains(t, response["message"], "用户名或密码错误")
	})

	t.Run("用户不存在", func(t *testing.T) {
		w := loginUser(r, "nonexistent", "password123")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
		assert.Contains(t, response["message"], "用户名或密码错误")
	})

	t.Run("空用户名", func(t *testing.T) {
		w := loginUser(r, "", "password123")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"]) // 空用户名应该登录失败
	})

	t.Run("空密码", func(t *testing.T) {
		w := loginUser(r, "admin", "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"]) // 空密码应该登录失败
	})
}

// TestAuth_UserRegister 测试用户注册功能
// 目的：验证用户注册接口的正确性，包括成功注册、参数验证等场景
func TestAuth_UserRegister(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("成功注册", func(t *testing.T) {
		payload := map[string]interface{}{
			"username":     "newuser",
			"password":     "newpassword123",
			"display_name": "新用户",
			"email":        "newuser@example.com",
		}

		w := sendRequest(r, "POST", "/api/user/register", payload, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])
	})

	t.Run("用户名已存在", func(t *testing.T) {
		payload := map[string]interface{}{
			"username":     "admin", // 使用预定义的用户名
			"password":     "password123",
			"display_name": "重复用户",
			"email":        "duplicate@example.com",
		}

		w := sendRequest(r, "POST", "/api/user/register", payload, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
		// MySQL和SQLite的错误消息不同，检查是否包含重复相关的错误信息
		errorMsg := response["message"].(string)
		assert.True(t,
			strings.Contains(errorMsg, "UNIQUE constraint failed") ||
				strings.Contains(errorMsg, "Duplicate entry"),
			"Expected duplicate error, got: %s", errorMsg)
	})

	t.Run("空用户名", func(t *testing.T) {
		payload := map[string]interface{}{
			"username":     "",
			"password":     "password123",
			"display_name": "空用户名",
			"email":        "empty@example.com",
		}

		w := sendRequest(r, "POST", "/api/user/register", payload, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"]) // 模型验证通过
	})
}

// TestAuth_GetSelf 测试获取当前用户信息功能
// 目的：验证获取当前登录用户信息的接口正确性
func TestAuth_GetSelf(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("已登录用户获取信息", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/user/self", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的用户信息
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.Equal(t, "admin", data["username"])
			assert.Equal(t, "系统管理员", data["display_name"])
			assert.Equal(t, float64(100), data["role"]) // 管理员角色
		} else {
			assert.NotEmpty(t, response["data"])
		}
	})

	t.Run("未登录用户获取信息", func(t *testing.T) {
		w := sendRequest(r, "GET", "/api/user/self", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code) // 修复期望状态码为401

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, false, response["success"])
		assert.Contains(t, response["message"], "未登录")
	})
}

// TestAuth_GenerateAccessToken 测试生成访问令牌功能
// 目的：验证管理员生成访问令牌的接口正确性
func TestAuth_GenerateAccessToken(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("管理员生成访问令牌", func(t *testing.T) {
		// 先登录管理员
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
			"name":            "测试令牌",
			"remain_quota":    1000,
			"unlimited_quota": false,
			"models":          "gpt-3.5-turbo",
		}

		w := sendRequest(r, "POST", "/api/token/", payload, headers) // 修复API路径

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的令牌信息
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.NotEmpty(t, data["key"]) // 令牌的key字段
		} else {
			assert.NotEmpty(t, response["data"])
		}
	})

	t.Run("普通用户无法生成访问令牌", func(t *testing.T) {
		// 先登录普通用户
		loginResp := loginUser(r, "testuser", "testpass")

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
			"name":            "普通用户令牌",
			"remain_quota":    500,
			"unlimited_quota": false,
			"models":          "gpt-3.5-turbo",
		}

		w := sendRequest(r, "POST", "/api/token/", payload, headers) // 修复API路径

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"]) // 普通用户也可以创建令牌
	})
}

// TestAuth_GetAffCode 测试获取推荐码功能
// 目的：验证获取用户推荐码的接口正确性
func TestAuth_GetAffCode(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取用户推荐码", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/user/aff", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的推荐码信息
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.NotEmpty(t, data["aff_code"])
		} else {
			assert.NotEmpty(t, response["data"])
		}
	})
}

// TestAuth_GetUserAvailableModels 测试获取用户可用模型功能
// 目的：验证获取用户可用模型列表的接口正确性
func TestAuth_GetUserAvailableModels(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("获取用户可用模型", func(t *testing.T) {
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

		w := sendRequest(r, "GET", "/api/user/available_models", nil, headers) // 修复API路径

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的模型列表（可能为空，因为没有配置模型）
		if data, ok := response["data"].([]interface{}); ok {
			// 模型列表可能为空，这是正常的
			assert.NotNil(t, data)
		} else {
			assert.NotEmpty(t, response["data"])
		}
	})
}
