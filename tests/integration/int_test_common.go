package integration

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/router"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 设置集成测试环境
func setupIntegrationTest() (*gin.Engine, *gorm.DB) {
	// 初始化Redis客户端（禁用Redis）
	common.RedisEnabled = false

	// 禁用限流，避免测试时触发限流
	os.Setenv("GLOBAL_WEB_RATE_LIMIT", "0")
	os.Setenv("GLOBAL_API_RATE_LIMIT", "0")

	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 设置SQLite模式
	common.UsingSQLite = true

	// 迁移所有表
	model.DB = db
	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Log{},
		&model.Ability{},
		&model.Option{},
		&model.Redemption{},
	)
	if err != nil {
		panic("failed to migrate database")
	}

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// 确保没有设置FRONTEND_BASE_URL环境变量，避免重定向
	os.Setenv("FRONTEND_BASE_URL", "")

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("one-api", store))

	// 设置数据库连接
	model.DB = db
	model.LOG_DB = db // 确保日志数据库也被设置

	// 设置完整的路由（使用空的buildFS）
	var buildFS embed.FS
	router.SetRouter(r, buildFS)

	return r, db
}

// 创建测试用户
func createTestUser(db *gorm.DB, username, password string, role int) *model.User {
	hashedPassword, _ := common.Password2Hash(password)
	user := &model.User{
		Username:    username,
		Password:    hashedPassword,
		Status:      1,
		Role:        role,
		Quota:       10000,
		AccessToken: "test-token-" + username,
		AffCode:     "test-aff-" + username, // 为每个用户生成唯一的推荐码
	}
	db.Create(user)
	return user
}

// 创建测试管理员用户
func createTestAdmin(db *gorm.DB) *model.User {
	return createTestUser(db, "admin", "admin123", model.RoleAdminUser)
}

// 创建测试普通用户
func createTestNormalUser(db *gorm.DB) *model.User {
	return createTestUser(db, "testuser", "password123", model.RoleCommonUser)
}

// 创建测试令牌
func createTestToken(db *gorm.DB, userId int, name string) *model.Token {
	token := &model.Token{
		UserId:      userId,
		Name:        name,
		Key:         "test-key-" + name,
		Status:      1,
		RemainQuota: 10000,
	}
	db.Create(token)
	return token
}

// 创建测试渠道
func createTestChannel(db *gorm.DB, name, key string) *model.Channel {
	channel := &model.Channel{
		Type:    1,
		Key:     key,
		Name:    name,
		Status:  1,
		Group:   "default",
		Models:  "gpt-3.5-turbo,gpt-4",
		Balance: 100.0,
	}
	db.Create(channel)
	return channel
}

// 创建测试日志
func createTestLog(db *gorm.DB, userId int, logType int, quota int, modelName string) *model.Log {
	log := &model.Log{
		UserId:    userId,
		Type:      logType,
		Quota:     quota,
		ModelName: modelName,
	}
	db.Create(log)
	return log
}

// 发送HTTP请求的辅助函数
func sendRequest(r *gin.Engine, method, path string, payload interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// 登录用户并返回session
func loginUser(r *gin.Engine, username, password string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"username": username,
		"password": password,
	}
	return sendRequest(r, "POST", "/api/user/login", payload, nil)
}

// 获取用户session的辅助函数
func getUserSession(r *gin.Engine, username, password string) *httptest.ResponseRecorder {
	return loginUser(r, username, password)
}

// 清理测试数据
func cleanupTestData(db *gorm.DB) {
	db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{}, &model.Ability{}, &model.Option{}, &model.Redemption{})
}
