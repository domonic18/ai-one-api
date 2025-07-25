package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/model"
)

// TokenCheckParams 令牌检查参数结构体
type TokenCheckParams struct {
	TokenKey string
	Format   string
}

func main() {
	// 解析命令行参数
	params := parseFlags()

	// 初始化数据库连接
	if err := initDatabase(); err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化Redis
	if err := initRedis(); err != nil {
		fmt.Printf("❌ Redis初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 检查令牌状态
	if err := checkTokenStatus(params); err != nil {
		fmt.Printf("❌ 令牌检查失败: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags 解析命令行参数
func parseFlags() *TokenCheckParams {
	var tokenKey, format string
	var showHelp bool
	
	flag.StringVar(&tokenKey, "token", "", "令牌密钥")
	flag.StringVar(&format, "format", "detailed", "输出格式: brief, detailed, json")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if tokenKey == "" {
		fmt.Println("❌ 请提供令牌密钥: -token <your-token>")
		printUsage()
		os.Exit(1)
	}

	return &TokenCheckParams{
		TokenKey: tokenKey,
		Format:   format,
	}
}

// initDatabase 初始化数据库连接
func initDatabase() error {
	fmt.Println("🔍 正在初始化数据库连接...")
	model.InitDB()
	fmt.Println("✅ 数据库连接成功！")
	return nil
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

// checkTokenStatus 检查令牌状态
func checkTokenStatus(params *TokenCheckParams) error {
	// 验证令牌
	token, err := model.ValidateUserToken(params.TokenKey)
	if err != nil {
		return fmt.Errorf("令牌验证失败: %w", err)
	}

	// 根据格式输出结果
	switch params.Format {
	case "brief":
		printBriefTokenInfo(token)
	case "json":
		printJSONTokenInfo(token)
	default:
		printDetailedTokenInfo(token)
	}

	return nil
}

// printDetailedTokenInfo 打印详细令牌信息
func printDetailedTokenInfo(token *model.Token) {
	fmt.Println("✅ 令牌验证成功")
	fmt.Println("========================")
	fmt.Printf("令牌ID: %d\n", token.Id)
	fmt.Printf("用户ID: %d\n", token.UserId)
	fmt.Printf("令牌名称: %s\n", token.Name)
	
	// 格式化配额显示
	if config.DisplayInCurrencyEnabled {
		fmt.Printf("剩余配额: %.6f 美元\n", float64(token.RemainQuota)/config.QuotaPerUnit)
		fmt.Printf("已用配额: %.6f 美元\n", float64(token.UsedQuota)/config.QuotaPerUnit)
	} else {
		fmt.Printf("剩余配额: %d 点\n", token.RemainQuota)
		fmt.Printf("已用配额: %d 点\n", token.UsedQuota)
	}
	
	fmt.Printf("无限配额: %t\n", token.UnlimitedQuota)
	
	// 格式化状态显示
	statusText := "未知"
	switch token.Status {
	case 1:
		statusText = "启用"
	case 2:
		statusText = "禁用"
	case 3:
		statusText = "已过期"
	}
	fmt.Printf("令牌状态: %s (%d)\n", statusText, token.Status)
	
	// 格式化过期时间显示
	if token.ExpiredTime == -1 {
		fmt.Printf("过期时间: 永不过期\n")
	} else {
		expireTime := time.Unix(token.ExpiredTime, 0)
		fmt.Printf("过期时间: %s\n", expireTime.Format("2006-01-02 15:04:05"))
	}
	
	// 检查配额是否足够
	if token.RemainQuota <= 0 && !token.UnlimitedQuota {
		fmt.Printf("⚠️  警告: 令牌配额已用尽\n")
	} else {
		fmt.Printf("✅ 令牌配额充足\n")
	}
}

// printBriefTokenInfo 打印简要令牌信息
func printBriefTokenInfo(token *model.Token) {
	fmt.Printf("令牌名称: %s | 状态: %d | 剩余配额: %d | 已用配额: %d\n", 
		token.Name, token.Status, token.RemainQuota, token.UsedQuota)
	
	if token.RemainQuota <= 0 && !token.UnlimitedQuota {
		fmt.Println("⚠️  配额已用尽")
	} else {
		fmt.Println("✅ 配额充足")
	}
}

// printJSONTokenInfo 打印JSON格式令牌信息
func printJSONTokenInfo(token *model.Token) {
	// 创建一个简化版本的令牌结构用于JSON输出
	type SimpleToken struct {
		ID           int    `json:"id"`
		UserID       int    `json:"user_id"`
		Name         string `json:"name"`
		RemainQuota  int64  `json:"remain_quota"`
		UsedQuota    int64  `json:"used_quota"`
		Unlimited    bool   `json:"unlimited_quota"`
		Status       int    `json:"status"`
		ExpiredTime  int64  `json:"expired_time"`
		QuotaSufficient bool  `json:"quota_sufficient"`
	}
	
	simpleToken := SimpleToken{
		ID:           token.Id,
		UserID:       token.UserId,
		Name:         token.Name,
		RemainQuota:  token.RemainQuota,
		UsedQuota:    token.UsedQuota,
		Unlimited:    token.UnlimitedQuota,
		Status:       token.Status,
		ExpiredTime:  token.ExpiredTime,
		QuotaSufficient: !(token.RemainQuota <= 0 && !token.UnlimitedQuota),
	}
	
	jsonData, err := json.MarshalIndent(simpleToken, "", "  ")
	if err != nil {
		fmt.Printf("序列化JSON失败: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

// printUsage 打印使用帮助
func printUsage() {
	fmt.Println(`
🤖 One-API 令牌状态检查工具
=========================

用法: go run token_checker.go [选项]

选项:
  -token string
        令牌密钥 (必需)
  -format string
        输出格式: brief, detailed, json (默认 "detailed")
  -help
        显示帮助信息

输出格式说明:
  brief   - 简要信息
  detailed - 详细信息 (默认)
  json    - JSON格式

示例:
  # 检查令牌状态 (详细格式)
  go run token_checker.go -token sk-xxxxxx

  # 检查令牌状态 (简要格式)
  go run token_checker.go -token sk-xxxxxx -format brief

  # 检查令牌状态 (JSON格式)
  go run token_checker.go -token sk-xxxxxx -format json
`)
}