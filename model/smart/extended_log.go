package smart

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"gorm.io/gorm"
)

// ExtendedLog 扩展日志结构体
type ExtendedLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	LogID       int64     `json:"log_id" gorm:"uniqueIndex"` // 关联原始日志ID
	TeacherID   string    `json:"teacher_id"`                // 老师ID
	SchoolID    int       `json:"school_id"`                 // 学校ID
	SchoolName  string    `json:"school_name"`               // 学校名称
	SubjectID   int       `json:"subject_id"`                // 学科组ID
	SubjectName string    `json:"subject_name"`              // 学科组名称
	CreatedAt   time.Time `json:"created_at"`                // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                // 更新时间
}

// TableName 指定表名
func (ExtendedLog) TableName() string {
	return "extended_logs"
}

// CreateExtendedLog 创建扩展日志
func CreateExtendedLog(ctx context.Context, logID int64, teacherID string, teacherInfo *TeacherInfo) (*ExtendedLog, error) {
	// 创建扩展日志
	extendedLog := &ExtendedLog{
		LogID:     logID,
		TeacherID: teacherID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果有老师信息，则填充学校和学科组信息
	if teacherInfo != nil {
		extendedLog.SchoolID = teacherInfo.SchoolId
		extendedLog.SchoolName = teacherInfo.SchoolName
		extendedLog.SubjectID = teacherInfo.SubjectId
		extendedLog.SubjectName = teacherInfo.SubjectName
	}

	// 保存到数据库
	err := model.DB.Create(extendedLog).Error
	if err != nil {
		logger.Errorf(ctx, "创建扩展日志失败: logID=%d, error=%v", logID, err)
		return nil, err
	}

	logger.Debugf(ctx, "创建扩展日志成功: logID=%d, teacherID=%s", logID, teacherID)
	return extendedLog, nil
}

// GetExtendedLogByLogID 根据日志ID获取扩展日志
func GetExtendedLogByLogID(ctx context.Context, logID int64) (*ExtendedLog, error) {
	var extendedLog ExtendedLog
	err := model.DB.Where("log_id = ?", logID).First(&extendedLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Debugf(ctx, "扩展日志不存在: logID=%d", logID)
			return nil, nil
		}
		logger.Errorf(ctx, "获取扩展日志失败: logID=%d, error=%v", logID, err)
		return nil, err
	}
	return &extendedLog, nil
}

// GetExtendedLogsByTeacherID 根据老师ID获取扩展日志
func GetExtendedLogsByTeacherID(ctx context.Context, teacherID string, page, pageSize int) ([]ExtendedLog, int64, error) {
	var extendedLogs []ExtendedLog
	var total int64

	// 计算总数
	err := model.DB.Model(&ExtendedLog{}).Where("teacher_id = ?", teacherID).Count(&total).Error
	if err != nil {
		logger.Errorf(ctx, "获取老师扩展日志总数失败: teacherID=%s, error=%v", teacherID, err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = model.DB.Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&extendedLogs).Error
	if err != nil {
		logger.Errorf(ctx, "获取老师扩展日志失败: teacherID=%s, error=%v", teacherID, err)
		return nil, 0, err
	}

	return extendedLogs, total, nil
}

// GetExtendedLogsBySubjectID 根据学科组ID获取扩展日志
func GetExtendedLogsBySubjectID(ctx context.Context, subjectID int, page, pageSize int) ([]ExtendedLog, int64, error) {
	var extendedLogs []ExtendedLog
	var total int64

	// 计算总数
	err := model.DB.Model(&ExtendedLog{}).Where("subject_id = ?", subjectID).Count(&total).Error
	if err != nil {
		logger.Errorf(ctx, "获取学科组扩展日志总数失败: subjectID=%d, error=%v", subjectID, err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = model.DB.Where("subject_id = ?", subjectID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&extendedLogs).Error
	if err != nil {
		logger.Errorf(ctx, "获取学科组扩展日志失败: subjectID=%d, error=%v", subjectID, err)
		return nil, 0, err
	}

	return extendedLogs, total, nil
}

// GetExtendedLogsBySchoolID 根据学校ID获取扩展日志
func GetExtendedLogsBySchoolID(ctx context.Context, schoolID int, page, pageSize int) ([]ExtendedLog, int64, error) {
	var extendedLogs []ExtendedLog
	var total int64

	// 计算总数
	err := model.DB.Model(&ExtendedLog{}).Where("school_id = ?", schoolID).Count(&total).Error
	if err != nil {
		logger.Errorf(ctx, "获取学校扩展日志总数失败: schoolID=%d, error=%v", schoolID, err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = model.DB.Where("school_id = ?", schoolID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&extendedLogs).Error
	if err != nil {
		logger.Errorf(ctx, "获取学校扩展日志失败: schoolID=%d, error=%v", schoolID, err)
		return nil, 0, err
	}

	return extendedLogs, total, nil
}

// DeleteExtendedLogsByLogIDs 根据日志ID列表删除扩展日志
func DeleteExtendedLogsByLogIDs(ctx context.Context, logIDs []int64) error {
	if len(logIDs) == 0 {
		return fmt.Errorf("日志ID列表为空")
	}

	err := model.DB.Where("log_id IN ?", logIDs).Delete(&ExtendedLog{}).Error
	if err != nil {
		logger.Errorf(ctx, "删除扩展日志失败: logIDs=%v, error=%v", logIDs, err)
		return err
	}

	logger.Debugf(ctx, "删除扩展日志成功: logIDs=%v", logIDs)
	return nil
}

// GetFullLogInfo 获取完整日志信息（原始日志 + 扩展日志）
func GetFullLogInfo(ctx context.Context, logID int64) (*model.Log, *ExtendedLog, error) {
	// 获取原始日志
	var log model.Log
	err := model.DB.First(&log, logID).Error
	if err != nil {
		logger.Errorf(ctx, "获取原始日志失败: logID=%d, error=%v", logID, err)
		return nil, nil, err
	}

	// 获取扩展日志
	extendedLog, err := GetExtendedLogByLogID(ctx, logID)
	if err != nil {
		logger.Errorf(ctx, "获取扩展日志失败: logID=%d, error=%v", logID, err)
		return &log, nil, err
	}

	return &log, extendedLog, nil
}
