package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// SubjectInfo 学科组信息
type SubjectInfo struct {
	SubjectId    int    `json:"subject_id"`    // 学科组ID
	SubjectName  string `json:"subject_name"`  // 学科组名称
	SchoolId     int    `json:"school_id"`     // 所属学校ID
	SchoolName   string `json:"school_name"`   // 所属学校名称
	DefaultModel string `json:"default_model"` // 默认使用的模型
	UpdatedAt    int64  `json:"updated_at"`    // 更新时间
}

// TeacherInfo 老师信息
type TeacherInfo struct {
	TeacherId   string `json:"teacher_id"`   // 老师ID
	TeacherName string `json:"teacher_name"` // 老师姓名
	SchoolId    int    `json:"school_id"`    // 所属学校ID
	SchoolName  string `json:"school_name"`  // 所属学校名称
	SubjectId   int    `json:"subject_id"`   // 所属学科组ID
	SubjectName string `json:"subject_name"` // 所属学科组名称
	GroupName   string `json:"group_name"`   // 用户组名称
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

	// 1. 尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("%s%d", SubjectInfoCachePrefix, subjectId)
	infoStr, err := common.RedisGet(cacheKey)

	if err == nil && infoStr != "" {
		// 缓存命中，解析并返回
		var info SubjectInfo
		err = json.Unmarshal([]byte(infoStr), &info)
		if err == nil {
			logger.Debugf(ctx, "学科组信息缓存命中: subjectId=%d", subjectId)
			return &info, nil
		}
		logger.Warnf(ctx, "学科组信息缓存解析失败: subjectId=%d, error=%v", subjectId, err)
	}

	// 2. 缓存未命中，从API获取
	info, err := GetSubjectInfoFromAPI(ctx, subjectId)
	if err != nil {
		logger.Warnf(ctx, "获取学科组信息失败: subjectId=%d, error=%v", subjectId, err)
		return nil, err
	}

	// 3. 更新缓存
	if info != nil {
		infoJson, _ := json.Marshal(info)
		common.RedisSet(cacheKey, string(infoJson), DefaultCacheTTL)
		logger.Debugf(ctx, "学科组信息已缓存: subjectId=%d, model=%s", subjectId, info.DefaultModel)
	}

	return info, nil
}

// GetSubjectInfoFromAPI 从课件平台API获取学科组信息
var GetSubjectInfoFromAPI = func(ctx context.Context, subjectId int) (*SubjectInfo, error) {
	// TODO: 实现从课件平台API获取学科组信息
	logger.Warnf(ctx, "GetSubjectInfoFromAPI未实现实际API调用: subjectId=%d", subjectId)
	return nil, nil
}

// GetTeacherInfoWithCache 获取老师信息（带缓存）
func GetTeacherInfoWithCache(ctx context.Context, teacherId string) (*TeacherInfo, error) {
	if !common.RedisEnabled {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 1. 尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("%s%s", TeacherInfoCachePrefix, teacherId)
	infoStr, err := common.RedisGet(cacheKey)

	if err == nil && infoStr != "" {
		// 缓存命中，解析并返回
		var info TeacherInfo
		err = json.Unmarshal([]byte(infoStr), &info)
		if err == nil {
			logger.Debugf(ctx, "老师信息缓存命中: teacherId=%s", teacherId)
			return &info, nil
		}
		logger.Warnf(ctx, "老师信息缓存解析失败: teacherId=%s, error=%v", teacherId, err)
	}

	// 2. 缓存未命中，从API获取
	info, err := GetTeacherInfoFromAPI(ctx, teacherId)
	if err != nil {
		logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
		return nil, err
	}

	// 3. 更新缓存
	if info != nil {
		infoJson, _ := json.Marshal(info)
		common.RedisSet(cacheKey, string(infoJson), DefaultCacheTTL)
		logger.Debugf(ctx, "老师信息已缓存: teacherId=%s, subjectId=%d", teacherId, info.SubjectId)
	}

	return info, nil
}

// GetTeacherInfoFromAPI 从课件平台API获取老师信息
var GetTeacherInfoFromAPI = func(ctx context.Context, teacherId string) (*TeacherInfo, error) {
	// TODO: 实现从课件平台API获取老师信息
	logger.Warnf(ctx, "GetTeacherInfoFromAPI未实现实际API调用: teacherId=%s", teacherId)
	return nil, nil
}

// InvalidateSubjectInfoCache 使指定学科组的信息缓存失效
func InvalidateSubjectInfoCache(ctx context.Context, subjectId int) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%d", SubjectInfoCachePrefix, subjectId)
	err := common.RedisDel(cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除学科组信息缓存失败: subjectId=%d, error=%v", subjectId, err)
		return err
	}
	logger.Debugf(ctx, "学科组信息缓存已删除: subjectId=%d", subjectId)
	return nil
}

// InvalidateTeacherInfoCache 使指定老师的信息缓存失效
func InvalidateTeacherInfoCache(ctx context.Context, teacherId string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	cacheKey := fmt.Sprintf("%s%s", TeacherInfoCachePrefix, teacherId)
	err := common.RedisDel(cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除老师信息缓存失败: teacherId=%s, error=%v", teacherId, err)
		return err
	}
	logger.Debugf(ctx, "老师信息缓存已删除: teacherId=%s", teacherId)
	return nil
}

// GetModelByTeacherId 根据老师ID获取应该使用的模型
// 优先级：用户偏好 > 学科组默认 > 学校默认 > 全局默认
func GetModelByTeacherId(ctx context.Context, teacherId string) (string, error) {
	// 1. 获取用户模型配置
	userConfig, err := GetUserModelConfigWithCache(ctx, teacherId)
	if err == nil && userConfig != nil && userConfig.ModelName != "" {
		logger.Debugf(ctx, "使用用户偏好模型: teacherId=%s, model=%s", teacherId, userConfig.ModelName)
		return userConfig.ModelName, nil
	}

	// 2. 获取老师所属学科组
	teacherInfo, err := GetTeacherInfoWithCache(ctx, teacherId)
	if err != nil {
		logger.Warnf(ctx, "获取老师信息失败: teacherId=%s, error=%v", teacherId, err)
		return "", err
	}

	// 3. 获取学科组默认模型
	if teacherInfo.SubjectId > 0 {
		subjectInfo, err := GetSubjectInfoWithCache(ctx, teacherInfo.SubjectId)
		if err == nil && subjectInfo != nil && subjectInfo.DefaultModel != "" {
			logger.Debugf(ctx, "使用学科组默认模型: teacherId=%s, subjectId=%d, model=%s",
				teacherId, teacherInfo.SubjectId, subjectInfo.DefaultModel)
			return subjectInfo.DefaultModel, nil
		}
	}

	// 4. 使用全局默认模型
	logger.Debugf(ctx, "未找到匹配模型，使用请求中指定的模型: teacherId=%s", teacherId)
	return "", nil
}
