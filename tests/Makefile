# One-API 测试 Makefile
.PHONY: test unit-test integration-test coverage clean setup teardown help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
GO_VERSION := $(shell go version)
TEST_TIMEOUT := 30m
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# 颜色输出
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## 显示帮助信息
	@echo "$(GREEN)One-API 测试命令$(NC)"
	@echo ""
	@echo "$(YELLOW)测试命令:$(NC)"
	@echo "  make test          - 运行所有测试"
	@echo "  make unit-test     - 运行单元测试"
	@echo "  make integration   - 运行集成测试"
	@echo "  make coverage      - 生成覆盖率报告"
	@echo "  make benchmark     - 运行性能测试"
	@echo "  make security      - 运行安全扫描"
	@echo "  make lint          - 运行代码检查"
	@echo ""
	@echo "$(YELLOW)环境管理:$(NC)"
	@echo "  make setup         - 设置测试环境"
	@echo "  make teardown      - 清理测试环境"
	@echo "  make clean         - 清理测试产物"

# 环境检查
check-env: ## 检查测试环境
	@echo "$(GREEN)检查测试环境...$(NC)"
	@echo "Go版本: $(GO_VERSION)"
	@go version >/dev/null 2>&1 || (echo "$(RED)错误: Go未安装$(NC)" && exit 1)
	@mysql --version >/dev/null 2>&1 && echo "MySQL: $$(mysql --version | head -1)" || echo "$(YELLOW)警告: MySQL未安装$(NC)"
	@redis-cli --version >/dev/null 2>&1 && echo "Redis: $$(redis-cli --version | head -1)" || echo "$(YELLOW)警告: Redis未安装$(NC)"

# 测试运行
test: check-env ## 运行所有测试
	@echo "$(GREEN)运行所有测试...$(NC)"
	go test -v -timeout=$(TEST_TIMEOUT) -race ./tests/...

unit-test: check-env ## 运行单元测试
	@echo "$(GREEN)运行单元测试...$(NC)"
	go test -v -timeout=$(TEST_TIMEOUT) -race ./tests/unit/...

integration-test: check-env ## 运行集成测试
	@echo "$(GREEN)运行集成测试...$(NC)"
	go test -v -timeout=$(TEST_TIMEOUT) -tags=integration ./tests/integration/...

# 覆盖率相关
coverage: ## 生成覆盖率报告
	@echo "$(GREEN)生成覆盖率报告...$(NC)"
	go test -v -race -coverprofile=$(COVERAGE_FILE) ./tests/...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)覆盖率报告已生成: $(COVERAGE_HTML)$(NC)"

coverage-summary: ## 显示覆盖率摘要
	@echo "$(GREEN)覆盖率摘要:$(NC)"
	go test -cover ./tests/... | grep -v "coverage: 0.0%"

# 性能测试
benchmark: check-env ## 运行性能基准测试
	@echo "$(GREEN)运行性能基准测试...$(NC)"
	go test -bench=. -benchmem ./tests/...

# 代码质量
lint: ## 运行代码检查
	@echo "$(GREEN)运行代码检查...$(NC)"
	@golangci-lint run --timeout=5m || echo "$(YELLOW)警告: golangci-lint未安装$(NC)"

fmt: ## 格式化代码
	@echo "$(GREEN)格式化代码...$(NC)"
	gofmt -s -w .

vet: ## 运行go vet
	@echo "$(GREEN)运行go vet...$(NC)"
	go vet ./...

# 安全扫描
security: ## 运行安全扫描
	@echo "$(GREEN)运行安全扫描...$(NC)"
	@gosec ./... || echo "$(YELLOW)警告: gosec未安装$(NC)"

# 环境管理
setup: ## 设置测试环境
	@echo "$(GREEN)设置测试环境...$(NC)"
	@echo "确保MySQL和Redis已启动"
	@echo "测试数据库: oneapi_test"
	@echo "Redis端口: 6379"

setup-docker: ## 使用Docker设置测试环境
	@echo "$(GREEN)使用Docker设置测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d || echo "$(RED)错误: Docker环境启动失败$(NC)"

teardown: ## 清理测试环境
	@echo "$(GREEN)清理测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml down || true

clean: ## 清理测试产物
	@echo "$(GREEN)清理测试产物...$(NC)"
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	go clean -testcache

# 数据库测试
db-test: check-env ## 运行数据库相关测试
	@echo "$(GREEN)运行数据库测试...$(NC)"
	go test -v ./tests/unit/user_test.go ./tests/unit/token_test.go ./tests/unit/channel_test.go

# 业务逻辑测试
business-test: check-env ## 运行业务逻辑测试
	@echo "$(GREEN)运行业务逻辑测试...$(NC)"
	go test -v ./tests/unit/quota_test.go ./tests/unit/log_test.go

# 错误处理测试
error-test: check-env ## 运行错误处理测试
	@echo "$(GREEN)运行错误处理测试...$(NC)"
	go test -v ./tests/unit/error_test.go

# 并行测试
parallel-test: check-env ## 运行并行测试
	@echo "$(GREEN)运行并行测试...$(NC)"
	go test -v -parallel=4 ./tests/...

# 详细测试输出
verbose-test: check-env ## 运行详细测试
	@echo "$(GREEN)运行详细测试...$(NC)"
	go test -v -race -cover -timeout=$(TEST_TIMEOUT) ./tests/...

# 快速测试（跳过集成测试）
fast-test: check-env ## 运行快速测试（仅单元测试）
	@echo "$(GREEN)运行快速测试...$(NC)"
	go test -short ./tests/unit/...

# 持续集成测试
ci-test: lint vet security unit-test integration-test coverage-summary ## CI完整测试流程
	@echo "$(GREEN)CI测试完成！$(NC)"

# 帮助信息
help-test: ## 显示测试相关信息
	@echo "$(GREEN)One-API 测试信息$(NC)"
	@echo ""
	@echo "$(YELLOW)测试结构:$(NC)"
	@echo "  tests/unit/        - 单元测试"
	@echo "  tests/integration/ - 集成测试"
	@echo "  tests/fixtures/    - 测试数据"
	@echo ""
	@echo "$(YELLOW)环境变量:$(NC)"
	@echo "  SQL_DSN           - MySQL连接字符串"
	@echo "  REDIS_CONN_STRING - Redis连接字符串"
	@echo "  SESSION_SECRET    - 会话密钥"
	@echo ""
	@echo "$(YELLOW)常用命令:$(NC)"
	@echo "  make unit-test     - 快速单元测试"
	@echo "  make coverage      - 生成覆盖率报告"
	@echo "  make benchmark     - 性能测试"
	@echo "  make clean         - 清理测试产物"

# 目标依赖关系
test: unit-test integration-test
ci: ci-test