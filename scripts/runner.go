package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScriptInfo 脚本信息结构体
type ScriptInfo struct {
	Name        string
	Description string
	File        string
	Examples    []string
}

// 所有可用脚本
var availableScripts = []ScriptInfo{
	{
		Name:        "redis",
		Description: "Redis助手工具 - 管理用户模型配置",
		File:        "redis/main.go",
		Examples: []string{
			"go run runner.go redis -action list",
			"go run runner.go redis -action batch",
			"go run runner.go redis -action set -user teacher_001 -model gpt-4-turbo",
		},
	},
	{
		Name:        "smart-model",
		Description: "智能模型选择测试工具 - 测试智能模型选择功能",
		File:        "smart-model/main.go",
		Examples: []string{
			"go run runner.go smart-model -action help",
			"go run runner.go smart-model -action setup -user teacher_001 -model gpt-4-turbo",
			"go run runner.go smart-model -action test -token sk-xxxxxx -user teacher_001",
		},
	},
	{
		Name:        "token",
		Description: "令牌状态检查工具 - 检查令牌状态和配额",
		File:        "token/main.go",
		Examples: []string{
			"go run runner.go token -token sk-xxxxxx",
			"go run runner.go token -token sk-xxxxxx -format json",
			"go run runner.go token -token sk-xxxxxx -format brief",
		},
	},
	{
		Name:        "all",
		Description: "运行所有测试脚本",
		File:        "",
		Examples: []string{
			"go run runner.go all",
		},
	},
	{
		Name:        "config",
		Description: "显示当前配置信息",
		File:        "",
		Examples: []string{
			"go run runner.go config",
		},
	},
}

func main() {
	// 解析命令行参数
	scriptName := parseRunnerFlags()

	if scriptName == "help" || scriptName == "" {
		printRunnerUsage()
		return
	}

	// 执行相应脚本
	if err := executeRunnerScript(scriptName); err != nil {
		fmt.Printf("❌ 执行脚本失败: %v\n", err)
		os.Exit(1)
	}
}

// parseRunnerFlags 解析运行器命令行参数
func parseRunnerFlags() string {
	flag.Usage = printRunnerUsage
	flag.Parse()

	if len(flag.Args()) == 0 {
		return "help"
	}

	return flag.Args()[0]
}

// executeRunnerScript 执行指定脚本
func executeRunnerScript(scriptName string) error {
	// 获取当前脚本目录
	scriptDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取脚本目录失败: %w", err)
	}

	// 特殊处理内置命令
	switch scriptName {
	case "all":
		return runAllRunnerScripts(scriptDir)
	case "config":
		return showRunnerConfig()
	}

	// 查找对应的脚本
	for _, script := range availableScripts {
		if script.Name == scriptName {
			scriptPath := filepath.Join(scriptDir, script.File)
			return runRunnerScript(scriptPath, os.Args[2:]...)
		}
	}

	return fmt.Errorf("未知脚本: %s", scriptName)
}

// runRunnerScript 运行指定脚本
func runRunnerScript(scriptPath string, args ...string) error {
	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", scriptPath)
	}

	// 构建命令
	cmd := exec.Command("go", append([]string{"run", scriptPath}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 执行命令
	return cmd.Run()
}

// runAllRunnerScripts 运行所有脚本
func runAllRunnerScripts(scriptDir string) error {
	fmt.Println("🚀 开始运行所有测试脚本...")

	successCount := 0
	for _, script := range availableScripts {
		// 跳过内置命令
		if script.File == "" {
			continue
		}

		fmt.Printf("\n====================\n")
		fmt.Printf("运行脚本: %s\n", script.Description)
		fmt.Printf("====================\n")

		scriptPath := filepath.Join(scriptDir, script.File)
		if err := runRunnerScript(scriptPath, "-help"); err != nil {
			fmt.Printf("⚠️  运行脚本 %s 失败: %v\n", script.Name, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n🎉 完成! 成功运行 %d/%d 个脚本\n", successCount, len(availableScripts)-1)
	return nil
}

// showRunnerConfig 显示当前配置信息
func showRunnerConfig() error {
	fmt.Println("🔧 显示当前配置信息...")

	// 尝试加载配置文件
	configFile := ".env"
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("✅ 找到配置文件: %s\n", configFile)
		// 简化展示，不实际加载内容
		fmt.Println("ℹ️  配置文件内容:")
		content, _ := os.ReadFile(configFile)
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "#") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if strings.Contains(strings.ToLower(key), "token") ||
						strings.Contains(strings.ToLower(key), "password") {
						fmt.Printf("  %s: ***隐藏***\n", key)
					} else {
						fmt.Printf("  %s: %s\n", key, value)
					}
				}
			}
		}
	} else {
		fmt.Println("ℹ️  未找到 .env 配置文件，使用环境变量")
		fmt.Println("💡  可以运行: cp config.env.example .env 来创建配置文件")
	}
	return nil
}

// printRunnerUsage 打印使用帮助
func printRunnerUsage() {
	fmt.Print(`🤖 One-API 辅助脚本运行器
=========================

用法: go run runner.go <脚本名称> [选项]

可用脚本:`)

	for _, script := range availableScripts {
		fmt.Printf("  %-12s - %s\n", script.Name, script.Description)
	}

	fmt.Print(`
示例:
  # 显示所有脚本的帮助信息
  go run runner.go all

  # 运行Redis助手工具
  go run runner.go redis -help

  # 运行智能模型选择测试工具
  go run runner.go smart-model -help

  # 运行令牌状态检查工具
  go run runner.go token -help

  # 显示当前配置
  go run runner.go config

配置文件:
  创建 .env 文件来配置默认值，示例:
  cp config.env.example .env
  
通用选项:
  -help  - 显示帮助信息
`)
}
