package unit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/controller"
	identity "github.com/songquanpeng/one-api/middleware/identity"
	testmocks "github.com/songquanpeng/one-api/tests/mocks"
	"github.com/stretchr/testify/mock"
)

// sign 生成签名
func sign(secret string, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(body)))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func setupResolverForTest() func() {
	// 使用项目内置Mock Redis
	mockRedis := testmocks.NewMockRedisClient()
	// 允许任意 Set 调用
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cfg := &identity.CoursewareConfig{CacheTTL: time.Minute}
	cache := identity.NewCoursewareCache(mockRedis, cfg)
	resolver := identity.NewCoursewareIdentityResolver(nil, cache, cfg)
	identity.SetIdentityResolver(resolver)
	return func() {
		// 恢复默认解析器
		identity.SetIdentityResolver(&identity.DefaultIdentityResolver{})
	}
}

func TestCoursewareWebhook_UpsertUser_签名通过更新成功(t *testing.T) {
	t.Cleanup(setupResolverForTest())
	os.Setenv("COURSEWARE_WEBHOOK_SECRET", "test-secret")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/courseware/webhook/user", controller.CoursewareWebhookUpsertUser)

	payload := map[string]interface{}{
		"teacher_id":      "t_001",
		"teacher_name":    "张老师",
		"group_name":      "beijing_math_group",
		"preferred_model": "gpt-4",
	}
	body, _ := json.Marshal(payload)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := sign("test-secret", ts, body)

	req, _ := http.NewRequest("POST", "/api/courseware/webhook/user", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Timestamp", ts)
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestCoursewareWebhook_UpsertUser_签名失败(t *testing.T) {
	t.Cleanup(setupResolverForTest())
	os.Setenv("COURSEWARE_WEBHOOK_SECRET", "test-secret")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/courseware/webhook/user", controller.CoursewareWebhookUpsertUser)

	body := []byte(`{"teacher_id":"t_001"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	// 使用错误密钥签名
	sig := sign("wrong", ts, body)

	req, _ := http.NewRequest("POST", "/api/courseware/webhook/user", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Timestamp", ts)
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
