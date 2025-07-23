package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   TokenUsage `json:"usage"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func main() {
	var (
		action = flag.String("action", "test", "操作类型: setup, test, test-fallback, cleanup")
		token  = flag.String("token", "", "API Token")
		userID = flag.String("user", "", "用户ID")
		model  = flag.String("model", "deepseek-chat", "模型名称")
		smart  = flag.String("smart", "true", "是否启用智能选择")
	)
	flag.Parse()

	// 初始化Redis连接
	common.InitRedisClient()

	switch *action {
	case "setup":
		setupUserConfig(*userID, *model)
	case "test":
		testSmartModelSelection(*token, *userID, *model, *smart)
	case "test-fallback":
		testFallbackMechanism(*token, *userID, *model, *smart)
	case "cleanup":
		cleanupUserConfig(*userID)
	default:
		fmt.Println("无效的操作类型，支持: setup, test, test-fallback, cleanup")
		os.Exit(1)
	}
}

func setupUserConfig(userID, modelName string) {
	if userID == "" || modelName == "" {
		fmt.Println("❌ 用户ID和模型名称不能为空")
		os.Exit(1)
	}

	selector := model_selection.NewSmartModelSelector()
	config := &model_selection.UserModelConfig{
		ModelName: modelName,
		Parameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  2048,
		},
	}

	err := selector.SetUserModelConfig(userID, config)
	if err != nil {
		fmt.Printf("❌ 设置用户配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 成功设置用户 %s 的模型配置为 %s\n", userID, modelName)
}

func testSmartModelSelection(token, userID, originalModel, smartEnabled string) {
	if token == "" || userID == "" {
		fmt.Println("❌ Token和用户ID不能为空")
		os.Exit(1)
	}

	fmt.Println("🧪 开始测试智能模型选择功能...")

	// 构建请求
	req := ChatRequest{
		Model: originalModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: "你好，请简单介绍一下自己",
			},
		},
		Stream: false,
	}

	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "http://localhost:3000/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		os.Exit(1)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Smart-Model-Selection", smartEnabled) // 新增：智能选择控制头

	// 打印调试信息
	fmt.Printf("📡 请求URL: %s\n", httpReq.URL.String())
	fmt.Printf("📡 请求头 X-User-ID: %s\n", userID)
	fmt.Printf("📡 请求头 X-Smart-Model-Selection: %s\n", smartEnabled)
	fmt.Printf("📡 请求体: %s\n", string(reqBody))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ 发送请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 请求状态码: %d\n", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("📋 错误详情: %s\n", string(respBody))
		os.Exit(1)
	}

	// 解析响应
	var chatResp ChatResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("📋 原始响应: %s\n", string(respBody))
		os.Exit(1)
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
	if smartEnabled == "true" {
		// 获取用户配置的模型
		selector := model_selection.NewSmartModelSelector()
		config, err := selector.GetUserModelConfig(userID)
		if err == nil && chatResp.Model == config.ModelName {
			fmt.Printf("✅ 智能选择验证成功: 使用了配置的模型 %s\n", config.ModelName)
		} else if chatResp.Model == originalModel {
			fmt.Printf("⚠️  使用了兜底模型: %s (可能Redis中无配置)\n", originalModel)
		} else {
			fmt.Printf("⚠️  模型替换验证异常: 期望配置模型或兜底模型 %s, 实际 %s\n", originalModel, chatResp.Model)
		}
	} else {
		// 未启用智能选择，应该使用原始模型
		if chatResp.Model == originalModel {
			fmt.Printf("✅ 非智能选择验证成功: 使用了原始模型 %s\n", originalModel)
		} else {
			fmt.Printf("⚠️  非智能选择验证失败: 期望 %s, 实际 %s\n", originalModel, chatResp.Model)
		}
	}
}

func testFallbackMechanism(token, userID, originalModel, smartEnabled string) {
	if token == "" || userID == "" {
		fmt.Println("❌ Token和用户ID不能为空")
		os.Exit(1)
	}

	fmt.Println("🧪 开始测试兜底机制...")

	// 先清除用户配置，模拟Redis中无配置的情况
	selector := model_selection.NewSmartModelSelector()
	selector.DeleteUserModelConfig(userID)

	// 构建请求
	req := ChatRequest{
		Model: originalModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: "测试兜底机制，这应该使用原始模型",
			},
		},
		Stream: false,
	}

	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "http://localhost:3000/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		os.Exit(1)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Smart-Model-Selection", smartEnabled)

	fmt.Printf("📡 测试兜底机制 - 原始模型: %s\n", originalModel)
	fmt.Printf("📡 智能选择状态: %s\n", smartEnabled)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ 发送请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 请求状态码: %d\n", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("📋 错误详情: %s\n", string(respBody))
		os.Exit(1)
	}

	// 解析响应
	var chatResp ChatResponse
	err = json.Unmarshal(respBody, &chatResp)
	if err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("📋 原始响应: %s\n", string(respBody))
		os.Exit(1)
	}

	// 验证兜底机制
	if chatResp.Model == originalModel {
		fmt.Printf("✅ 兜底机制测试成功: 使用了原始模型 %s\n", originalModel)
	} else {
		fmt.Printf("❌ 兜底机制测试失败: 期望 %s, 实际 %s\n", originalModel, chatResp.Model)
	}

	fmt.Printf("📋 Token使用情况: 输入=%d, 输出=%d, 总计=%d\n",
		chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)
}

func cleanupUserConfig(userID string) {
	if userID == "" {
		fmt.Println("❌ 用户ID不能为空")
		os.Exit(1)
	}

	selector := model_selection.NewSmartModelSelector()
	err := selector.DeleteUserModelConfig(userID)
	if err != nil {
		fmt.Printf("❌ 清理用户配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 成功清理用户 %s 的配置\n", userID)
}
