#!/bin/bash

# OneAPI 服务启动脚本

set -e

echo "🚀 启动OneAPI服务..."

# 获取脚本所在目录的上级目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 加载环境变量
echo "⚙️  加载环境变量..."
if [ -f "env.dev" ]; then
    set -a  # 自动导出所有变量
    source env.dev
    set +a  # 关闭自动导出
    echo "✅ 环境变量已加载"
    
    # 显示关键环境变量
    echo "📋 关键环境变量:"
    echo "   SQL_DSN: ${SQL_DSN:-未设置}"
    echo "   REDIS_CONN_STRING: ${REDIS_CONN_STRING:-未设置}"
    echo "   COURSEWARE_ENABLED: ${COURSEWARE_ENABLED:-未设置}"
else
    echo "⚠️  警告: env.dev 文件不存在，使用默认环境变量"
fi

# 启动OneAPI服务
echo "🚀 启动OneAPI服务..."
go run main.go 