package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware/identity"
	"github.com/stretchr/testify/assert"
)

func TestSmartModelSelection_Integration(t *testing.T) {
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
		requestBody           map[string]interface{}
		expectedModelReplaced bool
		expectedModel         string
		expectedStatusCode    int
		description           string
	}{
		{
			name:                "启用智能模型选择_正常替换",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: true,
			expectedModel:         "gpt-4", // 假设teacher_001偏好gpt-4
			expectedStatusCode:    200,
			description:           "启用智能模型选择时应该根据用户偏好替换模型",
		},
		{
			name:                "禁用智能模型选择_不替换",
			userID:              "teacher_001",
			smartModelSelection: "false",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			expectedStatusCode:    200,
			description:           "禁用智能模型选择时不应该替换模型",
		},
		{
			name:                "未设置智能模型选择_不替换",
			userID:              "teacher_001",
			smartModelSelection: "",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			expectedStatusCode:    200,
			description:           "未设置智能模型选择时不应该替换模型",
		},
		{
			name:                "大小写混合_true_启用",
			userID:              "teacher_001",
			smartModelSelection: "TrUe",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: true,
			expectedModel:         "gpt-4",
			expectedStatusCode:    200,
			description:           "大小写混合的'true'应该启用智能模型选择",
		},
		{
			name:                "带空格_true_启用",
			userID:              "teacher_001",
			smartModelSelection: " true ",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: true,
			expectedModel:         "gpt-4",
			expectedStatusCode:    200,
			description:           "带空格的'true'应该启用智能模型选择",
		},
		{
			name:                "未知用户_不替换",
			userID:              "unknown_user",
			smartModelSelection: "true",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			expectedStatusCode:    200,
			description:           "未知用户不应该替换模型",
		},
		{
			name:                "无效用户ID_不替换",
			userID:              "",
			smartModelSelection: "true",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			expectedModelReplaced: false,
			expectedModel:         "gpt-3.5-turbo",
			expectedStatusCode:    200,
			description:           "无效用户ID不应该替换模型",
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
				if actualModel == "" {
					// 如果没有从中间件获取到模型，从请求体中获取
					var reqBody map[string]interface{}
					if err := c.ShouldBindJSON(&reqBody); err == nil {
						if m, ok := reqBody["model"].(string); ok {
							actualModel = m
						}
					}
				}

				c.JSON(200, gin.H{
					"model": actualModel,
					"usage": gin.H{
						"prompt_tokens":     10,
						"completion_tokens": 20,
						"total_tokens":      30,
					},
				})
			})

			// 创建请求
			jsonBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// 设置请求头
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.smartModelSelection != "" {
				req.Header.Set("X-Smart-Model-Selection", tt.smartModelSelection)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatusCode, w.Code, "状态码应该匹配")

			// 解析响应
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 验证模型
			if model, ok := response["model"].(string); ok {
				if tt.expectedModelReplaced {
					assert.NotEqual(t, tt.requestBody["model"], model, "模型应该被替换")
				} else {
					assert.Equal(t, tt.requestBody["model"], model, "模型不应该被替换")
				}

				if tt.expectedModel != "" {
					assert.Equal(t, tt.expectedModel, model, tt.description)
				}
			}

			t.Logf("测试用例: %s, 实际模型: %s", tt.name, actualModel)
		})
	}
}

func TestSmartModelSelection_ErrorHandling(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		userID              string
		smartModelSelection string
		requestBody         string
		expectedStatusCode  int
		description         string
	}{
		{
			name:                "无效JSON请求体",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody:         "invalid json",
			expectedStatusCode:  400,
			description:         "无效JSON请求体应该返回400错误",
		},
		{
			name:                "缺少模型字段",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody:         `{"messages": [{"role": "user", "content": "Hello"}]}`,
			expectedStatusCode:  200,
			description:         "缺少模型字段时应该正常处理",
		},
		{
			name:                "空请求体",
			userID:              "teacher_001",
			smartModelSelection: "true",
			requestBody:         "",
			expectedStatusCode:  400,
			description:         "空请求体应该返回400错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(identity.Identity())

			// 模拟API端点
			router.POST("/v1/chat/completions", func(c *gin.Context) {
				var reqBody map[string]interface{}
				if err := c.ShouldBindJSON(&reqBody); err != nil {
					c.JSON(400, gin.H{"error": "Invalid JSON"})
					return
				}

				model := c.GetString(ctxkey.RequestModel)
				if model == "" {
					if m, ok := reqBody["model"].(string); ok {
						model = m
					}
				}

				c.JSON(200, gin.H{
					"model": model,
					"usage": gin.H{
						"prompt_tokens":     10,
						"completion_tokens": 20,
						"total_tokens":      30,
					},
				})
			})

			// 创建请求
			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 设置请求头
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.smartModelSelection != "" {
				req.Header.Set("X-Smart-Model-Selection", tt.smartModelSelection)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatusCode, w.Code, tt.description)

			t.Logf("测试用例: %s, 状态码: %d", tt.name, w.Code)
		})
	}
}

func TestSmartModelSelection_Performance(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()
	router.Use(identity.Identity())

	// 模拟API端点
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		model := c.GetString(ctxkey.RequestModel)
		if model == "" {
			var reqBody map[string]interface{}
			if err := c.ShouldBindJSON(&reqBody); err == nil {
				if m, ok := reqBody["model"].(string); ok {
					model = m
				}
			}
		}

		c.JSON(200, gin.H{
			"model": model,
			"usage": gin.H{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		})
	})

	// 准备测试数据
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	// 性能测试：并发请求
	t.Run("并发请求测试", func(t *testing.T) {
		const numRequests = 100
		results := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-User-ID", "teacher_001")
				req.Header.Set("X-Smart-Model-Selection", "true")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				results <- w.Code == 200
			}()
		}

		// 收集结果
		successCount := 0
		for i := 0; i < numRequests; i++ {
			if <-results {
				successCount++
			}
		}

		// 验证所有请求都成功
		assert.Equal(t, numRequests, successCount, "所有并发请求都应该成功")
		t.Logf("并发请求测试: %d/%d 成功", successCount, numRequests)
	})
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
