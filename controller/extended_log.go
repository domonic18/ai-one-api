package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/model/identity"
)

// GetAllExtendedLogs 获取所有扩展日志（管理员权限）
func GetAllExtendedLogs(c *gin.Context) {

	// 解析查询参数
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	pageSize := config.ItemsPerPage
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 解析筛选参数
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	externalUserId := c.Query("external_user_id")
	userGroup := c.Query("user_group")

	// 构建查询条件
	conditions := make(map[string]interface{})
	if externalUserId != "" {
		conditions["external_user_id"] = externalUserId
	}
	if userGroup != "" {
		conditions["user_group"] = userGroup
	}

	// 查询扩展日志
	extendedLogs, total, err := getExtendedLogsWithConditions(c, conditions, startTimestamp, endTimestamp, p*pageSize, pageSize)
	if err != nil {
		logger.Errorf(c.Request.Context(), "查询扩展日志失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "查询扩展日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"logs":  extendedLogs,
			"total": total,
		},
	})
}

// GetUserExtendedLogs 获取用户的扩展日志
func GetUserExtendedLogs(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)

	// 解析查询参数
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	pageSize := config.ItemsPerPage
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 解析筛选参数
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	_ = startTimestamp // 避免未使用变量错误
	_ = endTimestamp   // 避免未使用变量错误

	// 通过userId获取对应的external_user_id（如果存在）
	// 这里需要根据实际的用户映射逻辑来实现
	externalUserId := getUserExternalId(userId)
	if externalUserId == "" {
		// 如果没有external_user_id，返回空结果
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data": gin.H{
				"logs":  []interface{}{},
				"total": 0,
			},
		})
		return
	}

	// 查询用户的扩展日志
	extendedLogs, total, err := identity.GetExtendedLogsByExternalUserId(c.Request.Context(), externalUserId, p+1, pageSize)
	if err != nil {
		logger.Errorf(c.Request.Context(), "查询用户扩展日志失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "查询扩展日志失败: " + err.Error(),
		})
		return
	}

	// 关联原始日志信息
	enrichedLogs, err := enrichExtendedLogsWithOriginalLogs(c, extendedLogs)
	if err != nil {
		logger.Errorf(c.Request.Context(), "关联原始日志失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "关联原始日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"logs":  enrichedLogs,
			"total": total,
		},
	})
}

// GetExtendedLogDetail 获取扩展日志详情
func GetExtendedLogDetail(c *gin.Context) {
	logIdStr := c.Param("log_id")
	logId, err := strconv.ParseInt(logIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的日志ID",
		})
		return
	}

	// 获取扩展日志
	extendedLog, err := identity.GetExtendedLogByLogId(c.Request.Context(), logId)
	if err != nil {
		logger.Errorf(c.Request.Context(), "获取扩展日志详情失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取扩展日志详情失败: " + err.Error(),
		})
		return
	}

	if extendedLog == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "扩展日志不存在",
		})
		return
	}

	// 获取原始日志
	originalLog, err := model.GetLogById(logId)
	if err != nil {
		logger.Errorf(c.Request.Context(), "获取原始日志失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取原始日志失败: " + err.Error(),
		})
		return
	}

	// 解析维度信息
	dimensionInfo, err := extendedLog.GetDimensionInfo()
	if err != nil {
		logger.Warnf(c.Request.Context(), "解析维度信息失败: %v", err)
		dimensionInfo = &identity.DimensionInfo{}
	}

	// 构建详情响应
	detail := gin.H{
		"id":               extendedLog.Id,
		"log_id":           extendedLog.LogId,
		"external_user_id": extendedLog.ExternalUserId,
		"user_group":       extendedLog.UserGroup,
		"dimension_info":   dimensionInfo,
		"created_at":       extendedLog.CreatedAt,
		"updated_at":       extendedLog.UpdatedAt,
		"original_log":     originalLog,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    detail,
	})
}

// GetExtendedLogsStatistics 获取扩展日志统计信息
func GetExtendedLogsStatistics(c *gin.Context) {
	// 解析时间范围参数
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	// 设置默认时间范围（最近30天）
	if startTimestamp == 0 || endTimestamp == 0 {
		now := time.Now()
		endTime := now
		startTime := now.AddDate(0, 0, -30)

		if startTimestamp == 0 {
			startTimestamp = startTime.Unix()
		}
		if endTimestamp == 0 {
			endTimestamp = endTime.Unix()
		}
	}

	// 获取统计数据
	statistics, err := getExtendedLogsStatistics(c, startTimestamp, endTimestamp)
	if err != nil {
		logger.Errorf(c.Request.Context(), "获取扩展日志统计失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取统计数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    statistics,
	})
}

