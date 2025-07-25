package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 消息结构体
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   TokenUsage `json:"usage"`
}

// Choice 选择结构体
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// TokenUsage Token使用情况结构体
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TestParams 测试参数结构体
type TestParams struct {
	Action     string
	Token      string
	UserID     string
	Model      string
	Smart      string
	ServerURL  string
	TestPrompt string
}

func main() {
	// 解析命令行参数
	params := parseFlags()

	// 初始化Redis连接
	if err := initRedis(); err != nil {
		fmt.Printf("❌ Redis初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 执行相应操作
	if err := executeTestAction(params); err != nil {
		fmt.Printf("❌ 测试执行失败: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags 解析命令行参数
func parseFlags() *TestParams {
	var (
		action     = flag.String("action", "help", "操作类型: setup, test, test-fallback, cleanup, all, help")
		token      = flag.String("token", "", "API Token")
		userID     = flag.String("user", "teacher_001", "用户ID")
		model      = flag.String("model", "gpt-4-turbo", "模型名称")
		smart      = flag.String("smart", "true", "是否启用智能选择")
		serverURL  = flag.String("url", "http://localhost:3000", "服务器URL")
		testPrompt = flag.String("prompt", "你好，请简单介绍一下自己", "测试提示词")
	)
	flag.Parse()

	return &TestParams{
		Action:     *action,
		Token:      *token,
		UserID:     *userID,
		Model:      *model,
		Smart:      *smart,
		ServerURL:  *serverURL,
		TestPrompt: *testPrompt,
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
		fmt.Println("⚠️  Redis未启用，部分功能可能受限")
		return nil
	}

	fmt.Println("✅ Redis连接成功！")
	return nil
}

// executeTestAction 执行测试操作
func executeTestAction(params *TestParams) error {
	switch params.Action {
	case "setup":
		return setupUserConfig(params.UserID, params.Model)
	case "test":
		return testSmartModelSelection(params)
	case "test-fallback":
		return testFallbackMechanism(params)
	case "cleanup":
		return cleanupUserConfig(params.UserID)
	case "all":
		return runAllTests(params)
	case "help":
		printUsage()
		return nil
	default:
		fmt.Printf("❌ 未知操作: %s\n", params.Action)
		printUsage()
		return fmt.Errorf("未知操作: %s", params.Action)
	}
}

// setupUserConfig 设置用户配置
func setupUserConfig(userID, modelName string) error {
	if userID == "" || modelName == "" {
		return fmt.Errorf("用户ID和模型名称不能为空")
	}

	selector := model_selection.NewSmartModelSelector()
	config := &model_selection.UserModelConfig{
		UserID:    userID,
		ModelName: modelName,
		Parameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  2048,
		},
		Priority:  1,
		UpdatedAt: time.Now(),
	}

	err := selector.SetUserModelConfig(userID, config)
	if err != nil {
		return fmt.Errorf("设置用户配置失败: %w", err)
	}

	fmt.Printf("✅ 成功设置用户 %s 的模型配置为 %s\n", userID, modelName)
	return nil
}

// testSmartModelSelection 测试智能模型选择
func testSmartModelSelection(params *TestParams) error {
	if params.Token == "" {
		return fmt.Errorf("Token不能为空")
	}

	fmt.Println("🧪 开始测试智能模型选择功能...")

	// 构建请求
	req := ChatRequest{
		Model: params.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: params.TestPrompt,
			},
		},
		Stream: false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/v1/chat/completions", params.ServerURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+params.Token)
	httpReq.Header.Set("X-User-ID", params.UserID)
	httpReq.Header.Set("X-Smart-Model-Selection", params.Smart)

	// 打印调试信息
	fmt.Printf("📡 请求URL: %s\n", httpReq.URL.String())
	fmt.Printf("📡 请求头 X-User-ID: %s\n", params.UserID)
	fmt.Printf("📡 请求头 X-Smart-Model-Selection: %s\n", params.Smart)
	fmt.Printf("📡 请求体: %s\n", string(reqBody))

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 请求状态码: %d\n", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("📋 错误详情: %s\n", string(respBody))
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var chatResp ChatResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("📋 原始响应: %s\n", string(respBody))
		return fmt.Errorf("解析响应失败: %w", err)
	}

	fmt.Println("✅ 智能模型选择测试成功！")
	fmt.Printf("📋 响应模型: %s\n", chatResp.Model)
	fmt.Printf("📋 响应ID: %s\n", chatResp.ID)
	fmt.Printf("📋 Token使用情况: 输入=%d, 输出=%d, 总计=%d\n",
		chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)

	if len(chatResp.Choices) > 0 {
		fmt.Printf("📋 AI回复: %s\n", chatResp.Choices[0].Message.Content)
	}

	// 验证智能选择是否生效
	if params.Smart == "true" {
		// 获取用户配置的模型
		selector := model_selection.NewSmartModelSelector()
		config, err := selector.GetUserModelConfig(params.UserID)
		if err == nil && chatResp.Model == config.ModelName {
			fmt.Printf("✅ 智能选择验证成功: 使用了配置的模型 %s\n", config.ModelName)
		} else if chatResp.Model == params.Model {
			fmt.Printf("⚠️  使用了兜底模型: %s (可能Redis中无配置)\n", params.Model)
		} else {
			fmt.Printf("⚠️  模型替换验证异常: 期望配置模型或兜底模型 %s, 实际 %s\n", params.Model, chatResp.Model)
		}
	} else {
		// 未启用智能选择，应该使用原始模型
		if chatResp.Model == params.Model {
			fmt.Printf("✅ 非智能选择验证成功: 使用了原始模型 %s\n", params.Model)
		} else {
			fmt.Printf("⚠️  非智能选择验证失败: 期望 %s, 实际 %s\n", params.Model, chatResp.Model)
		}
	}

	return nil
}

// testFallbackMechanism 测试兜底机制
func testFallbackMechanism(params *TestParams) error {
	if params.Token == "" {
		return fmt.Errorf("Token不能为空")
	}

	fmt.Println("🧪 开始测试兜底机制...")

	// 先清除用户配置，模拟Redis中无配置的情况
	selector := model_selection.NewSmartModelSelector()
	if err := selector.DeleteUserModelConfig(params.UserID); err != nil {
		fmt.Printf("⚠️  清理用户配置时出错: %v\n", err)
	}

	// 构建请求
	req := ChatRequest{
		Model: params.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: "测试兜底机制，这应该使用原始模型",
			},
		},
		Stream: false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/v1/chat/completions", params.ServerURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+params.Token)
	httpReq.Header.Set("X-User-ID", params.UserID)
	httpReq.Header.Set("X-Smart-Model-Selection", params.Smart)

	fmt.Printf("📡 测试兜底机制 - 原始模型: %s\n", params.Model)
	fmt.Printf("📡 智能选择状态: %s\n", params.Smart)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 请求状态码: %d\n", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("📋 错误详情: %s\n", string(respBody))
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var chatResp ChatResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("📋 原始响应: %s\n", string(respBody))
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 验证兜底机制
	if chatResp.Model == params.Model {
		fmt.Printf("✅ 兜底机制测试成功: 使用了原始模型 %s\n", params.Model)
	} else {
		fmt.Printf("❌ 兜底机制测试失败: 期望 %s, 实际 %s\n", params.Model, chatResp.Model)
	}

	fmt.Printf("📋 Token使用情况: 输入=%d, 输出=%d, 总计=%d\n",
		chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)

	return nil
}

// cleanupUserConfig 清理用户配置
func cleanupUserConfig(userID string) error {
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	selector := model_selection.NewSmartModelSelector()
	err := selector.DeleteUserModelConfig(userID)
	if err != nil {
		return fmt.Errorf("清理用户配置失败: %w", err)
	}

	fmt.Printf("✅ 成功清理用户 %s 的配置\n", userID)
	return nil
}

// runAllTests 运行所有测试
func runAllTests(params *TestParams) error {
	fmt.Println("🚀 开始运行所有智能模型选择测试...")

	// 1. 设置用户配置
	fmt.Println("\n1️⃣ 设置用户配置测试")
	if err := setupUserConfig(params.UserID, params.Model); err != nil {
		return fmt.Errorf("设置用户配置测试失败: %w", err)
	}

	// 2. 测试智能模型选择
	fmt.Println("\n2️⃣ 智能模型选择测试")
	testParams := *params
	testParams.Smart = "true"
	if err := testSmartModelSelection(&testParams); err != nil {
		return fmt.Errorf("智能模型选择测试失败: %w", err)
	}

	// 3. 测试兜底机制
	fmt.Println("\n3️⃣ 兜底机制测试")
	if err := testFallbackMechanism(params); err != nil {
		return fmt.Errorf("兜底机制测试失败: %w", err)
	}

	// 4. 清理配置
	fmt.Println("\n4️⃣ 清理用户配置")
	if err := cleanupUserConfig(params.UserID); err != nil {
		return fmt.Errorf("清理用户配置失败: %w", err)
	}

	fmt.Println("\n🎉 所有测试完成！")
	return nil
}

// printUsage 打印使用帮助
func printUsage() {
	fmt.Println(`
🤖 One-API 智能模型选择测试工具
================================

用法: go run smart_model_test.go [选项]

选项:
  -action string
        操作类型: setup, test, test-fallback, cleanup, all, help (默认 "help")
  -token string
        API Token (必需)
  -user string
        用户ID (默认 "teacher_001")
  -model string
        模型名称 (默认 "gpt-4-turbo")
  -smart string
        是否启用智能选择 (默认 "true")
  -url string
        服务器URL (默认 "http://localhost:3000")
  -prompt string
        测试提示词 (默认 "你好，请简单介绍一下自己")

操作说明:
  setup         - 设置用户模型配置
  test          - 测试智能模型选择功能
  test-fallback - 测试兜底机制
  cleanup       - 清理用户配置
  all           - 运行所有测试
  help          - 显示帮助信息

示例:
  # 设置用户配置
  go run smart_model_test.go -action setup -user teacher_001 -model gpt-4-turbo

  # 测试智能模型选择
  go run smart_model_test.go -action test -token sk-xxxxxx -user teacher_001 -smart true

  # 测试兜底机制
  go run smart_model_test.go -action test-fallback -token sk-xxxxxx -user teacher_001

  # 运行所有测试
  go run smart_model_test.go -action all -token sk-xxxxxx -user teacher_001

  # 清理用户配置
  go run smart_model_test.go -action cleanup -user teacher_001
`)
}