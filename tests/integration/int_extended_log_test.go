package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ExtendedLog_APIEndpoints 测试扩展日志API端点的集成
func TestIntegration_ExtendedLog_APIEndpoints(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()

	// 创建测试用户和管理员
	adminName := "admin_" + t.Name()
	userName := "user_" + t.Name()
	admin := createTestUser(db, adminName, "admin123", model.RoleAdminUser)
	testUser := createTestUser(db, userName, "password123", model.RoleCommonUser)

	// 获取管理员session
	adminLoginResp := loginUser(router, adminName, "admin123")
	adminCookies := adminLoginResp.Result().Cookies()
	var adminSessionCookie *http.Cookie
	for _, cookie := range adminCookies {
		if cookie.Name == "one-api" {
			adminSessionCookie = cookie
			break
		}
	}
	adminHeaders := map[string]string{}
	if adminSessionCookie != nil {
		adminHeaders["Cookie"] = "one-api=" + adminSessionCookie.Value
	}

	// 获取普通用户session
	userLoginResp := loginUser(router, userName, "password123")
	userCookies := userLoginResp.Result().Cookies()
	var userSessionCookie *http.Cookie
	for _, cookie := range userCookies {
		if cookie.Name == "one-api" {
			userSessionCookie = cookie
			break
		}
	}
	userHeaders := map[string]string{}
	if userSessionCookie != nil {
		userHeaders["Cookie"] = "one-api=" + userSessionCookie.Value
	}

	// 创建测试数据
	testLogs := createTestExtendedLogs()

	tests := []struct {
		name           string
		method         string
		url            string
		headers        map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "管理员获取所有扩展日志",
			method:         "GET",
			url:            "/api/log/extended/",
			headers:        adminHeaders,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
			},
		},
		{
			name:           "普通用户获取自己的扩展日志",
			method:         "GET",
			url:            "/api/log/extended/self",
			headers:        userHeaders,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name:           "管理员获取扩展日志详情",
			method:         "GET",
			url:            fmt.Sprintf("/api/log/extended/%d", testLogs[0].LogId),
			headers:        adminHeaders,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
			},
		},
		{
			name:           "管理员获取扩展日志统计",
			method:         "GET",
			url:            "/api/log/extended/statistics",
			headers:        adminHeaders,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
			},
		},
		{
			name:           "普通用户无权访问所有扩展日志",
			method:         "GET",
			url:            "/api/log/extended/",
			headers:        userHeaders,
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := sendRequest(router, tt.method, tt.url, nil, tt.headers)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理测试数据
	_ = admin
	_ = testUser
}

// TestIntegration_ExtendedLog_RecordingFlow 测试扩展日志记录流程的集成
func TestIntegration_ExtendedLog_RecordingFlow(t *testing.T) {
	// 设置集成测试环境
	_, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()
	cleanupLogTestData()

	ctx := context.Background()

	// 设置模拟的身份解析器
	setupMockIdentityResolver()

	// 创建原始日志
	originalLog := &model.Log{
		UserId:           100,
		Type:             model.LogTypeConsume,
		Content:          "测试消费日志",
		Username:         "testuser",
		ModelName:        "gpt-3.5-turbo",
		TokenName:        "test-token",
		Quota:            1000,
		PromptTokens:     100,
		CompletionTokens: 200,
		ChannelId:        1,
		CreatedAt:        time.Now().Unix(),
	}

	// 记录原始日志
	logId := model.RecordConsumeLog(ctx, originalLog)
	require.Greater(t, logId, int64(0))

	// 手动记录扩展日志（模拟集成场景）
	externalUserId := "teacher_integration_test"
	identity.RecordExtendedLog(ctx, logId, externalUserId)

	// 等待异步操作完成
	time.Sleep(200 * time.Millisecond)

	// 验证扩展日志是否创建
	extendedLog, err := identity.GetExtendedLogByLogId(ctx, logId)
	assert.NoError(t, err)
	assert.NotNil(t, extendedLog)
	assert.Equal(t, logId, extendedLog.LogId)
	assert.Equal(t, externalUserId, extendedLog.ExternalUserId)

	// 验证可以通过原始日志ID查询到扩展日志
	retrievedLog, err := model.GetLogById(logId)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedLog)
	assert.Equal(t, originalLog.UserId, retrievedLog.UserId)
	assert.Equal(t, originalLog.ModelName, retrievedLog.ModelName)
}

