package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

func main() {
	// 解析命令行参数
	var (
		userID    = flag.String("user", "29be822a-b330-47f3-8dbf-90dab7c3c189", "用户ID")
		modelName = flag.String("model", "qwen-max", "模型名称")
		action    = flag.String("action", "set", "操作: set, get, delete, batch")
		temp      = flag.Float64("temp", 0.7, "温度参数")
		maxTokens = flag.Int("max-tokens", 2000, "最大token数")
		topP      = flag.Float64("top-p", 0.9, "top_p参数")
	)
	flag.Parse()

	// 初始化Redis连接
	fmt.Println("正在初始化Redis连接...")
	err := common.InitRedisClient()
	if err != nil {
		fmt.Printf("初始化Redis失败: %v\n", err)
		os.Exit(1)
	}

	if !common.RedisEnabled {
		fmt.Println("Redis未启用，请检查环境变量REDIS_CONN_STRING")
		os.Exit(1)
	}

	fmt.Println("Redis连接成功！")

	// 创建选择器
	selector := model_selection.NewSmartModelSelector()

	switch *action {
	case "set":
		err := setUserConfig(selector, *userID, *modelName, *temp, *maxTokens, *topP)
		if err != nil {
			fmt.Printf("设置用户配置失败: %v\n", err)
			os.Exit(1)
		}
	case "get":
		err := getUserConfig(selector, *userID)
		if err != nil {
			fmt.Printf("获取用户配置失败: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		err := deleteUserConfig(selector, *userID)
		if err != nil {
			fmt.Printf("删除用户配置失败: %v\n", err)
			os.Exit(1)
		}
	case "batch":
		err := batchSetTestConfigs(selector)
		if err != nil {
			fmt.Printf("批量设置测试配置失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("未知操作: %s\n", *action)
		printUsage()
		os.Exit(1)
	}
}

func setUserConfig(selector *model_selection.SmartModelSelector, userID, modelName string, temp float64, maxTokens int, topP float64) error {
	config := &model_selection.UserModelConfig{
		UserID:    userID,
		ModelName: modelName,
		Parameters: map[string]interface{}{
			"temperature": temp,
			"max_tokens":  maxTokens,
			"top_p":       topP,
		},
		Priority:  1,
		UpdatedAt: time.Now(),
	}

	err := selector.SetUserModelConfig(userID, config)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 成功为用户 %s 设置模型配置:\n", userID)
	printConfig(config)
	return nil
}

func getUserConfig(selector *model_selection.SmartModelSelector, userID string) error {
	config, err := selector.GetUserModelConfig(userID)
	if err != nil {
		return err
	}

	fmt.Printf("📋 用户 %s 的模型配置:\n", userID)
	printConfig(config)
	return nil
}

func deleteUserConfig(selector *model_selection.SmartModelSelector, userID string) error {
	err := selector.DeleteUserModelConfig(userID)
	if err != nil {
		return err
	}

	fmt.Printf("🗑️ 成功删除用户 %s 的模型配置\n", userID)
	return nil
}

func batchSetTestConfigs(selector *model_selection.SmartModelSelector) error {
	testConfigs := []struct {
		userID    string
		modelName string
		temp      float64
		maxTokens int
		topP      float64
	}{
		{"29be822a-b330-47f3-8dbf-90dab7c3c189", "qwen-max", 0.8, 4000, 0.9},
		{"test_user_002", "claude-3-sonnet", 0.7, 3000, 0.95},
		{"test_user_003", "gpt-4-turbo", 0.6, 8000, 0.85},
		{"user_12345", "deepseek-chat", 0.75, 2000, 0.9},
		{"user_67890", "moonshot-v1-8k", 0.5, 8000, 0.8},
	}

	fmt.Println("🚀 开始批量设置测试配置...")

	for _, tc := range testConfigs {
		config := &model_selection.UserModelConfig{
			UserID:    tc.userID,
			ModelName: tc.modelName,
			Parameters: map[string]interface{}{
				"temperature": tc.temp,
				"max_tokens":  tc.maxTokens,
				"top_p":       tc.topP,
			},
			Priority:  1,
			UpdatedAt: time.Now(),
		}

		err := selector.SetUserModelConfig(tc.userID, config)
		if err != nil {
			fmt.Printf("❌ 设置用户 %s 配置失败: %v\n", tc.userID, err)
		} else {
			fmt.Printf("✅ 设置用户 %s 配置成功\n", tc.userID)
		}
	}

	fmt.Println("🎉 批量设置完成！")
	return nil
}

func printConfig(config *model_selection.UserModelConfig) {
	jsonData, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println(string(jsonData))
}

func printUsage() {
	fmt.Print(`
用法: go run redis_test_tool.go [选项]

选项:
  -user string
        用户ID (默认 "test_user_001")
  -model string
        模型名称 (默认 "gpt-4")
  -action string
        操作类型: set, get, delete, batch (默认 "set")
  -temp float
        温度参数 (默认 0.7)
  -max-tokens int
        最大token数 (默认 2000)
  -top-p float
        top_p参数 (默认 0.9)

示例:
  # 设置用户配置
  go run redis_test_tool.go -user user_123 -model gpt-4 -temp 0.8 -max-tokens 4000

  # 获取用户配置
  go run redis_test_tool.go -action get -user user_123

  # 删除用户配置
  go run redis_test_tool.go -action delete -user user_123

  # 批量设置测试配置
  go run redis_test_tool.go -action batch
`)
}
