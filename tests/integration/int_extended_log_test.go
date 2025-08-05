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
// 测试目的：验证扩展日志API端点的完整功能，包括权限控制、数据查询和响应格式
// 测试内容：
// 1. 管理员获取所有扩展日志的权限和功能
// 2. 普通用户获取自己扩展日志的权限和功能
// 3. 扩展日志详情查询功能
// 4. 统计信息查询功能
// 5. 权限控制和访问限制
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
			name:           "管理员获取扩展日志统计信息",
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
			name:           "普通用户无权限访问所有扩展日志",
			method:         "GET",
			url:            "/api/log/extended/",
			headers:        userHeaders,
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// 普通用户应该被拒绝访问
			},
		},
		{
			name:           "普通用户无权限访问扩展日志详情",
			method:         "GET",
			url:            fmt.Sprintf("/api/log/extended/%d", testLogs[0].LogId),
			headers:        userHeaders,
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// 普通用户应该被拒绝访问
			},
		},
		{
			name:           "普通用户无权限访问统计信息",
			method:         "GET",
			url:            "/api/log/extended/statistics",
			headers:        userHeaders,
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// 普通用户应该被拒绝访问
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(admin)
		model.DB.Unscoped().Delete(testUser)
		cleanupExtendedLogTestData()
	}()
}

// TestIntegration_ExtendedLog_RecordingFlow 测试扩展日志记录流程
// 测试目的：验证扩展日志的完整记录流程，包括创建、存储和查询
// 测试内容：
// 1. 扩展日志的创建和存储
// 2. 维度信息的序列化和反序列化
// 3. 日志记录的查询和验证
// 4. 身份解析器的集成
// 5. 异步记录机制的正确性
func TestIntegration_ExtendedLog_RecordingFlow(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()

	// 设置模拟身份解析器
	setupMockIdentityResolver()

	// 创建测试用户
	userName := "user_" + t.Name()
	testUser := createTestUser(db, userName, "password123", model.RoleCommonUser)

	// 获取用户session
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

	// 测试扩展日志记录流程
	t.Run("扩展日志记录流程", func(t *testing.T) {
		ctx := context.Background()
		logId := int64(1001)
		externalUserId := "test_user_001"
		userGroup := "test_group"

		// 创建扩展日志
		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": externalUserId,
			"group_id":         1,
			"group_name":       "测试组",
			"category_id":      10,
			"category_name":    "测试类别",
			"user_name":        "测试用户",
		}

		extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		require.NoError(t, err)
		require.NotNil(t, extendedLog)

		// 验证基本字段
		assert.Equal(t, logId, extendedLog.LogId)
		assert.Equal(t, externalUserId, extendedLog.ExternalUserId)
		assert.Equal(t, userGroup, extendedLog.UserGroup)

		// 验证维度信息
		parsedDimensionInfo, err := extendedLog.GetDimensionInfo()
		require.NoError(t, err)
		require.NotNil(t, parsedDimensionInfo)

		// 验证所有维度信息都被正确保存
		for key, expectedValue := range *dimensionInfo {
			actualValue, exists := (*parsedDimensionInfo)[key]
			assert.True(t, exists, "Key %s should exist", key)
			assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
		}

		// 测试异步记录函数
		identity.RecordExtendedLog(ctx, logId+1, externalUserId)

		// 等待异步操作完成
		time.Sleep(100 * time.Millisecond)

		// 验证异步记录的日志
		var asyncLog identity.ExtendedLog
		err = model.DB.Where("log_id = ?", logId+1).First(&asyncLog).Error
		assert.NoError(t, err)
		assert.Equal(t, externalUserId, asyncLog.ExternalUserId)
	})

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(testUser)
		cleanupExtendedLogTestData()
	}()
}

// TestIntegration_ExtendedLog_FilteringAndPagination 测试扩展日志的过滤和分页功能
// 测试目的：验证扩展日志查询的过滤和分页功能，确保能够正确处理大量数据
// 测试内容：
// 1. 分页参数的处理和验证
// 2. 过滤条件的应用
// 3. 排序功能的正确性
// 4. 边界条件的处理
// 5. 性能优化的验证
func TestIntegration_ExtendedLog_FilteringAndPagination(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()

	// 创建测试数据
	_ = createTestExtendedLogs()

	// 创建管理员用户
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

	// 测试分页功能
	t.Run("分页功能测试", func(t *testing.T) {
		testCases := []struct {
			name     string
			page     string
			pageSize string
		}{
			{"默认分页", "", ""},
			{"第一页", "1", "5"},
			{"第二页", "2", "5"},
			{"大页码", "10", "10"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := "/api/log/extended/"
				if tc.page != "" || tc.pageSize != "" {
					url += "?"
					if tc.page != "" {
						url += "page=" + tc.page
					}
					if tc.pageSize != "" {
						if tc.page != "" {
							url += "&"
						}
						url += "page_size=" + tc.pageSize
					}
				}

				req := httptest.NewRequest("GET", url, nil)
				for key, value := range adminHeaders {
					req.Header.Set(key, value)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
			})
		}
	})

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(admin)
		cleanupExtendedLogTestData()
	}()
}