// 辅助函数：根据条件查询扩展日志
func getExtendedLogsWithConditions(c *gin.Context, conditions map[string]interface{}, startTimestamp, endTimestamp int64, offset, limit int) ([]*identity.ExtendedLog, int64, error) {
	// 构建基础查询
	query := model.DB.WithContext(c.Request.Context()).Model(&identity.ExtendedLog{})

	// 添加时间范围条件
	if startTimestamp > 0 {
		startTime := time.Unix(startTimestamp, 0)
		query = query.Where("created_at >= ?", startTime)
	}
	if endTimestamp > 0 {
		endTime := time.Unix(endTimestamp, 0)
		query = query.Where("created_at <= ?", endTime)
	}

	// 添加其他条件
	for key, value := range conditions {
		switch key {
		case "external_user_id", "user_group":
			query = query.Where(key+" = ?", value)
		}
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var extendedLogs []*identity.ExtendedLog
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&extendedLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return extendedLogs, total, nil
}

// 辅助函数：关联原始日志信息
func enrichExtendedLogsWithOriginalLogs(c *gin.Context, extendedLogs []*identity.ExtendedLog) ([]gin.H, error) {
	if len(extendedLogs) == 0 {
		return []gin.H{}, nil
	}

	// 收集所有日志ID
	logIds := make([]int64, len(extendedLogs))
	for i, el := range extendedLogs {
		logIds[i] = el.LogId
	}

	// 批量获取原始日志
	originalLogs, err := model.GetLogsByIds(logIds)
	if err != nil {
		return nil, err
	}

	// 创建日志ID到日志对象的映射
	logMap := make(map[int64]*model.Log)
	for _, log := range originalLogs {
		logMap[int64(log.Id)] = log
	}

	// 构建增强的日志列表
	enrichedLogs := make([]gin.H, len(extendedLogs))
	for i, el := range extendedLogs {
		// 解析维度信息
		dimensionInfo, err := el.GetDimensionInfo()
		if err != nil {
			logger.Warnf(c.Request.Context(), "解析维度信息失败: log_id=%d, error=%v", el.LogId, err)
			dimensionInfo = &identity.DimensionInfo{}
		}

		enrichedLog := gin.H{
			"id":               el.Id,
			"log_id":           el.LogId,
			"external_user_id": el.ExternalUserId,
			"user_group":       el.UserGroup,
			"dimension_info":   dimensionInfo,
			"created_at":       el.CreatedAt,
			"updated_at":       el.UpdatedAt,
		}

		// 添加原始日志信息
		if originalLog, exists := logMap[el.LogId]; exists {
			enrichedLog["original_log"] = originalLog
		}

		enrichedLogs[i] = enrichedLog
	}

	return enrichedLogs, nil
}

// 辅助函数：获取扩展日志统计信息
func getExtendedLogsStatistics(c *gin.Context, startTimestamp, endTimestamp int64) (gin.H, error) {
	startTime := time.Unix(startTimestamp, 0)
	endTime := time.Unix(endTimestamp, 0)

	// 总记录数统计
	var totalCount int64
	err := model.DB.WithContext(c.Request.Context()).
		Model(&identity.ExtendedLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// 按用户组统计
	var userGroupStats []struct {
		UserGroup string `json:"user_group"`
		Count     int64  `json:"count"`
	}
	err = model.DB.WithContext(c.Request.Context()).
		Model(&identity.ExtendedLog{}).
		Select("user_group, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("user_group").
		Order("count DESC").
		Limit(10).
		Find(&userGroupStats).Error
	if err != nil {
		return nil, err
	}

	// 按外部用户ID统计（Top活跃用户）
	var externalUserStats []struct {
		ExternalUserId string `json:"external_user_id"`
		Count          int64  `json:"count"`
	}
	err = model.DB.WithContext(c.Request.Context()).
		Model(&identity.ExtendedLog{}).
		Select("external_user_id, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Where("external_user_id != ''").
		Group("external_user_id").
		Order("count DESC").
		Limit(10).
		Find(&externalUserStats).Error
	if err != nil {
		return nil, err
	}

	return gin.H{
		"total_count":         totalCount,
		"user_group_stats":    userGroupStats,
		"external_user_stats": externalUserStats,
		"time_range": gin.H{
			"start_time": startTime,
			"end_time":   endTime,
		},
	}, nil
}

// 辅助函数：获取用户的外部ID
func getUserExternalId(userId int) string {
	// 这里需要根据实际的用户映射逻辑来实现
	// 可能需要查询用户表或者从其他地方获取映射关系
	// 暂时返回空字符串，表示没有外部ID映射
	return ""
}
