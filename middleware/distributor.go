package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)

		// 获取令牌信息以支持多用户组选择
		var userGroup string
		tokenId := c.GetInt(ctxkey.TokenId)

		// 检查是否已经通过身份解析设置了用户组
		identityGroup, hasIdentityGroup := c.Get(ctxkey.Group)

		if tokenId > 0 {
			// 获取令牌信息
			token, err := model.GetTokenById(tokenId)
			if err != nil {
				logger.Warnf(ctx, "获取令牌信息失败: tokenId=%d, error=%v", tokenId, err)
				// 回退到原有逻辑
				if hasIdentityGroup && identityGroup != "" {
					userGroup = identityGroup.(string)
				} else {
					userGroup, _ = model.CacheGetUserGroup(userId)
				}
			} else {
				// 根据身份解析结果选择合适的用户组
				if hasIdentityGroup && identityGroup != "" {
					resolvedGroup := identityGroup.(string)
					// 使用令牌的SelectGroupByUserGroup方法选择最合适的组
					userGroup = token.SelectGroupByUserGroup(resolvedGroup)
					logger.Debugf(ctx, "多用户组选择: 令牌ID=%d, 解析组=%s, 选择组=%s, 令牌组列表=%v",
						tokenId, resolvedGroup, userGroup, token.GetUserGroups())
				} else {
					// 没有身份解析结果，使用令牌的主要组
					userGroup = token.GetPrimaryGroup()
					logger.Debugf(ctx, "使用令牌主要组: 令牌ID=%d, 主要组=%s, 令牌组列表=%v",
						tokenId, userGroup, token.GetUserGroups())
				}
			}
		} else {
			// 没有令牌信息，使用原有逻辑
			if hasIdentityGroup && identityGroup != "" {
				userGroup = identityGroup.(string)
				logger.Debugf(ctx, "使用身份解析组: %s (用户ID: %d)", userGroup, userId)
			} else {
				userGroup, _ = model.CacheGetUserGroup(userId)
				logger.Debugf(ctx, "使用用户默认组: %s (用户ID: %d)", userGroup, userId)
			}
		}

		c.Set(ctxkey.Group, userGroup)
		var requestModel string
		var channel *model.Channel
		channelId, ok := c.Get(ctxkey.SpecificChannelId)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != model.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			requestModel = c.GetString(ctxkey.RequestModel)
			var err error
			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, requestModel, false)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
				if channel != nil {
					logger.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}
		logger.Debugf(ctx, "user id %d, user group: %s, request model: %s, using channel #%d", userId, userGroup, requestModel, channel.Id)
		SetupContextForSelectedChannel(c, channel, requestModel)
		c.Next()
	}
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	if channel.SystemPrompt != nil && *channel.SystemPrompt != "" {
		c.Set(ctxkey.SystemPrompt, *channel.SystemPrompt)
	}
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Set(ctxkey.OriginalModel, modelName) // for retry
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	// this is for backward compatibility
	if channel.Other != nil {
		switch channel.Type {
		case channeltype.Azure:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Xunfei:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Gemini:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.AIProxyLibrary:
			if cfg.LibraryID == "" {
				cfg.LibraryID = *channel.Other
			}
		case channeltype.Ali:
			if cfg.Plugin == "" {
				cfg.Plugin = *channel.Other
			}
		}
	}
	c.Set(ctxkey.Config, cfg)
}
