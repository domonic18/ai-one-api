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

		// Log key headers individually for easier filtering
		logKeyHeaders(c)

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

// logKeyHeaders logs important headers individually
func logKeyHeaders(c *gin.Context) {
	ctx := c.Request.Context()

	// Log User-Agent
	userAgent := c.GetHeader("User-Agent")
	if userAgent != "" {
		logger.Infof(ctx, "Header-User-Agent: %s", userAgent)
	}

	// Log Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType != "" {
		logger.Infof(ctx, "Header-Content-Type: %s", contentType)
	}

	// Log Accept
	accept := c.GetHeader("Accept")
	if accept != "" {
		logger.Infof(ctx, "Header-Accept: %s", accept)
	}

	// Log X-Forwarded-For (for proxy headers)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		logger.Infof(ctx, "Header-X-Forwarded-For: %s", xForwardedFor)
	}

	// Log X-Real-IP (for proxy headers)
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		logger.Infof(ctx, "Header-X-Real-IP: %s", xRealIP)
	}

	// Log Origin
	origin := c.GetHeader("Origin")
	if origin != "" {
		logger.Infof(ctx, "Header-Origin: %s", origin)
	}

	// Log Referer
	referer := c.GetHeader("Referer")
	if referer != "" {
		logger.Infof(ctx, "Header-Referer: %s", referer)
	}

	// Log Authorization (masked)
	authorization := c.GetHeader("Authorization")
	if authorization != "" {
		maskedAuth := authorization
		if len(maskedAuth) > 20 {
			maskedAuth = maskedAuth[:10] + "..." + maskedAuth[len(maskedAuth)-5:]
		} else if len(maskedAuth) > 8 {
			maskedAuth = maskedAuth[:4] + "****" + maskedAuth[len(maskedAuth)-2:]
		} else {
			maskedAuth = "****"
		}
		logger.Infof(ctx, "Header-Authorization: %s", maskedAuth)
	}
}
