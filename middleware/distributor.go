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

		var token *model.Token
		if tokenId > 0 {
			// 获取令牌信息
			t, err := model.GetTokenById(tokenId)
			if err != nil {
				logger.Warnf(ctx, "获取令牌信息失败: tokenId=%d, error=%v", tokenId, err)
				// 回退到原有逻辑
				if hasIdentityGroup && identityGroup != "" {
					userGroup = identityGroup.(string)
				} else {
					userGroup, _ = model.CacheGetUserGroup(userId)
				}
			} else {
				token = t
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

		// 构造候选用户组列表：
		// - 若存在令牌：按 token.user_groups 顺序尝试；如果有身份解析结果，则将匹配到的组置于首位，其余按原顺序排在后面
		// - 若不存在令牌：仅尝试当前 userGroup
		var groupsToTry []string
		if token != nil {
			all := token.GetUserGroups()
			if hasIdentityGroup && identityGroup != "" {
				resolvedGroup := identityGroup.(string)
				primary := token.SelectGroupByUserGroup(resolvedGroup)
				// 去重并保序：primary 优先，然后追加其余
				seen := map[string]bool{}
				if primary != "" {
					groupsToTry = append(groupsToTry, primary)
					seen[primary] = true
				}
				for _, g := range all {
					if !seen[g] {
						groupsToTry = append(groupsToTry, g)
						seen[g] = true
					}
				}
			} else {
				groupsToTry = append(groupsToTry, all...)
			}
		} else {
			groupsToTry = []string{userGroup}
		}

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
			// 按候选组顺序尝试获取可用渠道
			for _, grp := range groupsToTry {
				channel, err = model.CacheGetRandomSatisfiedChannel(grp, requestModel, false)
				if err == nil && channel != nil {
					userGroup = grp
					break
				}
				logger.Debugf(ctx, "分组回退: 组=%s 对于模型 %s 无可用渠道，继续尝试下一组", grp, requestModel)
			}
			if channel == nil {
				// 所有分组均无可用渠道
				message := fmt.Sprintf("当前令牌可用分组 %v 下对于模型 %s 均无可用渠道", groupsToTry, requestModel)
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}

		// 记录最终选择的组
		c.Set(ctxkey.Group, userGroup)
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
