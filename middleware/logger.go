package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"strings"
)

func SetUpLogger(server *gin.Engine) {
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var requestID string
		if param.Keys != nil {
			requestID = param.Keys[helper.RequestIdKey].(string)
		}

		// Log all request headers
		if param.Request != nil {
			headers := param.Request.Header
			headerLog := "[HEADERS] "
			for key, values := range headers {
				headerLog += fmt.Sprintf("%s: %s | ", key, strings.Join(values, ", "))
			}
			logger.SysLog(headerLog)
		}

		return fmt.Sprintf("[GIN] %s | %s | %3d | %13v | %15s | %7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))
}
