package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type webhookUpsertReq struct {
	TeacherId      string `json:"teacher_id" binding:"required"`
	TeacherName    string `json:"teacher_name"`
	GroupName      string `json:"group_name"`
	PreferredModel string `json:"preferred_model"`
	SchoolId       int    `json:"school_id"`
	SchoolName     string `json:"school_name"`
	SubjectId      int    `json:"subject_id"`
	SubjectName    string `json:"subject_name"`
}

type webhookBatchReq struct {
	Users []webhookUpsertReq `json:"users" binding:"required"`
}

// sign 生成 OneAPI webhook 签名（sha256=hex(hmacSHA256(ts+"."+body, secret)))
func sign(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(body)))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// WebhookUpsertUser 通过 webhook 调用 OneAPI 设置单用户缓存
func WebhookUpsertUser(c *gin.Context) {
	cfg := memoryStorage.GetConfig()
	if cfg == nil || cfg.API.OneAPIBaseURL == "" || cfg.API.OneAPIWebhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OneAPI webhook 配置缺失: oneapi_base_url/oneapi_webhook_secret", "config": cfg})
		return
	}

	var req webhookUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 组装请求体
	bodyBytes, _ := json.Marshal(req)
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sig := sign(cfg.API.OneAPIWebhookSecret, ts, bodyBytes)

	// 发送到 OneAPI
	url := fmt.Sprintf("%s/api/courseware/webhook/user", cfg.API.OneAPIBaseURL)
	httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Timestamp", ts)
	httpReq.Header.Set("X-Webhook-Signature", sig)
	httpReq.Header.Set("X-Webhook-Event", "courseware.user.upsert")

	// 显示更详细的日志，便于排查
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "调用 OneAPI 失败", "detail": err.Error(), "url": url})
		return
	}
	defer resp.Body.Close()

	c.JSON(resp.StatusCode, gin.H{"message": "forwarded", "status": resp.Status})
}

// WebhookBatchUpsertUsers 通过 webhook 批量设置用户缓存
func WebhookBatchUpsertUsers(c *gin.Context) {
	cfg := memoryStorage.GetConfig()
	if cfg == nil || cfg.API.OneAPIBaseURL == "" || cfg.API.OneAPIWebhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OneAPI webhook 配置缺失: oneapi_base_url/oneapi_webhook_secret", "config": cfg})
		return
	}

	var req webhookBatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bodyBytes, _ := json.Marshal(req)
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sig := sign(cfg.API.OneAPIWebhookSecret, ts, bodyBytes)

	url := fmt.Sprintf("%s/api/courseware/webhook/users", cfg.API.OneAPIBaseURL)
	httpReq, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Timestamp", ts)
	httpReq.Header.Set("X-Webhook-Signature", sig)
	httpReq.Header.Set("X-Webhook-Event", "courseware.users.batch_upsert")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "调用 OneAPI 失败", "detail": err.Error(), "url": url})
		return
	}
	defer resp.Body.Close()

	c.JSON(resp.StatusCode, gin.H{"message": "forwarded", "status": resp.Status})
}

// WebhookDeleteUser 通过 webhook 删除用户缓存
func WebhookDeleteUser(c *gin.Context) {
	cfg := memoryStorage.GetConfig()
	if cfg == nil || cfg.API.OneAPIBaseURL == "" || cfg.API.OneAPIWebhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OneAPI webhook 配置缺失: oneapi_base_url/oneapi_webhook_secret", "config": cfg})
		return
	}

	teacherId := c.Param("teacher_id")
	if teacherId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "teacher_id is required"})
		return
	}

	// 删除请求体可为空，但签名仍需包含 ts + "." + body
	bodyBytes := []byte{}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sig := sign(cfg.API.OneAPIWebhookSecret, ts, bodyBytes)

	url := fmt.Sprintf("%s/api/courseware/webhook/user/%s", cfg.API.OneAPIBaseURL, teacherId)
	httpReq, _ := http.NewRequest(http.MethodDelete, url, bytes.NewReader(bodyBytes))
	httpReq.Header.Set("X-Webhook-Timestamp", ts)
	httpReq.Header.Set("X-Webhook-Signature", sig)
	httpReq.Header.Set("X-Webhook-Event", "courseware.user.delete")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "调用 OneAPI 失败", "detail": err.Error(), "url": url})
		return
	}
	defer resp.Body.Close()

	c.JSON(resp.StatusCode, gin.H{"message": "forwarded", "status": resp.Status})
}
