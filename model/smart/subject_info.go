package smart

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/cache"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/logger"
)

// SubjectInfo 学科组信息结构
type SubjectInfo struct {
	SubjectId    int    `json:"subject_id"`
	SubjectName  string `json:"subject_name"`
	SchoolId     int    `json:"school_id"`
	SchoolName   string `json:"school_name"`
	DefaultModel string `json:"default_model"`
	GroupName    string `json:"group_name"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// TeacherInfo 老师信息结构
type TeacherInfo struct {
	TeacherId   string `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	SchoolId    int    `json:"school_id"`
	SchoolName  string `json:"school_name"`
	SubjectId   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	GroupName   string `json:"group_name"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// 缓存相关常量
const (
	// 学科组信息缓存前缀
	SubjectInfoCachePrefix = "subject_info:"
	// 老师信息缓存前缀
	TeacherInfoCachePrefix = "teacher_info:"
	// 缓存时间（24小时）
	DefaultCacheTTL = 24 * time.Hour
)

// GetSubjectInfoWithCache 获取学科组信息（带缓存）
func GetSubjectInfoWithCache(ctx context.Context, subjectId int) (*SubjectInfo, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 1. 尝试从缓存管理器获取
	cacheKey := fmt.Sprintf("%s%d", SubjectInfoCachePrefix, subjectId)
	var info SubjectInfo

	err := cache.Mgr.Get(ctx, cacheKey, &info)
	if err == nil {
		logger.Debugf(ctx, "学科组信息缓存命中: subjectId=%d", subjectId)
		return &info, nil
	}

	// 2. 缓存未命中，从API获取
	apiInfo, err := GetSubjectInfoFromAPI(ctx, subjectId)
	if err != nil {
		logger.Warnf(ctx, "从API获取学科组信息失败: subjectId=%d, error=%v", subjectId, err)
		return nil, err
	}

	// 3. 使用缓存管理器更新缓存
	if apiInfo != nil {
		err = cache.Mgr.Set(ctx, cacheKey, apiInfo, DefaultCacheTTL)
		if err != nil {
			logger.Warnf(ctx, "设置学科组信息缓存失败: subjectId=%d, error=%v", subjectId, err)
		} else {
			logger.Debugf(ctx, "学科组信息已缓存: subjectId=%d, model=%s", subjectId, apiInfo.DefaultModel)
		}
	}

	return apiInfo, nil
}

// GetSubjectInfoFromAPI 从课件平台API获取学科组信息
func GetSubjectInfoFromAPI(ctx context.Context, subjectId int) (*SubjectInfo, error) {
	// 使用课件平台API客户端获取学科组信息
	coursewareClient := client.GetCoursewareClient()
	if coursewareClient == nil {
		logger.Warnf(ctx, "课件平台API客户端未初始化: subjectId=%d", subjectId)
		return nil, fmt.Errorf("课件平台API客户端未初始化")
	}

	// 调用API获取学科组信息
	apiInfo, err := coursewareClient.GetSubjectInfo(ctx, subjectId)
	if err != nil {
		logger.Warnf(ctx, "从API获取学科组信息失败: subjectId=%d, error=%v", subjectId, err)
		return nil, err
	}

	// 如果API返回空，则返回nil
	if apiInfo == nil {
		return nil, nil
	}

	// 转换为内部SubjectInfo结构
	subjectInfo := &SubjectInfo{
		SubjectId:    apiInfo.SubjectId,
		SubjectName:  apiInfo.SubjectName,
		SchoolId:     apiInfo.SchoolId,
		SchoolName:   apiInfo.SchoolName,
		DefaultModel: apiInfo.DefaultModel,
		GroupName:    apiInfo.GroupName,
		CreatedAt:    apiInfo.CreatedAt,
		UpdatedAt:    apiInfo.UpdatedAt,
	}

	logger.Debugf(ctx, "从API获取学科组信息成功: subjectId=%d, schoolId=%d, defaultModel=%s",
		subjectId, subjectInfo.SchoolId, subjectInfo.DefaultModel)

	return subjectInfo, nil
}

// GetTeacherInfoWithCache 获取老师信息（使用简化的缓存接口）
func GetTeacherInfoWithCache(ctx context.Context, teacherId string) (*TeacherInfo, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 1. 尝试从缓存管理器获取
	cacheKey := fmt.Sprintf("%s%s", TeacherInfoCachePrefix, teacherId)
	var info TeacherInfo

	// 使用简化的GetWithFallback方法
	err := cache.Mgr.GetWithFallback(ctx, cacheKey, func() (interface{}, error) {
		return GetTeacherInfoFromAPI(ctx, teacherId)
	}, DefaultCacheTTL, &info)

	if err != nil {
		logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
		return nil, err
	}

	logger.Debugf(ctx, "老师信息获取成功: teacherId=%s, subjectId=%d", teacherId, info.SubjectId)
	return &info, nil
}

// GetTeacherInfoFromAPI 从课件平台API获取老师信息
func GetTeacherInfoFromAPI(ctx context.Context, teacherId string) (*TeacherInfo, error) {
	// 使用课件平台API客户端获取老师信息
	coursewareClient := client.GetCoursewareClient()
	if coursewareClient == nil {
		logger.Warnf(ctx, "课件平台API客户端未初始化: teacherId=%s", teacherId)
		return nil, fmt.Errorf("课件平台API客户端未初始化")
	}

	// 调用API获取老师信息
	apiInfo, err := coursewareClient.GetTeacherInfo(ctx, teacherId)
	if err != nil {
		logger.Warnf(ctx, "从API获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
		return nil, err
	}

	// 如果API返回空，则返回nil
	if apiInfo == nil {
		return nil, nil
	}

	// 转换为内部TeacherInfo结构
	teacherInfo := &TeacherInfo{
		TeacherId:   apiInfo.TeacherId,
		TeacherName: apiInfo.TeacherName,
		SchoolId:    apiInfo.SchoolId,
		SchoolName:  apiInfo.SchoolName,
		SubjectId:   apiInfo.SubjectId,
		SubjectName: apiInfo.SubjectName,
		CreatedAt:   apiInfo.CreatedAt,
		UpdatedAt:   apiInfo.UpdatedAt,
	}

	logger.Debugf(ctx, "从API获取老师信息成功: teacherId=%s, schoolId=%d, subjectId=%d",
		teacherId, teacherInfo.SchoolId, teacherInfo.SubjectId)

	return teacherInfo, nil
}

// InvalidateSubjectInfoCache 使指定学科组的信息缓存失效
func InvalidateSubjectInfoCache(ctx context.Context, subjectId int) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%d", SubjectInfoCachePrefix, subjectId)
	err := cache.Mgr.Delete(ctx, cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除学科组信息缓存失败: subjectId=%d, error=%v", subjectId, err)
		return err
	}

	logger.Debugf(ctx, "学科组信息缓存已删除: subjectId=%d", subjectId)
	return nil
}

// InvalidateTeacherInfoCache 失效老师信息缓存
func InvalidateTeacherInfoCache(ctx context.Context, teacherId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 使用简化的失效管理器
	return cache.Invalidator.InvalidateUserCache(ctx, teacherId)
}

// GetModelByTeacherId 根据老师ID获取推荐模型
func GetModelByTeacherId(ctx context.Context, teacherId string) (string, error) {
	if teacherId == "" {
		return "", fmt.Errorf("teacherId不能为空")
	}

	// 1. 首先尝试获取用户模型配置（用户偏好）
	userConfig, err := GetUserConfigWithCache(ctx, teacherId)
	if err == nil && userConfig != nil && userConfig.ModelName != "" {
		logger.Debugf(ctx, "使用用户偏好模型: teacherId=%s, model=%s", teacherId, userConfig.ModelName)
		return userConfig.ModelName, nil
	}

	// 2. 获取老师信息，查找学科组默认模型
	teacherInfo, err := GetTeacherInfoWithCache(ctx, teacherId)
	if err != nil {
		logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
		return "", err
	}

	// 检查teacherInfo是否为nil
	if teacherInfo == nil {
		logger.Warnf(ctx, "老师信息为空: teacherId=%s", teacherId)
		return "", fmt.Errorf("未找到老师信息: %s", teacherId)
	}

	// 3. 获取学科组信息
	if teacherInfo.SubjectId > 0 {
		subjectInfo, err := GetSubjectInfoWithCache(ctx, teacherInfo.SubjectId)
		if err == nil && subjectInfo != nil && subjectInfo.DefaultModel != "" {
			logger.Debugf(ctx, "使用学科组默认模型: teacherId=%s, subjectId=%d, model=%s",
				teacherId, teacherInfo.SubjectId, subjectInfo.DefaultModel)
			return subjectInfo.DefaultModel, nil
		}
	}

	// 4. 返回全局默认模型
	defaultModel := "gpt-3.5-turbo"
	logger.Debugf(ctx, "使用全局默认模型: teacherId=%s, model=%s", teacherId, defaultModel)
	return defaultModel, nil
}

// BatchInvalidateTeacherInfoCache 批量失效老师信息缓存
func BatchInvalidateTeacherInfoCache(ctx context.Context, teacherIds []string) error {
	if len(teacherIds) == 0 {
		return nil
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 构建缓存键列表
	cacheKeys := make([]string, len(teacherIds))
	for i, teacherId := range teacherIds {
		cacheKeys[i] = fmt.Sprintf("%s%s", TeacherInfoCachePrefix, teacherId)
	}

	return cache.Invalidator.InvalidateByKeys(ctx, cacheKeys)
}

// PreloadTeacherInfos 预加载老师信息到缓存
func PreloadTeacherInfos(ctx context.Context, teacherIds []string) error {
	if len(teacherIds) == 0 {
		return nil
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	successCount := 0
	for _, teacherId := range teacherIds {
		_, err := GetTeacherInfoWithCache(ctx, teacherId)
		if err != nil {
			logger.Warnf(ctx, "预加载老师信息失败: teacherId=%s, error=%v", teacherId, err)
		} else {
			successCount++
		}
	}

	logger.Debugf(ctx, "预加载老师信息完成: 请求=%d, 成功=%d", len(teacherIds), successCount)
	return nil
}

// GetSubjectInfoCacheStats 获取学科组信息缓存统计
func GetSubjectInfoCacheStats() *cache.SimpleStats {
	if cache.Mgr == nil {
		return &cache.SimpleStats{}
	}
	return cache.Mgr.GetStats()
}
