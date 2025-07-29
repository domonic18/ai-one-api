package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// SimpleInvalidator 简化的缓存失效管理器
type SimpleInvalidator struct {
	// 移除复杂的模式和日志记录
}

var Invalidator *SimpleInvalidator

// InitInvalidationManager 初始化简化的缓存失效管理器
func InitInvalidationManager() {
	Invalidator = &SimpleInvalidator{}
	logger.SysLog("缓存失效管理器初始化完成")
}

// InvalidateByKeys 根据键列表失效缓存
func (si *SimpleInvalidator) InvalidateByKeys(ctx context.Context, keys []string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	if Mgr == nil {
		return fmt.Errorf("缓存管理器未初始化")
	}

	if len(keys) == 0 {
		return nil
	}

	result := Mgr.BatchDelete(ctx, keys)

	logger.Debugf(ctx, "批量失效缓存完成: 成功=%d, 失败=%d", result.SuccessCount, result.FailedCount)

	if result.FailedCount > 0 {
		return fmt.Errorf("部分缓存失效失败: 失败数量=%d", result.FailedCount)
	}

	return nil
}

// InvalidateUserCache 失效用户相关缓存
func (si *SimpleInvalidator) InvalidateUserCache(ctx context.Context, userId string) error {
	if userId == "" {
		return fmt.Errorf("userId不能为空")
	}

	keys := []string{
		fmt.Sprintf("user_config:%s", userId),
		fmt.Sprintf("teacher_info:%s", userId),
	}

	return si.InvalidateByKeys(ctx, keys)
}

// InvalidateSubjectCache 失效学科组相关缓存
func (si *SimpleInvalidator) InvalidateSubjectCache(ctx context.Context, subjectId int) error {
	if subjectId <= 0 {
		return fmt.Errorf("subjectId必须大于0")
	}

	keys := []string{
		fmt.Sprintf("subject_info:%d", subjectId),
	}

	return si.InvalidateByKeys(ctx, keys)
}

// RefreshCache 刷新缓存（删除旧缓存，设置新缓存）
func (si *SimpleInvalidator) RefreshCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	if Mgr == nil {
		return fmt.Errorf("缓存管理器未初始化")
	}

	// 1. 删除旧缓存（忽略错误）
	_ = Mgr.Delete(ctx, key)

	// 2. 设置新缓存
	err := Mgr.Set(ctx, key, value, ttl)
	if err != nil {
		return fmt.Errorf("设置新缓存失败: %w", err)
	}

	logger.Debugf(ctx, "缓存刷新成功: key=%s", key)
	return nil
}
