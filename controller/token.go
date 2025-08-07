package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.Query("order")
	tokens, err := model.GetAllUserTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage, order)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	tokens, err := model.SearchUserTokens(userId, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	token, err := model.GetTokenByIds(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func validateToken(c *gin.Context, token model.Token) error {
	if len(token.Name) > 30 {
		return fmt.Errorf("令牌名称过长")
	}
	if token.Subnet != nil && *token.Subnet != "" {
		err := network.IsValidSubnets(*token.Subnet)
		if err != nil {
			return fmt.Errorf("无效的网段：%s", err.Error())
		}
	}
	return nil
}

func AddToken(c *gin.Context) {
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}

	cleanToken := model.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           token.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
		Models:         token.Models,
		Subnet:         token.Subnet,
	}
	err = cleanToken.Insert()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	err := model.DeleteTokenById(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}
	cleanToken, err := model.GetTokenByIds(token.Id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Status == model.TokenStatusEnabled {
		if cleanToken.Status == model.TokenStatusExpired && cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == model.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.Models = token.Models
		cleanToken.Subnet = token.Subnet
	}
	err = cleanToken.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}

// UpdateTokenGroup 修改令牌的用户组
func UpdateTokenGroup(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的令牌ID",
		})
		return
	}

	var req struct {
		Group string `json:"group" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 验证用户组是否存在
	if req.Group != "default" {
		// 检查用户组是否存在
		exists := false
		for groupName := range billingratio.GroupRatio {
			if groupName == req.Group {
				exists = true
				break
			}
		}
		if !exists {
			// 检查数据库中是否存在该用户组
			var count int64
			model.DB.Model(&model.User{}).Where("`group` = ?", req.Group).Count(&count)
			if count == 0 {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": fmt.Sprintf("用户组 '%s' 不存在", req.Group),
				})
				return
			}
		}
	}

	// 获取令牌
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌不存在",
		})
		return
	}

	// 检查权限（只能修改自己的令牌）
	userId := c.GetInt(ctxkey.Id)
	if token.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权限修改此令牌",
		})
		return
	}

	// 更新令牌的分组（转换为多用户组格式）
	err = token.SetUserGroups([]string{req.Group})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "设置令牌分组失败: " + err.Error(),
		})
		return
	}

	err = token.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "更新令牌分组失败: " + err.Error(),
		})
		return
	}

	// 清除相关缓存
	if common.RedisEnabled {
		common.RedisDel(fmt.Sprintf("token:%s", token.Key))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("令牌用户组已更新为 '%s'", req.Group),
	})
}

// UpdateTokenGroups 修改令牌的多个用户组（支持优先级顺序）
func UpdateTokenGroups(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的令牌ID",
		})
		return
	}

	var req struct {
		UserGroups []string `json:"user_groups" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 验证用户组列表
	if len(req.UserGroups) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户组列表不能为空",
		})
		return
	}

	// 验证每个用户组是否存在
	for _, group := range req.UserGroups {
		if group == "default" {
			continue // default组总是有效的
		}

		// 检查用户组是否存在于billing ratio中
		exists := false
		for groupName := range billingratio.GroupRatio {
			if groupName == group {
				exists = true
				break
			}
		}

		if !exists {
			// 检查数据库中是否存在该用户组
			var count int64
			model.DB.Model(&model.User{}).Where("`group` = ?", group).Count(&count)
			if count == 0 {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": fmt.Sprintf("用户组 '%s' 不存在", group),
				})
				return
			}
		}
	}

	// 获取令牌
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌不存在",
		})
		return
	}

	// 检查权限（只能修改自己的令牌）
	userId := c.GetInt(ctxkey.Id)
	if token.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权限修改此令牌",
		})
		return
	}

	// 更新令牌的用户组列表
	err = token.SetUserGroups(req.UserGroups)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "设置用户组失败: " + err.Error(),
		})
		return
	}

	err = token.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "更新令牌用户组失败: " + err.Error(),
		})
		return
	}

	// 清除相关缓存
	if common.RedisEnabled {
		common.RedisDel(fmt.Sprintf("token:%s", token.Key))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "令牌用户组更新成功",
		"data": gin.H{
			"id":          token.Id,
			"user_groups": token.GetUserGroups(),
		},
	})
}

// GetTokenGroups 获取令牌的用户组列表
func GetTokenGroups(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的令牌ID",
		})
		return
	}

	// 获取令牌
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌不存在",
		})
		return
	}

	// 检查权限（只能查看自己的令牌）
	userId := c.GetInt(ctxkey.Id)
	if token.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权限查看此令牌",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":          token.Id,
			"name":        token.Name,
			"user_groups": token.GetUserGroups(),
		},
	})
}
