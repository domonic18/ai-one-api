package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// Monitor 缓存监控器
type Monitor struct {
	mutex              sync.RWMutex
	performanceMetrics *PerformanceMetrics
	healthStatus       *HealthStatus
	alertThresholds    *AlertThresholds
	isMonitoring       bool
}

// PerformanceMetrics 缓存性能指标
type PerformanceMetrics struct {
	// 基础指标
	TotalRequests int64   `json:"total_requests"`
	CacheHits     int64   `json:"cache_hits"`
	CacheMisses   int64   `json:"cache_misses"`
	HitRate       float64 `json:"hit_rate"`
	MissRate      float64 `json:"miss_rate"`

	// 性能指标
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	MaxResponseTime float64 `json:"max_response_time_ms"`
	MinResponseTime float64 `json:"min_response_time_ms"`

	// 操作指标
	SetOperations    int64 `json:"set_operations"`
	GetOperations    int64 `json:"get_operations"`
	DeleteOperations int64 `json:"delete_operations"`

	// 错误指标
	ErrorCount int64   `json:"error_count"`
	ErrorRate  float64 `json:"error_rate"`

	// 时间窗口
	WindowStart    int64 `json:"window_start"`
	WindowEnd      int64 `json:"window_end"`
	WindowDuration int64 `json:"window_duration_seconds"`
}

// HealthStatus 缓存健康状态
type HealthStatus struct {
	IsHealthy           bool     `json:"is_healthy"`
	OverallScore        float64  `json:"overall_score"`
	RedisConnected      bool     `json:"redis_connected"`
	HitRateHealthy      bool     `json:"hit_rate_healthy"`
	ResponseTimeHealthy bool     `json:"response_time_healthy"`
	ErrorRateHealthy    bool     `json:"error_rate_healthy"`
	LastCheckTime       int64    `json:"last_check_time"`
	Issues              []string `json:"issues"`
}

// AlertThresholds 缓存告警阈值
type AlertThresholds struct {
	MinHitRate          float64 `json:"min_hit_rate"`          // 最小命中率
	MaxResponseTime     float64 `json:"max_response_time"`     // 最大响应时间（毫秒）
	MaxErrorRate        float64 `json:"max_error_rate"`        // 最大错误率
	HealthCheckInterval int64   `json:"health_check_interval"` // 健康检查间隔（秒）
}

var MonitorInstance *Monitor

// InitMonitor 初始化缓存监控器
func InitMonitor() {
	MonitorInstance = &Monitor{
		performanceMetrics: &PerformanceMetrics{
			WindowStart:     time.Now().Unix(),
			MinResponseTime: 999999, // 初始化为最大值
		},
		healthStatus: &HealthStatus{
			IsHealthy:    true,
			OverallScore: 100.0,
			Issues:       make([]string, 0),
		},
		alertThresholds: &AlertThresholds{
			MinHitRate:          95.0,  // 95%
			MaxResponseTime:     100.0, // 100ms
			MaxErrorRate:        1.0,   // 1%
			HealthCheckInterval: 60,    // 60秒
		},
		isMonitoring: false,
	}

	logger.SysLog("缓存监控器初始化完成")
}

// StartMonitoring 开始监控
func (cm *Monitor) StartMonitoring() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.isMonitoring {
		logger.SysLog("缓存监控已在运行中")
		return
	}

	cm.isMonitoring = true

	// 启动健康检查协程
	go cm.healthCheckRoutine()

	// 启动性能指标重置协程
	go cm.metricsResetRoutine()

	logger.SysLog("缓存监控已启动")
}

// StopMonitoring 停止监控
func (cm *Monitor) StopMonitoring() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.isMonitoring = false
	logger.SysLog("缓存监控已停止")
}

// healthCheckRoutine 健康检查协程
func (cm *Monitor) healthCheckRoutine() {
	ticker := time.NewTicker(time.Duration(cm.alertThresholds.HealthCheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !cm.isMonitoring {
			break
		}
		cm.performHealthCheck()
	}
}

// metricsResetRoutine 性能指标重置协程（每小时重置一次）
func (cm *Monitor) metricsResetRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if !cm.isMonitoring {
			break
		}
		cm.resetMetrics()
	}
}

