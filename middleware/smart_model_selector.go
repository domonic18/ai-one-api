package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// SmartModelSelection 智能模型选择中间件
// 根据请求头中的X-Smart-Model-Selection和X-User-ID字段，自动选择合适的模型
func SmartModelSelection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 检查是否启用智能模型选择
		smartModelSelection := c.GetHeader("X-Smart-Model-Selection")
		if !isSmartModelSelectionEnabled(smartModelSelection) {
			// 未启用智能模型选择，继续后续处理
			c.Next()
			return
		}

		// 2. 获取用户ID
		userId := c.GetHeader("X-User-ID")
		if userId == "" {
			logger.Warnf(c, "启用智能模型选择但未提供用户ID")
			c.Next()
			return
		}

		// 3. 获取原始请求模型
		originalModel := c.GetString(ctxkey.RequestModel)
		if originalModel == "" {
			logger.Warnf(c, "未找到请求模型")
			c.Next()
			return
		}

		// 4. 设置上下文信息
		c.Set(ctxkey.TeacherId, userId)
		c.Set(ctxkey.SmartModelSelection, true)
		c.Set(ctxkey.OriginalModel, originalModel)

		// 5. 创建智能模型选择器并选择模型
		selector := model.NewSmartModelSelector(context.Background(), userId, originalModel)
		selectedModel, changed := selector.SelectModel()

		// 6. 如果选择了新模型，替换请求模型
		if changed && selectedModel != "" {
			logger.Infof(c, "智能模型选择: %s -> %s", originalModel, selectedModel)
			c.Set(ctxkey.RequestModel, selectedModel)
			c.Set(ctxkey.SelectedModel, selectedModel)
		}

		c.Next()
	}
}

// isSmartModelSelectionEnabled 检查是否启用智能模型选择
func isSmartModelSelectionEnabled(value string) bool {
	if value == "" {
		return false
	}

	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}
