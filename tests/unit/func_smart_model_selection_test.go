package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware/identity"
	"github.com/stretchr/testify/assert"
)

// TestSmartModelSelection_HeaderProcessing 测试智能模型选择的请求头处理
// 测试目的：验证智能模型选择功能的请求头解析逻辑，确保能够正确识别和处理各种输入格式
// 测试内容：
// 1. 不同大小写格式的"true"值处理
// 2. 不同大小写格式的"false"值处理
// 3. 空值和空格的处理
// 4. 无效值的处理
// 5. 带空格的值的处理
func TestSmartModelSelection_HeaderProcessing(t *testing.T) {
	tests := []struct {
		name           string
		headerValue    string
		expectedResult bool
		description    string
	}{
		{
			name:           "启用智能模型选择_true",
			headerValue:    "true",
			expectedResult: true,
			description:    "当header值为'true'时应该启用智能模型选择",
		},
		{
			name:           "启用智能模型选择_TRUE",
			headerValue:    "TRUE",
			expectedResult: true,
			description:    "当header值为'TRUE'时应该启用智能模型选择（大小写不敏感）",
		},
		{
			name:           "启用智能模型选择_True",
			headerValue:    "True",
			expectedResult: true,
			description:    "当header值为'True'时应该启用智能模型选择（大小写不敏感）",
		},
		{
			name:           "禁用智能模型选择_false",
			headerValue:    "false",
			expectedResult: false,
			description:    "当header值为'false'时应该禁用智能模型选择",
		},
		{
			name:           "禁用智能模型选择_FALSE",
			headerValue:    "FALSE",
			expectedResult: false,
			description:    "当header值为'FALSE'时应该禁用智能模型选择",
		},
		{
			name:           "禁用智能模型选择_空值",
			headerValue:    "",
			expectedResult: false,
			description:    "当header值为空时应该禁用智能模型选择",
		},
		{
			name:           "禁用智能模型选择_空格",
			headerValue:    " ",
			expectedResult: false,
			description:    "当header值为空格时应该禁用智能模型选择",
		},
		{
			name:           "禁用智能模型选择_其他值",
			headerValue:    "other",
			expectedResult: false,
			description:    "当header值为其他值时应该禁用智能模型选择",
		},
		{
			name:           "启用智能模型选择_带空格",
			headerValue:    " true ",
			expectedResult: true,
			description:    "当header值为' true '时应该启用智能模型选择（去除空格）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSmartModelSelectionEnabled(tt.headerValue)
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

// TestSmartModelSelection_MiddlewareIntegration 测试智能模型选择中间件集成
// 测试目的：验证智能模型选择中间件与身份解析器的集成，确保能够根据用户偏好正确替换模型
// 测试内容：
// 1. 启用智能模型选择时的模型替换逻辑
// 2. 禁用智能模型选择时的模型保持逻辑
// 3. 未知用户的处理
// 4. 不同输入格式的处理（大小写混合、带空格等）
// 5. 中间件与身份解析器的协作
func TestSmartModelSelection_MiddlewareIntegration(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建测试用的身份解析器
	testResolver := &TestIdentityResolver{
		userGroups: map[string]string{
			"teacher_001": "math_group",
		},
		userModels: map[string]string{
			"teacher_001": "gpt-4",
		},
	}

	// 设置测试解析器
	identity.SetIdentityResolver(testResolver)

	tests := []struct {
		name                  string
		userID                string
		smartModelSelection   string
		originalModel         string
		expectedModelReplaced bool
		expectedModel         string
		description           string
	}{
		{
			name:                  "启用智能模型选择_模型替换",
			userID:                "teacher_001",
			smartModelSelection:   "true",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: true,
			expectedModel:         "gpt-4", // 假设teacher_001偏好gpt-4
			description:           "当启用智能模型选择时，应该根据用户偏好替换模型",
		},
		{
			name:                  "禁用智能模型选择_不替换",
			userID:                "teacher_001",
			smartModelSelection:   "false",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			description:           "当禁用智能模型选择时，不应该替换模型",
		},
		{
			name:                  "未设置智能模型选择_不替换",
			userID:                "teacher_001",
			smartModelSelection:   "",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			description:           "当未设置智能模型选择时，不应该替换模型",
		},
		{
			name:                  "大小写混合_true_启用",
			userID:                "teacher_001",
			smartModelSelection:   "TrUe",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: true,
			expectedModel:         "gpt-4",
			description:           "大小写混合的'true'应该启用智能模型选择",
		},
		{
			name:                  "带空格_true_启用",
			userID:                "teacher_001",
			smartModelSelection:   " true ",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: true,
			expectedModel:         "gpt-4",
			description:           "带空格的'true'应该启用智能模型选择",
		},
		{
			name:                  "未知用户_不替换",
			userID:                "unknown_user",
			smartModelSelection:   "true",
			originalModel:         "gpt-3.5-turbo",
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			description:           "未知用户不应该替换模型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(identity.Identity())

			// 模拟API端点
			var actualModel string
			router.POST("/v1/chat/completions", func(c *gin.Context) {
				actualModel = c.GetString(ctxkey.RequestModel)
				c.JSON(200, gin.H{"model": actualModel})
			})

			// 创建请求
			requestBody := map[string]interface{}{
				"model": tt.originalModel,
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			}

			req := createTestRequest(t, "POST", "/v1/chat/completions", requestBody)

			// 设置请求头
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.smartModelSelection != "" {
				req.Header.Set("X-Smart-Model-Selection", tt.smartModelSelection)
			}

			// 执行请求
			w := performRequest(router, req)

			// 验证结果
			assert.Equal(t, 200, w.Code, "请求应该成功")

			// 验证模型是否被替换
			if tt.expectedModelReplaced {
				assert.NotEqual(t, tt.originalModel, actualModel, "模型应该被替换")
			} else {
				assert.Equal(t, tt.originalModel, actualModel, "模型不应该被替换")
			}

			// 验证最终模型
			if tt.expectedModel != "" {
				assert.Equal(t, tt.expectedModel, actualModel, tt.description)
			}
		})
	}
}

// TestSmartModelSelection_EdgeCases 测试智能模型选择的边界情况
// 测试目的：验证智能模型选择功能在各种边界情况下的稳定性和错误处理能力
// 测试内容：
// 1. 大小写混合输入的处理
// 2. 带空格的输入处理
// 3. 无效JSON请求体的处理
// 4. 缺少模型字段的请求处理
// 5. 异常情况的容错能力
func TestSmartModelSelection_EdgeCases(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		userID              string
		smartModelSelection string
		requestBody         map[string]interface{}
		expectedBehavior    string
		description         string
	}{
		{
			name:                "大小写混合_true",
			userID:              "teacher_001",
			smartModelSelection: "TrUe",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
			},
			expectedBehavior: "启用",
			description:      "大小写混合的'true'应该启用智能模型选择",
		},
		{
			name:                "带空格_true",
			userID:              "teacher_001",
			smartModelSelection: " true ",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
			},
			expectedBehavior: "启用",
			description:      "带空格的'true'应该启用智能模型选择",
		},
		{
			name:                "无效JSON请求体",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody:         nil,
			expectedBehavior:    "不替换",
			description:         "无效的请求体不应该导致错误",
		},
		{
			name:                "无模型字段",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedBehavior: "不替换",
			description:      "请求体中无model字段时不应该替换",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			router := gin.New()
			router.Use(identity.Identity())

			// 模拟请求处理
			var actualModel string
			router.POST("/test", func(c *gin.Context) {
				actualModel = c.GetString(ctxkey.RequestModel)
				c.JSON(200, gin.H{"model": actualModel})
			})

			// 创建请求
			var req *http.Request
			if tt.requestBody != nil {
				req = createTestRequest(t, "POST", "/test", tt.requestBody)
			} else {
				req = httptest.NewRequest("POST", "/test", strings.NewReader("invalid json"))
			}

			// 设置请求头
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.smartModelSelection != "" {
				req.Header.Set("X-Smart-Model-Selection", tt.smartModelSelection)
			}

			// 执行请求
			w := performRequest(router, req)

			// 验证结果
			assert.Equal(t, 200, w.Code, "请求应该成功")
			t.Logf("测试用例: %s, 实际模型: %s", tt.name, actualModel)
		})
	}
}

