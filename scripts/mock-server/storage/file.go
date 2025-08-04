package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mock-courseware-platform/models"
)

// FileStorage 文件存储结构体
type FileStorage struct {
	usersFile  string
	configFile string
}

// NewFileStorage 创建新的文件存储实例
func NewFileStorage() *FileStorage {
	return &FileStorage{
		usersFile:  "./data/users.json",
		configFile: "./data/config.json",
	}
}

// LoadUsers 从文件加载用户数据
func (f *FileStorage) LoadUsers() ([]*models.TeacherInfo, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(f.usersFile), 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(f.usersFile); os.IsNotExist(err) {
		// 文件不存在，返回空列表
		return []*models.TeacherInfo{}, nil
	}

	// 读取文件
	data, err := os.ReadFile(f.usersFile)
	if err != nil {
		return nil, fmt.Errorf("读取用户数据文件失败: %v", err)
	}

	// 解析JSON
	var users []*models.TeacherInfo
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("解析用户数据失败: %v", err)
	}

	return users, nil
}

// SaveUsers 保存用户数据到文件
func (f *FileStorage) SaveUsers(users []*models.TeacherInfo) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(f.usersFile), 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 序列化数据
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化用户数据失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(f.usersFile, data, 0644); err != nil {
		return fmt.Errorf("写入用户数据文件失败: %v", err)
	}

	return nil
}

// LoadConfig 从文件加载配置
func (f *FileStorage) LoadConfig() (*models.ServerConfig, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(f.configFile), 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(f.configFile); os.IsNotExist(err) {
		// 文件不存在，返回默认配置
		// 从环境变量获取API密钥，如果没有则使用默认值
		apiKey := os.Getenv("MOCK_API_KEY")
		if apiKey == "" {
			apiKey = "mock_api_key_123"
		}

		return &models.ServerConfig{
			Server: struct {
				Port int    `json:"port"`
				Host string `json:"host"`
			}{
				Port: 8080,
				Host: "0.0.0.0",
			},
			API: struct {
				ResponseDelay string  `json:"response_delay"`
				ErrorRate     float64 `json:"error_rate"`
				APIKey        string  `json:"api_key"`
			}{
				ResponseDelay: "0ms",
				ErrorRate:     0.0,
				APIKey:        apiKey,
			},
			DefaultUsers: []*models.TeacherInfo{},
		}, nil
	}

	// 读取文件
	data, err := os.ReadFile(f.configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON
	var config models.ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func (f *FileStorage) SaveConfig(config *models.ServerConfig) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(f.configFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(f.configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// InitializeDefaultData 初始化默认数据
func (f *FileStorage) InitializeDefaultData() error {
	// 检查用户数据文件是否存在
	if _, err := os.Stat(f.usersFile); os.IsNotExist(err) {
		// 创建默认用户数据
		defaultUsers := []*models.TeacherInfo{
			{
				TeacherId:      "teacher_001",
				TeacherName:    "张老师",
				SchoolId:       1,
				SchoolName:     "北京中学",
				SubjectId:      10,
				SubjectName:    "数学组",
				GroupName:      "beijing_math_group",
				PreferredModel: "gpt-4",
			},
			{
				TeacherId:      "teacher_002",
				TeacherName:    "李老师",
				SchoolId:       1,
				SchoolName:     "北京中学",
				SubjectId:      2,
				SubjectName:    "语文组",
				GroupName:      "beijing_chinese_group",
				PreferredModel: "gemini-pro",
			},
		}

		if err := f.SaveUsers(defaultUsers); err != nil {
			return fmt.Errorf("保存默认用户数据失败: %v", err)
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(f.configFile); os.IsNotExist(err) {
		// 创建默认配置
		// 从环境变量获取API密钥，如果没有则使用默认值
		apiKey := os.Getenv("MOCK_API_KEY")
		if apiKey == "" {
			apiKey = "mock_api_key_123"
		}

		defaultConfig := &models.ServerConfig{
			Server: struct {
				Port int    `json:"port"`
				Host string `json:"host"`
			}{
				Port: 8080,
				Host: "0.0.0.0",
			},
			API: struct {
				ResponseDelay string  `json:"response_delay"`
				ErrorRate     float64 `json:"error_rate"`
				APIKey        string  `json:"api_key"`
			}{
				ResponseDelay: "0ms",
				ErrorRate:     0.0,
				APIKey:        apiKey,
			},
			DefaultUsers: []*models.TeacherInfo{},
		}

		if err := f.SaveConfig(defaultConfig); err != nil {
			return fmt.Errorf("保存默认配置失败: %v", err)
		}
	}

	return nil
}
