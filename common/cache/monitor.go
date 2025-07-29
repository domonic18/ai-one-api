package cache

import (
	"context"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// SimpleMonitor 简化的缓存监控器
type SimpleMonitor struct {
	mutex     sync.RWMutex
	isHealthy bool
	lastCheck int64
}

var MonitorInstance *SimpleMonitor

// InitMonitor 初始化简化的缓存监控器
func InitMonitor() {
	MonitorInstance = &SimpleMonitor{
		isHealthy: true,
		lastCheck: time.Now().Unix(),
	}
	logger.SysLog("缓存监控器初始化完成")
}

// IsMonitoring 检查是否正在监控（简化版本总是返回true）
func (sm *SimpleMonitor) IsMonitoring() bool {
	return true
}

// GetHealthStatus 获取健康状态
func (sm *SimpleMonitor) GetHealthStatus() map[string]interface{} {
	// 先执行健康检查（需要写锁）
	healthy := sm.checkHealth()

	// 然后获取读锁获取其他状态信息
	sm.mutex.RLock()
	lastCheck := sm.lastCheck
	sm.mutex.RUnlock()

	return map[string]interface{}{
		"is_healthy":    healthy,
		"last_check":    lastCheck,
		"redis_enabled": common.RedisEnabled,
		"cache_enabled": Mgr != nil && Mgr.IsEnabled(),
	}
}

// checkHealth 执行简单的健康检查
func (sm *SimpleMonitor) checkHealth() bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 更新检查时间
	sm.lastCheck = time.Now().Unix()

	// 检查Redis连接
	if !common.RedisEnabled {
		sm.isHealthy = false
		return false
	}

	// 检查RDB是否已初始化
	if common.RDB == nil {
		sm.isHealthy = false
		return false
	}

	// 简单的ping测试（使用超时避免阻塞）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := common.RDB.Get(ctx, "__health_check__").Result()
	sm.isHealthy = (err == nil || err.Error() == "redis: nil")

	return sm.isHealthy
}

// GetPerformanceMetrics 获取简化的性能指标
func (sm *SimpleMonitor) GetPerformanceMetrics() map[string]interface{} {
	if Mgr == nil {
		return map[string]interface{}{
			"total_requests": 0,
			"hit_rate":       0.0,
			"enabled":        false,
		}
	}

	stats := Mgr.GetStats()
	return map[string]interface{}{
		"total_requests": stats.TotalCount,
		"hit_count":      stats.HitCount,
		"miss_count":     stats.MissCount,
		"hit_rate":       stats.HitRate,
		"enabled":        true,
	}
}