// recordCacheOperation 记录缓存操作
func (cm *Monitor) recordCacheOperation(operationType string, duration time.Duration, success bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 更新基础指标
	cm.performanceMetrics.TotalRequests++

	switch operationType {
	case "get":
		cm.performanceMetrics.GetOperations++
		if success {
			cm.performanceMetrics.CacheHits++
		} else {
			cm.performanceMetrics.CacheMisses++
		}
	case "set":
		cm.performanceMetrics.SetOperations++
	case "delete":
		cm.performanceMetrics.DeleteOperations++
	}

	// 更新性能指标
	durationMs := float64(duration.Nanoseconds()) / 1e6
	if cm.performanceMetrics.TotalRequests == 1 {
		cm.performanceMetrics.AvgResponseTime = durationMs
		cm.performanceMetrics.MaxResponseTime = durationMs
		cm.performanceMetrics.MinResponseTime = durationMs
	} else {
		// 更新平均响应时间
		cm.performanceMetrics.AvgResponseTime =
			(cm.performanceMetrics.AvgResponseTime*float64(cm.performanceMetrics.TotalRequests-1) + durationMs) /
				float64(cm.performanceMetrics.TotalRequests)

		// 更新最大响应时间
		if durationMs > cm.performanceMetrics.MaxResponseTime {
			cm.performanceMetrics.MaxResponseTime = durationMs
		}

		// 更新最小响应时间
		if durationMs < cm.performanceMetrics.MinResponseTime {
			cm.performanceMetrics.MinResponseTime = durationMs
		}
	}

	// 更新错误指标
	if !success {
		cm.performanceMetrics.ErrorCount++
	}

	// 计算命中率和错误率
	if cm.performanceMetrics.TotalRequests > 0 {
		cm.performanceMetrics.HitRate = float64(cm.performanceMetrics.CacheHits) / float64(cm.performanceMetrics.TotalRequests) * 100
		cm.performanceMetrics.MissRate = float64(cm.performanceMetrics.CacheMisses) / float64(cm.performanceMetrics.TotalRequests) * 100
		cm.performanceMetrics.ErrorRate = float64(cm.performanceMetrics.ErrorCount) / float64(cm.performanceMetrics.TotalRequests) * 100
	}
}

// performHealthCheck 执行健康检查
func (cm *Monitor) performHealthCheck() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ctx := context.Background()
	issues := make([]string, 0)
	score := 100.0

	// 检查Redis连接
	cm.healthStatus.RedisConnected = cm.checkRedisConnection(ctx)
	if !cm.healthStatus.RedisConnected {
		issues = append(issues, "Redis连接异常")
		score -= 30
	}

	// 检查命中率
	cm.healthStatus.HitRateHealthy = cm.performanceMetrics.HitRate >= cm.alertThresholds.MinHitRate
	if !cm.healthStatus.HitRateHealthy {
		issues = append(issues, fmt.Sprintf("缓存命中率过低: %.2f%% < %.2f%%",
			cm.performanceMetrics.HitRate, cm.alertThresholds.MinHitRate))
		score -= 25
	}

	// 检查响应时间
	cm.healthStatus.ResponseTimeHealthy = cm.performanceMetrics.AvgResponseTime <= cm.alertThresholds.MaxResponseTime
	if !cm.healthStatus.ResponseTimeHealthy {
		issues = append(issues, fmt.Sprintf("平均响应时间过高: %.2fms > %.2fms",
			cm.performanceMetrics.AvgResponseTime, cm.alertThresholds.MaxResponseTime))
		score -= 20
	}

	// 检查错误率
	cm.healthStatus.ErrorRateHealthy = cm.performanceMetrics.ErrorRate <= cm.alertThresholds.MaxErrorRate
	if !cm.healthStatus.ErrorRateHealthy {
		issues = append(issues, fmt.Sprintf("错误率过高: %.2f%% > %.2f%%",
			cm.performanceMetrics.ErrorRate, cm.alertThresholds.MaxErrorRate))
		score -= 25
	}

	// 更新健康状态
	cm.healthStatus.IsHealthy = len(issues) == 0
	cm.healthStatus.OverallScore = score
	cm.healthStatus.Issues = issues
	cm.healthStatus.LastCheckTime = time.Now().Unix()

	// 记录健康检查结果
	if !cm.healthStatus.IsHealthy {
		logger.Warnf(ctx, "缓存健康检查失败: score=%.1f, issues=%v", score, issues)
	} else {
		logger.Debugf(ctx, "缓存健康检查通过: score=%.1f", score)
	}
}

// checkRedisConnection 检查Redis连接
func (cm *Monitor) checkRedisConnection(ctx context.Context) bool {
	if !common.RedisEnabled {
		return false
	}

	// 执行简单的ping测试
	_, err := common.RedisGet("__health_check__")
	return err == nil || err.Error() != "redis: nil" // 键不存在也算连接正常
}

