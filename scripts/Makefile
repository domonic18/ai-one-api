# One-API 辅助脚本 Makefile
# 用于快速执行各种测试和管理脚本

.PHONY: help redis smart-model token all clean config setup

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## 显示帮助信息
	@echo "$(GREEN)One-API 辅助脚本工具箱$(NC)"
	@echo ""
	@echo "$(YELLOW)使用方法:$(NC)"
	@echo "  make <target>"
	@echo ""
	@echo "$(YELLOW)可用目标:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

redis: ## 运行Redis助手工具
	@echo "$(GREEN)运行Redis助手工具...$(NC)"
	go run runner.go redis -help

smart-model: ## 运行智能模型选择测试工具
	@echo "$(GREEN)运行智能模型选择测试工具...$(NC)"
	go run runner.go smart-model -help

token: ## 运行令牌状态检查工具
	@echo "$(GREEN)运行令牌状态检查工具...$(NC)"
	go run runner.go token -help

all: ## 显示所有脚本帮助信息
	@echo "$(GREEN)显示所有脚本帮助信息...$(NC)"
	go run runner.go all

clean: ## 清理生成的文件
	@echo "$(GREEN)清理生成的文件...$(NC)"
	@find . -name "*.log" -type f -delete
	@echo "$(GREEN)清理完成!$(NC)"

# 测试相关目标
test-redis-set: ## 测试设置Redis用户配置 (需要提供参数)
	@echo "$(YELLOW)示例: make test-redis-set USER_ID=teacher_001 MODEL=gpt-4-turbo$(NC)"
	go run redis_helper.go -action set -user $(USER_ID) -model $(MODEL)

test-redis-get: ## 测试获取Redis用户配置 (需要提供参数)
	@echo "$(YELLOW)示例: make test-redis-get USER_ID=teacher_001$(NC)"
	go run redis_helper.go -action get -user $(USER_ID)

test-redis-batch: ## 测试批量设置Redis配置
	@echo "$(GREEN)批量设置Redis测试配置...$(NC)"
	go run redis_helper.go -action batch



test-token-check: ## 检查令牌状态 (需要提供参数)
	@echo "$(YELLOW)示例: make test-token-check TOKEN=sk-xxxxxx$(NC)"
	go run token_checker.go -token $(TOKEN)

config: ## 显示当前配置信息
	@echo "$(GREEN)显示当前配置信息...$(NC)"
	go run runner.go config

setup: ## 设置配置文件
	@echo "$(GREEN)设置配置文件...$(NC)"
	@if [ ! -f .env ]; then \
		cp config.env.example .env; \
		echo "$(GREEN)已创建 .env 配置文件，请根据需要修改$(NC)"; \
	else \
		echo "$(YELLOW).env 文件已存在$(NC)"; \
	fi

# 便捷目标
run-all-tests: ## 运行所有测试脚本的帮助信息
	@echo "$(GREEN)运行所有测试脚本...$(NC)"
	go run runner.go all

list-scripts: ## 列出所有可用脚本
	@echo "$(GREEN)可用脚本:$(NC)"
	@echo "  - cmd/redis_helper.go      Redis助手工具"

	@echo "  - cmd/token_checker.go     令牌状态检查工具"
	@echo "  - runner.go               统一入口脚本"
	@echo "  - cmd/config_loader.go    配置加载工具"

script-info: ## 显示脚本目录信息
	@echo "$(GREEN)脚本目录信息:$(NC)"
	@ls -la

check-go: ## 检查Go环境
	@echo "$(GREEN)检查Go环境...$(NC)"
	@if command -v go &> /dev/null; then \
		echo "Go版本: $(shell go version)"; \
	else \
		echo "$(RED)未找到Go环境$(NC)"; \
	fi

check-deps: ## 检查依赖环境
	@echo "$(GREEN)检查依赖环境...$(NC)"
	@echo "检查Go环境:"
	@if command -v go &> /dev/null; then \
		echo "  ✅ Go已安装: $(shell go version)"; \
	else \
		echo "  ❌ Go未安装"; \
	fi
	@echo "检查Redis连接:"
	@if [ -n "$$REDIS_CONN_STRING" ]; then \
		echo "  ✅ Redis连接字符串已设置"; \
	else \
		echo "  ⚠️  Redis连接字符串未设置"; \
	fi

env-info: ## 显示环境信息
	@echo "$(GREEN)环境信息:$(NC)"
	@echo "当前目录: $$(pwd)"
	@echo "Go版本: $(shell go version 2>/dev/null || echo '未安装')"
	@echo "Redis连接: $${REDIS_CONN_STRING:-未设置}"
	@echo "数据库连接: $${SQL_DSN:-未设置}"

# 快速测试目标 (需要设置相应的环境变量)
quick-redis-test: ## 快速Redis测试
	@echo "$(GREEN)快速Redis测试...$(NC)"
	go run cmd/redis_helper.go -action list



quick-token-test: ## 快速令牌测试
	@echo "$(GREEN)快速令牌测试...$(NC)"
	go run cmd/token_checker.go -help