package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// InvalidationManager 缓存失效管理器
type InvalidationManager struct {
	mutex           sync.RWMutex
	invalidationLog map[string]int64 // 记录缓存失效时间戳
	patterns        []InvalidationPattern
}

// InvalidationPattern 缓存失效模式
type InvalidationPattern struct {
	Pattern     string        `json:"pattern"`     // 缓存键模式（支持通配符）
	TTL         time.Duration `json:"ttl"`         // 生存时间
	Description string        `json:"description"` // 描述
}

// InvalidationEvent 缓存失效事件
type InvalidationEvent struct {
	EventType  string            `json:"event_type"`  // 事件类型
	ObjectType string            `json:"object_type"` // 对象类型
	ObjectID   string            `json:"object_id"`   // 对象ID
	ChangeType string            `json:"change_type"` // 变更类型
	Timestamp  int64             `json:"timestamp"`   // 时间戳
	Metadata   map[string]string `json:"metadata"`    // 元数据
}

var Invalidator *InvalidationManager

// InitInvalidationManager 初始化缓存失效管理器
func InitInvalidationManager() {
	Invalidator = &InvalidationManager{
		invalidationLog: make(map[string]int64),
		patterns: []InvalidationPattern{
			{
				Pattern:     "user_model_config:*",
				TTL:         24 * time.Hour,
				Description: "用户模型配置缓存",
			},
			{
				Pattern:     "teacher_info:*",
				TTL:         24 * time.Hour,
				Description: "老师信息缓存",
			},
			{
				Pattern:     "subject_info:*",
				TTL:         24 * time.Hour,
				Description: "学科组信息缓存",
			},
		},
	}

	// 启动定期清理过期失效记录的协程
	go Invalidator.cleanupExpiredRecords()

	logger.SysLog("缓存失效管理器初始化完成")
}

// cleanupExpiredRecords 定期清理过期的失效记录
func (cim *InvalidationManager) cleanupExpiredRecords() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cim.mutex.Lock()
		now := time.Now().Unix()
		expiredKeys := make([]string, 0)

		for key, timestamp := range cim.invalidationLog {
			// 如果记录超过24小时，则删除
			if now-timestamp > 24*3600 {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(cim.invalidationLog, key)
		}

		if len(expiredKeys) > 0 {
			logger.Debugf(context.Background(), "清理过期失效记录: count=%d", len(expiredKeys))
		}
		cim.mutex.Unlock()
	}
}

// InvalidateByPattern 根据模式失效缓存
func (cim *InvalidationManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 获取匹配的键
	keys, err := cim.getKeysByPattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("获取匹配键失败: %w", err)
	}

	if len(keys) == 0 {
		logger.Debugf(ctx, "没有找到匹配的缓存键: pattern=%s", pattern)
		return nil
	}

	// 批量删除
	result := Mgr.BatchDelete(ctx, keys)

	// 记录失效操作
	cim.mutex.Lock()
	cim.invalidationLog[pattern] = time.Now().Unix()
	cim.mutex.Unlock()

	logger.Debugf(ctx, "按模式失效缓存完成: pattern=%s, 成功=%d, 失败=%d",
		pattern, result.SuccessCount, result.FailedCount)

	return nil
}

// getKeysByPattern 根据模式获取键列表（简化实现）
func (cim *InvalidationManager) getKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	// 这里是简化实现，实际应该使用Redis的SCAN命令
	// 由于Redis客户端限制，这里返回空列表
	// 在实际实现中，应该使用Redis的KEYS或SCAN命令
	logger.Warnf(ctx, "getKeysByPattern简化实现，实际应使用Redis SCAN: pattern=%s", pattern)
	return []string{}, nil
}

// InvalidateByEvent 根据事件失效相关缓存
func (cim *InvalidationManager) InvalidateByEvent(ctx context.Context, event InvalidationEvent) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	var keysToInvalidate []string

	switch event.ObjectType {
	case "teacher":
		// 老师信息变更，失效相关缓存
		teacherKey := fmt.Sprintf("teacher_info:%s", event.ObjectID)
		userConfigKey := fmt.Sprintf("user_model_config:%s", event.ObjectID)
		keysToInvalidate = append(keysToInvalidate, teacherKey, userConfigKey)

	case "subject":
		// 学科组信息变更，失效相关缓存
		subjectKey := fmt.Sprintf("subject_info:%s", event.ObjectID)
		keysToInvalidate = append(keysToInvalidate, subjectKey)

	case "school":
		// 学校信息变更，可能需要失效相关的学科组和老师缓存
		// 这里需要根据实际业务逻辑实现
		logger.Debugf(ctx, "学校信息变更事件: schoolId=%s", event.ObjectID)

	default:
		logger.Warnf(ctx, "未知的对象类型: %s", event.ObjectType)
		return nil
	}

	if len(keysToInvalidate) > 0 {
		result := Mgr.BatchDelete(ctx, keysToInvalidate)

		// 记录失效操作
		cim.mutex.Lock()
		eventKey := fmt.Sprintf("%s:%s:%s", event.ObjectType, event.ChangeType, event.ObjectID)
		cim.invalidationLog[eventKey] = event.Timestamp
		cim.mutex.Unlock()

		logger.Debugf(ctx, "根据事件失效缓存完成: type=%s, id=%s, 成功=%d, 失败=%d",
			event.ObjectType, event.ObjectID, result.SuccessCount, result.FailedCount)
	}

	return nil
}

