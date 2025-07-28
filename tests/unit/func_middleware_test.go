package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/middleware/smart"
	"github.com/stretchr/testify/assert"
)

func TestIdentityAuth(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("无用户ID测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证没有设置TeacherId
			teacherId := smart.GetTeacherIdFromContext(c)
			assert.Equal(t, "", teacherId)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("有用户ID但Redis禁用测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证TeacherId被正确设置
			teacherId := smart.GetTeacherIdFromContext(c)
			assert.Equal(t, "test_teacher", teacherId)

			// 由于Redis禁用，学校和学科组信息不会被设置
			schoolId := smart.GetSchoolIdFromContext(c)
			assert.Equal(t, 0, schoolId)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSmartModelSelection(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("未启用智能模型选择测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择未启用
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.False(t, enabled)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("启用智能模型选择但无用户ID测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择已启用但没有选择模型
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.True(t, enabled)

			selectedModel := smart.GetSelectedModel(c)
			assert.Equal(t, "", selectedModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("启用智能模型选择但无原始模型测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择已启用但没有选择模型
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.True(t, enabled)

			selectedModel := smart.GetSelectedModel(c)
			assert.Equal(t, "", selectedModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher")
		req.Header.Set("X-Smart-Model-Selection", "true")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("完整智能模型选择测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择已启用
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.True(t, enabled)

			// 由于Redis禁用，模型选择应该失败，不会设置选择的模型
			selectedModel := smart.GetSelectedModel(c)
			assert.Equal(t, "", selectedModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher")
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-Original-Model", "gpt-3.5-turbo")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSmartModelSelectionEnabled(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("智能模型选择功能启用测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择功能正常工作
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.True(t, enabled)

			teacherId := smart.GetTeacherIdFromContext(c)
			assert.Equal(t, "test_teacher", teacherId)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher")
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-Original-Model", "gpt-3.5-turbo")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExtendedLogRecorder(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("扩展日志记录测试", func(t *testing.T) {
		// 跳过此测试，因为它需要数据库连接
		t.Skip("扩展日志记录测试需要数据库连接，在集成测试中进行")
	})
}