// resetMetrics 重置性能指标
func (cm *Monitor) resetMetrics() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	logger.Debugf(context.Background(), "重置缓存性能指标: 总请求=%d, 命中率=%.2f%%",
		cm.performanceMetrics.TotalRequests, cm.performanceMetrics.HitRate)

	cm.performanceMetrics = &PerformanceMetrics{
		WindowStart:     time.Now().Unix(),
		MinResponseTime: 999999,
	}
}

// GetPerformanceMetrics 获取性能指标
func (cm *Monitor) GetPerformanceMetrics() *PerformanceMetrics {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建副本
	metrics := *cm.performanceMetrics
	metrics.WindowEnd = time.Now().Unix()
	metrics.WindowDuration = metrics.WindowEnd - metrics.WindowStart

	return &metrics
}

// GetHealthStatus 获取健康状态
func (cm *Monitor) GetHealthStatus() *HealthStatus {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建副本
	status := *cm.healthStatus
	return &status
}

// SetAlertThresholds 设置告警阈值
func (cm *Monitor) SetAlertThresholds(thresholds *AlertThresholds) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.alertThresholds = thresholds
	logger.Debugf(context.Background(), "更新缓存告警阈值: hitRate=%.1f%%, responseTime=%.1fms, errorRate=%.1f%%",
		thresholds.MinHitRate, thresholds.MaxResponseTime, thresholds.MaxErrorRate)
}

// GetAlertThresholds 获取告警阈值
func (cm *Monitor) GetAlertThresholds() *AlertThresholds {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建副本
	thresholds := *cm.alertThresholds
	return &thresholds
}

// RunPerformanceTest 运行缓存性能测试
func (cm *Monitor) RunPerformanceTest(ctx context.Context, testConfig *TestConfig) *TestResult {
	logger.Debugf(ctx, "开始缓存性能测试: operations=%d, concurrency=%d",
		testConfig.Operations, testConfig.Concurrency)

	start := time.Now()
	result := &TestResult{
		TestConfig: *testConfig,
		StartTime:  start.Unix(),
		Operations: make([]OperationResult, 0, testConfig.Operations),
	}

	// 执行测试操作
	for i := 0; i < testConfig.Operations; i++ {
		operationStart := time.Now()

		// 模拟缓存操作
		testKey := fmt.Sprintf("test_key_%d", i)
		testValue := fmt.Sprintf("test_value_%d", i)

		var err error
		var operationType string

		switch i % 3 {
		case 0: // Set操作
			operationType = "set"
			err = Mgr.Set(ctx, testKey, testValue, 5*time.Minute)
		case 1: // Get操作
			operationType = "get"
			var value string
			err = Mgr.Get(ctx, testKey, &value)
		case 2: // Delete操作
			operationType = "delete"
			err = Mgr.Delete(ctx, testKey)
		}

		operationDuration := time.Since(operationStart)

		operationResult := OperationResult{
			Operation: operationType,
			Duration:  operationDuration,
			Success:   err == nil,
			Error:     "",
		}

		if err != nil {
			operationResult.Error = err.Error()
		}

		result.Operations = append(result.Operations, operationResult)

		// 记录到监控系统
		cm.recordCacheOperation(operationType, operationDuration, err == nil)
	}

	// 计算测试结果
	totalDuration := time.Since(start)
	result.EndTime = time.Now().Unix()
	result.TotalDuration = totalDuration
	result.OperationsPerSecond = float64(testConfig.Operations) / totalDuration.Seconds()

	// 统计成功率
	successCount := 0
	for _, op := range result.Operations {
		if op.Success {
			successCount++
		}
	}
	result.SuccessRate = float64(successCount) / float64(len(result.Operations)) * 100

	logger.Debugf(ctx, "缓存性能测试完成: 耗时=%v, 成功率=%.2f%%, QPS=%.2f",
		totalDuration, result.SuccessRate, result.OperationsPerSecond)

	return result
}

// TestConfig 缓存测试配置
type TestConfig struct {
	Operations  int `json:"operations"`  // 操作数量
	Concurrency int `json:"concurrency"` // 并发数
}

// TestResult 缓存测试结果
type TestResult struct {
	TestConfig          TestConfig        `json:"test_config"`
	StartTime           int64             `json:"start_time"`
	EndTime             int64             `json:"end_time"`
	TotalDuration       time.Duration     `json:"total_duration"`
	OperationsPerSecond float64           `json:"operations_per_second"`
	SuccessRate         float64           `json:"success_rate"`
	Operations          []OperationResult `json:"operations"`
}

// OperationResult 缓存操作结果
type OperationResult struct {
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}

// IsMonitoring 检查是否正在监控
func (cm *Monitor) IsMonitoring() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.isMonitoring
}
