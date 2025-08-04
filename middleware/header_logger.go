package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
)

// HeaderLogger middleware to log all request headers
func HeaderLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log all request headers in a structured format
		headers := c.Request.Header
		if len(headers) > 0 {
			headerLog := "[REQUEST_HEADERS] "
			for key, values := range headers {
				// Mask sensitive headers
				maskedValues := values
				if isSensitiveHeader(key) {
					maskedValues = maskSensitiveValues(values)
				}
				headerLog += fmt.Sprintf("%s: [%s] | ", key, strings.Join(maskedValues, ", "))
			}
			logger.SysLog(headerLog)
		}

		c.Next()
	}
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(key string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"x-access-token",
		"x-csrf-token",
		"set-cookie",
		"proxy-authorization",
		"www-authenticate",
	}

	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveHeaders {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
}

// maskSensitiveValues masks sensitive values in headers
func maskSensitiveValues(values []string) []string {
	masked := make([]string, len(values))
	for i, value := range values {
		if len(value) > 8 {
			// Show first 4 and last 2 characters
			masked[i] = value[:4] + "****" + value[len(value)-2:]
		} else if len(value) > 4 {
			// Show first 2 and last 1 characters
			masked[i] = value[:2] + "***" + value[len(value)-1:]
		} else {
			// Show only first character
			masked[i] = string(value[0]) + "***"
		}
	}
	return masked
}
