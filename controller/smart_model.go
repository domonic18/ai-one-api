package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

// 智能模型请求
type SmartModelRequest struct {
	UserID     string                 `json:"user_id" binding:"required"`
	ModelName  string                 `json:"model_name" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
}

// GetUserSmartModelConfig 获取用户的智能模型配置
func GetUserSmartModelConfig(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	if !common.RedisEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Redis未启用",
		})
		return
	}

	selector := model_selection.NewSmartModelSelector()
	config, err := selector.GetUserModelConfig(userID)
	if err != nil {
		logger.Errorf(ctx, "获取用户智能模型配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    config,
	})
}

// SetUserSmartModelConfig 设置用户的智能模型配置
func SetUserSmartModelConfig(c *gin.Context) {
	ctx := c.Request.Context()
	var req SmartModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if !common.RedisEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Redis未启用",
		})
		return
	}

	config := &model_selection.UserModelConfig{
		UserID:     req.UserID,
		ModelName:  req.ModelName,
		Parameters: req.Parameters,
		Priority:   req.Priority,
		UpdatedAt:  time.Now(),
	}

	selector := model_selection.NewSmartModelSelector()
	err := selector.SetUserModelConfig(req.UserID, config)
	if err != nil {
		logger.Errorf(ctx, "设置用户智能模型配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Infof(ctx, "成功为用户 %s 设置智能模型选择配置: %s", req.UserID, req.ModelName)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成功设置智能模型配置",
	})
}

// DeleteUserSmartModelConfig 删除用户的智能模型配置
func DeleteUserSmartModelConfig(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	if !common.RedisEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Redis未启用",
		})
		return
	}

	selector := model_selection.NewSmartModelSelector()
	err := selector.DeleteUserModelConfig(userID)
	if err != nil {
		logger.Errorf(ctx, "删除用户智能模型配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	logger.Infof(ctx, "成功删除用户 %s 的智能模型选择配置", userID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成功删除智能模型配置",
	})
}
