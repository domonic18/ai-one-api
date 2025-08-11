package integration

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

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/controller"
	identity "github.com/songquanpeng/one-api/middleware/identity"
	testmocks "github.com/songquanpeng/one-api/tests/mocks"
	"github.com/stretchr/testify/mock"
)

func signWebhook(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(body)))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestCoursewareWebhook_Upsert_Delete_端到端(t *testing.T) {
	// 启用 webhook
	os.Setenv("COURSEWARE_WEBHOOK_SECRET", "int-secret")

	// 准备依赖：启用解析器与缓存（使用Mock Redis）
	mockRedis := testmocks.NewMockRedisClient()
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRedis.On("Del", mock.Anything, mock.Anything).Return(nil)
	cfg := &identity.CoursewareConfig{CacheTTL: time.Minute}
	cache := identity.NewCoursewareCache(mockRedis, cfg)
	resolver := identity.NewCoursewareIdentityResolver(nil, cache, cfg)
	identity.SetIdentityResolver(resolver)
	t.Cleanup(func() { identity.SetIdentityResolver(&identity.DefaultIdentityResolver{}) })

	// 路由
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("session", store))
	r.POST("/api/courseware/webhook/user", controller.CoursewareWebhookUpsertUser)
	r.DELETE("/api/courseware/webhook/user/:teacher_id", controller.CoursewareWebhookDeleteUser)

	// Upsert
	up := map[string]interface{}{
		"teacher_id": "t_100",
		"group_name": "group_a",
	}
	body, _ := json.Marshal(up)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := signWebhook("int-secret", ts, body)

	req, _ := http.NewRequest("POST", "/api/courseware/webhook/user", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Timestamp", ts)
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	req2, _ := http.NewRequest("DELETE", "/api/courseware/webhook/user/t_100", bytes.NewReader([]byte{}))
	ts2 := strconv.FormatInt(time.Now().Unix(), 10)
	sig2 := signWebhook("int-secret", ts2, []byte{})
	req2.Header.Set("X-Webhook-Timestamp", ts2)
	req2.Header.Set("X-Webhook-Signature", sig2)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
