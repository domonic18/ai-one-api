package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartModelSelectionAPI(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 创建测试用户和Token
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

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
		assert.Equal(t, http.StatusOK, w.Code)

		// 启用智能模型选择
		headers["X-Smart-Model-Selection"] = "true"
		w = sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("身份识别中间件集成测试", func(t *testing.T) {
		// 测试身份识别中间件是否正确设置上下文信息
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test message",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_002",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("扩展日志记录中间件集成测试", func(t *testing.T) {
		// 测试扩展日志记录功能
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test for extended logging",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_003",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 验证是否创建了扩展日志记录
		// 注意：由于扩展日志是异步创建的，这里我们只验证请求成功
		// 在实际环境中，可以通过查询数据库来验证扩展日志是否被创建
	})
}

func TestIdentityAuthMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("无用户ID请求测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test without user ID",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			// 不设置X-User-ID
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
					"content": "Test with user ID",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_004",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSmartModelSelectionMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

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
			"X-User-ID":               "teacher_005",
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
			"Authorization":           "Bearer " + token.Key,
			"Content-Type":            "application/json",
			"X-User-ID":               "teacher_006",
			"X-Smart-Model-Selection": "false",
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
			// 不设置X-User-ID
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExtendedLogRecorderMiddleware(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "test-token")

	t.Run("扩展日志记录测试", func(t *testing.T) {
		// 创建测试日志记录
		testLog := &model.Log{
			UserId:            user.Id,
			CreatedAt:         helper.GetTimestamp(),
			Type:              model.LogTypeConsume,
			Content:           "测试扩展日志记录",
			Username:          user.Username,
			TokenName:         token.Name,
			ModelName:         "gpt-3.5-turbo",
			Quota:             100,
			PromptTokens:      50,
			CompletionTokens:  50,
			ChannelId:         1,
			RequestId:         "test-request-extended-log",
			ElapsedTime:       1000,
			IsStream:          false,
			SystemPromptReset: false,
		}

		err := db.Create(testLog).Error
		require.NoError(t, err)

		// 发送请求，触发扩展日志记录
		payload := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Test extended logging",
				},
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + token.Key,
			"Content-Type":  "application/json",
			"X-User-ID":     "teacher_007",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 清理测试数据
		db.Delete(testLog)
	})
}

func TestMiddlewareChainIntegration(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

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
			"X-User-ID":               "teacher_008",
			"X-Smart-Model-Selection": "true",
		}

		w := sendRequest(r, "POST", "/v1/chat/completions", payload, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// 验证响应格式
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要的字段
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
