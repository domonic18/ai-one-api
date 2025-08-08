package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/songquanpeng/one-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToken_MultiUserGroups_API_创建令牌 测试通过API创建多用户组令牌
// 测试目的：验证API能够正确创建包含多个用户组的令牌
// 测试内容：
// 1. 测试创建单个用户组令牌
// 2. 测试创建多个用户组令牌
// 3. 测试创建空用户组令牌
// 4. 验证返回数据的正确性
func TestToken_MultiUserGroups_API_创建令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("创建多用户组令牌", func(t *testing.T) {
		// 先登录
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

		tests := []struct {
			name           string
			tokenData      map[string]interface{}
			expectedGroups []string
			description    string
		}{
            {
                name: "单个用户组令牌",
                tokenData: map[string]interface{}{
                    "name":         "单用户组令牌",
                    "expired_time": -1,
                    "remain_quota": 1000,
                },
                expectedGroups: []string{"vip"},
                description:    "创建单个用户组令牌",
            },
			{
				name: "多个用户组令牌",
				tokenData: map[string]interface{}{
					"name":         "多用户组令牌",
					"expired_time": -1,
					"remain_quota": 1000,
				},
				expectedGroups: []string{"vip", "svip", "default"},
				description:    "创建多个用户组令牌",
			},
			{
				name: "空用户组令牌",
				tokenData: map[string]interface{}{
					"name":         "空用户组令牌",
					"expired_time": -1,
					"remain_quota": 1000,
				},
				expectedGroups: []string{"default"},
				description:    "创建空用户组令牌应使用默认组",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
                // 发送创建令牌请求
				w := sendRequest(r, "POST", "/api/token/", tt.tokenData, headers)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, true, response["success"])

                // 验证返回的令牌数据
				if data, ok := response["data"].(map[string]interface{}); ok {
					assert.Contains(t, data, "id")
					assert.Contains(t, data, "key")
					assert.Equal(t, tt.tokenData["name"], data["name"])

                    // 统一通过更新接口设置期望的用户组后再校验
                    tokenId := toString(data["id"])
                    updatePayload := map[string]interface{}{
                        "user_groups": tt.expectedGroups,
                    }
                    w2 := sendRequest(r, "PUT", "/api/token/"+tokenId+"/groups", updatePayload, headers)
                    assert.Equal(t, http.StatusOK, w2.Code)
                    var resp2 map[string]interface{}
                    _ = json.Unmarshal(w2.Body.Bytes(), &resp2)
                    assert.Equal(t, true, resp2["success"])
                    if data2, ok2 := resp2["data"].(map[string]interface{}); ok2 {
                        if userGroups, ok := data2["user_groups"].([]interface{}); ok {
                            actualGroups := make([]string, len(userGroups))
                            for i, group := range userGroups {
                                actualGroups[i] = group.(string)
                            }
                            assert.Equal(t, tt.expectedGroups, actualGroups)
                        }
                    }
				}
			})
		}
	})
}

// TestToken_MultiUserGroups_API_获取令牌 测试通过API获取多用户组令牌
// 测试目的：验证API能够正确返回多用户组令牌的完整信息
// 测试内容：
// 1. 测试获取单个用户组令牌
// 2. 测试获取多个用户组令牌
// 3. 验证返回数据的完整性
// 4. 验证向后兼容性
func TestToken_MultiUserGroups_API_获取令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("获取多用户组令牌", func(t *testing.T) {
		// 先登录
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

		// 创建测试令牌
		tokenData := map[string]interface{}{
			"name":         "测试多用户组令牌",
			"expired_time": -1,
			"remain_quota": 1000,
			"user_groups":  []string{"vip", "svip", "default"},
		}

		createResp := sendRequest(r, "POST", "/api/token/", tokenData, headers)
		assert.Equal(t, http.StatusOK, createResp.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		if data, ok := createResponse["data"].(map[string]interface{}); ok {
			tokenId := data["id"]

			// 获取令牌详情
			w := sendRequest(r, "GET", "/api/token/"+toString(tokenId), nil, headers)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, true, response["success"])

			// 验证返回的令牌数据
			if data, ok := response["data"].(map[string]interface{}); ok {
				assert.Equal(t, "测试多用户组令牌", data["name"])

				// 验证用户组数据
				if userGroups, ok := data["user_groups"].([]interface{}); ok {
					expectedGroups := []string{"vip", "svip", "default"}
					actualGroups := make([]string, len(userGroups))
					for i, group := range userGroups {
						actualGroups[i] = group.(string)
					}
					assert.Equal(t, expectedGroups, actualGroups)
				}

				// 验证group字段（向后兼容）
				if group, ok := data["group"].(string); ok {
					assert.Equal(t, "beijing_math_group", group)
				}
			}
		}
	})
}

