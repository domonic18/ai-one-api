# One-API 辅助测试脚本

这个目录包含了用于测试和调试 One-API 项目的辅助脚本工具，经过优化后提供了更好的可用性和统一的使用体验。

## 📁 优化后的目录结构

```
scripts/
├── main.go                 # 统一入口脚本（主控制台）
├── Makefile               # 快速命令构建工具
├── config.env.example     # 配置文件模板
├── README.md             # 使用文档（本文件）
├── .env                  # 配置文件（用户创建）
└── cmd/                  # 独立功能脚本
    ├── redis_helper.go   # Redis助手工具

    ├── token_checker.go  # 令牌状态检查
    └── config_loader.go  # 配置加载工具
```

## 🚀 快速开始

### 1. 初始化配置
```bash
cd scripts/
make setup  # 自动创建 .env 配置文件
```

### 2. 使用统一入口（推荐）
```bash
# 查看所有可用命令
go run main.go help

# 使用Makefile快捷方式
make help           # 显示帮助
make redis          # Redis助手

make token          # 令牌检查
make config         # 显示配置
```

## 📋 功能概览

### 统一入口脚本 (`main.go`)
提供统一的命令行界面，避免直接运行独立脚本的冲突问题。

**使用方式：**
```bash
go run main.go <脚本名称> [参数...]

# 示例
go run main.go redis -action list
go run main.go token -token sk-xxxxxx
go run main.go all                    # 运行所有脚本帮助
go run main.go config                # 显示当前配置
```

### 1. Redis助手工具 (`redis`)
**文件位置：** `cmd/redis_helper.go`

**功能特性：**
- ✅ 设置用户模型配置
- ✅ 获取用户模型配置
- ✅ 删除用户模型配置
- ✅ 批量设置预置测试配置
- ✅ 列出预置测试配置

**使用示例：**
```bash
# 设置配置
go run main.go redis -action set -user teacher_001 -model gpt-4-turbo

# 获取配置
go run main.go redis -action get -user teacher_001

# 批量设置测试数据
go run main.go redis -action batch

# 查看预置配置
go run main.go redis -action list
```



### 3. 令牌状态检查工具 (`token`)
**文件位置：** `cmd/token_checker.go`

**功能特性：**
- ✅ 验证令牌有效性
- ✅ 显示详细令牌信息
- ✅ 支持多种输出格式
- ✅ 配额状态检查

**使用示例：**
```bash
# 详细格式检查
go run main.go token -token sk-xxxxxx

# JSON格式输出
go run main.go token -token sk-xxxxxx -format json

# 简要格式输出
go run main.go token -token sk-xxxxxx -format brief
```

## ⚙️ 配置文件系统

### 配置文件模板 (`config.env.example`)
包含所有可配置参数的示例：

```bash
# Redis 配置
REDIS_CONN_STRING=redis://localhost:6379

# 数据库配置
SQL_DSN=root:password@tcp(localhost:3306)/oneapi

# API 配置
API_BASE_URL=http://localhost:3000
API_TOKEN=sk-your-token-here

# 测试用户配置
TEST_USER_ID=teacher_001
TEST_MODEL=gpt-4-turbo

# 智能模型选择配置
DEFAULT_TEMPERATURE=0.7
DEFAULT_MAX_TOKENS=2000
DEFAULT_TOP_P=0.9
```

### 使用Makefile快捷操作

**可用目标：**
```bash
make help           # 显示所有可用目标
make setup          # 创建 .env 配置文件
make redis          # 运行Redis助手帮助
make smart-model    # 运行智能模型测试帮助
make token          # 运行令牌检查帮助
make config         # 显示当前配置
make all            # 运行所有脚本帮助
make clean          # 清理临时文件
make check-deps     # 检查依赖环境
make env-info       # 显示环境信息
```

## 🎯 使用场景

### 开发阶段
- **Redis调试**：验证用户配置存储功能
- **模型测试**：测试智能选择逻辑
- **令牌验证**：检查API权限和配额

### 测试阶段
- **功能测试**：验证各模块功能完整性
- **集成测试**：测试系统整体流程
- **性能测试**：评估系统响应时间

### 生产维护
- **配置管理**：批量更新用户配置
- **状态监控**：定期检查系统健康状态
- **故障诊断**：快速定位问题根源

## 🛠️ 环境要求

### 必需组件
- **Go 1.19+**：脚本运行环境
- **One-API项目**：完整的项目依赖

### 可选组件
- **Redis**：智能模型选择功能
- **MySQL/SQLite**：数据存储
- **网络连接**：API测试需要

## 🐛 故障排除

### 常见问题及解决方案

#### 1. 脚本运行冲突
```bash
# 问题：main函数重复定义
# 解决：使用统一入口脚本

# ❌ 错误方式
go run redis_helper.go

# ✅ 正确方式  
go run main.go redis -help
```

#### 2. Redis连接失败
```bash
# 检查配置
make config

# 手动设置
export REDIS_CONN_STRING="redis://localhost:6379"
```

#### 3. 数据库连接失败
```bash
# 检查数据库配置
export SQL_DSN="root:password@tcp(localhost:3306)/oneapi"
```

#### 4. 令牌验证失败
```bash
# 检查令牌有效性
go run main.go token -token your-actual-token
```

### 调试技巧

#### 查看详细日志
```bash
# 设置调试模式
export LOG_LEVEL=debug

# 运行带详细输出的脚本
go run main.go redis -action list -v
```

#### 验证环境
```bash
# 检查所有环境组件
make check-deps

# 显示完整环境信息
make env-info
```

## 🔍 文件命名规范

优化后的文件命名遵循以下规则：

- **统一入口**：`main.go`（主控制台）
- **功能模块**：`cmd/xxx_helper.go`（具体工具）
- **配置文件**：`config.env.example`（模板）
- **构建工具**：`Makefile`（快捷命令）
- **文档**：`README.md`（使用说明）

## 🤝 贡献指南

### 添加新脚本
1. 创建脚本文件到 `cmd/` 目录
2. 更新 `main.go` 中的脚本列表
3. 更新 `Makefile` 中的快捷目标
4. 更新本README文档

### 改进建议
- 提交Issue描述具体问题
- 提供改进方案或PR
- 确保遵循当前命名规范
- 测试所有相关功能

## 📊 优化总结

本次优化完成了以下改进：

✅ **代码质量**：修复所有linter错误
✅ **命名规范**：统一文件命名规则
✅ **使用体验**：提供统一入口和快捷命令
✅ **配置管理**：支持.env配置文件
✅ **文档完善**：提供详细使用说明
✅ **错误处理**：改进错误提示和诊断
✅ **测试验证**：确保所有脚本可运行

现在所有脚本都可以通过统一入口使用，支持配置文件预设，提供更好的用户体验。