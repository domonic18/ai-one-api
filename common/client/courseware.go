package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
)

// CoursewareClient 课件平台API客户端
type CoursewareClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// 课件平台API响应结构
type TeacherInfo struct {
	TeacherId      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	SchoolId       int    `json:"school_id"`
	SchoolName     string `json:"school_name"`
	SubjectId      int    `json:"subject_id"`
	SubjectName    string `json:"subject_name"`
	GroupName      string `json:"group_name"`      // OneAPI用户组
	PreferredModel string `json:"preferred_model"` // 用户偏好模型
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

type SubjectInfo struct {
	SubjectId    int    `json:"subject_id"`
	SubjectName  string `json:"subject_name"`
	SchoolId     int    `json:"school_id"`
	SchoolName   string `json:"school_name"`
	DefaultModel string `json:"default_model"`
	GroupName    string `json:"group_name"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type UserConfig struct {
	UserId     string `json:"user_id"`
	ModelName  string `json:"model_name"`
	Parameters string `json:"parameters,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

var coursewareClient *CoursewareClient

// InitCoursewareClient 初始化课件平台API客户端
func InitCoursewareClient() {
	if coursewareClient != nil {
		return
	}

	// 获取配置
	baseURL := config.CoursewarePlatformBaseURL
	apiKey := config.CoursewarePlatformAPIKey
	timeout := time.Duration(config.CoursewarePlatformTimeout) * time.Second

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: timeout,
	}

	coursewareClient = &CoursewareClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}

	logger.SysLog("课件平台API客户端初始化完成")
}

// GetCoursewareClient 获取课件平台API客户端实例
func GetCoursewareClient() *CoursewareClient {
	if coursewareClient == nil {
		InitCoursewareClient()
	}
	return coursewareClient
}

// SetBaseURL 设置API基础URL（用于测试）
func (c *CoursewareClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// SetAPIKey 设置API密钥（用于测试）
func (c *CoursewareClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// GetTeacherInfo 获取老师信息
func (c *CoursewareClient) GetTeacherInfo(ctx context.Context, teacherId string) (*TeacherInfo, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("课件平台API基础URL未配置")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/teacher/%s/info", c.baseURL, teacherId)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var teacherInfo TeacherInfo
	if err := json.NewDecoder(resp.Body).Decode(&teacherInfo); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &teacherInfo, nil
}

// GetSubjectInfo 获取学科组信息
func (c *CoursewareClient) GetSubjectInfo(ctx context.Context, subjectId int) (*SubjectInfo, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("课件平台API基础URL未配置")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/subject/%d", c.baseURL, subjectId)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var subjectInfo SubjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&subjectInfo); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &subjectInfo, nil
}

// GetUserConfig 获取用户模型配置
func (c *CoursewareClient) GetUserConfig(ctx context.Context, userId string) (*UserConfig, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("课件平台API基础URL未配置")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/teacher/%s/model", c.baseURL, userId)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var userConfig UserConfig
	if err := json.NewDecoder(resp.Body).Decode(&userConfig); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &userConfig, nil
}

// PreloadTeacherInfos 批量预加载老师信息
func (c *CoursewareClient) PreloadTeacherInfos(ctx context.Context, teacherIds []string) (map[string]*TeacherInfo, error) {
	if len(teacherIds) == 0 {
		return make(map[string]*TeacherInfo), nil
	}

	results := make(map[string]*TeacherInfo)
	for _, teacherId := range teacherIds {
		info, err := c.GetTeacherInfo(ctx, teacherId)
		if err != nil {
			logger.Warnf(ctx, "预加载老师信息失败: teacherId=%s, error=%v", teacherId, err)
			continue
		}
		if info != nil {
			results[teacherId] = info
		}
	}

	return results, nil
}

// PreloadSubjectInfos 批量预加载学科组信息
func (c *CoursewareClient) PreloadSubjectInfos(ctx context.Context, subjectIds []int) (map[int]*SubjectInfo, error) {
	if len(subjectIds) == 0 {
		return make(map[int]*SubjectInfo), nil
	}

	results := make(map[int]*SubjectInfo)
	for _, subjectId := range subjectIds {
		info, err := c.GetSubjectInfo(ctx, subjectId)
		if err != nil {
			logger.Warnf(ctx, "预加载学科组信息失败: subjectId=%d, error=%v", subjectId, err)
			continue
		}
		if info != nil {
			results[subjectId] = info
		}
	}

	return results, nil
}

// PreloadUserConfigs 批量预加载用户模型配置
func (c *CoursewareClient) PreloadUserConfigs(ctx context.Context, userIds []string) (map[string]*UserConfig, error) {
	if len(userIds) == 0 {
		return make(map[string]*UserConfig), nil
	}

	results := make(map[string]*UserConfig)
	for _, userId := range userIds {
		config, err := c.GetUserConfig(ctx, userId)
		if err != nil {
			logger.Warnf(ctx, "预加载用户模型配置失败: userId=%s, error=%v", userId, err)
			continue
		}
		if config != nil {
			results[userId] = config
		}
	}

	return results, nil
}

// GetAllTeacherIds 获取所有老师ID列表（预加载用）
func (c *CoursewareClient) GetAllTeacherIds(ctx context.Context) ([]string, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("课件平台API基础URL未配置")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/teachers/ids", c.baseURL)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TeacherIds []string `json:"teacher_ids"`
			Total      int      `json:"total"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", response.Message)
	}

	return response.Data.TeacherIds, nil
}

// BatchGetUserInfo 批量获取用户信息（预加载用）
func (c *CoursewareClient) BatchGetUserInfo(ctx context.Context, teacherIds []string) ([]*TeacherInfo, error) {
	if len(teacherIds) == 0 {
		return []*TeacherInfo{}, nil
	}

	if c.baseURL == "" {
		return nil, fmt.Errorf("课件平台API基础URL未配置")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/teachers/batch", c.baseURL)

	// 构建请求体
	requestBody := map[string]interface{}{
		"teacher_ids": teacherIds,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败: 状态码=%d, 响应=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Users        []*TeacherInfo `json:"users"`
			SuccessCount int            `json:"success_count"`
			ErrorCount   int            `json:"error_count"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", response.Message)
	}

	return response.Data.Users, nil
}
