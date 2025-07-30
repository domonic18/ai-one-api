package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/model"
)

// CostReport 费用报表查询结构
type CostReport struct {
	Dimension    string  `json:"dimension"`     // 维度名称
	DimensionId  int     `json:"dimension_id"`  // 维度ID
	UserGroup    string  `json:"user_group"`    // 用户组
	TotalCost    float64 `json:"total_cost"`    // 总费用
	TotalTokens  int64   `json:"total_tokens"`  // 总令牌数
	RequestCount int64   `json:"request_count"` // 请求次数
}

// ComprehensiveCostReport 综合维度费用报表
type ComprehensiveCostReport struct {
	SchoolName   string  `json:"school_name"`
	SubjectName  string  `json:"subject_name"`
	TeacherId    string  `json:"teacher_id"`
	TeacherName  string  `json:"teacher_name"`
	UserGroup    string  `json:"user_group"`
	TotalCost    float64 `json:"total_cost"`
	TotalTokens  int64   `json:"total_tokens"`
	RequestCount int64   `json:"request_count"`
}

// GetCostReportBySchool 按学校维度查询费用报表
func GetCostReportBySchool(ctx context.Context, schoolId int, startTime, endTime time.Time) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
				SELECT
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')) as dimension,
			CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED) as dimension_id,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE 
			l.created_at BETWEEN ? AND ?
			AND CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED) = ?
		GROUP BY 
			JSON_EXTRACT(el.dimension_info, '$.school_id'),
			JSON_EXTRACT(el.dimension_info, '$.school_name'),
			el.user_group,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')),
			CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED)
		ORDER BY total_cost DESC
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, schoolId).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询学校费用报表失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}

// GetCostReportBySubject 按学科组维度查询费用报表
func GetCostReportBySubject(ctx context.Context, subjectId int, startTime, endTime time.Time) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()
	query := `
				SELECT
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.subject_name')) as dimension,
			CAST(JSON_EXTRACT(el.dimension_info, '$.subject_id') AS UNSIGNED) as dimension_id,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE 
			l.created_at BETWEEN ? AND ?
			AND CAST(JSON_EXTRACT(el.dimension_info, '$.subject_id') AS UNSIGNED) = ?
		GROUP BY 
			JSON_EXTRACT(el.dimension_info, '$.subject_id'),
			JSON_EXTRACT(el.dimension_info, '$.subject_name'),
			el.user_group,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.subject_name')),
			CAST(JSON_EXTRACT(el.dimension_info, '$.subject_id') AS UNSIGNED)
		ORDER BY total_cost DESC
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, subjectId).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询学科组费用报表失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}

// GetCostReportByTeacher 按老师维度查询费用报表
func GetCostReportByTeacher(ctx context.Context, teacherId string, startTime, endTime time.Time) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
		SELECT 
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name')) as dimension,
			el.external_user_id as dimension_id,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE 
			l.created_at BETWEEN ? AND ?
			AND el.external_user_id = ?
		GROUP BY el.external_user_id, 
				 JSON_EXTRACT(el.dimension_info, '$.teacher_name'),
				 el.user_group,
				 JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name')),
				 el.external_user_id
		ORDER BY total_cost DESC
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, teacherId).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询老师费用报表失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}

// GetCostReportByUserGroup 按用户组维度查询费用报表
func GetCostReportByUserGroup(ctx context.Context, userGroup string, startTime, endTime time.Time) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
		SELECT 
			el.user_group as dimension,
			0 as dimension_id,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE 
			l.created_at BETWEEN ? AND ?
			AND el.user_group = ?
		GROUP BY el.user_group
		ORDER BY total_cost DESC
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, userGroup).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户组费用报表失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}

// GetComprehensiveCostReport 综合维度费用报表（支持多维度组合查询）
func GetComprehensiveCostReport(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time) ([]*ComprehensiveCostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	baseQuery := `
		SELECT 
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')) as school_name,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.subject_name')) as subject_name,
			el.external_user_id as teacher_id,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name')) as teacher_name,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE l.created_at BETWEEN ? AND ?
	`

	args := []interface{}{startTimestamp, endTimestamp}

	// 根据过滤条件动态构建查询
	if schoolId, ok := filters["school_id"]; ok {
		baseQuery += " AND CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED) = ?"
		args = append(args, schoolId)
	}

	if subjectId, ok := filters["subject_id"]; ok {
		baseQuery += " AND CAST(JSON_EXTRACT(el.dimension_info, '$.subject_id') AS UNSIGNED) = ?"
		args = append(args, subjectId)
	}

	if teacherId, ok := filters["teacher_id"]; ok {
		baseQuery += " AND el.external_user_id = ?"
		args = append(args, teacherId)
	}

	if userGroup, ok := filters["user_group"]; ok {
		baseQuery += " AND el.user_group = ?"
		args = append(args, userGroup)
	}

	baseQuery += `
		GROUP BY
			JSON_EXTRACT(el.dimension_info, '$.school_id'),
			JSON_EXTRACT(el.dimension_info, '$.school_name'),
			JSON_EXTRACT(el.dimension_info, '$.subject_id'),
			JSON_EXTRACT(el.dimension_info, '$.subject_name'),
			el.external_user_id,
			JSON_EXTRACT(el.dimension_info, '$.teacher_name'),
			el.user_group,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')),
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.subject_name')),
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name'))
		ORDER BY total_cost DESC
		LIMIT 1000
	`

	var reports []*ComprehensiveCostReport
	err := model.DB.WithContext(ctx).Raw(baseQuery, args...).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询综合费用报表失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*ComprehensiveCostReport{}
	}

	return reports, nil
}

// GetCostReportSummary 获取费用报表汇总信息
func GetCostReportSummary(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
		SELECT
			COUNT(DISTINCT el.external_user_id) as total_teachers,
			COUNT(DISTINCT JSON_EXTRACT(el.dimension_info, '$.school_id')) as total_schools,
			COUNT(DISTINCT JSON_EXTRACT(el.dimension_info, '$.subject_id')) as total_subjects,
			COUNT(DISTINCT el.user_group) as total_groups,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as total_requests
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE l.created_at BETWEEN ? AND ?
	`

	var result map[string]interface{}
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("查询费用报表汇总失败: %w", err)
	}

	// 确保返回空map而不是nil
	if result == nil {
		result = make(map[string]interface{})
	}

	return result, nil
}

