package integration

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/cache"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/smart"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/router"
	"github.com/songquanpeng/one-api/tests/fixtures"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 设置集成测试环境
func setupIntegrationTest() (*gin.Engine, *gorm.DB) {
	// 设置测试环境变量
	setupTestEnvironment()

	// 直接设置配置变量来禁用限流
	config.DebugEnabled = true
	config.GlobalWebRateLimitNum = 0
	config.GlobalApiRateLimitNum = 0

	// 初始化tiktoken编码器
	openai.InitTokenEncoders()

	// 初始化HTTP客户端
	client.Init()

	// 初始化Redis客户端
	err := common.InitRedisClient()
	if err != nil {
		panic("failed to initialize Redis client: " + err.Error())
	}

	// 初始化缓存系统
	cache.Init()

	// 连接MySQL数据库
	db, err := connectTestMySQL()
	if err != nil {
		panic("failed to connect to test MySQL database: " + err.Error())
	}

	// 设置MySQL模式
	common.UsingMySQL = true
	common.UsingSQLite = false

	// 设置数据库连接
	model.DB = db
	model.LOG_DB = db

	// 清理现有数据和表结构，避免迁移冲突
	cleanupTestData(db)

	// 删除可能存在的扩展日志表以避免外键约束问题
	db.Exec("DROP TABLE IF EXISTS extended_logs")

	// 初始化智能模型系统（需要在设置DB之后）
	smart.Init()

	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Log{},
		&model.Ability{},
		&model.Option{},
		&model.Redemption{},
		&smart.ExtendedLog{},
	)
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// 再次清理数据
	cleanupTestData(db)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// 确保没有设置FRONTEND_BASE_URL环境变量，避免重定向
	os.Setenv("FRONTEND_BASE_URL", "")

	// 设置session中间件
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("one-api", store))

	// 插入预定义测试数据
	if err := fixtures.InsertTestData(db); err != nil {
		panic("failed to insert test data: " + err.Error())
	}

	// 设置完整的路由（使用空的buildFS）
	var buildFS embed.FS
	router.SetRouter(r, buildFS)

	return r, db
}

// 设置测试环境变量
func setupTestEnvironment() {
	// 启用DEBUG模式来禁用限流
	os.Setenv("DEBUG", "true")

	// 禁用限流，避免测试时触发限流
	os.Setenv("GLOBAL_WEB_RATE_LIMIT", "0")
	os.Setenv("GLOBAL_API_RATE_LIMIT", "0")

	// 设置MySQL连接字符串
	if os.Getenv("SQL_DSN") == "" {
		os.Setenv("SQL_DSN", "testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local")
	}

	// 设置Redis连接字符串
	if os.Getenv("REDIS_CONN_STRING") == "" {
		os.Setenv("REDIS_CONN_STRING", "redis://localhost:6379")
	}

	// 设置同步频率
	if os.Getenv("SYNC_FREQUENCY") == "" {
		os.Setenv("SYNC_FREQUENCY", "60")
	}

	// 设置会话密钥
	if os.Getenv("SESSION_SECRET") == "" {
		os.Setenv("SESSION_SECRET", "test-secret-key")
	}
}

// 连接测试MySQL数据库
func connectTestMySQL() (*gorm.DB, error) {
	dsn := os.Getenv("SQL_DSN")
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})
}

// 创建测试用户（保留原有函数以兼容现有测试）
func createTestUser(db *gorm.DB, username, password string, role int) *model.User {
	hashedPassword, _ := common.Password2Hash(password)
	// 生成较短的access_token避免MySQL字段长度限制
	accessToken := "test-" + username
	if len(accessToken) > 32 {
		accessToken = accessToken[:32]
	}

	// 生成较短的aff_code
	affCode := "aff-" + username
	if len(affCode) > 8 {
		affCode = affCode[:8]
	}

	user := &model.User{
		Username:    username,
		Password:    hashedPassword,
		Status:      1,
		Role:        role,
		Quota:       10000,
		AccessToken: accessToken,
		AffCode:     affCode,
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
	// 生成唯一的用户名，使用时间戳
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("testuser_%d", timestamp)
	return createTestUser(db, username, "password123", model.RoleCommonUser)
}

// 创建测试令牌
func createTestToken(db *gorm.DB, userId int, name string) *model.Token {
	// 生成不包含连字符的简单key，避免被TokenAuth中间件分割
	timestamp := time.Now().UnixNano()
	tokenKey := fmt.Sprintf("testkey%d", timestamp)

	token := &model.Token{
		UserId:         userId,
		Key:            tokenKey,
		Name:           name,
		Status:         1,
		RemainQuota:    1000,
		UnlimitedQuota: false,
		Models:         nil,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
	}
	db.Create(token)
	return token
}

// 创建测试渠道
func createTestChannel(db *gorm.DB, name, key string) *model.Channel {
	channel := &model.Channel{
		Type:        1, // OpenAI
		Key:         key,
		Name:        name,
		Status:      1,
		Weight:      uintPtr(100),
		Group:       "default",
		Models:      "gpt-3.5-turbo",
		Balance:     100.0,
		CreatedTime: time.Now().Unix(),
	}
	db.Create(channel)
	return channel
}

// 创建测试日志
func createTestLog(db *gorm.DB, userId int, logType int, quota int, modelName string) *model.Log {
	log := &model.Log{
		UserId:           userId,
		Type:             logType,
		Username:         "testuser",
		TokenName:        "test-token",
		ModelName:        modelName,
		Quota:            quota,
		PromptTokens:     100,
		CompletionTokens: 50,
		ChannelId:        1,
		ElapsedTime:      1000,
		Content:          "测试日志",
		CreatedAt:        time.Now().Unix(),
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

	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// 用户登录辅助函数
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
	fixtures.ClearTestData(db)
}

// 辅助函数：返回uint指针
func uintPtr(u uint) *uint {
	return &u
}