// InvalidateRelatedCaches 失效相关缓存
func (cim *InvalidationManager) InvalidateRelatedCaches(ctx context.Context, objectType, objectID string) error {
	event := InvalidationEvent{
		EventType:  "manual_invalidation",
		ObjectType: objectType,
		ObjectID:   objectID,
		ChangeType: "update",
		Timestamp:  time.Now().Unix(),
	}

	return cim.InvalidateByEvent(ctx, event)
}

// RefreshCache 刷新缓存（先失效再预热）
func (cim *InvalidationManager) RefreshCache(ctx context.Context, cacheKey string, refreshFunc func() (interface{}, error), ttl time.Duration) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	// 1. 删除旧缓存
	err := Mgr.Delete(ctx, cacheKey)
	if err != nil {
		logger.Warnf(ctx, "删除旧缓存失败: key=%s, error=%v", cacheKey, err)
	}

	// 2. 获取新数据
	newValue, err := refreshFunc()
	if err != nil {
		return fmt.Errorf("获取新数据失败: %w", err)
	}

	// 3. 设置新缓存
	err = Mgr.Set(ctx, cacheKey, newValue, ttl)
	if err != nil {
		return fmt.Errorf("设置新缓存失败: %w", err)
	}

	logger.Debugf(ctx, "缓存刷新成功: key=%s", cacheKey)
	return nil
}

// GetInvalidationStats 获取失效统计信息
func (cim *InvalidationManager) GetInvalidationStats() map[string]interface{} {
	cim.mutex.RLock()
	defer cim.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_invalidations":  len(cim.invalidationLog),
		"patterns":             cim.patterns,
		"recent_invalidations": make([]map[string]interface{}, 0),
	}

	// 获取最近的失效记录
	now := time.Now().Unix()
	recentCount := 0
	for key, timestamp := range cim.invalidationLog {
		if now-timestamp < 3600 && recentCount < 10 { // 最近1小时内的，最多10条
			stats["recent_invalidations"] = append(stats["recent_invalidations"].([]map[string]interface{}), map[string]interface{}{
				"key":       key,
				"timestamp": timestamp,
				"age":       now - timestamp,
			})
			recentCount++
		}
	}

	return stats
}

// AddInvalidationPattern 添加失效模式
func (cim *InvalidationManager) AddInvalidationPattern(pattern InvalidationPattern) {
	cim.mutex.Lock()
	defer cim.mutex.Unlock()

	cim.patterns = append(cim.patterns, pattern)
	logger.Debugf(context.Background(), "添加缓存失效模式: pattern=%s, ttl=%v", pattern.Pattern, pattern.TTL)
}

// RemoveInvalidationPattern 移除失效模式
func (cim *InvalidationManager) RemoveInvalidationPattern(pattern string) {
	cim.mutex.Lock()
	defer cim.mutex.Unlock()

	for i, p := range cim.patterns {
		if p.Pattern == pattern {
			cim.patterns = append(cim.patterns[:i], cim.patterns[i+1:]...)
			logger.Debugf(context.Background(), "移除缓存失效模式: pattern=%s", pattern)
			break
		}
	}
}

// ScheduledInvalidation 定时失效缓存
func (cim *InvalidationManager) ScheduledInvalidation(ctx context.Context, cacheKey string, delay time.Duration) {
	go func() {
		time.Sleep(delay)

		err := Mgr.Delete(ctx, cacheKey)
		if err != nil {
			logger.Warnf(ctx, "定时失效缓存失败: key=%s, error=%v", cacheKey, err)
		} else {
			logger.Debugf(ctx, "定时失效缓存成功: key=%s, delay=%v", cacheKey, delay)
		}
	}()
}

// ConditionalInvalidation 条件性失效缓存
func (cim *InvalidationManager) ConditionalInvalidation(ctx context.Context, cacheKey string, condition func() bool) error {
	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用")
	}

	if !condition() {
		logger.Debugf(ctx, "条件不满足，跳过缓存失效: key=%s", cacheKey)
		return nil
	}

	err := Mgr.Delete(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("条件性失效缓存失败: %w", err)
	}

	logger.Debugf(ctx, "条件性失效缓存成功: key=%s", cacheKey)
	return nil
}

// WarmupCache 预热缓存
func (cim *InvalidationManager) WarmupCache(ctx context.Context, warmupFunc func(ctx context.Context) error) error {
	logger.Debugf(ctx, "开始预热缓存")

	start := time.Now()
	err := warmupFunc(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf(ctx, "预热缓存失败: error=%v, duration=%v", err, duration)
		return err
	}

	logger.Debugf(ctx, "预热缓存完成: duration=%v", duration)
	return nil
}

// IsRecentlyInvalidated 检查是否最近被失效过
func (cim *InvalidationManager) IsRecentlyInvalidated(key string, within time.Duration) bool {
	cim.mutex.RLock()
	defer cim.mutex.RUnlock()

	timestamp, exists := cim.invalidationLog[key]
	if !exists {
		return false
	}

	return time.Now().Unix()-timestamp < int64(within.Seconds())
}
