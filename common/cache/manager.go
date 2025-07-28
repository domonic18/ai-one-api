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

// Manager 缓存管理器
type Manager struct {
	mutex   sync.RWMutex
	stats   *Stats
	enabled bool
}

// Stats 缓存统计信息
type Stats struct {
	HitCount      int64   `json:"hit_count"`
	MissCount     int64   `json:"miss_count"`
	TotalCount    int64   `json:"total_count"`
	HitRate       float64 `json:"hit_rate"`
	LastResetTime int64   `json:"last_reset_time"`
}

// Item 缓存项结构
type Item struct {
	Key        string        `json:"key"`
	Value      interface{}   `json:"value"`
	Expiration time.Duration `json:"expiration"`
	CreatedAt  time.Time     `json:"created_at"`
}

// BatchResult 批量缓存操作结果
type BatchResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedKeys   []string `json:"failed_keys"`
}

var Mgr *Manager

// InitManager 初始化缓存管理器
func InitManager() {
	Mgr = &Manager{
		stats: &Stats{
			LastResetTime: time.Now().Unix(),
		},
		enabled: common.RedisEnabled,
	}

	// 启动统计更新协程
	go Mgr.updateStatsRoutine()

	logger.SysLog("缓存管理器初始化完成")
}

// updateStatsRoutine 定期更新缓存统计信息
func (cm *Manager) updateStatsRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cm.updateHitRate()
	}
}

// updateHitRate 更新缓存命中率
func (cm *Manager) updateHitRate() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.stats.TotalCount > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(cm.stats.TotalCount) * 100
	}
}

// recordHit 记录缓存命中
func (cm *Manager) recordHit() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.stats.HitCount++
	cm.stats.TotalCount++
}

// recordMiss 记录缓存未命中
func (cm *Manager) recordMiss() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.stats.MissCount++
	cm.stats.TotalCount++
}

// GetStats 获取缓存统计信息
func (cm *Manager) GetStats() *Stats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建副本以避免并发读写
	statsCopy := &Stats{
		HitCount:      cm.stats.HitCount,
		MissCount:     cm.stats.MissCount,
		TotalCount:    cm.stats.TotalCount,
		HitRate:       cm.stats.HitRate,
		LastResetTime: cm.stats.LastResetTime,
	}

	return statsCopy
}

// ResetStats 重置缓存统计信息
func (cm *Manager) ResetStats() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.stats.HitCount = 0
	cm.stats.MissCount = 0
	cm.stats.TotalCount = 0
	cm.stats.HitRate = 0
	cm.stats.LastResetTime = time.Now().Unix()

	logger.SysLog("缓存统计信息已重置")
}

// Set 设置缓存项（带统计）
func (cm *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !cm.enabled {
		return fmt.Errorf("缓存未启用")
	}

	// 序列化值
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	// 设置到Redis
	err = common.RedisSet(key, string(valueBytes), expiration)
	if err != nil {
		logger.Errorf(ctx, "设置缓存失败: key=%s, error=%v", key, err)
		return err
	}

	logger.Debugf(ctx, "缓存设置成功: key=%s, ttl=%v", key, expiration)
	return nil
}

// Get 获取缓存项（带统计）
func (cm *Manager) Get(ctx context.Context, key string, result interface{}) error {
	if !cm.enabled {
		cm.recordMiss()
		return fmt.Errorf("缓存未启用")
	}

	// 从Redis获取
	valueStr, err := common.RedisGet(key)
	if err != nil {
		cm.recordMiss()
		logger.Debugf(ctx, "缓存未命中: key=%s", key)
		return err
	}

	// 反序列化
	err = json.Unmarshal([]byte(valueStr), result)
	if err != nil {
		cm.recordMiss()
		logger.Warnf(ctx, "缓存值反序列化失败: key=%s, error=%v", key, err)
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
		logger.Warnf(ctx, "删除缓存失败: key=%s, error=%v", key, err)
		return err
	}

	logger.Debugf(ctx, "缓存删除成功: key=%s", key)
	return nil
}

// BatchSet 批量设置缓存项
func (cm *Manager) BatchSet(ctx context.Context, items []Item) *BatchResult {
	result := &BatchResult{
		FailedKeys: make([]string, 0),
	}

	if !cm.enabled {
		result.FailedCount = len(items)
		for _, item := range items {
			result.FailedKeys = append(result.FailedKeys, item.Key)
		}
		return result
	}

	for _, item := range items {
		err := cm.Set(ctx, item.Key, item.Value, item.Expiration)
		if err != nil {
			result.FailedCount++
			result.FailedKeys = append(result.FailedKeys, item.Key)
		} else {
			result.SuccessCount++
		}
	}

	logger.Debugf(ctx, "批量设置缓存完成: 成功=%d, 失败=%d", result.SuccessCount, result.FailedCount)
	return result
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

// GetMultiple 批量获取缓存项
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

// GetWithDefault 获取缓存项，如果不存在则返回默认值
func (cm *Manager) GetWithDefault(ctx context.Context, key string, defaultValue interface{}, result interface{}) error {
	err := cm.Get(ctx, key, result)
	if err != nil {
		// 使用默认值
		if defaultValue != nil {
			// 将默认值赋给result
			defaultBytes, marshalErr := json.Marshal(defaultValue)
			if marshalErr != nil {
				return marshalErr
			}

			unmarshalErr := json.Unmarshal(defaultBytes, result)
			if unmarshalErr != nil {
				return unmarshalErr
			}
		}

		logger.Debugf(ctx, "缓存未命中，使用默认值: key=%s", key)
		return err
	}

	return nil
}

// Refresh 刷新缓存项（重新设置TTL）
func (cm *Manager) Refresh(ctx context.Context, key string, expiration time.Duration) error {
	if !cm.enabled {
		return fmt.Errorf("缓存未启用")
	}

	// 先获取当前值
	currentValue, err := common.RedisGet(key)
	if err != nil {
		return fmt.Errorf("获取缓存值失败: %w", err)
	}

	// 重新设置相同的值但更新TTL
	err = common.RedisSet(key, currentValue, expiration)
	if err != nil {
		return fmt.Errorf("刷新缓存失败: %w", err)
	}

	logger.Debugf(ctx, "缓存刷新成功: key=%s, new_ttl=%v", key, expiration)
	return nil
}

// IsEnabled 检查缓存是否启用
func (cm *Manager) IsEnabled() bool {
	return cm.enabled
}

// EnableCache 启用缓存
func (cm *Manager) EnableCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.enabled = true
	logger.SysLog("缓存已启用")
}

// DisableCache 禁用缓存
func (cm *Manager) DisableCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.enabled = false
	logger.SysLog("缓存已禁用")
}
