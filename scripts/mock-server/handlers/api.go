package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"mock-courseware-platform/config"
	"mock-courseware-platform/models"
	"mock-courseware-platform/storage"

	"github.com/gin-gonic/gin"
)

var (
	memoryStorage *storage.MemoryStorage
	fileStorage   *storage.FileStorage
)

// InitHandlers 初始化处理器
func InitHandlers() {
	memoryStorage = storage.NewMemoryStorage()
	fileStorage = storage.NewFileStorage()

	// 初始化默认数据
	if err := fileStorage.InitializeDefaultData(); err != nil {
		panic(err)
	}

	// 加载用户数据到内存
	users, err := fileStorage.LoadUsers()
	if err != nil {
		panic(err)
	}
	memoryStorage.LoadDefaultUsers(users)

	// 优先使用config包中的配置（包含环境变量）
	serverConfig := config.GetConfig()
	if serverConfig == nil {
		// 如果config包中的配置为空，尝试从文件存储加载
		fileConfig, err := fileStorage.LoadConfig()
		if err != nil {
			panic(err)
		}

		if fileConfig.API.APIKey == "" {
			// 如果文件配置也为空，创建默认配置
			serverConfig = &models.ServerConfig{
				Server: struct {
					Port int    `json:"port"`
					Host string `json:"host"`
				}{
					Port: 8080,
					Host: "0.0.0.0",
				},
				API: struct {
					ResponseDelay string  `json:"response_delay"`
					ErrorRate     float64 `json:"error_rate"`
					APIKey        string  `json:"api_key"`
				}{
					ResponseDelay: "0ms",
					ErrorRate:     0.0,
					APIKey:        "mock_api_key_123",
				},
				DefaultUsers: []*models.TeacherInfo{},
			}
		} else {
			serverConfig = fileConfig
		}
	}

	fmt.Printf("Debug: InitHandlers - Final API Key: %s\n", serverConfig.API.APIKey)
	memoryStorage.SetConfig(serverConfig)
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 从配置中获取API密钥
		apiKey := config.GetAPIKey()
		expectedToken := "Bearer " + apiKey

		// 添加调试日志
		fmt.Printf("Debug: API Key from config: %s\n", apiKey)
		fmt.Printf("Debug: Auth Header: %s\n", authHeader)
		fmt.Printf("Debug: Expected Token: %s\n", expectedToken)

		if authHeader != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid api key",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetTeacherInfo 获取教师信息
func GetTeacherInfo(c *gin.Context) {
	teacherId := c.Param("teacher_id")

	// 模拟网络延迟
	if delay := config.GetResponseDelay(); delay > 0 {
		time.Sleep(delay)
	}

	// 模拟错误率
	if rand.Float64() < config.GetErrorRate() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "internal server error",
			"data":    nil,
		})
		return
	}

	user, err := memoryStorage.GetUser(teacherId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "teacher not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    user,
	})
}

// GetTeacherIds 获取所有教师ID列表
func GetTeacherIds(c *gin.Context) {
	// 模拟网络延迟
	if delay := config.GetResponseDelay(); delay > 0 {
		time.Sleep(delay)
	}

	// 模拟错误率
	if rand.Float64() < config.GetErrorRate() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "internal server error",
			"data":    nil,
		})
		return
	}

	teacherIds, err := memoryStorage.GetAllUserIds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "failed to get teacher ids",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"teacher_ids": teacherIds,
			"total":       len(teacherIds),
		},
	})
}

// BatchGetTeachers 批量获取教师信息
func BatchGetTeachers(c *gin.Context) {
	var req models.BatchTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
			"data":    nil,
		})
		return
	}

	// 模拟网络延迟
	if delay := config.GetResponseDelay(); delay > 0 {
		time.Sleep(delay)
	}

	// 模拟错误率
	if rand.Float64() < config.GetErrorRate() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "internal server error",
			"data":    nil,
		})
		return
	}

	users, err := memoryStorage.BatchGetUsers(req.TeacherIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "failed to get teachers",
			"data":    nil,
		})
		return
	}

	successCount := len(users)
	errorCount := len(req.TeacherIds) - successCount

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"users":         users,
			"success_count": successCount,
			"error_count":   errorCount,
		},
	})
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "mock courseware platform api server is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}
