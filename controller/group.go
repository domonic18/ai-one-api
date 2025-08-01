package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
)

// GroupInfo 用户组信息结构
type GroupInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCount   int64  `json:"user_count"`
	TotalQuota  int64  `json:"total_quota"`
	UsedQuota   int64  `json:"used_quota"`
}

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range billingratio.GroupRatio {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}

// GetGroupsDetail 获取用户组详细信息
func GetGroupsDetail(c *gin.Context) {
	var groups []GroupInfo

	// 获取所有用户组名称
	for groupName := range billingratio.GroupRatio {
		groupInfo := GroupInfo{
			Name:        groupName,
			Description: getGroupDescription(groupName),
		}

		// 统计该用户组的用户数量和配额
		var userCount int64
		var totalQuota int64
		var usedQuota int64

		// 查询用户数量
		model.DB.Model(&model.User{}).Where("`group` = ? AND status = ?", groupName, model.UserStatusEnabled).Count(&userCount)

		// 查询总配额和已使用配额
		model.DB.Model(&model.User{}).Where("`group` = ? AND status = ?", groupName, model.UserStatusEnabled).
			Select("COALESCE(SUM(quota), 0) as total_quota, COALESCE(SUM(used_quota), 0) as used_quota").
			Scan(&struct {
				TotalQuota int64 `json:"total_quota"`
				UsedQuota  int64 `json:"used_quota"`
			}{TotalQuota: totalQuota, UsedQuota: usedQuota})

		groupInfo.UserCount = userCount
		groupInfo.TotalQuota = totalQuota
		groupInfo.UsedQuota = usedQuota

		groups = append(groups, groupInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groups,
	})
}

// getGroupDescription 获取用户组描述
func getGroupDescription(groupName string) string {
	descriptions := map[string]string{
		"default": "默认用户组",
		"vip":     "VIP用户组",
		"svip":    "SVIP用户组",
	}

	if desc, exists := descriptions[groupName]; exists {
		return desc
	}
	return "用户组"
}
