package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/middleware/identity"
)

// CoursewareStatus 课件平台状态信息
type CoursewareStatus struct {
	IsConnected  bool    `json:"is_connected"`
	LastSyncTime *int64  `json:"last_sync_time"`
	TotalUsers   int     `json:"total_users"`
	CachedUsers  int     `json:"cached_users"`
	SyncProgress float64 `json:"sync_progress"`
	IsSyncing    bool    `json:"is_syncing"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// CoursewareConfig 课件平台配置信息
type CoursewareConfig struct {
	Enabled          bool   `json:"enabled"`
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key,omitempty"` // 不返回实际的API Key
	Timeout          int    `json:"timeout"`
	CacheTTL         int    `json:"cache_ttl"`
	DefaultGroup     string `json:"default_group"`
	PreloadBatchSize int    `json:"preload_batch_size"`
	RefreshInterval  int    `json:"refresh_interval"`
}

// CacheItem 缓存项信息
type CacheItem struct {
	TeacherID      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	SchoolName     string `json:"school_name"`
	SubjectName    string `json:"subject_name"`
	GroupName      string `json:"group_name"`
	PreferredModel string `json:"preferred_model"`
	CacheTime      int64  `json:"cache_time"`
	ExpiryTime     int64  `json:"expiry_time"`
}

// GetCoursewareStatus 获取课件平台集成状态
func GetCoursewareStatus(c *gin.Context) {
	// 检查课件平台集成是否启用
	resolver := identity.GetIdentityResolver()

	status := &CoursewareStatus{
		IsConnected:  false,
		LastSyncTime: nil,
		TotalUsers:   0,
		CachedUsers:  0,
		SyncProgress: 0.0,
		IsSyncing:    false,
	}

	// 如果是课件平台解析器，获取详细状态
	if coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver); ok {
		// 测试连接
		apiClient := client.GetCoursewareClient()
		if apiClient != nil && apiClient.IsValid() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 尝试获取教师ID列表来测试连接
			_, err := apiClient.GetTeacherIds(ctx)
			if err == nil {
				status.IsConnected = true
			} else {
				status.ErrorMessage = err.Error()
			}
		}

		// 获取缓存统计信息
		if cache := coursewareResolver.GetCache(); cache != nil {
			stats := cache.GetStats()
			status.CachedUsers = stats.CachedUsers
			status.LastSyncTime = stats.LastSyncTime
		}
	} else {
		status.ErrorMessage = "课件平台集成未启用或使用默认解析器"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetCoursewareConfig 获取课件平台配置信息
func GetCoursewareConfig(c *gin.Context) {
	coursewareConfig := &CoursewareConfig{
		Enabled:          config.CoursewareEnabled,
		BaseURL:          config.CoursewarePlatformBaseURL,
		APIKey:           "***", // 不返回实际的API Key，只显示掩码
		Timeout:          config.CoursewarePlatformTimeout,
		CacheTTL:         int(config.CoursewareCacheTTL.Minutes()),
		DefaultGroup:     config.CoursewareDefaultGroup,
		PreloadBatchSize: config.CoursewarePreloadBatchSize,
		RefreshInterval:  int(config.CoursewareRefreshInterval.Minutes()),
	}

	// 如果API Key为空，显示为空字符串
	if config.CoursewarePlatformAPIKey == "" {
		coursewareConfig.APIKey = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    coursewareConfig,
	})
}

// GetCoursewareCache 获取课件平台缓存数据
func GetCoursewareCache(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var cacheItems []CacheItem
	var total int

	// 检查是否启用了课件平台集成
	resolver := identity.GetIdentityResolver()
	if coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver); ok {
		if cache := coursewareResolver.GetCache(); cache != nil {
			// 获取缓存数据
			items, totalCount := cache.GetCacheItems(page, size, search)
			total = totalCount

			for _, item := range items {
				cacheItems = append(cacheItems, CacheItem{
					TeacherID:      item.TeacherID,
					TeacherName:    item.TeacherName,
					SchoolName:     item.SchoolName,
					SubjectName:    item.SubjectName,
					GroupName:      item.GroupName,
					PreferredModel: item.PreferredModel,
					CacheTime:      item.CacheTime,
					ExpiryTime:     item.ExpiryTime,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": cacheItems,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// SyncCoursewareUsers 同步课件平台用户数据
func SyncCoursewareUsers(c *gin.Context) {
	// 检查是否启用了课件平台集成
	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "课件平台集成未启用",
		})
		return
	}

	// 检查是否正在同步
	if cache := coursewareResolver.GetCache(); cache != nil {
		if cache.IsSyncing() {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "正在同步中，请稍后再试",
			})
			return
		}
	}

	// 异步执行同步
	go func() {
		ctx := context.Background()
		if preloadManager := coursewareResolver.GetPreloadManager(); preloadManager != nil {
			logger.SysLog("开始手动同步课件平台用户数据")
			if err := preloadManager.PreloadUserInfos(ctx); err != nil {
				logger.SysError("手动同步失败: " + err.Error())
			} else {
				logger.SysLog("手动同步完成")
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "同步任务已启动，请稍后查看状态",
	})
}

// ClearCoursewareCache 清理课件平台缓存
func ClearCoursewareCache(c *gin.Context) {
	// 检查是否启用了课件平台集成
	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "课件平台集成未启用",
		})
		return
	}

	// 清理缓存
	if cache := coursewareResolver.GetCache(); cache != nil {
		cleared := cache.ClearCache()
		logger.SysLogf("手动清理课件平台缓存，清理了 %d 个缓存项", cleared)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "缓存清理完成",
			"data": gin.H{
				"cleared_count": cleared,
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "缓存管理器不可用",
		})
	}
}

// RefreshCoursewareCache 刷新课件平台缓存
func RefreshCoursewareCache(c *gin.Context) {
	// 检查是否启用了课件平台集成
	resolver := identity.GetIdentityResolver()
	coursewareResolver, ok := resolver.(*identity.CoursewareIdentityResolver)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "课件平台集成未启用",
		})
		return
	}

	// 检查是否正在同步
	if cache := coursewareResolver.GetCache(); cache != nil {
		if cache.IsSyncing() {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "正在同步中，请稍后再试",
			})
			return
		}
	}

	// 异步执行刷新
	go func() {
		ctx := context.Background()
		if preloadManager := coursewareResolver.GetPreloadManager(); preloadManager != nil {
			logger.SysLog("开始手动刷新课件平台缓存")
			if err := preloadManager.RefreshCache(ctx); err != nil {
				logger.SysError("手动刷新失败: " + err.Error())
			} else {
				logger.SysLog("手动刷新完成")
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "刷新任务已启动，请稍后查看状态",
	})
}

// TestCoursewareConnection 测试课件平台连接
func TestCoursewareConnection(c *gin.Context) {
	apiClient := client.GetCoursewareClient()
	if apiClient == nil || !apiClient.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "课件平台API客户端未正确配置",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 测试连接 - 尝试获取教师ID列表
	teacherIds, err := apiClient.GetTeacherIds(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "连接测试失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "连接测试成功",
		"data": gin.H{
			"teacher_count": len(teacherIds),
			"test_time":     time.Now().Unix(),
		},
	})
}
