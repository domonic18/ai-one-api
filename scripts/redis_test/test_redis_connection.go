package main

import (
	"fmt"
	"os"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/model_selection"
)

func main() {
	fmt.Println("🔍 测试Redis连接和智能模型选择功能...")

	// 检查Redis是否启用
	fmt.Printf("Redis启用状态: %v\n", common.RedisEnabled)
	fmt.Printf("Redis连接字符串: %s\n", os.Getenv("REDIS_CONN_STRING"))
	fmt.Printf("SYNC_FREQUENCY: %s\n", os.Getenv("SYNC_FREQUENCY"))

	// 初始化Redis客户端
	err := common.InitRedisClient()
	if err != nil {
		fmt.Printf("❌ Redis初始化失败: %v\n", err)
		return
	}

	fmt.Printf("Redis客户端初始化后状态: %v\n", common.RedisEnabled)

	if !common.RedisEnabled {
		fmt.Println("❌ Redis未启用，请检查环境变量")
		return
	}

	// 测试智能模型选择器
	selector := model_selection.NewSmartModelSelector()

	// 测试获取用户配置
	userID := "test_user_001"
	config, err := selector.GetUserModelConfig(userID)
	if err != nil {
		fmt.Printf("❌ 获取用户配置失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 成功获取用户 %s 的配置:\n", userID)
	fmt.Printf("   模型名称: %s\n", config.ModelName)
	fmt.Printf("   参数: %+v\n", config.Parameters)
	fmt.Printf("   优先级: %d\n", config.Priority)
	fmt.Printf("   更新时间: %s\n", config.UpdatedAt)

	// 测试其他用户
	testUsers := []string{"test_user_002", "test_user_003"}
	for _, user := range testUsers {
		config, err := selector.GetUserModelConfig(user)
		if err != nil {
			fmt.Printf("❌ 获取用户 %s 配置失败: %v\n", user, err)
		} else {
			fmt.Printf("✅ 用户 %s 配置: %s\n", user, config.ModelName)
		}
	}

	fmt.Println("🎉 Redis连接和智能模型选择功能测试完成！")
}
