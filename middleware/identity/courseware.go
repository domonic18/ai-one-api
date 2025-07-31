package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/logger"
)

// CoursewareAPIClient 课件平台API客户端接口
type CoursewareAPIClient interface {
	GetTeacherInfo(ctx context.Context, teacherId string) (*client.TeacherInfo, error)
	GetAllTeacherIds(ctx context.Context) ([]string, error)
	BatchGetUserInfo(ctx context.Context, teacherIds []string) ([]*client.TeacherInfo, error)
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Pipeline() redis.Pipeliner
}

// CoursewareIdentityResolver 课件平台身份解析器
// 实现Redis缓存和预加载机制
type CoursewareIdentityResolver struct {
	apiClient CoursewareAPIClient
	cache     *CoursewareCache
	config    *CoursewareConfig
}

// UserInfo 用户信息结构
type UserInfo struct {
	TeacherId      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	GroupName      string `json:"group_name"`      // OneAPI用户组
	PreferredModel string `json:"preferred_model"` // 用户偏好模型
	SchoolId       int    `json:"school_id"`       // 元数据：学校ID
	SchoolName     string `json:"school_name"`     // 元数据：学校名称
	SubjectId      int    `json:"subject_id"`      // 元数据：学科ID
	SubjectName    string `json:"subject_name"`    // 元数据：学科名称
	UpdatedAt      int64  `json:"updated_at"`      // 更新时间戳
}

