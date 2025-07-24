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
