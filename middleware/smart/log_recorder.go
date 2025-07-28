package smart

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	smartModel "github.com/songquanpeng/one-api/model/smart"
)

// LogRecorder 扩展日志记录中间件
// 记录请求的扩展信息（学校、学科组等）
func LogRecorder() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续中间件和处理函数
		c.Next()

		// 获取日志ID
		logIdValue, exists := c.Get(ctxkey.LogId)
		if !exists {
			// 没有日志ID，不记录扩展日志
			return
		}

		// 转换为int64类型
		logId, ok := logIdValue.(int64)
		if !ok || logId == 0 {
			logger.Warnf(c, "无效的日志ID: %v", logIdValue)
			return
		}

		// 获取老师ID
		teacherId := GetTeacherIdFromContext(c)
		if teacherId == "" {
			// 没有老师ID，不记录扩展日志
			return
		}

		// 获取老师信息
		ctx := c.Request.Context()
		teacherInfo, err := smartModel.GetTeacherInfoWithCache(ctx, teacherId)
		if err != nil {
			logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
			// 即使没有老师信息，也记录基本的扩展日志
		}

		// 创建扩展日志
		_, err = smartModel.CreateExtendedLog(ctx, logId, teacherId, teacherInfo)
		if err != nil {
			logger.Errorf(ctx, "创建扩展日志失败: logId=%d, teacherId=%s, error=%v",
				logId, teacherId, err)
		} else {
			logger.Debugf(ctx, "创建扩展日志成功: logId=%d, teacherId=%s",
				logId, teacherId)
		}
	}
}