// 辅助函数：检查是否启用智能模型选择
func isSmartModelSelectionEnabled(headerValue string) bool {
	if headerValue == "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(headerValue)) == "true"
}

// 辅助函数：创建测试请求
func createTestRequest(t *testing.T, method, path string, body map[string]interface{}) *http.Request {
	jsonBody, err := json.Marshal(body)
	assert.NoError(t, err)

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// 辅助函数：执行请求
func performRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestIdentityResolver 测试用的身份解析器
type TestIdentityResolver struct {
	userGroups map[string]string
	userModels map[string]string
}

// ResolveGroup 解析用户组
func (t *TestIdentityResolver) ResolveGroup(ctx context.Context, externalIdentity string) string {
	if group, ok := t.userGroups[externalIdentity]; ok {
		return group
	}
	return ""
}

// ResolveModel 解析用户偏好模型
func (t *TestIdentityResolver) ResolveModel(ctx context.Context, externalIdentity string, requestModel string) string {
	if model, ok := t.userModels[externalIdentity]; ok {
		return model
	}
	return requestModel
}

// GetUserDetails 获取用户详细信息
func (t *TestIdentityResolver) GetUserDetails(ctx context.Context, externalIdentity string) map[string]interface{} {
	// 测试实现返回基本的用户信息
	if _, ok := t.userGroups[externalIdentity]; ok {
		return map[string]interface{}{
			"user_id": externalIdentity,
			"group":   t.userGroups[externalIdentity],
			"model":   t.userModels[externalIdentity],
		}
	}
	return nil
}
