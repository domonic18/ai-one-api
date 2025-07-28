package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmartModelSelectionAPI 测试智能模型选择API集成功能
// 测试目的：验证智能模型选择系统在完整API调用链路中的正确集成和行为表现
// 测试内容：
// 1. 验证智能模型选择中间件在真实API调用中的启用和禁用机制
// 2. 验证用户身份识别与模型选择的关联性
// 3. 验证扩展日志记录与API调用的集成效果
// 4. 验证渠道可用性和模型支持的完整性
// 5. 验证响应格式的正确性和API兼容性
func TestSmartModelSelectionAPI(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	// 创建测试渠道，支持gpt-3.5-turbo模型
	channel := createTestChannel(db, "Test OpenAI", "test-key")

	// 创建能力记录，使渠道支持gpt-3.5-turbo模型
	priority := int64(1)
	ability := &model.Ability{
		Group:     "default",
		Model:     "gpt-3.5-turbo",
		ChannelId: channel.Id,
		Enabled:   true,
		Priority:  &priority,
	}
	db.Create(ability)

	// 调试信息：打印令牌key
	t.Logf("Created token with key: %s", token.Key)
	t.Logf("Created channel with id: %d", channel.Id)

	t.Run("智能模型选择中间件集成测试", func(t *testing.T) {
		// 测试聊天完成接口
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		// 不启用智能模型选择
		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		// 由于没有实际的OpenAI API，我们期望某种错误，但不是503无可用渠道
		// 可能是400或500等其他错误
		assert.NotEqual(t, http.StatusServiceUnavailable, w.Code)

		// 启用智能模型选择
		headers["X-Smart-Model-Selection"] = "true"
		w = sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.NotEqual(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("身份识别中间件集成测试", func(t *testing.T) {
		// 测试聊天完成接口
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要字段
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "object")
		assert.Contains(t, response, "created")
		assert.Contains(t, response, "model")
		assert.Contains(t, response, "choices")
	})

	t.Run("扩展日志记录中间件集成测试", func(t *testing.T) {
		// 测试聊天完成接口
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要字段
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "object")
		assert.Contains(t, response, "created")
		assert.Contains(t, response, "model")
		assert.Contains(t, response, "choices")
	})
}

func TestIdentityAuthMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("无用户ID请求测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello without user ID",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("有用户ID请求测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello with user ID",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSmartModelSelectionMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("启用智能模型选择测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test smart model selection",
				},
			},
		}

		headers := map[string]string{
			"Authorization":           "Bearer " + token.Key,
			"Content-Type":            "application/json",
			"X-User-ID":               "teacher_001",
			"X-Smart-Model-Selection": "true",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("禁用智能模型选择测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test without smart model selection",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("启用但无用户ID测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test smart selection without user ID",
				},
			},
		}

		headers := map[string]string{
			"Authorization":           "Bearer " + token.Key,
			"Content-Type":            "application/json",
			"X-Smart-Model-Selection": "true",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExtendedLogRecorderMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("扩展日志记录测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test extended log recording",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_001",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要字段
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "object")
		assert.Contains(t, response, "created")
		assert.Contains(t, response, "model")
		assert.Contains(t, response, "choices")
	})
}

func TestMiddlewareChainIntegration(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("完整中间件链测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test complete middleware chain",
				},
			},
		}

		headers := map[string]string{
			"Authorization":           "Bearer " + token.Key,
			"Content-Type":            "application/json",
			"X-User-ID":               "teacher_001",
			"X-Smart-Model-Selection": "true",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要字段
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "object")
		assert.Contains(t, response, "created")
		assert.Contains(t, response, "model")
		assert.Contains(t, response, "choices")
	})
}

// 辅助函数：创建测试扩展日志信息
func createTestExtendedLogInfo() *model.ExtendedLogInfo {
	return &model.ExtendedLogInfo{
		SchoolId:    1,
		SchoolName:  "测试学校",
		SubjectId:   101,
		SubjectName: "数学组",
		TeacherId:   "teacher_test",
		TeacherName: "测试老师",
		GroupName:   "数学组",
	}
}