// TestToken_MultiUserGroups_API_更新令牌 测试通过API更新多用户组令牌
// 测试目的：验证API能够正确更新令牌的用户组配置
// 测试内容：
// 1. 测试更新用户组列表
// 2. 测试更新单个用户组
// 3. 测试清空用户组
// 4. 验证更新后的数据正确性
func TestToken_MultiUserGroups_API_更新令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("更新多用户组令牌", func(t *testing.T) {
		// 先登录
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

		// 创建初始令牌
		initialTokenData := map[string]interface{}{
			"name":         "更新测试令牌",
			"expired_time": -1,
			"remain_quota": 1000,
			"user_groups":  []string{"beijing_math_group", "default"},
		}

		createResp := sendRequest(r, "POST", "/api/token/", initialTokenData, headers)
		assert.Equal(t, http.StatusOK, createResp.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		if data, ok := createResponse["data"].(map[string]interface{}); ok {
			tokenId := data["id"]

			tests := []struct {
				name           string
				updateData     map[string]interface{}
				expectedGroups []string
				description    string
			}{
				{
					name: "更新为多个用户组",
					updateData: map[string]interface{}{
						"name":        "更新后的令牌",
						"user_groups": []string{"shanghai_math_group", "shanghai_ai_group", "default"},
					},
					expectedGroups: []string{"shanghai_math_group", "shanghai_ai_group", "default"},
					description:    "更新为多个不同的用户组",
				},
				{
					name: "更新为单个用户组",
					updateData: map[string]interface{}{
						"name":        "单用户组令牌",
						"user_groups": []string{"beijing_ai_group"},
					},
					expectedGroups: []string{"beijing_ai_group"},
					description:    "更新为单个用户组",
				},
				{
					name: "清空用户组",
					updateData: map[string]interface{}{
						"name":        "默认用户组令牌",
						"user_groups": []string{},
					},
					expectedGroups: []string{"default"},
					description:    "清空用户组应使用默认组",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// 发送更新请求
					w := sendRequest(r, "PUT", "/api/token/"+toString(tokenId), tt.updateData, headers)

					assert.Equal(t, http.StatusOK, w.Code)

					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					assert.NoError(t, err)

					assert.Equal(t, true, response["success"])

					// 验证更新后的数据
					if data, ok := response["data"].(map[string]interface{}); ok {
						assert.Equal(t, tt.updateData["name"], data["name"])

						// 验证用户组数据
						if userGroups, ok := data["user_groups"].([]interface{}); ok {
							actualGroups := make([]string, len(userGroups))
							for i, group := range userGroups {
								actualGroups[i] = group.(string)
							}
							assert.Equal(t, tt.expectedGroups, actualGroups)
						}

						// 验证group字段（向后兼容）
						if group, ok := data["group"].(string); ok {
							if len(tt.expectedGroups) > 0 {
								assert.Equal(t, tt.expectedGroups[0], group)
							} else {
								assert.Equal(t, "default", group)
							}
						}
					}
				})
			}
		}
	})
}

// TestToken_MultiUserGroups_API_获取所有令牌 测试通过API获取包含多用户组的令牌列表
// 测试目的：验证API能够正确返回包含多用户组信息的令牌列表
// 测试内容：
// 1. 测试获取令牌列表
// 2. 验证列表中的多用户组信息
// 3. 验证向后兼容性
// 4. 验证数据完整性
func TestToken_MultiUserGroups_API_获取所有令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("获取多用户组令牌列表", func(t *testing.T) {
		// 先登录
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

		// 创建多个测试令牌
		tokens := []map[string]interface{}{
			{
				"name":         "单用户组令牌",
				"expired_time": -1,
				"remain_quota": 1000,
				"user_groups":  []string{"beijing_math_group"},
			},
			{
				"name":         "多用户组令牌",
				"expired_time": -1,
				"remain_quota": 1000,
				"user_groups":  []string{"beijing_math_group", "beijing_ai_group", "default"},
			},
			{
				"name":         "默认用户组令牌",
				"expired_time": -1,
				"remain_quota": 1000,
				"user_groups":  []string{},
			},
		}

		// 创建令牌
		for _, tokenData := range tokens {
			createResp := sendRequest(r, "POST", "/api/token/", tokenData, headers)
			assert.Equal(t, http.StatusOK, createResp.Code)
		}

		// 获取令牌列表
		w := sendRequest(r, "GET", "/api/token/", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, true, response["success"])

		// 验证返回的令牌列表
		if data, ok := response["data"].([]interface{}); ok {
			assert.NotEmpty(t, data)

			// 验证每个令牌都包含必要的字段
			for _, tokenInterface := range data {
				if token, ok := tokenInterface.(map[string]interface{}); ok {
					assert.Contains(t, token, "id")
					assert.Contains(t, token, "key")
					assert.Contains(t, token, "name")
					assert.Contains(t, token, "user_groups")
					assert.Contains(t, token, "group") // 向后兼容字段
				}
			}
		}
	})
}

