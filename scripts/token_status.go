package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
)

func main() {
	var tokenKey string
	flag.StringVar(&tokenKey, "token", "", "令牌密钥")
	flag.Parse()

	if tokenKey == "" {
		log.Fatal("请提供令牌密钥: -token <your-token>")
	}

	// 初始化数据库连接
	model.InitDB()

	// 初始化Redis
	common.InitRedisClient()

	// 验证令牌
	token, err := model.ValidateUserToken(tokenKey)
	if err != nil {
		fmt.Printf("❌ 令牌验证失败: %s\n", err.Error())
		return
	}

	fmt.Printf("✅ 令牌验证成功\n")
	fmt.Printf("令牌ID: %d\n", token.Id)
	fmt.Printf("用户ID: %d\n", token.UserId)
	fmt.Printf("令牌名称: %s\n", token.Name)
	fmt.Printf("剩余配额: %d\n", token.RemainQuota)
	fmt.Printf("已用配额: %d\n", token.UsedQuota)
	fmt.Printf("无限配额: %t\n", token.UnlimitedQuota)
	fmt.Printf("令牌状态: %d\n", token.Status)
	fmt.Printf("过期时间: %d\n", token.ExpiredTime)

	// 检查配额是否足够
	if token.RemainQuota <= 0 && !token.UnlimitedQuota {
		fmt.Printf("⚠️  警告: 令牌配额已用尽\n")
	} else {
		fmt.Printf("✅ 令牌配额充足\n")
	}
}
