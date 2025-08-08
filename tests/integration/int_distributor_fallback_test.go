package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	midIdentity "github.com/songquanpeng/one-api/middleware/identity"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestDistributor_GroupFallback 当主组无可用渠道时，应按 user_groups 顺序回退到后续组
func TestDistributor_GroupFallback(t *testing.T) {
	// 复用公共集成测试初始化工具（与其他API集成测试保持一致）
	_, db := setupIntegrationTest()
	cleanupTestData(db)

	// 创建测试用户与令牌
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "fallback-token")
	// 令牌用户组：优先 beijing_math_group，其次 default
	_ = token.SetUserGroups([]string{"beijing_math_group", "default"})
	_ = token.Update()

	// 在 default 组创建支持 qwen-max 的渠道；在 beijing_math_group 不创建对应渠道
	ch := &model.Channel{
		Type:   1, // OpenAI 兼容适配器
		Key:    "sk-test",
		Name:   "default-qwen",
		Status: model.ChannelStatusEnabled,
		Group:  "default",
		Models: "qwen-max",
	}
	err := model.DB.Create(ch).Error
	assert.NoError(t, err)
	// 建立能力表记录，便于 CacheGetRandomSatisfiedChannel 命中
	_ = ch.AddAbilities()

	// 构造仅含鉴权、身份、分发的最小路由，使用桩处理器
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TokenAuth(), midIdentity.Identity(), middleware.Distribute())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		// 验证分发后选择的组应为 default（因为主组无可用渠道，应回退）
		selectedGroup := c.GetString(ctxkey.Group)
		selectedChannelId := c.GetInt(ctxkey.ChannelId)
		c.JSON(http.StatusOK, gin.H{
			"selected_group":   selectedGroup,
			"selected_channel": selectedChannelId,
		})
	})

	// 组织请求
	body := map[string]interface{}{
		"model":    "qwen-max",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Key)

	// 执行
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 断言：应成功且选中 default 组
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "default", resp["selected_group"])
}

// 三组用户组（A、B、C）回滚验证：A 组有可用渠道 → 直接使用 A
func TestDistributor_GroupFallback_ABC_UseA(t *testing.T) {
	_, db := setupIntegrationTest()
	cleanupTestData(db)

	// 创建用户与令牌，用户组优先级：A > B > C
	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "fallback-abc-a")
	_ = token.SetUserGroups([]string{"group_a", "group_b", "group_c"})
	_ = token.Update()

	// 仅在 A 组创建可用渠道
	chA := &model.Channel{Type: 1, Key: "sk-a", Name: "chan-a", Status: model.ChannelStatusEnabled, Group: "group_a", Models: "qwen-max"}
	err := model.DB.Create(chA).Error
	assert.NoError(t, err)
	_ = chA.AddAbilities()

	// 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TokenAuth(), midIdentity.Identity(), middleware.Distribute())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"selected_group": c.GetString(ctxkey.Group),
		})
	})

	// 请求
	body := map[string]interface{}{"model": "qwen-max", "messages": []map[string]string{{"role": "user", "content": "hello"}}}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Key)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "group_a", resp["selected_group"]) // 直接命中 A
}

// 三组用户组（A、B、C）回滚验证：仅 B 组有可用渠道 → 回滚到 B
func TestDistributor_GroupFallback_ABC_UseB(t *testing.T) {
	_, db := setupIntegrationTest()
	cleanupTestData(db)

	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "fallback-abc-b")
	_ = token.SetUserGroups([]string{"group_a", "group_b", "group_c"})
	_ = token.Update()

	// 仅在 B 组创建可用渠道
	chB := &model.Channel{Type: 1, Key: "sk-b", Name: "chan-b", Status: model.ChannelStatusEnabled, Group: "group_b", Models: "qwen-max"}
	err := model.DB.Create(chB).Error
	assert.NoError(t, err)
	_ = chB.AddAbilities()

	// 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TokenAuth(), midIdentity.Identity(), middleware.Distribute())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"selected_group": c.GetString(ctxkey.Group)})
	})

	body := map[string]interface{}{"model": "qwen-max", "messages": []map[string]string{{"role": "user", "content": "hello"}}}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Key)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "group_b", resp["selected_group"]) // 回滚到 B
}

// 三组用户组（A、B、C）回滚验证：仅 C 组有可用渠道 → 最终回滚到 C
func TestDistributor_GroupFallback_ABC_UseC(t *testing.T) {
	_, db := setupIntegrationTest()
	cleanupTestData(db)

	user := createTestNormalUser(db)
	token := createTestToken(db, user.Id, "fallback-abc-c")
	_ = token.SetUserGroups([]string{"group_a", "group_b", "group_c"})
	_ = token.Update()

	// 仅在 C 组创建可用渠道
	chC := &model.Channel{Type: 1, Key: "sk-c", Name: "chan-c", Status: model.ChannelStatusEnabled, Group: "group_c", Models: "qwen-max"}
	err := model.DB.Create(chC).Error
	assert.NoError(t, err)
	_ = chC.AddAbilities()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TokenAuth(), midIdentity.Identity(), middleware.Distribute())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"selected_group": c.GetString(ctxkey.Group)})
	})

	body := map[string]interface{}{"model": "qwen-max", "messages": []map[string]string{{"role": "user", "content": "hello"}}}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Key)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "group_c", resp["selected_group"]) // 最终回滚到 C
}
