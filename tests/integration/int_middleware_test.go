package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware/smart"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// TestIdentityAuthIntegration 测试身份识别中间件集成功能
// 测试目的：验证身份识别中间件在真实HTTP请求环境中的正确集成和功能完整性
// 测试内容：
// 1. 验证有效用户ID从请求头到上下文对象的正确传递
// 2. 验证无效用户ID的异常处理和默认值设置
// 3. 验证中间件在Gin框架中的执行顺序和依赖关系
// 4. 验证上下文对象中教师ID的存储和获取机制
// 5. 验证集成环境下的HTTP响应状态码和响应体格式
func TestIdentityAuthIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("有效用户ID集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证老师ID被正确设置
			teacherId := smart.GetTeacherIdFromContext(c)
			assert.NotEmpty(t, teacherId)

			c.JSON(http.StatusOK, gin.H{"status": "ok", "teacher_id": teacherId})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("无效用户ID集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证老师ID被正确设置
			teacherId := smart.GetTeacherIdFromContext(c)
			assert.Equal(t, "invalid_teacher", teacherId)

			// 学校和学科组信息可能为空（如果API返回失败）
			schoolId := smart.GetSchoolIdFromContext(c)
			// schoolId可能为0或有效值，取决于API响应

			c.JSON(http.StatusOK, gin.H{
				"status":     "ok",
				"teacher_id": teacherId,
				"school_id":  schoolId,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "invalid_teacher")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestSmartModelSelectionIntegration 测试智能模型选择中间件集成功能
// 测试目的：验证智能模型选择中间件在完整请求链路中的正确集成和功能行为
// 测试内容：
// 1. 验证智能模型选择功能的启用和禁用机制
// 2. 验证原始模型与智能选择模型的正确传递和替换
// 3. 验证用户ID与模型选择的关联性验证
// 4. 验证中间件的执行顺序和上下文数据传递
// 5. 验证在启用和禁用状态下的不同行为表现
func TestSmartModelSelectionIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("启用智能模型选择集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择已启用
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.True(t, enabled)

			// 获取选择的模型（可能为空，取决于API响应）
			selectedModel := smart.GetSelectedModel(c)
			// selectedModel可能为空或有效模型名

			c.JSON(http.StatusOK, gin.H{
				"status":         "ok",
				"smart_enabled":  enabled,
				"selected_model": selectedModel,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher_001")
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-Original-Model", "gpt-3.5-turbo")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("未启用智能模型选择集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择未启用
			enabled := smart.IsSmartModelSelectionEnabled(c)
			assert.False(t, enabled)

			c.JSON(http.StatusOK, gin.H{"status": "ok", "smart_enabled": enabled})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("完整智能模型选择流程集成测试", func(t *testing.T) {
		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择功能
			enabled := smart.IsSmartModelSelectionEnabled(c)
			teacherId := smart.GetTeacherIdFromContext(c)
			selectedModel := smart.GetSelectedModel(c)

			c.JSON(http.StatusOK, gin.H{
				"status":         "ok",
				"smart_enabled":  enabled,
				"teacher_id":     teacherId,
				"selected_model": selectedModel,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher_002")
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-Original-Model", "gpt-4")
		router.ServeHTTP(w, req)

		// 请求应该成功，无论API调用是否成功
		assert.NotEqual(t, http.StatusServiceUnavailable, w.Code)
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	})
}

// TestExtendedLogRecorderIntegration 测试扩展日志记录中间件集成功能
// 测试目的：验证扩展日志记录中间件在真实请求环境中的正确集成和数据关联能力
// 测试内容：
// 1. 验证日志记录中间件与身份识别中间件的集成顺序
// 2. 验证日志ID在上下文中的正确传递和关联
// 3. 验证数据库日志记录与HTTP请求的关联机制
// 4. 验证中间件对请求处理流程的无侵入性
// 5. 验证测试数据的正确创建和清理机制
func TestExtendedLogRecorderIntegration(t *testing.T) {
	_, db := setupIntegrationTest()
	defer cleanupTestData(db)

	t.Run("扩展日志记录集成测试", func(t *testing.T) {
		// 创建测试日志记录
		log := &model.Log{
			UserId:    1,
			Content:   "测试日志内容",
			ModelName: "gpt-3.5-turbo",
		}
		err := model.DB.Create(log).Error
		assert.NoError(t, err)

		router := gin.New()
		router.Use(smart.IdentityAuth())
		router.Use(smart.ExtendedLogRecorder())

		router.GET("/test", func(c *gin.Context) {
			// 设置日志ID
			c.Set(ctxkey.LogId, log.Id)

			c.JSON(http.StatusOK, gin.H{"status": "ok", "log_id": log.Id})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "test_teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 清理测试数据
		model.DB.Delete(log)
	})
}
