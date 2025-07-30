package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	config map[string]string
}

// NewConfigLoader 创建新的配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		config: make(map[string]string),
	}
}

// LoadConfig 加载配置文件
func (cl *ConfigLoader) LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("无法打开配置文件: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			cl.config[key] = value
		}
	}

	return scanner.Err()
}

// Get 获取配置值
func (cl *ConfigLoader) Get(key string) string {
	if value, ok := cl.config[key]; ok {
		return value
	}
	return os.Getenv(key)
}

// GetWithDefault 获取配置值，如果不存在则使用默认值
func (cl *ConfigLoader) GetWithDefault(key, defaultValue string) string {
	if value := cl.Get(key); value != "" {
		return value
	}
	return defaultValue
}

// PrintConfig 打印所有配置（用于调试）
func (cl *ConfigLoader) PrintConfig() {
	fmt.Println("当前配置:")
	for key, value := range cl.config {
		// 隐藏敏感信息
		if strings.Contains(strings.ToLower(key), "token") ||
			strings.Contains(strings.ToLower(key), "password") {
			fmt.Printf("  %s: ***隐藏***\n", key)
		} else {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}
