package model

import (
	"context"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"gorm.io/gorm"
)

// ExtendedLog 扩展日志表，用于记录学校、学科组、老师等维度信息
type ExtendedLog struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	LogId       int    `json:"log_id" gorm:"uniqueIndex:idx_log_id;not null;comment:关联原日志表ID"`
	Log         Log    `json:"-" gorm:"foreignKey:LogId;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	SchoolId    int    `json:"school_id" gorm:"index:idx_school_id;default:0;comment:学校ID"`
	SchoolName  string `json:"school_name" gorm:"type:varchar(100);default:'';comment:学校名称"`
	SubjectId   int    `json:"subject_id" gorm:"index:idx_subject_id;default:0;comment:学科组ID"`
	SubjectName string `json:"subject_name" gorm:"type:varchar(100);default:'';comment:学科组名称"`
	TeacherId   string `json:"teacher_id" gorm:"type:varchar(100);index:idx_teacher_id;default:'';comment:老师ID"`
	TeacherName string `json:"teacher_name" gorm:"type:varchar(100);default:'';comment:老师姓名"`
	GroupName   string `json:"group_name" gorm:"type:varchar(100);default:'';comment:用户组名称"`
	CreatedAt   int64  `json:"created_at" gorm:"index:idx_created_at;not null;comment:创建时间"`
}

// ExtendedLogInfo 扩展日志信息结构体，用于传递扩展日志数据
type ExtendedLogInfo struct {
	SchoolId    int    `json:"school_id"`
	SchoolName  string `json:"school_name"`
	SubjectId   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	TeacherId   string `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	GroupName   string `json:"group_name"`
}

// CompleteLogInfo 完整日志信息，包含原日志和扩展日志
type CompleteLogInfo struct {
	Log         `json:"log"`                    // 原日志信息
	ExtendedLog `json:"extended_log,omitempty"` // 扩展日志信息
}

// CreateExtendedLog 创建扩展日志记录
func CreateExtendedLog(ctx context.Context, logId int, extendedInfo *ExtendedLogInfo) error {
	if extendedInfo == nil {
		return nil
	}

	extendedLog := &ExtendedLog{
		LogId:       logId,
		SchoolId:    extendedInfo.SchoolId,
		SchoolName:  extendedInfo.SchoolName,
		SubjectId:   extendedInfo.SubjectId,
		SubjectName: extendedInfo.SubjectName,
		TeacherId:   extendedInfo.TeacherId,
		TeacherName: extendedInfo.TeacherName,
		GroupName:   extendedInfo.GroupName,
		CreatedAt:   helper.GetTimestamp(),
	}

	err := DB.WithContext(ctx).Create(extendedLog).Error
	if err != nil {
		logger.Errorf(ctx, "创建扩展日志失败: logId=%d, error=%v", logId, err)
		return err
	}

	logger.Debugf(ctx, "扩展日志创建成功: logId=%d, teacherId=%s, subjectId=%d",
		logId, extendedInfo.TeacherId, extendedInfo.SubjectId)
	return nil
}

// GetExtendedLogByLogId 根据日志ID获取扩展日志
func GetExtendedLogByLogId(ctx context.Context, logId int) (*ExtendedLog, error) {
	var extendedLog ExtendedLog
	err := DB.WithContext(ctx).Where("log_id = ?", logId).First(&extendedLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &extendedLog, nil
}

// GetCompleteLogInfo 获取完整的日志信息（原日志 + 扩展信息）
func GetCompleteLogInfo(ctx context.Context, logId int) (*CompleteLogInfo, error) {
	var result CompleteLogInfo

	// 1. 获取原日志信息
	err := DB.WithContext(ctx).Where("id = ?", logId).First(&result.Log).Error
	if err != nil {
		return nil, err
	}

	// 2. 获取扩展日志信息（可能不存在）
	extendedLog, err := GetExtendedLogByLogId(ctx, logId)
	if err != nil {
		return nil, err
	}
	if extendedLog != nil {
		result.ExtendedLog = *extendedLog
	}

	return &result, nil
}

// GetExtendedLogsByTeacherId 根据老师ID获取扩展日志列表
func GetExtendedLogsByTeacherId(ctx context.Context, teacherId string, startIdx, num int) ([]*ExtendedLog, error) {
	var logs []*ExtendedLog
	err := DB.WithContext(ctx).
		Where("teacher_id = ?", teacherId).
		Order("id desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error
	return logs, err
}

// GetExtendedLogsBySubjectId 根据学科组ID获取扩展日志列表
func GetExtendedLogsBySubjectId(ctx context.Context, subjectId int, startIdx, num int) ([]*ExtendedLog, error) {
	var logs []*ExtendedLog
	err := DB.WithContext(ctx).
		Where("subject_id = ?", subjectId).
		Order("id desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error
	return logs, err
}

// GetExtendedLogsBySchoolId 根据学校ID获取扩展日志列表
func GetExtendedLogsBySchoolId(ctx context.Context, schoolId int, startIdx, num int) ([]*ExtendedLog, error) {
	var logs []*ExtendedLog
	err := DB.WithContext(ctx).
		Where("school_id = ?", schoolId).
		Order("id desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error
	return logs, err
}

// DeleteExtendedLogsByLogIds 根据日志ID列表删除扩展日志
func DeleteExtendedLogsByLogIds(ctx context.Context, logIds []int) error {
	if len(logIds) == 0 {
		return nil
	}

	result := DB.WithContext(ctx).Where("log_id IN ?", logIds).Delete(&ExtendedLog{})
	return result.Error
}

// DeleteOldExtendedLogs 删除旧的扩展日志记录
func DeleteOldExtendedLogs(ctx context.Context, targetTimestamp int64) (int64, error) {
	result := DB.WithContext(ctx).Where("created_at < ?", targetTimestamp).Delete(&ExtendedLog{})
	return result.RowsAffected, result.Error
}
