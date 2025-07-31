package identity

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
)

// Identity 身份解析中间件
// 根据实现方案v3.0版本设计，集成身份解析器
func Identity() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		externalUserId := c.GetHeader("X-User-ID")
		smartModelSelection := c.GetHeader("X-Smart-Model-Selection")

		// 如果有X-User-ID且课件平台集成已启用，则使用身份解析器
		if externalUserId != "" && isCoursewareEnabled() {
			resolver := GetIdentityResolver()

			// 新增：使用身份解析器获取用户组 (核心修改)
			userGroup := resolver.ResolveGroup(ctx, externalUserId)

			// 只有当解析器返回有效用户组时才设置
			if userGroup != "" {
				c.Set(ctxkey.Group, userGroup)

				// 新增：如果启用智能模型选择，解析用户偏好模型
				if isSmartModelSelectionEnabled(smartModelSelection) {
					originalModel := c.GetString(ctxkey.RequestModel)
					if originalModel == "" {
						// 尝试从请求体中获取模型
						originalModel = getModelFromRequest(c)
					}

					if originalModel != "" {
						// 先设置原始模型到上下文
						c.Set(ctxkey.RequestModel, originalModel)

						preferredModel := resolver.ResolveModel(ctx, externalUserId, originalModel)
						if preferredModel != originalModel {
							c.Set(ctxkey.RequestModel, preferredModel)
							logger.Infof(ctx, "智能模型选择: 用户=%s, 原始模型=%s, 替换模型=%s", externalUserId, originalModel, preferredModel)
						} else {
							logger.Debugf(ctx, "智能模型选择: 用户=%s, 模型=%s (无需替换)", externalUserId, originalModel)
						}
					} else {
						logger.Warnf(ctx, "智能模型选择: 用户=%s, 无法获取原始模型", externalUserId)
					}
				} else {
					// 当智能模型选择被禁用时，也要设置原始模型到上下文
					originalModel := c.GetString(ctxkey.RequestModel)
					if originalModel == "" {
						originalModel = getModelFromRequest(c)
					}
					if originalModel != "" {
						c.Set(ctxkey.RequestModel, originalModel)
					}
					logger.Debugf(ctx, "智能模型选择: 用户=%s, 未启用 (header=%s)", externalUserId, smartModelSelection)
				}

				logger.Debugf(ctx, "身份解析: externalUserId=%s, group=%s", externalUserId, userGroup)
			}
		} else {
			// 如果没有X-User-ID或课件平台集成未启用，尝试设置原始模型到上下文
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
	return !isDefault
}

// isSmartModelSelectionEnabled 检查是否启用智能模型选择
// 支持大小写不敏感的处理
func isSmartModelSelectionEnabled(headerValue string) bool {
	if headerValue == "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(headerValue)) == "true"
}

// getModelFromRequest 从请求中获取模型名称
func getModelFromRequest(c *gin.Context) string {
	// 保存原始请求体以便后续中间件使用
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}

	// 重新设置请求体
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// 解析JSON获取模型名称
	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err == nil {
		if model, ok := reqBody["model"].(string); ok {
			return model
		}
	}
	return ""
}
