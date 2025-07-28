package smart

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model/smart"
)

// Identity 身份识别中间件
// 从请求头中获取用户ID，并关联学校和学科组信息
func Identity() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取用户ID
		teacherId := c.GetHeader("X-User-ID")
		if teacherId == "" {
			// 没有用户ID，跳过身份识别
			c.Next()
			return
		}

		// 将用户ID存入上下文
		c.Set(ctxkey.TeacherId, teacherId)

		// 获取老师信息
		ctx := c.Request.Context()
		teacherInfo, err := smart.GetTeacherInfoWithCache(ctx, teacherId)
		if err != nil || teacherInfo == nil {
			logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
			// 获取失败也继续处理请求，但不设置学校和学科组信息
			c.Next()
			return
		}

		// 将学校和学科组信息存入上下文
		c.Set(ctxkey.SchoolId, teacherInfo.SchoolId)
		c.Set(ctxkey.SchoolName, teacherInfo.SchoolName)
		c.Set(ctxkey.SubjectId, teacherInfo.SubjectId)
		c.Set(ctxkey.SubjectName, teacherInfo.SubjectName)

		logger.Debugf(ctx, "身份识别成功: teacherId=%s, schoolId=%d, subjectId=%d",
			teacherId, teacherInfo.SchoolId, teacherInfo.SubjectId)

		c.Next()
	}
}

// GetTeacherIdFromContext 从上下文中获取老师ID
func GetTeacherIdFromContext(c *gin.Context) string {
	teacherId, exists := c.Get(ctxkey.TeacherId)
	if !exists {
		return ""
	}
	return teacherId.(string)
}

// GetSchoolIdFromContext 从上下文中获取学校ID
func GetSchoolIdFromContext(c *gin.Context) int {
	schoolId, exists := c.Get(ctxkey.SchoolId)
	if !exists {
		return 0
	}
	return schoolId.(int)
}

// GetSubjectIdFromContext 从上下文中获取学科组ID
func GetSubjectIdFromContext(c *gin.Context) int {
	subjectId, exists := c.Get(ctxkey.SubjectId)
	if !exists {
		return 0
	}
	return subjectId.(int)
}

// RequireIdentity 要求请求必须包含身份信息
func RequireIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherId := GetTeacherIdFromContext(c)
		if teacherId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求必须包含X-User-ID头",
			})
			c.Abort()
			return
		}

		schoolId := GetSchoolIdFromContext(c)
		if schoolId == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无法识别用户所属学校",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