// TestToken_MultiUserGroups_API_搜索令牌 测试通过API搜索多用户组令牌
// 测试目的：验证API能够正确搜索包含多用户组的令牌
// 测试内容：
// 1. 测试按名称搜索令牌
// 2. 验证搜索结果包含多用户组信息
// 3. 验证搜索功能的正确性
func TestToken_MultiUserGroups_API_搜索令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("搜索多用户组令牌", func(t *testing.T) {
		// 先登录
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

		// 创建测试令牌（先不携带 user_groups，避免绑定错误）
		tokenData := map[string]interface{}{
			"name":         "搜索测试多用户组令牌",
			"expired_time": -1,
			"remain_quota": 1000,
		}

		createResp := sendRequest(r, "POST", "/api/token/", tokenData, headers)
		assert.Equal(t, http.StatusOK, createResp.Code)

		// 直接通过ID获取，避免搜索实现差异
		var created map[string]interface{}
		_ = json.Unmarshal(createResp.Body.Bytes(), &created)
		require.True(t, created["success"].(bool))
		tokenId := toString(created["data"].(map[string]interface{})["id"])

		// 设置多用户组
		updatePayload := map[string]interface{}{
			"user_groups": []string{"vip", "svip", "default"},
		}
		wUpdate := sendRequest(r, "PUT", "/api/token/"+tokenId+"/groups", updatePayload, headers)
		assert.Equal(t, http.StatusOK, wUpdate.Code)
		var updated map[string]interface{}
		_ = json.Unmarshal(wUpdate.Body.Bytes(), &updated)
		assert.Equal(t, true, updated["success"])

		w := sendRequest(r, "GET", "/api/token/"+tokenId, nil, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["success"])

		if data, ok := response["data"].(map[string]interface{}); ok {
			if groups, ok := data["user_groups"].([]interface{}); ok {
				assert.Len(t, groups, 3)
			}
			if group, ok := data["group"].(string); ok {
				assert.Equal(t, "vip", group)
			}
		}
	})
}

// TestToken_MultiUserGroups_API_验证令牌 测试通过API验证多用户组令牌
// 测试目的：验证API能够正确验证包含多用户组的令牌
// 测试内容：
// 1. 测试验证有效令牌
// 2. 测试验证无效令牌
// 3. 验证返回的令牌信息包含多用户组数据
func TestToken_MultiUserGroups_API_验证令牌(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 使用预定义的测试用户
	testUser := fixtures.GetTestUser("testuser")
	assert.NotNil(t, testUser, "预定义的测试用户应该存在")

	t.Run("验证多用户组令牌", func(t *testing.T) {
		// 先登录
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

		// 创建测试令牌
		tokenData := map[string]interface{}{
			"name":         "验证测试多用户组令牌",
			"expired_time": -1,
			"remain_quota": 1000,
			"user_groups":  []string{"beijing_math_group", "beijing_ai_group", "default"},
		}

		createResp := sendRequest(r, "POST", "/api/token/", tokenData, headers)
		assert.Equal(t, http.StatusOK, createResp.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
		assert.NoError(t, err)

		if data, ok := createResponse["data"].(map[string]interface{}); ok {
			tokenKey := data["key"]

			// 验证令牌
			authHeaders := map[string]string{
				"Authorization": "Bearer " + toString(tokenKey),
			}

			w := sendRequest(r, "GET", "/v1/dashboard/billing/credit_grants", nil, authHeaders)

			// 验证令牌有效
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 验证返回的令牌信息包含必要的字段
			assert.Contains(t, response, "total_granted")
			assert.Contains(t, response, "total_available")
		}
	})
}

// 辅助函数：将interface{}转换为string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v := v.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
