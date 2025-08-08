package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/songquanpeng/one-api/middleware/identity"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// mock resolver: 返回固定组与模型
type relayMockResolver struct{}

func (m *relayMockResolver) ResolveGroup(_ context.Context, _ string) string {
	return "beijing_math_group"
}
func (m *relayMockResolver) ResolveModel(_ context.Context, _ string, requestModel string) string {
	// 模拟智能选择：无论请求模型为何，一律替换为 deepseek-chat
	return "deepseek-chat"
}
func (m *relayMockResolver) GetUserDetails(_ context.Context, externalIdentity string) map[string]interface{} {
	return map[string]interface{}{"external_user_id": externalIdentity}
}

// TestRelay_ModelReplaced_OnMainPath 通过主路径 /v1/chat/completions 验证模型替换与转发
func TestRelay_ModelReplaced_OnMainPath(t *testing.T) {
	router, db := setupIntegrationTest()
	cleanupTestData(db)

	// 安装测试身份解析器
	identity.SetIdentityResolver(&relayMockResolver{})

	// 创建测试用户与令牌
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "relay-model-replace")
	_ = token.SetUserGroups([]string{"beijing_math_group", "default"})
	_ = token.Update()

	// 启动下游Mock服务，校验转发时的模型字段是否已替换为 deepseek-chat
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		modelVal, _ := body["model"].(string)
		if modelVal != "deepseek-chat" {
			http.Error(w, `{"error":{"message":"Model Not Exist"}}`, http.StatusBadRequest)
			return
		}
		// 返回OpenAI兼容的最小成功响应
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
            "id":"cmpl-xyz",
            "object":"chat.completion",
            "choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],
            "usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
        }`))
	}))
	defer downstream.Close()

	// 在 beijing_math_group 组创建可用渠道（支持 deepseek-chat），并将BaseURL指向下游Mock
	baseURL := downstream.URL
	ch := &model.Channel{Type: 1, Key: "sk-relay", Name: "relay-deepseek", Status: model.ChannelStatusEnabled, Group: "beijing_math_group", Models: "deepseek-chat"}
	ch.BaseURL = &baseURL
	err := model.DB.Create(ch).Error
	assert.NoError(t, err)
	_ = ch.AddAbilities()

	// 组织请求：请求体中使用 qwen-max，但应被替换为 deepseek-chat
	reqBody := map[string]interface{}{
		"model":    "qwen-max",
		"messages": []map[string]interface{}{{"role": "user", "content": "hello"}},
	}
	payload, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Key)
	req.Header.Set("X-User-ID", "teacher_abc")
	req.Header.Set("X-Smart-Model-Selection", "true")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
