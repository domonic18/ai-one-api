package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
)

// GroupInfo 用户组信息结构
type GroupInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCount   int64  `json:"user_count"`
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

	// 从数据库中获取所有用户组名称（包括新创建的）
	var groupNames []string

	// 首先获取预定义的用户组
	for groupName := range billingratio.GroupRatio {
		groupNames = append(groupNames, groupName)
	}

	// 然后从数据库中获取所有用户组（包括新创建的）
	var dbGroups []string
	model.DB.Model(&model.User{}).
		Where("`group` != '' AND `group` IS NOT NULL").
		Distinct().
		Pluck("`group`", &dbGroups)

	// 合并并去重
	groupMap := make(map[string]bool)
	for _, name := range groupNames {
		groupMap[name] = true
	}
	for _, name := range dbGroups {
		if name != "" && !groupMap[name] {
			groupNames = append(groupNames, name)
			groupMap[name] = true
		}
	}

	// 为每个用户组创建详细信息
	for _, groupName := range groupNames {
		groupInfo := GroupInfo{
			Name:        groupName,
			Description: getGroupDescription(groupName),
		}

		// 统计该用户组的用户数量
		var userCount int64

		// 查询用户数量（排除用户组代表用户）
		model.DB.Model(&model.User{}).
			Where("`group` = ? AND username NOT LIKE 'group_%'", groupName).
			Count(&userCount)

		groupInfo.UserCount = userCount

		groups = append(groups, groupInfo)
	}

	logger.SysLog(fmt.Sprintf("GetGroupsDetail 返回 %d 个用户组", len(groups)))
	for i, group := range groups {
		logger.SysLog(fmt.Sprintf("用户组 %d: name=%s, description=%s", i, group.Name, group.Description))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groups,
	})
}

// getGroupDescription 获取用户组描述
func getGroupDescription(groupName string) string {
	// 预定义用户组描述
	descriptions := map[string]string{
		"default": "默认用户组",
		"vip":     "VIP用户组",
		"svip":    "SVIP用户组",
	}

	if desc, exists := descriptions[groupName]; exists {
		return desc
	}

	// 对于动态创建的用户组，尝试从数据库中获取描述
	// 查找该用户组的代表用户（用户名为 "group_" + groupName 的用户）
	var user model.User
	if err := model.DB.Where("username = ?", "group_"+groupName).First(&user).Error; err == nil {
		// 直接返回显示名称作为用户组描述
		if user.DisplayName != "" {
			return user.DisplayName
		}
	}

	return "自定义用户组"
}

// CreateGroupRequest 创建用户组请求结构
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	UserCount   int64  `json:"user_count"`
}

// UpdateGroupRequest 更新用户组请求结构
type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	UserCount   int64  `json:"user_count"`
}

// CreateGroup 创建用户组
func CreateGroup(c *gin.Context) {
	// 读取原始请求体用于调试
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "读取请求体失败: " + err.Error(),
		})
		return
	}

	// 打印请求体用于调试
	logger.SysLog("CreateGroup 请求体: " + string(body))

	// 重新设置请求体，因为GetRawData会消费掉
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SysLog("CreateGroup 参数绑定失败: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户组名称是否已存在（包括所有状态的用户）
	var existingCount int64
	model.DB.Model(&model.User{}).Where("`group` = ?", req.Name).Count(&existingCount)
	if existingCount > 0 {
		logger.SysLog("CreateGroup 用户组名称已存在: " + req.Name)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户组名称已存在",
		})
		return
	}

	// 创建用户组（通过创建默认用户来实现）
	// 注意：这里我们创建一个隐藏的管理员用户来代表用户组
	// 实际项目中可能需要单独的用户组表
	displayName := req.Name
	if req.Description != "" {
		displayName = req.Description
	}

	defaultUser := &model.User{
		Username:    "group_" + req.Name,
		Password:    "group_password",        // 实际项目中应该使用更安全的方式
		DisplayName: displayName,             // 使用描述作为显示名称
		Role:        model.RoleCommonUser,    // 默认普通用户
		Status:      model.UserStatusEnabled, // 默认启用状态，仅用于标识用户组
		Group:       req.Name,
		Quota:       0,                         // 用户组代表用户不需要配额
		AffCode:     random.GetRandomString(8), // 生成唯一的推荐码
		AccessToken: random.GetUUID(),          // 生成唯一的访问令牌
	}

	if err := model.DB.Create(defaultUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建用户组失败: " + err.Error(),
		})
		return
	}

	// 将新用户组添加到GroupRatio映射中
	billingratio.GroupRatio[req.Name] = 1.0

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户组创建成功",
		"data": gin.H{
			"id":   defaultUser.Id,
			"name": req.Name,
		},
	})
}

// UpdateGroup 更新用户组
func UpdateGroup(c *gin.Context) {
	groupName := c.Param("id")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户组名称不能为空",
		})
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户组是否存在
	var userCount int64
	model.DB.Model(&model.User{}).Where("`group` = ? AND status = ?", groupName, model.UserStatusEnabled).Count(&userCount)
	if userCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户组不存在",
		})
		return
	}

	// 如果更改了名称，检查新名称是否已存在
	if req.Name != groupName {
		var existingCount int64
		model.DB.Model(&model.User{}).Where("`group` = ? AND status = ?", req.Name, model.UserStatusEnabled).Count(&existingCount)
		if existingCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "用户组名称已存在",
			})
			return
		}
	}

	// 更新用户组信息
	// 用户组本身不管理配额，只更新描述信息
	// 这里可以更新用户组代表用户的显示名称
	if err := model.DB.Model(&model.User{}).
		Where("username = ?", "group_"+groupName).
		Update("display_name", req.Description).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户组失败: " + err.Error(),
		})
		return
	}

	// 如果更改了名称，更新用户组名称
	if req.Name != groupName {
		if err := model.DB.Model(&model.User{}).
			Where("`group` = ? AND status = ?", groupName, model.UserStatusEnabled).
			Update("`group`", req.Name).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "更新用户组名称失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户组更新成功",
	})
}

// DeleteGroup 删除用户组
func DeleteGroup(c *gin.Context) {
	groupName := c.Param("id")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户组名称不能为空",
		})
		return
	}

	logger.SysLog("DeleteGroup 尝试删除用户组: " + groupName)

	// 检查是否为预定义用户组
	if groupName == "default" || groupName == "vip" || groupName == "svip" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无法删除系统预定义用户组",
		})
		return
	}

	// 检查用户组是否存在（查找用户组代表用户）
	var groupUser model.User
	if err := model.DB.Where("username = ?", "group_"+groupName).First(&groupUser).Error; err != nil {
		logger.SysLog("DeleteGroup 用户组不存在: " + groupName)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户组不存在",
		})
		return
	}

	// 检查用户组是否有活跃用户（排除用户组代表用户）
	var activeUserCount int64
	model.DB.Model(&model.User{}).Where("`group` = ? AND status = ? AND username NOT LIKE 'group_%'", groupName, model.UserStatusEnabled).Count(&activeUserCount)
	if activeUserCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("用户组中还有 %d 个活跃用户，无法删除", activeUserCount),
		})
		return
	}

	// 删除用户组代表用户
	if err := model.DB.Delete(&groupUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除用户组失败: " + err.Error(),
		})
		return
	}

	// 从GroupRatio映射中移除
	delete(billingratio.GroupRatio, groupName)

	logger.SysLog("DeleteGroup 用户组删除成功: " + groupName)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户组删除成功",
	})
}