// GetTopSchoolsByCost 获取费用最高的学校列表
func GetTopSchoolsByCost(ctx context.Context, startTime, endTime time.Time, limit int) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
		SELECT
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')) as dimension,
			CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED) as dimension_id,
			'' as user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE l.created_at BETWEEN ? AND ?
		GROUP BY
			JSON_EXTRACT(el.dimension_info, '$.school_id'),
			JSON_EXTRACT(el.dimension_info, '$.school_name'),
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.school_name')),
			CAST(JSON_EXTRACT(el.dimension_info, '$.school_id') AS UNSIGNED)
		ORDER BY total_cost DESC
		LIMIT ?
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, limit).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询学校费用排行失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}

// GetTopTeachersByCost 获取费用最高的老师列表
func GetTopTeachersByCost(ctx context.Context, startTime, endTime time.Time, limit int) ([]*CostReport, error) {
	// 检查数据库连接
	if model.DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	// 转换为Unix时间戳
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	query := `
		SELECT
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name')) as dimension,
			el.external_user_id as dimension_id,
			el.user_group,
			SUM(l.quota) as total_cost,
			SUM(l.prompt_tokens + l.completion_tokens) as total_tokens,
			COUNT(*) as request_count
		FROM logs l
		JOIN extended_logs el ON l.id = el.log_id
		WHERE l.created_at BETWEEN ? AND ?
		GROUP BY
			el.external_user_id,
			JSON_EXTRACT(el.dimension_info, '$.teacher_name'),
			el.user_group,
			JSON_UNQUOTE(JSON_EXTRACT(el.dimension_info, '$.teacher_name')),
			el.external_user_id
		ORDER BY total_cost DESC
		LIMIT ?
	`

	var reports []*CostReport
	err := model.DB.WithContext(ctx).Raw(query, startTimestamp, endTimestamp, limit).Scan(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("查询老师费用排行失败: %w", err)
	}

	// 确保返回空切片而不是nil
	if reports == nil {
		reports = []*CostReport{}
	}

	return reports, nil
}
