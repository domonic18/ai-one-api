package smart

import (
	"github.com/gin-gonic/gin"
)

// IdentityAuth 身份识别中间件（向后兼容）
func IdentityAuth() gin.HandlerFunc {
	return Identity()
}

// SmartModelSelection 智能模型选择中间件（向后兼容）
func SmartModelSelection() gin.HandlerFunc {
	return ModelSelection()
}

// ExtendedLogRecorder 扩展日志记录中间件（向后兼容）
func ExtendedLogRecorder() gin.HandlerFunc {
	return LogRecorder()
}
