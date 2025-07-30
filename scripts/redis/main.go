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

// ConfigParams 配置参数结构体
type ConfigParams struct {
	UserID    string
	ModelName string
	Temp      float64
	MaxTokens int
	TopP      float64
	Action    string
}

// TestUserConfig 测试用户配置
type TestUserConfig struct {
	UserID    string
	ModelName string
	Temp      float64
	MaxTokens int
	TopP      float64
}

// 预置测试数据
var presetTestConfigs = []TestUserConfig{
	{"teacher_001", "gpt-4-turbo", 0.7, 4000, 0.9},
	{"teacher_002", "claude-3-sonnet", 0.6, 3000, 0.85},
	{"teacher_003", "qwen-max", 0.8, 2000, 0.95},
	{"student_001", "deepseek-chat", 0.5, 1000, 0.8},
	{"student_002", "moonshot-v1-8k", 0.7, 8000, 0.9},
}

func main() {
	// 解析命令行参数
	params := parseFlags()

	// 如果是帮助命令，直接显示帮助信息
	if params.Action == "help" {
		printUsage()
		return
	}

	// 初始化Redis连接
	if err := initRedis(); err != nil {
		fmt.Printf("❌ Redis初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 创建选择器
	selector := model_selection.NewSmartModelSelector()

	// 执行相应操作
	if err := executeAction(selector, params); err != nil {
		fmt.Printf("❌ 操作执行失败: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags 解析命令行参数
func parseFlags() *ConfigParams {
	var (
		userID    = flag.String("user", "teacher_001", "用户ID")
		modelName = flag.String("model", "gpt-4-turbo", "模型名称")
		action    = flag.String("action", "help", "操作: set, get, delete, batch, list, help")
		temp      = flag.Float64("temp", 0.7, "温度参数")
		maxTokens = flag.Int("max-tokens", 2000, "最大token数")
		topP      = flag.Float64("top-p", 0.9, "top_p参数")
	)
	flag.Parse()

	return &ConfigParams{
		UserID:    *userID,
		ModelName: *modelName,
		Temp:      *temp,
		MaxTokens: *maxTokens,
		TopP:      *topP,
		Action:    *action,
	}
}

// initRedis 初始化Redis连接
func initRedis() error {
	fmt.Println("🔍 正在初始化Redis连接...")
	err := common.InitRedisClient()
	if err != nil {
		return fmt.Errorf("初始化Redis客户端失败: %w", err)
	}

	if !common.RedisEnabled {
		return fmt.Errorf("Redis未启用，请检查环境变量REDIS_CONN_STRING")
	}

	fmt.Println("✅ Redis连接成功！")
	return nil
}

// executeAction 执行相应操作
func executeAction(selector *model_selection.SmartModelSelector, params *ConfigParams) error {
	switch params.Action {
	case "set":
		return setUserConfig(selector, params)
	case "get":
		return getUserConfig(selector, params.UserID)
	case "delete":
		return deleteUserConfig(selector, params.UserID)
	case "batch":
		return batchSetTestConfigs(selector)
	case "list":
		return listPresetConfigs()
	case "help":
		printUsage()
		return nil
	default:
		fmt.Printf("❌ 未知操作: %s\n", params.Action)
		printUsage()
		return fmt.Errorf("未知操作: %s", params.Action)
	}
}

// setUserConfig 设置用户配置
func setUserConfig(selector *model_selection.SmartModelSelector, params *ConfigParams) error {
	config := &model_selection.UserModelConfig{
		UserID:    params.UserID,
		ModelName: params.ModelName,
		Parameters: map[string]interface{}{
			"temperature": params.Temp,
			"max_tokens":  params.MaxTokens,
			"top_p":       params.TopP,
		},
		Priority:  1,
		UpdatedAt: time.Now(),
	}

	err := selector.SetUserModelConfig(params.UserID, config)
	if err != nil {
		return fmt.Errorf("设置用户配置失败: %w", err)
	}

	fmt.Printf("✅ 成功为用户 %s 设置模型配置:\n", params.UserID)
	printConfig(config)
	return nil
}

// getUserConfig 获取用户配置
func getUserConfig(selector *model_selection.SmartModelSelector, userID string) error {
	config, err := selector.GetUserModelConfig(userID)
	if err != nil {
		return fmt.Errorf("获取用户配置失败: %w", err)
	}

	fmt.Printf("📋 用户 %s 的模型配置:\n", userID)
	printConfig(config)
	return nil
}

// deleteUserConfig 删除用户配置
func deleteUserConfig(selector *model_selection.SmartModelSelector, userID string) error {
	err := selector.DeleteUserModelConfig(userID)
	if err != nil {
		return fmt.Errorf("删除用户配置失败: %w", err)
	}

	fmt.Printf("🗑️ 成功删除用户 %s 的模型配置\n", userID)
	return nil
}

// batchSetTestConfigs 批量设置测试配置
func batchSetTestConfigs(selector *model_selection.SmartModelSelector) error {
	fmt.Println("🚀 开始批量设置测试配置...")
	successCount := 0

	for _, tc := range presetTestConfigs {
		config := &model_selection.UserModelConfig{
			UserID:    tc.UserID,
			ModelName: tc.ModelName,
			Parameters: map[string]interface{}{
				"temperature": tc.Temp,
				"max_tokens":  tc.MaxTokens,
				"top_p":       tc.TopP,
			},
			Priority:  1,
			UpdatedAt: time.Now(),
		}

		err := selector.SetUserModelConfig(tc.UserID, config)
		if err != nil {
			fmt.Printf("❌ 设置用户 %s 配置失败: %v\n", tc.UserID, err)
		} else {
			fmt.Printf("✅ 设置用户 %s 配置成功 (模型: %s)\n", tc.UserID, tc.ModelName)
			successCount++
		}
	}

	fmt.Printf("🎉 批量设置完成！成功: %d, 失败: %d\n", successCount, len(presetTestConfigs)-successCount)
	return nil
}

// listPresetConfigs 列出预置配置
func listPresetConfigs() error {
	fmt.Println("📋 预置测试配置列表:")
	for i, tc := range presetTestConfigs {
		fmt.Printf("%d. 用户ID: %s, 模型: %s, 温度: %.1f, 最大Token: %d, TopP: %.1f\n",
			i+1, tc.UserID, tc.ModelName, tc.Temp, tc.MaxTokens, tc.TopP)
	}
	return nil
}

// printConfig 打印配置信息
func printConfig(config *model_selection.UserModelConfig) {
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("序列化配置信息失败: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// printUsage 打印使用帮助
func printUsage() {
	fmt.Print(`
🤖 One-API Redis助手工具
========================

用法: go run redis_helper.go [选项]

选项:
  -user string
        用户ID (默认 "teacher_001")
  -model string
        模型名称 (默认 "gpt-4-turbo")
  -action string
        操作类型: set, get, delete, batch, list, help (默认 "help")
  -temp float
        温度参数 (默认 0.7)
  -max-tokens int
        最大token数 (默认 2000)
  -top-p float
        top_p参数 (默认 0.9)

操作说明:
  set    - 设置用户模型配置
  get    - 获取用户模型配置
  delete - 删除用户模型配置
  batch  - 批量设置预置测试配置
  list   - 列出预置测试配置
  help   - 显示帮助信息

示例:
  # 设置用户配置
  go run redis_helper.go -action set -user teacher_001 -model gpt-4-turbo -temp 0.8 -max-tokens 4000

  # 获取用户配置
  go run redis_helper.go -action get -user teacher_001

  # 删除用户配置
  go run redis_helper.go -action delete -user teacher_001

  # 批量设置测试配置
  go run redis_helper.go -action batch

  # 列出预置配置
  go run redis_helper.go -action list
`)
}
