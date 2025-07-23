package middleware

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

// SmartModelSelection 智能模型选择中间件
// 通过HTTP头 X-Smart-Model-Selection 控制是否启用智能选择
// 如果启用且Redis中有用户配置，则使用配置的模型；否则使用请求中的原始模型作为兜底
func SmartModelSelection() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 检查是否启用了Redis
		if !common.RedisEnabled {
			c.Next()
			return
		}

		// 检查是否启用智能模型选择
		smartSelectionHeader := c.GetHeader(constant.SmartModelSelectionHeader)
		if !isSmartSelectionEnabled(smartSelectionHeader) {
			// 如果未启用智能选择，直接处理下一个中间件
			c.Next()
			return
		}

		// 获取请求体
		var requestBody map[string]interface{}
		bodyBytes, err := common.GetRequestBody(c)
		if err != nil {
			c.Next()
			return
		}

		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			c.Next()
			return
		}

		// 获取原始模型名称作为兜底
		originalModel, ok := requestBody["model"].(string)
		if !ok {
			logger.Warnf(c.Request.Context(), "智能模型选择: 请求体中未找到model字段")
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		ctx := c.Request.Context()
		logger.Debugf(ctx, "检测到智能模型选择请求，原始模型: %s", originalModel)

		// 从请求头获取用户ID
		userID := c.GetHeader(constant.UserIdHeader)
		if userID == "" {
			logger.Warnf(ctx, "使用智能模型选择功能时未提供用户ID，使用原始模型: %s", originalModel)
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 创建选择器
		selector := model_selection.NewSmartModelSelector()

		// 获取用户配置的模型信息
		config, err := selector.GetUserModelConfig(userID)
		if err != nil {
			logger.Infof(ctx, "未找到用户模型配置或获取失败，使用原始模型 %s 作为兜底: %v", originalModel, err)
			// 使用原始模型作为兜底，不需要修改请求体
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 替换模型名称为配置的模型
		requestBody["model"] = config.ModelName

		// 合并模型参数
		if len(config.Parameters) > 0 {
			mergeModelParameters(requestBody, config.Parameters)
		}

		logger.Infof(ctx, "智能模型选择: 将模型 %s 替换为 %s", originalModel, config.ModelName)

		// 更新请求体
		newBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			logger.Errorf(ctx, "序列化更新后的请求体失败，使用原始模型: %v", err)
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 更新上下文中的RequestModel值
		c.Set(ctxkey.RequestModel, config.ModelName)

		// 更新缓存的请求体内容
		c.Set(ctxkey.KeyRequestBody, newBodyBytes)

		c.Request.Body = common.GetRequestBodyReader(newBodyBytes)
		c.Next()
	}
}

// isSmartSelectionEnabled 检查是否启用智能模型选择
// 支持多种格式: true, 1, on, enabled, yes (不区分大小写)
func isSmartSelectionEnabled(headerValue string) bool {
	if headerValue == "" {
		return false
	}

	value := strings.ToLower(strings.TrimSpace(headerValue))
	return value == "true" || value == "1" || value == "on" || value == "enabled" || value == "yes"
}

// mergeModelParameters 合并模型参数
func mergeModelParameters(requestBody map[string]interface{}, configParams map[string]interface{}) {
	// 检查请求体中是否已有参数
	if existingParams, exists := requestBody["parameters"]; exists {
		if params, ok := existingParams.(map[string]interface{}); ok {
			// 合并参数，配置的参数优先级更高
			for k, v := range configParams {
				params[k] = v
			}
			requestBody["parameters"] = params
		} else {
			// 如果现有参数格式不正确，直接使用配置参数
			requestBody["parameters"] = configParams
		}
	} else {
		// 如果请求中没有parameters字段，直接使用配置参数
		requestBody["parameters"] = configParams
	}
}
