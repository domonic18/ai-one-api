package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"gorm.io/gorm"
)

// ExtendedLog 扩展日志数据结构
// 符合实现方案v3.0版本设计，使用JSON字段存储维度信息
type ExtendedLog struct {
	Id             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	LogId          int64     `json:"log_id" gorm:"not null;index;uniqueIndex"`        // 关联OneAPI原始日志ID
	ExternalUserId string    `json:"external_user_id" gorm:"type:varchar(100);index"` // 外部用户ID（如teacher_id）
	UserGroup      string    `json:"user_group" gorm:"type:varchar(100);index"`       // OneAPI用户组
	DimensionInfo  string    `json:"dimension_info" gorm:"type:json"`                 // 多维度统计维度信息（JSON格式）
	CreatedAt      time.Time `json:"created_at"`                                      // 创建时间
	UpdatedAt      time.Time `json:"updated_at"`                                      // 更新时间
}

// TableName 指定表名
func (ExtendedLog) TableName() string {
	return "extended_logs"
}

// DimensionInfo 维度信息结构（灵活的JSON格式）
type DimensionInfo struct {
	SchoolId    int    `json:"school_id,omitempty"`
	SchoolName  string `json:"school_name,omitempty"`
	SubjectId   int    `json:"subject_id,omitempty"`
	SubjectName string `json:"subject_name,omitempty"`
	TeacherName string `json:"teacher_name,omitempty"`

	// 可扩展其他维度信息
	Department string `json:"department,omitempty"` // 部门（企业场景）
	Project    string `json:"project,omitempty"`    // 项目（项目场景）
	Region     string `json:"region,omitempty"`     // 地区
}

// CreateExtendedLog 创建扩展日志
func CreateExtendedLog(ctx context.Context, logId int64, externalUserId string, userGroup string, dimensionInfo *DimensionInfo) (*ExtendedLog, error) {
	// 序列化维度信息
	var dimensionBytes []byte
	var err error
	if dimensionInfo != nil {
		dimensionBytes, err = json.Marshal(dimensionInfo)
		if err != nil {
			logger.Warnf(ctx, "序列化维度信息失败: externalUserId=%s, error=%v", externalUserId, err)
			return nil, fmt.Errorf("序列化维度信息失败: %w", err)
		}
	}

	// 创建扩展日志
	extendedLog := &ExtendedLog{
		LogId:          logId,
		ExternalUserId: externalUserId,
		UserGroup:      userGroup,
		DimensionInfo:  string(dimensionBytes),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存到数据库
	err = model.DB.WithContext(ctx).Create(extendedLog).Error
	if err != nil {
		logger.Errorf(ctx, "创建扩展日志失败: logId=%d, externalUserId=%s, error=%v", logId, externalUserId, err)
		return nil, fmt.Errorf("创建扩展日志失败: %w", err)
	}

	logger.Debugf(ctx, "创建扩展日志成功: logId=%d, externalUserId=%s, userGroup=%s", logId, externalUserId, userGroup)
	return extendedLog, nil
}

// GetExtendedLogByLogId 根据日志ID获取扩展日志
func GetExtendedLogByLogId(ctx context.Context, logId int64) (*ExtendedLog, error) {
	var extendedLog ExtendedLog
	err := model.DB.WithContext(ctx).Where("log_id = ?", logId).First(&extendedLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取扩展日志失败: %w", err)
	}
	return &extendedLog, nil
}

// GetExtendedLogsByExternalUserId 根据外部用户ID获取扩展日志（分页）
func GetExtendedLogsByExternalUserId(ctx context.Context, externalUserId string, page, pageSize int) ([]*ExtendedLog, int64, error) {
	var extendedLogs []*ExtendedLog
	var total int64

	// 计算总数
	err := model.DB.WithContext(ctx).Model(&ExtendedLog{}).Where("external_user_id = ?", externalUserId).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取扩展日志总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = model.DB.WithContext(ctx).Where("external_user_id = ?", externalUserId).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&extendedLogs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取扩展日志失败: %w", err)
	}

	return extendedLogs, total, nil
}

// GetExtendedLogsByUserGroup 根据用户组获取扩展日志（分页）
func GetExtendedLogsByUserGroup(ctx context.Context, userGroup string, page, pageSize int) ([]*ExtendedLog, int64, error) {
	var extendedLogs []*ExtendedLog
	var total int64

	// 计算总数
	err := model.DB.WithContext(ctx).Model(&ExtendedLog{}).Where("user_group = ?", userGroup).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户组扩展日志总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = model.DB.WithContext(ctx).Where("user_group = ?", userGroup).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&extendedLogs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户组扩展日志失败: %w", err)
	}

	return extendedLogs, total, nil
}

// DeleteExtendedLogsByLogIds 根据日志ID列表删除扩展日志
func DeleteExtendedLogsByLogIds(ctx context.Context, logIds []int64) error {
	if len(logIds) == 0 {
		return nil
	}

	err := model.DB.WithContext(ctx).Where("log_id IN ?", logIds).Delete(&ExtendedLog{}).Error
	if err != nil {
		return fmt.Errorf("删除扩展日志失败: %w", err)
	}

	logger.Debugf(ctx, "删除扩展日志成功: 删除数量=%d", len(logIds))
	return nil
}

// GetDimensionInfo 获取维度信息
func (el *ExtendedLog) GetDimensionInfo() (*DimensionInfo, error) {
	if el.DimensionInfo == "" {
		return nil, nil
	}

	var dimensionInfo DimensionInfo
	err := json.Unmarshal([]byte(el.DimensionInfo), &dimensionInfo)
	if err != nil {
		return nil, fmt.Errorf("解析维度信息失败: %w", err)
	}

	return &dimensionInfo, nil
}

// RecordExtendedLog 异步记录扩展日志
// 根据实现方案v3.0版本设计，从Redis缓存获取用户信息
func RecordExtendedLog(ctx context.Context, logId int64, externalUserId string) {
	if externalUserId == "" {
		return // 没有外部用户ID，跳过扩展日志
	}

	// 异步记录扩展日志，不影响主流程
	go func() {
		// TODO: 在迭代二中实现从Redis缓存获取用户信息的逻辑
		// 这里先使用简化实现，仅记录基本信息
		userGroup := "default" // 临时使用默认值
		dimensionInfo := &DimensionInfo{
			TeacherName: externalUserId, // 临时使用用户ID作为教师名称
		}

		_, err := CreateExtendedLog(ctx, logId, externalUserId, userGroup, dimensionInfo)
		if err != nil {
			logger.Warnf(ctx, "记录扩展日志失败: logId=%d, externalUserId=%s, error=%v", logId, externalUserId, err)
		} else {
			logger.Debugf(ctx, "扩展日志记录成功: logId=%d, externalUserId=%s", logId, externalUserId)
		}
	}()
}