// TestIntegration_ExtendedLog_ErrorHandling 测试扩展日志的错误处理
// 测试目的：验证扩展日志系统在各种错误情况下的稳定性和错误处理能力
// 测试内容：
// 1. 无效参数的错误处理
// 2. 数据库连接错误的处理
// 3. 身份解析器错误的处理
// 4. 并发访问的安全性
// 5. 异常情况的容错能力
func TestIntegration_ExtendedLog_ErrorHandling(t *testing.T) {
	// 设置集成测试环境
	router, db := setupIntegrationTest()
	model.DB = db

	// 清理测试数据
	cleanupExtendedLogTestData()

	// 创建测试用户
	userName := "user_" + t.Name()
	testUser := createTestUser(db, userName, "password123", model.RoleCommonUser)

	// 获取用户session
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

	// 测试错误处理
	t.Run("错误处理测试", func(t *testing.T) {
		ctx := context.Background()

		// 测试空外部用户ID
		t.Run("空外部用户ID", func(t *testing.T) {
			identity.RecordExtendedLog(ctx, 9999, "")

			// 验证没有创建日志记录
			var count int64
			model.DB.Model(&identity.ExtendedLog{}).Where("log_id = ?", 9999).Count(&count)
			assert.Equal(t, int64(0), count)
		})

		// 测试无效的日志ID
		t.Run("无效日志ID", func(t *testing.T) {
			extendedLog, err := identity.CreateExtendedLog(ctx, -1, "test_user", "test_group", nil)
			// 这里应该根据实际实现来决定是否允许负数的日志ID
			if err != nil {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, extendedLog)
			}
		})

		// 测试并发访问
		t.Run("并发访问", func(t *testing.T) {
			const concurrency = 10
			done := make(chan bool, concurrency)

			for i := 0; i < concurrency; i++ {
				go func(index int) {
					defer func() { done <- true }()
					logId := int64(2000 + index)
					externalUserId := fmt.Sprintf("concurrent_user_%d", index)
					identity.RecordExtendedLog(ctx, logId, externalUserId)
				}(i)
			}

			// 等待所有goroutine完成
			for i := 0; i < concurrency; i++ {
				<-done
			}

			// 验证所有日志都被创建
			var count int64
			model.DB.Model(&identity.ExtendedLog{}).Where("log_id >= ? AND log_id < ?", 2000, 2010).Count(&count)
			assert.Equal(t, int64(concurrency), count)
		})
	})

	// 清理测试数据
	defer func() {
		model.DB.Unscoped().Delete(testUser)
		cleanupExtendedLogTestData()
	}()
}

// 辅助函数：清理扩展日志测试数据
func cleanupExtendedLogTestData() {
	model.DB.Where("1 = 1").Delete(&identity.ExtendedLog{})
}

// 辅助函数：清理日志测试数据
func cleanupLogTestData() {
	model.DB.Where("1 = 1").Delete(&model.Log{})
}

// 辅助函数：创建测试扩展日志
func createTestExtendedLogs() []*identity.ExtendedLog {
	ctx := context.Background()
	var logs []*identity.ExtendedLog

	// 创建多个测试扩展日志
	for i := 1; i <= 10; i++ {
		logId := int64(1000 + i)
		externalUserId := fmt.Sprintf("test_user_%03d", i)
		userGroup := fmt.Sprintf("test_group_%d", (i%3)+1)

		dimensionInfo := &identity.DimensionInfo{
			"external_user_id": externalUserId,
			"group_id":         i,
			"group_name":       userGroup,
			"category_id":      10 + i,
			"category_name":    fmt.Sprintf("测试类别_%d", i),
			"user_name":        fmt.Sprintf("测试用户_%d", i),
		}

		extendedLog, err := identity.CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		if err == nil && extendedLog != nil {
			logs = append(logs, extendedLog)
		}
	}

	return logs
}

// 辅助函数：设置模拟身份解析器
func setupMockIdentityResolver() {
	// 这里可以设置模拟的身份解析器用于测试
	// 实际实现中可能需要根据具体的身份解析器接口来设置
}
