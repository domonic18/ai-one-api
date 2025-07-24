package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupIntegrationTest() (*gin.Engine, *gorm.DB) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移所有表
	model.DB = db
	db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Log{},
	)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// 设置路由
	setupTestRoutes(r)

	return r, db
}

func setupTestRoutes(r *gin.Engine) {
	// API路由
	api := r.Group("/api")
	{
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 用户路由
		api.POST("/user/login", controller.Login)
		api.GET("/user/self", controller.GetSelf)
		
		// 令牌路由
		api.GET("/token", controller.GetAllTokens)
		api.POST("/token", controller.AddToken)
		
		// 渠道路由
		api.GET("/channel", controller.GetAllChannels)
		api.POST("/channel", controller.AddChannel)
		
		// 日志路由
		api.GET("/log", controller.GetAllLogs)
		api.GET("/log/self", controller.GetUserLogs)
	}
}

func TestAPI_UserLogin(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: common.Password2Hash("password123"),
		Status:   1,
	}
	db.Create(&user)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "正确登录",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "错误密码",
			payload: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "用户不存在",
			payload: map[string]interface{}{
				"username": "nonexistent",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "空用户名",
			payload: map[string]interface{}{
				"username": "",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedMsg != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedMsg)
			}
		})
	}
}

func TestAPI_TokenManagement(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建测试用户
	user := model.User{
		Username: "testuser",
		Password: common.Password2Hash("password123"),
		Status:   1,
	}
	db.Create(&user)

	// 登录获取session
	jsonData, _ := json.Marshal(map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	})
	
	loginReq := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonData))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	assert.Equal(t, http.StatusOK, loginW.Code)

	tests := []struct {
		name           string
		method         string
		path           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "创建令牌",
			method:         "POST",
			path:           "/api/token",
			payload:        map[string]interface{}{
				"name": "Test Token",
				"remain_quota": 1000,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取令牌列表",
			method:         "GET",
			path:           "/api/token",
			payload:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "创建空名称令牌",
			method:         "POST",
			path:           "/api/token",
			payload:        map[string]interface{}{
				"name": "",
				"remain_quota": 1000,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.payload != nil {
				jsonData, _ := json.Marshal(tt.payload)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPI_ChannelManagement(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建管理员用户
	admin := model.User{
		Username: "admin",
		Password: common.Password2Hash("admin123"),
		Status:   1,
		Role:     10, // 管理员
	}
	db.Create(&admin)

	tests := []struct {
		name           string
		method         string
		path           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "创建渠道",
			method:         "POST",
			path:           "/api/channel",
			payload: map[string]interface{}{
				"name":     "OpenAI Channel",
				"type":     1,
				"key":      "sk-test-key",
				"status":   1,
				"models":   "gpt-3.5-turbo,gpt-4",
				"group":    "default",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取渠道列表",
			method:         "GET",
			path:           "/api/channel",
			payload:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "创建无效渠道",
			method:         "POST",
			path:           "/api/channel",
			payload: map[string]interface{}{
				"name": "",
				"type": 999,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.payload != nil {
				jsonData, _ := json.Marshal(tt.payload)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPI_LogQuery(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建测试数据
	user := model.User{Username: "testuser"}
	db.Create(&user)

	logs := []model.Log{
		{UserId: user.Id, Type: model.LogTypeConsume, Quota: 100, ModelName: "gpt-3.5-turbo"},
		{UserId: user.Id, Type: model.LogTypeConsume, Quota: 200, ModelName: "gpt-4"},
		{UserId: user.Id, Type: model.LogTypeRecharge, Quota: 1000},
	}
	for _, log := range logs {
		db.Create(&log)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "获取所有日志",
			path:           "/api/log",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "获取用户日志",
			path:           "/api/log/self",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "带分页参数",
			path:           "/api/log?p=1&size=10",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPI_ErrorHandling(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	tests := []struct {
		name           string
		method         string
		path           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "无效JSON格式",
			method:         "POST",
			path:           "/api/user/login",
			payload:        nil, // 将发送无效JSON
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid",
		},
		{
			name:           "404错误",
			method:         "GET",
			path:           "/api/nonexistent",
			payload:        nil,
			expectedStatus: http.StatusNotFound,
			expectedError:  "404",
		},
		{
			name:           "方法不允许",
			method:         "DELETE",
			path:           "/api/status",
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.payload == nil {
				body = []byte(`{invalid json}`)
			} else {
				body, _ = json.Marshal(tt.payload)
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
			if tt.payload != nil || tt.name == "无效JSON格式" {
				req.Header.Set("Content-Type", "application/json")
			}
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPI_Performance(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建大量测试数据
	for i := 0; i < 100; i++ {
		user := model.User{
			Username: "user" + string(rune(i)),
			Password: common.Password2Hash("password"),
			Status:   1,
		}
		db.Create(&user)

		log := model.Log{
			UserId:    user.Id,
			Type:      model.LogTypeConsume,
			Quota:     100,
			ModelName: "gpt-3.5-turbo",
		}
		db.Create(&log)
	}

	// 测试性能
	start := time.Now()
	req := httptest.NewRequest("GET", "/api/log", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	duration := time.Since(start)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, duration, 100*time.Millisecond, "API响应时间应该小于100ms")
}

func TestAPI_ConcurrentAccess(t *testing.T) {
	r, db := setupIntegrationTest()
	defer db.Migrator().DropTable(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{})

	// 创建测试用户
	user := model.User{
		Username: "concurrentuser",
		Password: common.Password2Hash("password123"),
		Status:   1,
		Quota:    10000,
	}
	db.Create(&user)

	// 并发测试
	concurrency := 10
	results := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			jsonData, _ := json.Marshal(map[string]interface{}{
				"username": "concurrentuser",
				"password": "password123",
			})
			
			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			results <- w.Code == http.StatusOK
		}()
	}

	// 验证所有并发请求都成功
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-results {
			successCount++
		}
	}
	
	assert.Equal(t, concurrency, successCount, "所有并发请求都应该成功")
}