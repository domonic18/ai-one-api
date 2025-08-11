package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	identity "github.com/songquanpeng/one-api/middleware/identity"
)

// 课件平台 Webhook 事件头
const (
	headerSignature = "X-Webhook-Signature"
	headerTimestamp = "X-Webhook-Timestamp"
	headerEvent     = "X-Webhook-Event"
)

// 事件名称约定
const (
	eventUserUpsert = "courseware.user.upsert"
	eventUsersBatch = "courseware.users.batch_upsert"
	eventUserDelete = "courseware.user.delete"
)

// webhookEnabled 判断是否开启
func webhookEnabled() bool {
	return os.Getenv("COURSEWARE_WEBHOOK_SECRET") != ""
}

// verifySignature 校验签名（HMAC-SHA256，原文: timestamp + "." + body，头部: X-Webhook-Signature=sha256=HEX）
func verifySignature(c *gin.Context, body []byte) bool {
	secret := os.Getenv("COURSEWARE_WEBHOOK_SECRET")
	if secret == "" {
		return false
	}
	ts := c.GetHeader(headerTimestamp)
	sig := c.GetHeader(headerSignature)
	if ts == "" || sig == "" {
		return false
	}
	// 防重放: 默认允许5分钟内
	if unix, err := strconv.ParseInt(ts, 10, 64); err == nil {
		now := time.Now().Unix()
		if now-unix > 300 || unix-now > 300 {
			return false
		}
	} else {
		return false
	}

	payload := []byte(ts + "." + string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)

	// 签名格式: sha256=HEX
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}
	hexPart := strings.TrimPrefix(sig, "sha256=")
	given, err := hex.DecodeString(hexPart)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, given)
}

// WebhookUserPayload 单用户变更负载
type WebhookUserPayload struct {
	TeacherId      string `json:"teacher_id" binding:"required"`
	TeacherName    string `json:"teacher_name"`
	GroupName      string `json:"group_name"`
	PreferredModel string `json:"preferred_model"`
	SchoolId       int    `json:"school_id"`
	SchoolName     string `json:"school_name"`
	SubjectId      int    `json:"subject_id"`
	SubjectName    string `json:"subject_name"`
}

// WebhookUsersBatchPayload 批量用户变更负载
type WebhookUsersBatchPayload struct {
	Users []WebhookUserPayload `json:"users" binding:"required"`
}

// CoursewareWebhookUpsertUser 处理单用户新增/更新
func CoursewareWebhookUpsertUser(c *gin.Context) {
	if !webhookEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "webhook 未启用"})
		return
	}

	// 读取原始请求体用于签名校验
	raw, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "读取请求失败"})
		return
	}
	// 恢复请求体以供绑定
	c.Request.Body = ioutil.NopCloser(strings.NewReader(string(raw)))

	if !verifySignature(c, raw) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "签名验证失败"})
		return
	}

	// 事件校验（可选）
	if evt := c.GetHeader(headerEvent); evt != "" && evt != eventUserUpsert {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "事件类型不匹配"})
		return
	}

	var payload WebhookUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 仅在课件平台解析器启用时接收
	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "课件平台集成未启用"})
		return
	}

	userInfo := &identity.UserInfo{
		TeacherId:      payload.TeacherId,
		TeacherName:    payload.TeacherName,
		GroupName:      payload.GroupName,
		PreferredModel: payload.PreferredModel,
		SchoolId:       payload.SchoolId,
		SchoolName:     payload.SchoolName,
		SubjectId:      payload.SubjectId,
		SubjectName:    payload.SubjectName,
		UpdatedAt:      time.Now().Unix(),
	}

	if err := coursewareResolver.GetCache().SetUserInfo(c.Request.Context(), userInfo); err != nil {
		logger.Warnf(c.Request.Context(), "webhook写入缓存失败: teacher=%s, err=%v", payload.TeacherId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "缓存更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CoursewareWebhookBatchUpsertUsers 处理批量新增/更新
func CoursewareWebhookBatchUpsertUsers(c *gin.Context) {
	if !webhookEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "webhook 未启用"})
		return
	}

	raw, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "读取请求失败"})
		return
	}
	c.Request.Body = ioutil.NopCloser(strings.NewReader(string(raw)))

	if !verifySignature(c, raw) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "签名验证失败"})
		return
	}
	if evt := c.GetHeader(headerEvent); evt != "" && evt != eventUsersBatch {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "事件类型不匹配"})
		return
	}

	var payload WebhookUsersBatchPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求格式错误"})
		return
	}
	if len(payload.Users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "users 不能为空"})
		return
	}

	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "课件平台集成未启用"})
		return
	}

	userInfos := make([]*identity.UserInfo, 0, len(payload.Users))
	for _, u := range payload.Users {
		userInfos = append(userInfos, &identity.UserInfo{
			TeacherId:      u.TeacherId,
			TeacherName:    u.TeacherName,
			GroupName:      u.GroupName,
			PreferredModel: u.PreferredModel,
			SchoolId:       u.SchoolId,
			SchoolName:     u.SchoolName,
			SubjectId:      u.SubjectId,
			SubjectName:    u.SubjectName,
			UpdatedAt:      time.Now().Unix(),
		})
	}

	if err := coursewareResolver.GetCache().BatchSetUserInfo(c.Request.Context(), userInfos); err != nil {
		logger.Warnf(c.Request.Context(), "webhook批量写入缓存失败: count=%d, err=%v", len(userInfos), err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "缓存更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"updated": len(userInfos)}})
}

// CoursewareWebhookDeleteUser 处理删除缓存
func CoursewareWebhookDeleteUser(c *gin.Context) {
	if !webhookEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "webhook 未启用"})
		return
	}

	raw, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "读取请求失败"})
		return
	}
	// 删除可以无体，但若有体也参与签名
	if !verifySignature(c, raw) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "签名验证失败"})
		return
	}
	if evt := c.GetHeader(headerEvent); evt != "" && evt != eventUserDelete {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "事件类型不匹配"})
		return
	}

	teacherId := c.Param("teacher_id")
	if teacherId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "teacher_id 不能为空"})
		return
	}

	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "课件平台集成未启用"})
		return
	}

	if err := coursewareResolver.GetCache().DeleteUserInfo(c.Request.Context(), teacherId); err != nil {
		logger.Warnf(c.Request.Context(), "webhook删除缓存失败: teacher=%s, err=%v", teacherId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "缓存删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
