package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
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
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证设置为默认分组
			group := c.GetString(ctxkey.Group)
			assert.Equal(t, "default", group)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("有用户ID但Redis禁用测试", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证设置了老师ID
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			// 由于Redis禁用，最终会设置为默认分组
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
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证没有设置智能模型选择相关字段
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.False(t, smartSelection)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("启用但无用户ID测试", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证没有设置智能模型选择标志（因为没有用户ID）
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.False(t, smartSelection)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("启用但无请求模型测试", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证没有设置智能模型选择标志（因为没有请求模型）
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.False(t, smartSelection)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-User-ID", "teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("完整参数但Redis禁用测试", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证设置了智能模型选择标志
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.True(t, smartSelection)

			// 验证设置了老师ID
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			// 验证设置了原始模型
			originalModel := c.GetString(ctxkey.OriginalModel)
			assert.Equal(t, "gpt-3.5-turbo", originalModel)

			// 验证请求模型没有改变（因为Redis禁用，无法获取推荐模型）
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

func TestSmartModelSelectionEnabled(t *testing.T) {
	// 由于isSmartModelSelectionEnabled是私有函数，我们通过中间件的整体行为来测试
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 保存原始Redis状态
	originalRedisEnabled := common.RedisEnabled
	defer func() {
		common.RedisEnabled = originalRedisEnabled
	}()

	// 禁用Redis，确保单元测试不依赖外部环境
	common.RedisEnabled = false

	t.Run("启用智能模型选择测试", func(t *testing.T) {
		testCases := []struct {
			headerValue string
			expected    bool
		}{
			{"true", true},
			{"True", true},
			{"TRUE", true},
			{"1", true},
			{"yes", true},
			{"Yes", true},
			{"y", true},
			{"Y", true},
			{"false", false},
			{"0", false},
			{"no", false},
			{"", false},
		}

		for _, tc := range testCases {
			t.Run("Header值: "+tc.headerValue, func(t *testing.T) {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					// 设置请求模型和用户ID，确保其他条件满足
					c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
					c.Next()
				})
				router.Use(middleware.SmartModelSelection())

				router.GET("/test", func(c *gin.Context) {
					smartSelection := c.GetBool(ctxkey.SmartModelSelection)
					assert.Equal(t, tc.expected, smartSelection)
					c.JSON(http.StatusOK, gin.H{"status": "ok"})
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				if tc.headerValue != "" {
					req.Header.Set("X-Smart-Model-Selection", tc.headerValue)
				}
				req.Header.Set("X-User-ID", "teacher_001")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})
}

func TestExtendedLogRecorder(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	t.Run("扩展日志记录中间件基本测试", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.ExtendedLogRecorder())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
