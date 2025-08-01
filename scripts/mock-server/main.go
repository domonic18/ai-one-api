package main

import (
	"fmt"
	"log"
	"net/http"

	"mock-courseware-platform/config"
	"mock-courseware-platform/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化处理器
	handlers.InitHandlers()

	// 创建Gin引擎
	r := gin.Default()

	// 设置HTML模板
	r.LoadHTMLGlob("static/*.html")
	r.Static("/static", "./static")

	// 设置路由
	setupRoutes(r)

	// 获取配置
	cfg := config.GetConfig()
	if cfg == nil {
		log.Fatal("配置为空")
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("模拟课件平台API服务器启动在: http://%s", addr)
	log.Printf("Web管理界面: http://%s", addr)
	log.Printf("健康检查: http://%s/health", addr)
	log.Printf("API文档: http://%s/api/docs", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine) {
	// Web管理界面路由
	r.GET("/", handlers.ServeWebInterface)

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// API文档
	r.GET("/api/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "api-docs.html", gin.H{
			"title": "API文档",
		})
	})

	// Web管理API路由组
	web := r.Group("/web")
	{
		web.GET("/users", handlers.GetUsers)
		web.GET("/users/:teacher_id", handlers.GetUser)
		web.POST("/users", handlers.AddUser)
		web.PUT("/users", handlers.UpdateUser)
		web.DELETE("/users/:teacher_id", handlers.DeleteUser)
		web.GET("/config", handlers.GetConfig)
		web.PUT("/config", handlers.UpdateConfig)
		web.POST("/reset", handlers.ResetToDefault)
		web.GET("/test/:type", handlers.TestAPI)
	}

	// 课件平台API路由组（需要认证）
	api := r.Group("/api/v1")
	api.Use(handlers.AuthMiddleware())
	{
		// 获取教师信息
		api.GET("/teacher/:teacher_id/info", handlers.GetTeacherInfo)

		// 获取所有教师ID列表
		api.GET("/teachers/ids", handlers.GetTeacherIds)

		// 批量获取教师信息
		api.POST("/teachers/batch", handlers.BatchGetTeachers)
	}

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}
