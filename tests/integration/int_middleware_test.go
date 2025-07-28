package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_IdentityAuthIntegration(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 启用Redis用于测试
	common.RedisEnabled = true

	// 保存原始函数
	originalGetTeacherInfoFromAPI := model.GetTeacherInfoFromAPI
	defer func() {
		model.GetTeacherInfoFromAPI = originalGetTeacherInfoFromAPI
	}()

	t.Run("身份认证中间件与API集成测试", func(t *testing.T) {
		// Mock API函数返回成功
		model.GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*model.TeacherInfo, error) {
			return &model.TeacherInfo{
				TeacherId:   teacherId,
				TeacherName: "集成测试老师",
				SchoolId:    1,
				SchoolName:  "集成测试学校",
				SubjectId:   101,
				SubjectName: "数学",
				GroupName:   "数学组",
			}, nil
		}

		router := gin.New()
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			// 验证设置了正确的信息
			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			schoolId := c.GetInt(ctxkey.SchoolId)
			assert.Equal(t, 1, schoolId)

			schoolName := c.GetString(ctxkey.SchoolName)
			assert.Equal(t, "集成测试学校", schoolName)

			subjectId := c.GetInt(ctxkey.SubjectId)
			assert.Equal(t, 101, subjectId)

			subjectName := c.GetString(ctxkey.SubjectName)
			assert.Equal(t, "数学", subjectName)

			teacherName := c.GetString(ctxkey.TeacherName)
			assert.Equal(t, "集成测试老师", teacherName)

			group := c.GetString(ctxkey.Group)
			assert.Equal(t, "数学组", group)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("身份认证中间件缓存测试", func(t *testing.T) {
		// 第一次调用，数据会被缓存
		model.GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*model.TeacherInfo, error) {
			return &model.TeacherInfo{
				TeacherId:   teacherId,
				TeacherName: "缓存测试老师",
				SchoolId:    2,
				SchoolName:  "缓存测试学校",
				SubjectId:   102,
				SubjectName: "英语",
				GroupName:   "英语组",
			}, nil
		}

		router := gin.New()
		router.Use(middleware.IdentityAuth())

		router.GET("/test", func(c *gin.Context) {
			teacherName := c.GetString(ctxkey.TeacherName)
			assert.Equal(t, "缓存测试老师", teacherName)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// 第一次请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "teacher_002")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 修改API函数，但由于缓存，应该返回缓存的数据
		model.GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*model.TeacherInfo, error) {
			return &model.TeacherInfo{
				TeacherId:   teacherId,
				TeacherName: "新的老师名称",
				SchoolId:    3,
				SchoolName:  "新的学校",
				SubjectId:   103,
				SubjectName: "新的学科",
				GroupName:   "新的组",
			}, nil
		}

		router2 := gin.New()
		router2.Use(middleware.IdentityAuth())

		router2.GET("/test", func(c *gin.Context) {
			// 应该还是返回缓存的数据
			teacherName := c.GetString(ctxkey.TeacherName)
			assert.Equal(t, "缓存测试老师", teacherName)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// 第二次请求，使用相同的teacher_002
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-User-ID", "teacher_002")
		router2.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

func TestMiddleware_SmartModelSelectionIntegration(t *testing.T) {
	r, db := setupIntegrationTest()
	defer cleanupTestData(db)

	// 启用Redis用于测试
	common.RedisEnabled = true

	// 保存原始函数
	originalGetUserModelConfigFromAPI := model.GetUserModelConfigFromAPI
	originalGetTeacherInfoFromAPI := model.GetTeacherInfoFromAPI
	originalGetSubjectInfoFromAPI := model.GetSubjectInfoFromAPI
	defer func() {
		model.GetUserModelConfigFromAPI = originalGetUserModelConfigFromAPI
		model.GetTeacherInfoFromAPI = originalGetTeacherInfoFromAPI
		model.GetSubjectInfoFromAPI = originalGetSubjectInfoFromAPI
	}()

	t.Run("智能模型选择完整流程测试", func(t *testing.T) {
		// Mock用户模型配置API
		model.GetUserModelConfigFromAPI = func(ctx context.Context, userId string) (*model.UserModelConfig, error) {
			return &model.UserModelConfig{
				UserId:    userId,
				ModelName: "gpt-4-turbo",
				UpdatedAt: 1640995200, // 2022-01-01
				Source:    "user_preference",
			}, nil
		}

		// Mock老师信息API
		model.GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*model.TeacherInfo, error) {
			return &model.TeacherInfo{
				TeacherId:   teacherId,
				TeacherName: "智能选择测试老师",
				SchoolId:    1,
				SchoolName:  "智能选择测试学校",
				SubjectId:   101,
				SubjectName: "数学",
				GroupName:   "数学组",
			}, nil
		}

		// Mock学科组信息API
		model.GetSubjectInfoFromAPI = func(ctx context.Context, subjectId int) (*model.SubjectInfo, error) {
			return &model.SubjectInfo{
				SubjectId:    subjectId,
				SubjectName:  "数学",
				SchoolId:     1,
				SchoolName:   "智能选择测试学校",
				DefaultModel: "claude-3-sonnet",
				UpdatedAt:    1640995200,
			}, nil
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择生效
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.True(t, smartSelection)

			teacherId := c.GetString(ctxkey.TeacherId)
			assert.Equal(t, "teacher_001", teacherId)

			originalModel := c.GetString(ctxkey.OriginalModel)
			assert.Equal(t, "gpt-3.5-turbo", originalModel)

			// 应该选择用户偏好的模型
			requestModel := c.GetString(ctxkey.RequestModel)
			assert.Equal(t, "gpt-4-turbo", requestModel)

			selectedModel := c.GetString(ctxkey.SelectedModel)
			assert.Equal(t, "gpt-4-turbo", selectedModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-User-ID", "teacher_001")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("智能模型选择降级测试", func(t *testing.T) {
		// Mock用户模型配置API返回空
		model.GetUserModelConfigFromAPI = func(ctx context.Context, userId string) (*model.UserModelConfig, error) {
			return nil, nil // 没有用户偏好
		}

		// Mock老师信息API
		model.GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*model.TeacherInfo, error) {
			return &model.TeacherInfo{
				TeacherId:   teacherId,
				TeacherName: "降级测试老师",
				SchoolId:    1,
				SchoolName:  "降级测试学校",
				SubjectId:   101,
				SubjectName: "数学",
				GroupName:   "数学组",
			}, nil
		}

		// Mock学科组信息API返回默认模型
		model.GetSubjectInfoFromAPI = func(ctx context.Context, subjectId int) (*model.SubjectInfo, error) {
			return &model.SubjectInfo{
				SubjectId:    subjectId,
				SubjectName:  "数学",
				SchoolId:     1,
				SchoolName:   "降级测试学校",
				DefaultModel: "claude-3-sonnet",
				UpdatedAt:    1640995200,
			}, nil
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			// 设置请求模型
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")
			c.Next()
		})
		router.Use(middleware.SmartModelSelection())

		router.GET("/test", func(c *gin.Context) {
			// 验证智能模型选择生效
			smartSelection := c.GetBool(ctxkey.SmartModelSelection)
			assert.True(t, smartSelection)

			// 应该使用学科组默认模型
			requestModel := c.GetString(ctxkey.RequestModel)
			assert.Equal(t, "claude-3-sonnet", requestModel)

			selectedModel := c.GetString(ctxkey.SelectedModel)
			assert.Equal(t, "claude-3-sonnet", selectedModel)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Smart-Model-Selection", "true")
		req.Header.Set("X-User-ID", "teacher_002")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMiddleware_ExtendedLogRecorderIntegration(t *testing.T) {
	r, db := setupIntegrationTest()
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
