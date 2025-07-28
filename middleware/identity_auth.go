package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// IdentityAuth 身份识别中间件
// 根据请求头中的X-User-ID字段，识别老师所属的学校和学科组
func IdentityAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取老师ID
		teacherId := c.GetHeader("X-User-ID")
		if teacherId == "" {
			// 如果没有提供老师ID，使用默认分组
			c.Set(ctxkey.Group, "default")
			c.Next()
			return
		}

		// 2. 设置老师ID到上下文
		c.Set(ctxkey.TeacherId, teacherId)

		// 3. 通过老师ID查找所属学科组和学校
		teacherInfo, err := model.GetTeacherInfoWithCache(context.Background(), teacherId)
		if err != nil || teacherInfo == nil {
			logger.Warnf(c, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
			// 如果获取失败或返回nil，使用默认分组
			c.Set(ctxkey.Group, "default")
			c.Next()
			return
		}

		// 4. 设置学校和学科组信息到上下文
		c.Set(ctxkey.SchoolId, teacherInfo.SchoolId)
		c.Set(ctxkey.SchoolName, teacherInfo.SchoolName)
		c.Set(ctxkey.SubjectId, teacherInfo.SubjectId)
		c.Set(ctxkey.SubjectName, teacherInfo.SubjectName)
		c.Set(ctxkey.TeacherName, teacherInfo.TeacherName)

		// 5. 设置用户组
		// 使用学科组名称作为用户组，方便按学科组选择渠道
		if teacherInfo.GroupName != "" {
			c.Set(ctxkey.Group, teacherInfo.GroupName)
		} else {
			// 如果没有学科组名称，使用默认分组
			c.Set(ctxkey.Group, "default")
		}

		c.Next()
	}
}
