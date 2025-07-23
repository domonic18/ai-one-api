package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

// SmartModelSelection 智能模型选择中间件
// 如果请求中的模型名称为"smart_select"，则根据请求头中的X-User-ID查询Redis获取用户配置的模型信息
func SmartModelSelection() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 检查是否启用了Redis
		if !common.RedisEnabled {
			c.Next()
			return
		}

		// 判断请求体中的模型是否为smart_select
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

		modelName, ok := requestBody["model"].(string)
		if !ok || modelName != constant.SmartSelect {
			// 如果不是智能选择模型，直接处理下一个中间件
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		ctx := c.Request.Context()
		logger.Debugf(ctx, "检测到智能模型选择请求: %s", constant.SmartSelect)

		// 从请求头获取用户ID
		userID := c.GetHeader(constant.UserIdHeader)
		if userID == "" {
			logger.Warnf(ctx, "使用智能模型选择功能时未提供用户ID")
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 创建选择器
		selector := model_selection.NewSmartModelSelector()

		// 获取用户配置的模型信息
		config, err := selector.GetUserModelConfig(userID)
		if err != nil {
			logger.Warnf(ctx, "获取用户模型配置失败: %v", err)
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 替换模型名称
		requestBody["model"] = config.ModelName

		// 合并模型参数
		if len(config.Parameters) > 0 {
			params, paramsExist := requestBody["parameters"].(map[string]interface{})
			if !paramsExist {
				// 如果请求中没有parameters字段，则直接使用配置中的parameters
				requestBody["parameters"] = config.Parameters
			} else {
				// 如果请求中有parameters字段，则合并配置中的parameters
				for k, v := range config.Parameters {
					params[k] = v
				}
				requestBody["parameters"] = params
			}
		}

		logger.Infof(ctx, "智能模型选择: 将模型 %s 替换为 %s", constant.SmartSelect, config.ModelName)

		// 更新请求体
		newBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			logger.Errorf(ctx, "序列化更新后的请求体失败: %v", err)
			c.Request.Body = common.GetRequestBodyReader(bodyBytes)
			c.Next()
			return
		}

		// 关键修复：同时更新上下文中的RequestModel值
		// 这样Distribute中间件就会使用替换后的模型名称而不是原始的smart_select
		c.Set(ctxkey.RequestModel, config.ModelName)

		c.Request.Body = common.GetRequestBodyReader(newBodyBytes)
		c.Next()
	}
}
