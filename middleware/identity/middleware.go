package identity

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/constant"
)

// Identity 身份解析中间件
// 根据实现方案v3.0版本设计，集成身份解析器
func Identity() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		externalUserId := c.GetHeader(constant.UserIdHeader)
		smartModelSelection := c.GetHeader(constant.SmartModelSelectionHeader)

		// 记录请求开始
		logger.Debugf(ctx, "身份解析中间件开始处理: externalUserId=%s, smartModelSelection=%s", externalUserId, smartModelSelection)

		// 如果有X-User-ID，将其设置到context中（无论是否启用课件平台集成）
		if externalUserId != "" {
			c.Set(ctxkey.ExternalUserId, externalUserId)
		}

		// 如果有X-User-ID且课件平台集成已启用，则使用身份解析器
		if externalUserId != "" && isCoursewareEnabled() {
			logger.Debugf(ctx, "启用课件平台身份解析: externalUserId=%s", externalUserId)
			resolver := GetIdentityResolver()

			// 新增：使用身份解析器获取用户组 (核心修改)
			userGroup := resolver.ResolveGroup(ctx, externalUserId)

			// 设置原始模型到上下文（无论用户组是否有效）
			originalModel := c.GetString(ctxkey.RequestModel)
			if originalModel == "" {
				originalModel = getModelFromRequest(c)
				if originalModel != "" {
					logger.Debugf(ctx, "从请求体获取模型: model=%s", originalModel)
				}
			}
			if originalModel != "" {
				c.Set(ctxkey.RequestModel, originalModel)
			}

			// 只有当解析器返回有效用户组时才设置用户组
			if userGroup != "" {
				c.Set(ctxkey.Group, userGroup)
				logger.Infof(ctx, "身份解析成功: externalUserId=%s, group=%s, originalModel=%s", externalUserId, userGroup, originalModel)

				// 新增：如果启用智能模型选择，解析用户偏好模型
				if isSmartModelSelectionEnabled(smartModelSelection) {
					logger.Debugf(ctx, "启用智能模型选择: externalUserId=%s, originalModel=%s", externalUserId, originalModel)
					preferredModel := resolver.ResolveModel(ctx, externalUserId, originalModel)
					if preferredModel != originalModel {
						c.Set(ctxkey.RequestModel, preferredModel)
						logger.Infof(ctx, "智能模型选择: 用户=%s, 原始模型=%s, 替换模型=%s", externalUserId, originalModel, preferredModel)
					} else {
						logger.Debugf(ctx, "智能模型选择: 用户=%s, 模型=%s (无需替换)", externalUserId, originalModel)
					}
				} else {
					logger.Debugf(ctx, "智能模型选择: 用户=%s, 未启用 (header=%s)", externalUserId, smartModelSelection)
				}
			} else {
				logger.Warnf(ctx, "身份解析失败: externalUserId=%s, 未找到有效用户组", externalUserId)
			}
		} else {
			// 如果没有X-User-ID或课件平台集成未启用，尝试设置原始模型到上下文
			if externalUserId == "" {
				logger.Debugf(ctx, "未提供X-User-ID，跳过身份解析")
			} else {
				logger.Debugf(ctx, "课件平台集成未启用，跳过身份解析: externalUserId=%s", externalUserId)
			}

			originalModel := c.GetString(ctxkey.RequestModel)
			if originalModel == "" {
				originalModel = getModelFromRequest(c)
			}
			if originalModel != "" {
				c.Set(ctxkey.RequestModel, originalModel)
			}
		}

		// 如果没有X-User-ID或课件平台集成未启用，完全保持OneAPI原有逻辑

		c.Next()
	}
}

// isCoursewareEnabled 检查课件平台是否启用
func isCoursewareEnabled() bool {
	resolver := GetIdentityResolver()
	// 检查是否使用DefaultIdentityResolver
	_, isDefault := resolver.(*DefaultIdentityResolver)
	enabled := !isDefault
	if enabled {
		logger.SysLog("课件平台集成已启用")
	} else {
		logger.SysLog("课件平台集成未启用，使用默认身份解析器")
	}
	return enabled
}

// isSmartModelSelectionEnabled 检查是否启用智能模型选择
// 支持大小写不敏感的处理
func isSmartModelSelectionEnabled(headerValue string) bool {
	if headerValue == "" {
		return false
	}
	normalizedValue := strings.ToLower(strings.TrimSpace(headerValue))
	return normalizedValue == strings.ToLower(constant.SmartModelSelectionEnabled)
}

// getModelFromRequest 从请求中获取模型名称
func getModelFromRequest(c *gin.Context) string {
	// 保存原始请求体以便后续中间件使用
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.SysErrorf("读取请求体失败: %v", err)
		return ""
	}

	// 重新设置请求体
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// 解析JSON获取模型名称
	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		logger.SysErrorf("解析请求体JSON失败: %v", err)
		return ""
	}

	if model, ok := reqBody["model"].(string); ok {
		return model
	}

	return ""
}
