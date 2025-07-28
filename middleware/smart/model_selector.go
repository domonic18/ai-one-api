package smart

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	smartModel "github.com/songquanpeng/one-api/model/smart"
)

// ModelSelection 智能模型选择中间件
// 根据请求头中的X-Smart-Model-Selection字段和用户ID，智能选择最合适的模型
func ModelSelection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用智能模型选择
		smartSelection := c.GetHeader("X-Smart-Model-Selection")
		if smartSelection != "true" {
			// 未启用智能模型选择，继续处理请求
			c.Next()
			return
		}

		// 设置智能模型选择标志
		c.Set(ctxkey.SmartModelSelection, true)

		// 获取用户ID
		teacherId := GetTeacherIdFromContext(c)
		if teacherId == "" {
			// 没有用户ID，无法进行智能选择
			c.Next()
			return
		}

		// 获取原始模型
		originalModel := c.GetHeader("X-Original-Model")
		if originalModel == "" {
			// 没有原始模型，无法进行智能选择
			c.Next()
			return
		}

		// 进行智能模型选择
		ctx := c.Request.Context()
		selectedModel, replaced := smartModel.DefaultSelector.SelectModel(ctx, teacherId, originalModel)
		if replaced {
			// 设置选择的模型到上下文
			c.Set(ctxkey.SelectedModel, selectedModel)
			logger.Debugf(ctx, "智能模型选择: 用户=%s, 原始模型=%s, 选择模型=%s",
				teacherId, originalModel, selectedModel)
		}

		c.Next()
	}
}

// IsSmartModelSelectionEnabled 检查是否启用了智能模型选择
func IsSmartModelSelectionEnabled(c *gin.Context) bool {
	enabled, exists := c.Get(ctxkey.SmartModelSelection)
	if !exists {
		return false
	}
	return enabled.(bool)
}

// GetSelectedModel 获取智能选择的模型
func GetSelectedModel(c *gin.Context) string {
	model, exists := c.Get(ctxkey.SelectedModel)
	if !exists {
		return ""
	}
	return model.(string)
}