// TestIntegration_ExtendedLog_FilteringAndPagination 测试扩展日志筛选和分页的集成
func TestIntegration_ExtendedLog_FilteringAndPagination(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()

	// 创建管理员
	adminName := "admin_" + t.Name()
	admin := createTestUser(db, adminName, "admin123", model.RoleAdminUser)

	// 获取管理员session
	adminLoginResp := loginUser(router, adminName, "admin123")
	adminCookies := adminLoginResp.Result().Cookies()
	var adminSessionCookie *http.Cookie
	for _, cookie := range adminCookies {
		if cookie.Name == "one-api" {
			adminSessionCookie = cookie
			break
		}
	}
	adminHeaders := map[string]string{}
	if adminSessionCookie != nil {
		adminHeaders["Cookie"] = "one-api=" + adminSessionCookie.Value
	}

	// 创建测试数据
	createTestExtendedLogs()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "分页获取扩展日志",
			url:            "/api/log/extended/?page=1&page_size=5",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))

				data := response["data"].(map[string]interface{})
				logs := data["logs"].([]interface{})
				assert.LessOrEqual(t, len(logs), 5)
			},
		},
		{
			name:           "按外部用户ID筛选",
			url:            "/api/log/extended/?external_user_id=teacher_001",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name:           "按用户组筛选",
			url:            "/api/log/extended/?user_group=test_group",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name: "按时间范围筛选",
			url: fmt.Sprintf("/api/log/extended/?start_timestamp=%d&end_timestamp=%d",
				time.Now().Add(-24*time.Hour).Unix(), time.Now().Unix()),
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := sendRequest(router, "GET", tt.url, nil, adminHeaders)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理测试数据
	_ = admin
}

// TestIntegration_ExtendedLog_ErrorHandling 测试扩展日志错误处理的集成
func TestIntegration_ExtendedLog_ErrorHandling(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 创建管理员
	adminName := "admin_" + t.Name()
	admin := createTestUser(db, adminName, "admin123", model.RoleAdminUser)

	// 获取管理员session
	adminLoginResp := loginUser(router, adminName, "admin123")
	adminCookies := adminLoginResp.Result().Cookies()
	var adminSessionCookie *http.Cookie
	for _, cookie := range adminCookies {
		if cookie.Name == "one-api" {
			adminSessionCookie = cookie
			break
		}
	}
	adminHeaders := map[string]string{}
	if adminSessionCookie != nil {
		adminHeaders["Cookie"] = "one-api=" + adminSessionCookie.Value
	}

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "获取不存在的扩展日志详情",
			method:         "GET",
			url:            "/api/log/extended/999999",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:           "无效的日志ID格式",
			method:         "GET",
			url:            "/api/log/extended/invalid_id",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:           "无效的分页参数",
			method:         "GET",
			url:            "/api/log/extended/?page=-1&page_size=0",
			expectedStatus: http.StatusOK, // 服务器会使用默认值
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := sendRequest(router, tt.method, tt.url, nil, adminHeaders)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理测试数据
	_ = admin
}

// 辅助函数：清理扩展日志测试数据
func cleanupExtendedLogTestData() {
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})
}

// 辅助函数：清理日志测试数据
func cleanupLogTestData() {
	model.DB.Where("1 = 1").Delete(&model.Log{})
}

// 辅助函数：创建测试扩展日志数据
func createTestExtendedLogs() []*identity.ExtendedLog {
	ctx := context.Background()
	var logs []*identity.ExtendedLog

	// 创建多个测试日志
	for i := 0; i < 10; i++ {
		logId := int64(10000 + i)
		externalUserId := fmt.Sprintf("user_%03d", i+1)
		userGroup := "test_group"

		// 使用抽象化的维度信息，不耦合具体业务逻辑
		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": externalUserId,
			"group_id":         i%3 + 1,
			"group_name":       fmt.Sprintf("测试组%d", i%3+1),
			"category_id":      i%5 + 10,
			"category_name":    fmt.Sprintf("测试类别%d", i%5+1),
			"user_name":        fmt.Sprintf("测试用户%d", i+1),
		}

		log, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		if err == nil {
			logs = append(logs, log)
		}
	}

	return logs
}

// 辅助函数：设置模拟的身份解析器
func setupMockIdentityResolver() {
	// 注意：在实际的集成测试中，身份解析器需要通过配置或初始化来设置
	// 这里我们跳过设置，因为RecordExtendedLog函数会处理解析器未初始化的情况
}