// CoursewareConfig 课件平台配置
type CoursewareConfig struct {
	Enabled          bool          `json:"enabled"`
	BaseURL          string        `json:"base_url"`
	APIKey           string        `json:"api_key"`
	Timeout          time.Duration `json:"timeout"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	DefaultGroup     string        `json:"default_group"`
	PreloadBatchSize int           `json:"preload_batch_size"`
	RefreshInterval  time.Duration `json:"refresh_interval"`
}

// NewCoursewareIdentityResolver 创建课件平台身份解析器
func NewCoursewareIdentityResolver(apiClient CoursewareAPIClient, cache *CoursewareCache, config *CoursewareConfig) *CoursewareIdentityResolver {
	return &CoursewareIdentityResolver{
		apiClient: apiClient,
		cache:     cache,
		config:    config,
	}
}

// ResolveGroup 解析用户组
func (c *CoursewareIdentityResolver) ResolveGroup(ctx context.Context, externalIdentity string) string {
	logger.Debugf(ctx, "开始解析用户组: teacher=%s", externalIdentity)

	// 1. 从Redis缓存获取用户信息
	userInfo, err := c.cache.GetUserInfo(ctx, externalIdentity)
	if err != nil {
		logger.Warnf(ctx, "获取用户信息缓存失败: teacher=%s, error=%v", externalIdentity, err)
		logger.Infof(ctx, "使用默认分组: teacher=%s, group=%s", externalIdentity, c.config.DefaultGroup)
		return c.config.DefaultGroup
	}

	// 2. 缓存命中，直接返回
	if userInfo != nil {
		logger.Debugf(ctx, "缓存命中: teacher=%s, group=%s", externalIdentity, userInfo.GroupName)
		return userInfo.GroupName
	}

	// 3. 缓存未命中，调用API获取（这种情况应该很少发生，因为有预加载）
	logger.Warnf(ctx, "缓存未命中，调用API获取用户信息: teacher=%s", externalIdentity)

	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	logger.Debugf(ctx, "调用课件平台API: teacher=%s, timeout=%v", externalIdentity, c.config.Timeout)
	apiUserInfo, err := c.apiClient.GetTeacherInfo(ctx, externalIdentity)
	if err != nil {
		// 4. API调用失败，降级处理
		logger.Warnf(ctx, "课件平台API调用失败，使用默认分组: teacher=%s, error=%v", externalIdentity, err)
		return c.config.DefaultGroup
	}

	// 5. 更新缓存
	if apiUserInfo != nil {
		// 转换为UserInfo格式
		userInfo = &UserInfo{
			TeacherId:      apiUserInfo.TeacherId,
			TeacherName:    apiUserInfo.TeacherName,
			GroupName:      apiUserInfo.GroupName,
			PreferredModel: apiUserInfo.PreferredModel,
			SchoolId:       apiUserInfo.SchoolId,
			SchoolName:     apiUserInfo.SchoolName,
			SubjectId:      apiUserInfo.SubjectId,
			SubjectName:    apiUserInfo.SubjectName,
			UpdatedAt:      time.Now().Unix(),
		}

		logger.Debugf(ctx, "API返回用户信息: teacher=%s, group=%s, preferredModel=%s",
			externalIdentity, apiUserInfo.GroupName, apiUserInfo.PreferredModel)

		err = c.cache.SetUserInfo(ctx, userInfo)
		if err != nil {
			logger.Warnf(ctx, "更新缓存失败: teacher=%s, error=%v", externalIdentity, err)
		} else {
			logger.Debugf(ctx, "用户信息已缓存: teacher=%s", externalIdentity)
		}

		logger.Infof(ctx, "身份解析成功: teacher=%s, group=%s", externalIdentity, apiUserInfo.GroupName)
		return apiUserInfo.GroupName
	}

	// 6. API返回空，使用默认分组
	logger.Warnf(ctx, "API返回空用户信息，使用默认分组: teacher=%s", externalIdentity)
	return c.config.DefaultGroup
}

// ResolveModel 解析用户偏好模型
func (c *CoursewareIdentityResolver) ResolveModel(ctx context.Context, externalIdentity string, requestModel string) string {
	logger.Debugf(ctx, "开始解析用户偏好模型: teacher=%s, requestModel=%s", externalIdentity, requestModel)

	// 1. 从Redis缓存获取用户信息
	userInfo, err := c.cache.GetUserInfo(ctx, externalIdentity)
	if err != nil {
		logger.Warnf(ctx, "获取用户偏好模型缓存失败，使用请求模型: teacher=%s, error=%v", externalIdentity, err)
		return requestModel
	}

	// 2. 缓存命中且有偏好模型
	if userInfo != nil && userInfo.PreferredModel != "" {
		logger.Debugf(ctx, "使用用户偏好模型: teacher=%s, preferred=%s, original=%s",
			externalIdentity, userInfo.PreferredModel, requestModel)
		return userInfo.PreferredModel
	}

	// 3. 缓存未命中或无偏好模型，使用原始请求模型
	if userInfo == nil {
		logger.Debugf(ctx, "用户信息缓存未命中，使用请求模型: teacher=%s, model=%s", externalIdentity, requestModel)
	} else {
		logger.Debugf(ctx, "用户无偏好模型配置，使用请求模型: teacher=%s, model=%s", externalIdentity, requestModel)
	}

	return requestModel
}

// CoursewareCache Redis缓存管理器
type CoursewareCache struct {
	redisClient RedisClient
	config      *CoursewareConfig
}

// NewCoursewareCache 创建缓存管理器
func NewCoursewareCache(redisClient RedisClient, config *CoursewareConfig) *CoursewareCache {
	return &CoursewareCache{
		redisClient: redisClient,
		config:      config,
	}
}

// GetUserInfo 从Redis获取用户信息
func (c *CoursewareCache) GetUserInfo(ctx context.Context, teacherId string) (*UserInfo, error) {
	key := fmt.Sprintf("courseware:teacher:%s", teacherId)
	logger.Debugf(ctx, "从Redis获取用户信息: key=%s", key)

	result, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debugf(ctx, "Redis缓存未命中: key=%s", key)
			return nil, nil // 缓存未命中
		}
		logger.Warnf(ctx, "Redis获取用户信息失败: key=%s, error=%v", key, err)
		return nil, err
	}

	var userInfo UserInfo
	err = json.Unmarshal([]byte(result), &userInfo)
	if err != nil {
		logger.Warnf(ctx, "解析Redis用户信息失败: key=%s, error=%v", key, err)
		return nil, err
	}

	logger.Debugf(ctx, "Redis缓存命中: key=%s, teacherId=%s, group=%s", key, userInfo.TeacherId, userInfo.GroupName)
	return &userInfo, nil
}

// SetUserInfo 设置用户信息到Redis
func (c *CoursewareCache) SetUserInfo(ctx context.Context, userInfo *UserInfo) error {
	key := fmt.Sprintf("courseware:teacher:%s", userInfo.TeacherId)
	logger.Debugf(ctx, "设置用户信息到Redis: key=%s, ttl=%v", key, c.config.CacheTTL)

	data, err := json.Marshal(userInfo)
	if err != nil {
		logger.Warnf(ctx, "序列化用户信息失败: teacherId=%s, error=%v", userInfo.TeacherId, err)
		return err
	}

	err = c.redisClient.Set(ctx, key, data, c.config.CacheTTL).Err()
	if err != nil {
		logger.Warnf(ctx, "Redis设置用户信息失败: key=%s, error=%v", key, err)
	} else {
		logger.Debugf(ctx, "Redis设置用户信息成功: key=%s", key)
	}
	return err
}

// BatchSetUserInfo 批量设置用户信息
func (c *CoursewareCache) BatchSetUserInfo(ctx context.Context, userInfos []*UserInfo) error {
	logger.Debugf(ctx, "批量设置用户信息到Redis: count=%d, ttl=%v", len(userInfos), c.config.CacheTTL)

	pipe := c.redisClient.Pipeline()
	successCount := 0

	for _, userInfo := range userInfos {
		key := fmt.Sprintf("courseware:teacher:%s", userInfo.TeacherId)
		data, err := json.Marshal(userInfo)
		if err != nil {
			logger.Warnf(ctx, "序列化用户信息失败，跳过: teacherId=%s, error=%v", userInfo.TeacherId, err)
			continue // 跳过序列化失败的数据
		}
		pipe.Set(ctx, key, data, c.config.CacheTTL)
		successCount++
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Warnf(ctx, "Redis批量设置用户信息失败: count=%d, error=%v", successCount, err)
	} else {
		logger.Debugf(ctx, "Redis批量设置用户信息成功: count=%d", successCount)
	}
	return err
}

// DeleteUserInfo 删除用户信息缓存
func (c *CoursewareCache) DeleteUserInfo(ctx context.Context, teacherId string) error {
	key := fmt.Sprintf("courseware:teacher:%s", teacherId)
	return c.redisClient.Del(ctx, key).Err()
}

// PreloadManager 预加载管理器
type PreloadManager struct {
	apiClient CoursewareAPIClient
	cache     *CoursewareCache
	config    *CoursewareConfig
}

// NewPreloadManager 创建预加载管理器
func NewPreloadManager(apiClient CoursewareAPIClient, cache *CoursewareCache, config *CoursewareConfig) *PreloadManager {
	return &PreloadManager{
		apiClient: apiClient,
		cache:     cache,
		config:    config,
	}
}

// PreloadUserInfos 执行预加载
func (p *PreloadManager) PreloadUserInfos(ctx context.Context) error {
	logger.Infof(ctx, "开始预加载用户信息到Redis缓存")

	// 1. 获取所有需要预加载的老师ID列表
	logger.Debugf(ctx, "调用课件平台API获取所有老师ID列表")
	teacherIds, err := p.apiClient.GetAllTeacherIds(ctx)
	if err != nil {
		logger.Errorf(ctx, "获取老师ID列表失败: %v", err)
		return err
	}

	// 2. 分批预加载，避免单次请求过大
	batchSize := p.config.PreloadBatchSize
	if batchSize <= 0 {
		batchSize = 100 // 默认批次大小
		logger.Debugf(ctx, "使用默认批次大小: %d", batchSize)
	}

	totalBatches := (len(teacherIds) + batchSize - 1) / batchSize
	logger.Infof(ctx, "预加载用户信息: 总计%d个用户，分%d批处理", len(teacherIds), totalBatches)

	successBatches := 0
	failedBatches := 0

	for i := 0; i < len(teacherIds); i += batchSize {
		end := i + batchSize
		if end > len(teacherIds) {
			end = len(teacherIds)
		}

		batch := teacherIds[i:end]
		if err := p.preloadBatch(ctx, batch, i/batchSize+1, totalBatches); err != nil {
			logger.Warnf(ctx, "预加载第%d批失败: %v", i/batchSize+1, err)
			failedBatches++
			// 继续处理下一批，不中断整个预加载过程
		} else {
			successBatches++
		}
	}

	logger.Infof(ctx, "用户信息预加载完成: 成功%d批，失败%d批", successBatches, failedBatches)
	return nil
}

// preloadBatch 预加载单个批次
func (p *PreloadManager) preloadBatch(ctx context.Context, teacherIds []string, batchNum, totalBatches int) error {
	logger.Debugf(ctx, "预加载第%d/%d批，包含%d个用户", batchNum, totalBatches, len(teacherIds))

	// 1. 批量获取用户信息
	logger.Debugf(ctx, "调用课件平台API批量获取用户信息: batchNum=%d, count=%d", batchNum, len(teacherIds))
	apiUserInfos, err := p.apiClient.BatchGetUserInfo(ctx, teacherIds)
	if err != nil {
		logger.Errorf(ctx, "批量获取用户信息失败: batchNum=%d, error=%v", batchNum, err)
		return fmt.Errorf("批量获取用户信息失败: %w", err)
	}

	logger.Debugf(ctx, "API返回用户信息: batchNum=%d, count=%d", batchNum, len(apiUserInfos))

	// 2. 转换为UserInfo格式
	userInfos := make([]*UserInfo, 0, len(apiUserInfos))
	for _, apiUserInfo := range apiUserInfos {
		userInfo := &UserInfo{
			TeacherId:      apiUserInfo.TeacherId,
			TeacherName:    apiUserInfo.TeacherName,
			GroupName:      apiUserInfo.GroupName,
			PreferredModel: apiUserInfo.PreferredModel,
			SchoolId:       apiUserInfo.SchoolId,
			SchoolName:     apiUserInfo.SchoolName,
			SubjectId:      apiUserInfo.SubjectId,
			SubjectName:    apiUserInfo.SubjectName,
			UpdatedAt:      time.Now().Unix(),
		}
		userInfos = append(userInfos, userInfo)
	}

	// 3. 批量写入Redis缓存
	if len(userInfos) > 0 {
		logger.Debugf(ctx, "批量写入Redis缓存: batchNum=%d, count=%d", batchNum, len(userInfos))
		err = p.cache.BatchSetUserInfo(ctx, userInfos)
		if err != nil {
			logger.Errorf(ctx, "批量写入缓存失败: batchNum=%d, error=%v", batchNum, err)
			return fmt.Errorf("批量写入缓存失败: %w", err)
		}
	} else {
		logger.Warnf(ctx, "第%d批无有效用户信息，跳过缓存写入", batchNum)
	}

	logger.Debugf(ctx, "第%d批预加载完成，成功缓存%d个用户信息", batchNum, len(userInfos))
	return nil
}

// StartPeriodicRefresh 定期刷新缓存
func (p *PreloadManager) StartPeriodicRefresh(ctx context.Context) {
	if p.config.RefreshInterval <= 0 {
		logger.Infof(ctx, "未配置定期刷新间隔，跳过定期刷新")
		return
	}

	ticker := time.NewTicker(p.config.RefreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				logger.Infof(ctx, "开始定期刷新用户信息缓存")
				if err := p.PreloadUserInfos(ctx); err != nil {
					logger.Errorf(ctx, "定期刷新失败: %v", err)
				}
			case <-ctx.Done():
				logger.Infof(ctx, "定期刷新任务停止")
				return
			}
		}
	}()

	logger.Infof(ctx, "定期刷新任务已启动，间隔: %v", p.config.RefreshInterval)
}
