package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_IdentityAuthIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("身份认证中间件集成测试_无用户ID", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证默认值设置
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "", teacherId)

			group := c.GetString(ctxkey.Group)
			assert.Equal(t, "default", group)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("身份认证中间件集成测试_有用户ID", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证用户ID被正确设置
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			// 由于没有实际的API实现，应该使用默认值
			group := c.GetString(ctxkey.Group)
			assert.Equal(t, "default", group)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMiddleware_SmartModelSelectionIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("智能模型选择集成测试_未启用", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择未启用
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.False(t, smartSelection)

			// 原始模型应该保持不变
			requestModel := c.GetString(ctxkey.RequestModel)
			assert.Equal(t, "gpt-3.5-turbo", requestModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("智能模型选择集成测试_启用但无用户ID", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择未启用（因为没有用户ID）
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.False(t, smartSelection)

			// 原始模型应该保持不变
			requestModel := c.GetString(ctxkey.RequestModel)
			assert.Equal(t, "gpt-3.5-turbo", requestModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("智能模型选择集成测试_完整参数", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择启用
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.True(t, smartSelection)

			// 验证用户ID被设置
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			// 由于没有实际的用户配置，模型应该保持原样
			requestModel := c.GetString(ctxkey.RequestModel)
			assert.Equal(t, "gpt-3.5-turbo", requestModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-User-ID", "teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMiddleware_ExtendedLogRecorderIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("扩展日志记录中间件集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置扩展信息到上下文
			c.Set(ctxkey.SchoolId, 1)
			c.Set(ctxkey.SchoolName, "集成测试学校")
			c.Set(ctxkey.SubjectId, 101)
			c.Set(ctxkey.SubjectName, "数学组")
			c.Set(ctxkey.TeacherId, "teacher_001")
			c.Set(ctxkey.TeacherName, "张老师")
			c.Set(ctxkey.Group, "数学组")
			c.Next()
		})
		router.Use(middleware.ExtendedLogRecorder())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// 模拟设置请求ID
		req.Header.Set("X-Request-ID", "integration-test-001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证响应
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})
}
