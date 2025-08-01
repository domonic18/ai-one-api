package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"mock-courseware-platform/models"

	"github.com/spf13/viper"
)

var (
	config *models.ServerConfig
)

// LoadConfig 加载配置文件
func LoadConfig() error {
	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("api.response_delay", "0ms")
	viper.SetDefault("api.error_rate", 0.0)
	viper.SetDefault("api.api_key", "mock_api_key_123")

	// 从环境变量读取配置
	viper.SetEnvPrefix("MOCK")
	viper.AutomaticEnv()

	// 绑定环境变量到配置键
	viper.BindEnv("server.port", "MOCK_SERVER_PORT")
	viper.BindEnv("server.host", "MOCK_SERVER_HOST")
	viper.BindEnv("api.api_key", "MOCK_API_KEY")
	viper.BindEnv("api.response_delay", "MOCK_RESPONSE_DELAY")
	viper.BindEnv("api.error_rate", "MOCK_ERROR_RATE")

	// 从配置文件读取
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./data")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
		// 配置文件不存在，使用默认配置
		fmt.Println("配置文件不存在，使用默认配置")
	}

	// 解析配置
	config = &models.ServerConfig{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 如果配置为空，设置默认值
	if config.API.APIKey == "" {
		config.API.APIKey = "mock_api_key_123"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.API.ResponseDelay == "" {
		config.API.ResponseDelay = "0ms"
	}

	// 从环境变量覆盖配置
	if port := os.Getenv("MOCK_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("MOCK_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if apiKey := os.Getenv("MOCK_API_KEY"); apiKey != "" {
		config.API.APIKey = apiKey
		fmt.Printf("Debug: Environment API Key: %s\n", apiKey)
	}
	if responseDelay := os.Getenv("MOCK_RESPONSE_DELAY"); responseDelay != "" {
		config.API.ResponseDelay = responseDelay
	}
	if errorRate := os.Getenv("MOCK_ERROR_RATE"); errorRate != "" {
		if rate, err := strconv.ParseFloat(errorRate, 64); err == nil {
			config.API.ErrorRate = rate
		}
	}

	// 打印最终配置
	fmt.Printf("Debug: Final API Key: %s\n", config.API.APIKey)

	return nil
}

// GetConfig 获取配置
func GetConfig() *models.ServerConfig {
	return config
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	configPath := "./data/config.json"

	// 确保目录存在
	if err := os.MkdirAll("./data", 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetResponseDelay 获取响应延迟时间
func GetResponseDelay() time.Duration {
	if config == nil {
		return 0
	}

	duration, err := time.ParseDuration(config.API.ResponseDelay)
	if err != nil {
		return 0
	}
	return duration
}

// GetErrorRate 获取错误率
func GetErrorRate() float64 {
	if config == nil {
		return 0
	}
	return config.API.ErrorRate
}

// GetAPIKey 获取API密钥
func GetAPIKey() string {
	if config == nil {
		fmt.Printf("Debug: Config is nil, returning default API key\n")
		return "mock_api_key_123"
	}
	fmt.Printf("Debug: GetAPIKey called, returning: %s\n", config.API.APIKey)
	return config.API.APIKey
}

// GetDefaultUsers 获取默认用户列表
func GetDefaultUsers() []*models.TeacherInfo {
	if config == nil {
		return []*models.TeacherInfo{
			{
				TeacherId:      "teacher_001",
				TeacherName:    "张老师",
				SchoolId:       1,
				SchoolName:     "北京中学",
				SubjectId:      10,
				SubjectName:    "数学组",
				OneapiGroup:    "beijing_math_group",
				PreferredModel: "gpt-4",
			},
			{
				TeacherId:      "teacher_002",
				TeacherName:    "李老师",
				SchoolId:       1,
				SchoolName:     "北京中学",
				SubjectId:      2,
				SubjectName:    "语文组",
				OneapiGroup:    "beijing_chinese_group",
				PreferredModel: "gemini-pro",
			},
		}
	}
	return config.DefaultUsers
}
