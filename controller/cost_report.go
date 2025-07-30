package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model/identity"
)

// CostReportController 费用报表控制器
type CostReportController struct{}

// NewCostReportController 创建费用报表控制器
func NewCostReportController() *CostReportController {
	return &CostReportController{}
}

// GetCostReportBySchool 按学校维度查询费用报表
func (c *CostReportController) GetCostReportBySchool(ctx *gin.Context) {
	// 获取查询参数
	schoolIdStr := ctx.Query("school_id")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 验证必要参数
	if schoolIdStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "school_id参数不能为空",
		})
		return
	}

	// 解析学校ID
	schoolId, err := strconv.Atoi(schoolIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "school_id参数格式错误",
		})
		return
	}

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 查询费用报表
	reports, err := identity.GetCostReportBySchool(ctx.Request.Context(), schoolId, startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询学校费用报表失败: schoolId=%d, error=%v", schoolId, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"school_id":   schoolId,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetCostReportBySubject 按学科组维度查询费用报表
func (c *CostReportController) GetCostReportBySubject(ctx *gin.Context) {
	// 获取查询参数
	subjectIdStr := ctx.Query("subject_id")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 验证必要参数
	if subjectIdStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "subject_id参数不能为空",
		})
		return
	}

	// 解析学科组ID
	subjectId, err := strconv.Atoi(subjectIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "subject_id参数格式错误",
		})
		return
	}

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 查询费用报表
	reports, err := identity.GetCostReportBySubject(ctx.Request.Context(), subjectId, startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询学科组费用报表失败: subjectId=%d, error=%v", subjectId, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subject_id":  subjectId,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetCostReportByTeacher 按老师维度查询费用报表
func (c *CostReportController) GetCostReportByTeacher(ctx *gin.Context) {
	// 获取查询参数
	teacherId := ctx.Query("teacher_id")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 验证必要参数
	if teacherId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "teacher_id参数不能为空",
		})
		return
	}

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 查询费用报表
	reports, err := identity.GetCostReportByTeacher(ctx.Request.Context(), teacherId, startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询老师费用报表失败: teacherId=%s, error=%v", teacherId, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"teacher_id":  teacherId,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetCostReportByUserGroup 按用户组维度查询费用报表
func (c *CostReportController) GetCostReportByUserGroup(ctx *gin.Context) {
	// 获取查询参数
	userGroup := ctx.Query("user_group")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 验证必要参数
	if userGroup == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "user_group参数不能为空",
		})
		return
	}

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 查询费用报表
	reports, err := identity.GetCostReportByUserGroup(ctx.Request.Context(), userGroup, startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询用户组费用报表失败: userGroup=%s, error=%v", userGroup, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_group":  userGroup,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetComprehensiveCostReport 综合维度费用报表查询
func (c *CostReportController) GetComprehensiveCostReport(ctx *gin.Context) {
	// 获取查询参数
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if schoolIdStr := ctx.Query("school_id"); schoolIdStr != "" {
		if schoolId, err := strconv.Atoi(schoolIdStr); err == nil {
			filters["school_id"] = schoolId
		}
	}
	if subjectIdStr := ctx.Query("subject_id"); subjectIdStr != "" {
		if subjectId, err := strconv.Atoi(subjectIdStr); err == nil {
			filters["subject_id"] = subjectId
		}
	}
	if teacherId := ctx.Query("teacher_id"); teacherId != "" {
		filters["teacher_id"] = teacherId
	}
	if userGroup := ctx.Query("user_group"); userGroup != "" {
		filters["user_group"] = userGroup
	}

	// 查询费用报表
	reports, err := identity.GetComprehensiveCostReport(ctx.Request.Context(), filters, startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询综合费用报表失败: filters=%v, error=%v", filters, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"filters":     filters,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetCostReportSummary 获取费用报表汇总信息
func (c *CostReportController) GetCostReportSummary(ctx *gin.Context) {
	// 获取查询参数
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 查询汇总信息
	summary, err := identity.GetCostReportSummary(ctx.Request.Context(), startTime, endTime)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询费用报表汇总失败: error=%v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询费用报表汇总失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   endTime.Format(time.RFC3339),
			"summary":    summary,
		},
	})
}

// GetTopSchoolsByCost 获取费用最高的学校列表
func (c *CostReportController) GetTopSchoolsByCost(ctx *gin.Context) {
	// 获取查询参数
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")
	limitStr := ctx.Query("limit")

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 解析限制数量
	limit := 10 // 默认限制10个
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// 查询学校排行
	reports, err := identity.GetTopSchoolsByCost(ctx.Request.Context(), startTime, endTime, limit)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询学校费用排行失败: error=%v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询学校费用排行失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"limit":       limit,
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// GetTopTeachersByCost 获取费用最高的老师列表
func (c *CostReportController) GetTopTeachersByCost(ctx *gin.Context) {
	// 获取查询参数
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")
	limitStr := ctx.Query("limit")

	// 解析时间范围
	startTime, endTime, err := c.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "时间范围参数格式错误",
		})
		return
	}

	// 解析限制数量
	limit := 10 // 默认限制10个
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// 查询老师排行
	reports, err := identity.GetTopTeachersByCost(ctx.Request.Context(), startTime, endTime, limit)
	if err != nil {
		logger.Errorf(ctx.Request.Context(), "查询老师费用排行失败: error=%v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询老师费用排行失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
			"limit":       limit,
			"reports":     reports,
			"total_count": len(reports),
		},
	})
}

// parseTimeRange 解析时间范围参数
func (c *CostReportController) parseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	// 解析开始时间
	if startTimeStr == "" {
		// 默认使用30天前
		startTime = time.Now().AddDate(0, 0, -30)
	} else {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// 解析结束时间
	if endTimeStr == "" {
		// 默认使用当前时间
		endTime = time.Now()
	} else {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// 验证时间范围
	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始时间不能晚于结束时间")
	}

	return startTime, endTime, nil
}
