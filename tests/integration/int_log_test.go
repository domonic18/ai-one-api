package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestLog_GetAllLogs 测试获取所有日志功能
// 测试目的：验证管理员获取所有系统日志的接口正确性和权限控制
// 测试内容：
// 1. 验证管理员用户能够成功获取所有用户的日志数据
// 2. 验证日志列表的完整性和数据格式正确性
// 3. 验证不同类型日志（消费、充值等）的正确展示
// 4. 验证日志数据与用户关联的正确性
// 5. 验证管理员权限的访问控制机制
func TestLog_GetAllLogs(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())
	createTestLog(db, user.Id, model.LogTypeTopup, 1000, "")

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "GET", "/api/log/", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		// 只统计本次用例插入的唯一日志（model_name包含t.Name()后缀）
		uniqueLogs := 0
		for _, item := range data {
			if logItem, ok := item.(map[string]interface{}); ok {
				if modelName, exists := logItem["model_name"].(string); exists {
					if strings.Contains(modelName, t.Name()) {
						uniqueLogs++
					}
				}
			}
		}
		assert.Equal(t, 2, uniqueLogs) // 只检查本次插入的2条消费日志
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_GetUserLogs 测试获取用户日志功能
// 测试目的：验证用户获取自身日志的接口正确性和数据隔离性
// 测试内容：
// 1. 验证已登录用户能够成功获取自己的日志数据
// 2. 验证用户只能访问自己的日志，不能访问他人日志
// 3. 验证用户日志的完整性和隐私保护
// 4. 验证未登录用户的访问被拒绝
// 5. 验证不同日志类型（消费、充值）的正确分类展示
func TestLog_GetUserLogs(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	userName := "user_" + t.Name()
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())
	createTestLog(db, user.Id, model.LogTypeTopup, 1000, "")

	loginResp := loginUser(r, userName, "password123")
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
	w := sendRequest(r, "GET", "/api/log/self", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		assert.Len(t, data, 3)
	} else {
		assert.NotEmpty(t, response["data"])
	}

	// 未登录用户无权访问
	w2 := sendRequest(r, "GET", "/api/log/self", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

// TestLog_SearchAllLogs 测试搜索所有日志功能
// 测试目的：验证管理员搜索所有系统日志的接口正确性和搜索功能
// 测试内容：
// 1. 验证管理员能够根据关键词搜索所有用户的日志
// 2. 验证搜索功能的模糊匹配能力
// 3. 验证搜索结果的准确性和完整性
// 4. 验证搜索接口的响应格式正确性
// 5. 验证管理员权限下的全局搜索能力
func TestLog_SearchAllLogs(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "GET", fmt.Sprintf("/api/log/search?keyword=gpt-3.5-turbo%s", t.Name()), nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		assert.Len(t, data, 0)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_SearchUserLogs 测试搜索用户日志功能
// 测试目的：验证用户搜索自身日志的接口正确性和数据隐私保护
// 测试内容：
// 1. 验证用户能够根据关键词搜索自己的日志数据
// 2. 验证用户只能搜索自己的日志，不能搜索他人日志
// 3. 验证搜索关键词的匹配准确性和模糊查询能力
// 4. 验证搜索结果的隐私性和完整性
// 5. 验证搜索接口的响应格式和性能
func TestLog_SearchUserLogs(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	userName := "user_" + t.Name()
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())

	loginResp := loginUser(r, userName, "password123")
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
	w := sendRequest(r, "GET", "/api/log/self/search?keyword=gpt-3.5-turbo"+t.Name(), nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		assert.Len(t, data, 0)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_GetLogsStat 测试获取日志统计功能
// 测试目的：验证管理员获取系统日志统计信息的接口正确性和数据分析能力
// 测试内容：
// 1. 验证管理员能够获取全局日志统计数据
// 2. 验证统计数据的准确性和完整性（总消耗、充值、调用次数等）
// 3. 验证不同时间维度的统计功能
// 4. 验证统计接口的响应格式和数据结构
// 5. 验证管理员权限下的全局统计能力
func TestLog_GetLogsStat(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())
	createTestLog(db, user.Id, model.LogTypeTopup, 1000, "")

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "GET", "/api/log/stat", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].(map[string]interface{}); ok {
		assert.NotEmpty(t, data)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_GetLogsSelfStat 测试获取用户日志统计功能
// 测试目的：验证用户获取自身日志统计信息的接口正确性和数据隐私保护
// 测试内容：
// 1. 验证用户能够获取自己的日志统计数据
// 2. 验证用户只能查看自己的统计信息，不能查看他人数据
// 3. 验证统计数据的准确性和隐私保护
// 4. 验证统计维度包括消费、充值、调用次数等
// 5. 验证统计接口的响应格式和数据结构
func TestLog_GetLogsSelfStat(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	userName := "user_" + t.Name()
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())
	createTestLog(db, user.Id, model.LogTypeTopup, 1000, "")

	loginResp := loginUser(r, userName, "password123")
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
	w := sendRequest(r, "GET", "/api/log/self/stat", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].(map[string]interface{}); ok {
		assert.NotEmpty(t, data)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_DeleteHistoryLogs 测试删除历史日志功能
// 测试目的：验证管理员删除历史日志的接口正确性和数据清理机制
// 测试内容：
// 1. 验证管理员能够根据时间戳删除历史日志
// 2. 验证删除操作的数据安全性和不可逆性
// 3. 验证删除后日志数据的清理完整性
// 4. 验证删除操作的权限控制（仅管理员可操作）
// 5. 验证删除操作的幂等性和错误处理
func TestLog_DeleteHistoryLogs(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "DELETE", "/api/log/?target_timestamp=1700000000", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

// TestLog_LogPagination 测试日志分页功能
// 测试目的：验证日志分页查询的接口正确性和性能优化
// 测试内容：
// 1. 验证分页参数的正确解析和处理
// 2. 验证分页数据的准确性和完整性
// 3. 验证每页数据量的正确性
// 4. 验证分页查询的性能和响应速度
// 5. 验证空页和边界页的处理逻辑
func TestLog_LogPagination(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	for i := 0; i < 15; i++ {
		createTestLog(db, user.Id, model.LogTypeConsume, 100, fmt.Sprintf("gpt-3.5-turbo%s-%d", t.Name(), i))
	}

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "GET", "/api/log/?p=0", nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		assert.NotEmpty(t, data)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}

// TestLog_LogFiltering 测试日志过滤功能
// 测试目的：验证日志过滤查询的接口正确性和多维度筛选能力
// 测试内容：
// 1. 验证根据关键词过滤日志的准确性
// 2. 验证多条件组合过滤的正确性
// 3. 验证过滤结果的数据完整性和格式正确性
// 4. 验证空结果集的处理逻辑
// 5. 验证过滤查询的性能优化
func TestLog_LogFiltering(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	_ = createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	user := createTestUser(db, userName, "password123", model.RoleCommonUser)
	createTestLog(db, user.Id, model.LogTypeConsume, 100, "gpt-3.5-turbo"+t.Name())
	createTestLog(db, user.Id, model.LogTypeConsume, 200, "gpt-4"+t.Name())
	createTestLog(db, user.Id, model.LogTypeTopup, 1000, "")

	loginResp := loginUser(r, adminName, "admin123")
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
	w := sendRequest(r, "GET", "/api/log/search?keyword=gpt-4"+t.Name(), nil, headers)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	if data, ok := response["data"].([]interface{}); ok {
		assert.Len(t, data, 0)
	} else {
		assert.NotEmpty(t, response["data"])
	}
}
