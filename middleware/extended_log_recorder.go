package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// ExtendedLogRecorder 扩展日志记录中间件
// 用于记录学校、学科组和老师信息
func ExtendedLogRecorder() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续中间件和处理器
		c.Next()

		// 检查是否需要记录扩展日志
		if !shouldRecordExtendedLog(c) {
			return
		}

		// 获取请求ID，用于查询日志记录
		requestId := c.GetString(helper.RequestIdKey)
		if requestId == "" {
			logger.Warnf(c, "未找到请求ID，无法记录扩展日志")
			return
		}

		// 构建扩展日志信息
		extendedInfo := &model.ExtendedLogInfo{
			SchoolId:    getIntFromContext(c, ctxkey.SchoolId),
			SchoolName:  getStringFromContext(c, ctxkey.SchoolName),
			SubjectId:   getIntFromContext(c, ctxkey.SubjectId),
			SubjectName: getStringFromContext(c, ctxkey.SubjectName),
			TeacherId:   getStringFromContext(c, ctxkey.TeacherId),
			TeacherName: getStringFromContext(c, ctxkey.TeacherName),
			GroupName:   getStringFromContext(c, ctxkey.Group),
		}

		// 异步记录扩展日志
		go func(requestId string, info *model.ExtendedLogInfo) {
			ctx := context.Background()

			// 通过请求ID查询日志记录
			var log model.Log
			err := model.LOG_DB.Where("request_id = ?", requestId).First(&log).Error
			if err != nil {
				logger.Errorf(ctx, "查询日志记录失败: requestId=%s, error=%v", requestId, err)
				return
			}

			// 创建扩展日志
			err = model.CreateExtendedLog(ctx, log.Id, info)
			if err != nil {
				logger.Errorf(ctx, "创建扩展日志失败: logId=%d, error=%v", log.Id, err)
			} else {
				logger.Debugf(ctx, "扩展日志创建成功: logId=%d, teacherId=%s, subjectId=%d",
					log.Id, info.TeacherId, info.SubjectId)
			}
		}(requestId, extendedInfo)
	}
}

// shouldRecordExtendedLog 判断是否需要记录扩展日志
func shouldRecordExtendedLog(c *gin.Context) bool {
	// 只有在有老师ID或学科组ID时才记录扩展日志
	teacherId := getStringFromContext(c, ctxkey.TeacherId)
	subjectId := getIntFromContext(c, ctxkey.SubjectId)
	schoolId := getIntFromContext(c, ctxkey.SchoolId)

	return teacherId != "" || subjectId > 0 || schoolId > 0
}

// getStringFromContext 从上下文中获取字符串值
func getStringFromContext(c *gin.Context, key string) string {
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

// getIntFromContext 从上下文中获取整数值
func getIntFromContext(c *gin.Context, key string) int {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}
	intVal, ok := value.(int)
	if !ok {
		return 0
	}
	return intVal
}
