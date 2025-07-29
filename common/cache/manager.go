package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// Manager 简化的缓存管理器
type Manager struct {
	mutex   sync.RWMutex
	enabled bool
	stats   *SimpleStats
}

// SimpleStats 简化的统计信息
type SimpleStats struct {
	HitCount   int64   `json:"hit_count"`
	MissCount  int64   `json:"miss_count"`
	TotalCount int64   `json:"total_count"`
	HitRate    float64 `json:"hit_rate"`
}

// BatchResult 批量操作结果
type BatchResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedKeys   []string `json:"failed_keys"`
}

var Mgr *Manager

// InitManager 初始化缓存管理器
func InitManager() {
	Mgr = &Manager{
		enabled: common.RedisEnabled,
		stats:   &SimpleStats{},
	}
	logger.SysLog("缓存管理器初始化完成")
}

// recordHit 记录缓存命中
func (cm *Manager) recordHit() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.stats.HitCount++
	cm.stats.TotalCount++
	if cm.stats.TotalCount > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(cm.stats.TotalCount) * 100
	}
}

// recordMiss 记录缓存未命中
func (cm *Manager) recordMiss() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.stats.MissCount++
	cm.stats.TotalCount++
	if cm.stats.TotalCount > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(cm.stats.TotalCount) * 100
	}
}

// GetStats 获取统计信息
func (cm *Manager) GetStats() *SimpleStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return &SimpleStats{
		HitCount:   cm.stats.HitCount,
		MissCount:  cm.stats.MissCount,
		TotalCount: cm.stats.TotalCount,
		HitRate:    cm.stats.HitRate,
	}
}

// Set 设置缓存项
func (cm *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !cm.enabled {
		return fmt.Errorf("缓存未启用")
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	err = common.RedisSet(key, string(valueBytes), expiration)
	if err != nil {
		logger.Debugf(ctx, "设置缓存失败: key=%s, error=%v", key, err)
		return err
	}

	logger.Debugf(ctx, "缓存设置成功: key=%s", key)
	return nil
}

// Get 获取缓存项
func (cm *Manager) Get(ctx context.Context, key string, result interface{}) error {
	if !cm.enabled {
		cm.recordMiss()
		return fmt.Errorf("缓存未启用")
	}

	valueStr, err := common.RedisGet(key)
	if err != nil {
		cm.recordMiss()
		logger.Debugf(ctx, "缓存未命中: key=%s", key)
		return err
	}

	err = json.Unmarshal([]byte(valueStr), result)
	if err != nil {
		cm.recordMiss()
		logger.Debugf(ctx, "反序列化失败: key=%s, error=%v", key, err)
		return err
	}

	cm.recordHit()
	logger.Debugf(ctx, "缓存命中: key=%s", key)
	return nil
}

// Delete 删除缓存项
func (cm *Manager) Delete(ctx context.Context, key string) error {
	if !cm.enabled {
		return fmt.Errorf("缓存未启用")
	}

	err := common.RedisDel(key)
	if err != nil {
		logger.Debugf(ctx, "删除缓存失败: key=%s, error=%v", key, err)
		return err
	}

	logger.Debugf(ctx, "缓存删除成功: key=%s", key)
	return nil
}

// GetWithFallback 获取缓存，如果不存在则执行回调函数并缓存结果
func (cm *Manager) GetWithFallback(ctx context.Context, key string, fallbackFunc func() (interface{}, error), ttl time.Duration, result interface{}) error {
	// 1. 尝试从缓存获取
	err := cm.Get(ctx, key, result)
	if err == nil {
		return nil
	}

	// 2. 缓存未命中，执行回调函数
	value, err := fallbackFunc()
	if err != nil {
		return err
	}

	// 3. 设置缓存
	if err := cm.Set(ctx, key, value, ttl); err != nil {
		logger.Warnf(ctx, "设置缓存失败: key=%s, error=%v", key, err)
	}

	// 4. 将结果复制到result
	valueBytes, _ := json.Marshal(value)
	return json.Unmarshal(valueBytes, result)
}

// BatchDelete 批量删除缓存项
func (cm *Manager) BatchDelete(ctx context.Context, keys []string) *BatchResult {
	result := &BatchResult{
		FailedKeys: make([]string, 0),
	}

	if !cm.enabled {
		result.FailedCount = len(keys)
		result.FailedKeys = keys
		return result
	}

	for _, key := range keys {
		err := cm.Delete(ctx, key)
		if err != nil {
			result.FailedCount++
			result.FailedKeys = append(result.FailedKeys, key)
		} else {
			result.SuccessCount++
		}
	}

	logger.Debugf(ctx, "批量删除缓存完成: 成功=%d, 失败=%d", result.SuccessCount, result.FailedCount)
	return result
}

// InvalidatePattern 失效匹配模式的缓存键（简化版）
func (cm *Manager) InvalidatePattern(ctx context.Context, pattern string) error {
	if !cm.enabled {
		return fmt.Errorf("缓存未启用")
	}

	// 简化实现：只支持基础的模式匹配
	// 实际项目中可以根据需要扩展
	logger.Debugf(ctx, "模式失效缓存: pattern=%s", pattern)

	// 这里可以实现具体的模式匹配逻辑
	// 当前简化版本只记录日志
	return nil
}

// IsEnabled 检查缓存是否启用
func (cm *Manager) IsEnabled() bool {
	return cm.enabled
}

// Exists 检查缓存项是否存在
func (cm *Manager) Exists(ctx context.Context, key string) bool {
	if !cm.enabled {
		return false
	}

	_, err := common.RedisGet(key)
	exists := err == nil
	logger.Debugf(ctx, "缓存存在性检查: key=%s, exists=%v", key, exists)
	return exists
}

// GetMultiple 批量获取缓存项（简化版）
func (cm *Manager) GetMultiple(ctx context.Context, keys []string) map[string]interface{} {
	result := make(map[string]interface{})

	if !cm.enabled {
		return result
	}

	for _, key := range keys {
		var value interface{}
		err := cm.Get(ctx, key, &value)
		if err == nil {
			result[key] = value
		}
	}

	logger.Debugf(ctx, "批量获取缓存完成: 请求=%d, 命中=%d", len(keys), len(result))
	return result
}
