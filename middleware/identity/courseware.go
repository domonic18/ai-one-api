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
	// 1. 从Redis缓存获取用户信息
	userInfo, err := c.cache.GetUserInfo(ctx, externalIdentity)
	if err != nil {
		logger.Warnf(ctx, "获取用户信息缓存失败: teacher=%s, error=%v", externalIdentity, err)
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

		err = c.cache.SetUserInfo(ctx, userInfo)
		if err != nil {
			logger.Warnf(ctx, "更新缓存失败: teacher=%s, error=%v", externalIdentity, err)
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

	result, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var userInfo UserInfo
	err = json.Unmarshal([]byte(result), &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// SetUserInfo 设置用户信息到Redis
func (c *CoursewareCache) SetUserInfo(ctx context.Context, userInfo *UserInfo) error {
	key := fmt.Sprintf("courseware:teacher:%s", userInfo.TeacherId)

	data, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	return c.redisClient.Set(ctx, key, data, c.config.CacheTTL).Err()
}

// BatchSetUserInfo 批量设置用户信息
func (c *CoursewareCache) BatchSetUserInfo(ctx context.Context, userInfos []*UserInfo) error {
	pipe := c.redisClient.Pipeline()

	for _, userInfo := range userInfos {
		key := fmt.Sprintf("courseware:teacher:%s", userInfo.TeacherId)
		data, err := json.Marshal(userInfo)
		if err != nil {
			continue // 跳过序列化失败的数据
		}
		pipe.Set(ctx, key, data, c.config.CacheTTL)
	}

	_, err := pipe.Exec(ctx)
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
	teacherIds, err := p.apiClient.GetAllTeacherIds(ctx)
	if err != nil {
		logger.Errorf(ctx, "获取老师ID列表失败: %v", err)
		return err
	}

	// 2. 分批预加载，避免单次请求过大
	batchSize := p.config.PreloadBatchSize
	if batchSize <= 0 {
		batchSize = 100 // 默认批次大小
	}

	totalBatches := (len(teacherIds) + batchSize - 1) / batchSize
	logger.Infof(ctx, "预加载用户信息: 总计%d个用户，分%d批处理", len(teacherIds), totalBatches)

	for i := 0; i < len(teacherIds); i += batchSize {
		end := i + batchSize
		if end > len(teacherIds) {
			end = len(teacherIds)
		}

		batch := teacherIds[i:end]
		if err := p.preloadBatch(ctx, batch, i/batchSize+1, totalBatches); err != nil {
			logger.Warnf(ctx, "预加载第%d批失败: %v", i/batchSize+1, err)
			// 继续处理下一批，不中断整个预加载过程
		}
	}

	logger.Infof(ctx, "用户信息预加载完成")
	return nil
}

// preloadBatch 预加载单个批次
func (p *PreloadManager) preloadBatch(ctx context.Context, teacherIds []string, batchNum, totalBatches int) error {
	logger.Debugf(ctx, "预加载第%d/%d批，包含%d个用户", batchNum, totalBatches, len(teacherIds))

	// 1. 批量获取用户信息
	apiUserInfos, err := p.apiClient.BatchGetUserInfo(ctx, teacherIds)
	if err != nil {
		return fmt.Errorf("批量获取用户信息失败: %w", err)
	}

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
		err = p.cache.BatchSetUserInfo(ctx, userInfos)
		if err != nil {
			return fmt.Errorf("批量写入缓存失败: %w", err)
		}
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
