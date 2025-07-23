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
)

// 测试配置
type TestConfig struct {
	ServerURL string
	Token     string
	UserID    string
	ModelName string
	Temp      float64
	MaxTokens int
	TopP      float64
}

// 请求结构
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Stream      bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	// 解析命令行参数
	var (
		serverURL = flag.String("server", "http://localhost:3000", "OneAPI服务器地址")
		token     = flag.String("token", "", "sk-v1oRcgk6BrXGd0MN61D0Ad09D2E841D7B9B1078328Da270b")
		userID    = flag.String("user", "29be822a-b330-47f3-8dbf-90dab7c3c189", "用户ID")
		modelName = flag.String("model", "smart_select", "要配置的模型名称")
		temp      = flag.Float64("temp", 0.7, "温度参数")
		maxTokens = flag.Int("max-tokens", 2000, "最大token数")
		topP      = flag.Float64("top-p", 0.9, "top_p参数")
		action    = flag.String("action", "test", "操作: setup, test, cleanup")
	)
	flag.Parse()

	if *token == "" {
		fmt.Println("❌ 错误: 必须提供API Token")
		fmt.Println("用法: go run test_smart_model_selection.go -token YOUR_TOKEN")
		os.Exit(1)
	}

	config := &TestConfig{
		ServerURL: *serverURL,
		Token:     *token,
		UserID:    *userID,
		ModelName: *modelName,
		Temp:      *temp,
		MaxTokens: *maxTokens,
		TopP:      *topP,
	}

	switch *action {
	case "setup":
		err := setupSmartModelConfig(config)
		if err != nil {
			fmt.Printf("❌ 设置智能模型配置失败: %v\n", err)
			os.Exit(1)
		}
	case "test":
		err := testSmartModelSelection(config)
		if err != nil {
			fmt.Printf("❌ 测试智能模型选择失败: %v\n", err)
			os.Exit(1)
		}
	case "cleanup":
		err := cleanupSmartModelConfig(config)
		if err != nil {
			fmt.Printf("❌ 清理智能模型配置失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 未知操作: %s\n", *action)
		fmt.Println("支持的操作: setup, test, cleanup")
		os.Exit(1)
	}
}

// 设置智能模型配置
func setupSmartModelConfig(config *TestConfig) error {
	fmt.Println("🔧 开始设置智能模型配置...")

	// 构建请求体
	requestBody := map[string]interface{}{
		"user_id":    config.UserID,
		"model_name": config.ModelName,
		"parameters": map[string]interface{}{
			"temperature": config.Temp,
			"max_tokens":  config.MaxTokens,
			"top_p":       config.TopP,
		},
		"priority": 1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/api/smart-model/", config.ServerURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("设置配置失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ 成功为用户 %s 设置智能模型配置: %s\n", config.UserID, config.ModelName)
	return nil
}

// 测试智能模型选择
func testSmartModelSelection(config *TestConfig) error {
	fmt.Println("🧪 开始测试智能模型选择功能...")

	// 构建聊天请求
	request := ChatRequest{
		Model: "smart_select", // 使用智能选择模型
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "你好，请简单介绍一下自己",
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/v1/chat/completions", config.ServerURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("X-User-ID", config.UserID) // 关键：设置用户ID

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("📡 请求状态码: %d\n", resp.StatusCode)
	fmt.Printf("📡 请求URL: %s\n", url)
	fmt.Printf("📡 请求头 X-User-ID: %s\n", config.UserID)
	fmt.Printf("📡 请求体: %s\n", string(jsonData))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	err = json.Unmarshal(body, &chatResp)
	if err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	fmt.Printf("✅ 智能模型选择测试成功！\n")
	fmt.Printf("📋 响应模型: %s\n", chatResp.Model)
	fmt.Printf("📋 响应ID: %s\n", chatResp.ID)
	fmt.Printf("📋 Token使用情况: 输入=%d, 输出=%d, 总计=%d\n",
		chatResp.Usage.PromptTokens,
		chatResp.Usage.CompletionTokens,
		chatResp.Usage.TotalTokens)

	if len(chatResp.Choices) > 0 {
		fmt.Printf("📋 AI回复: %s\n", chatResp.Choices[0].Message.Content)
	}

	// 验证模型是否被正确替换
	if chatResp.Model == config.ModelName {
		fmt.Printf("✅ 模型替换验证成功: smart_select -> %s\n", config.ModelName)
	} else {
		fmt.Printf("⚠️  模型替换验证失败: 期望 %s, 实际 %s\n", config.ModelName, chatResp.Model)
	}

	return nil
}

// 清理智能模型配置
func cleanupSmartModelConfig(config *TestConfig) error {
	fmt.Println("🧹 开始清理智能模型配置...")

	// 发送删除请求
	url := fmt.Sprintf("%s/api/smart-model/%s", config.ServerURL, config.UserID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("删除配置失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ 成功删除用户 %s 的智能模型配置\n", config.UserID)
	return nil
}
