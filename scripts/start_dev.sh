#!/bin/bash

# OneAPI 开发环境快速启动脚本

set -e

echo "🚀 启动OneAPI开发环境..."

# 获取脚本所在目录的上级目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 1. 启动数据库和Redis
echo "📦 启动数据库和Redis..."
docker-compose -f "$PROJECT_ROOT/docker-compose.dev.yml" up -d

# 2. 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 3. 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f "$PROJECT_ROOT/docker-compose.dev.yml" ps

# 4. 加载环境变量
echo "⚙️  加载环境变量..."
if [ -f "$PROJECT_ROOT/env.dev" ]; then
    # 使用source命令加载环境变量，确保在当前shell中生效
    set -a  # 自动导出所有变量
    source "$PROJECT_ROOT/env.dev"
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

# 5. 显示配置信息
echo ""
echo "✅ 开发环境已启动！"
echo ""
echo "📊 服务信息:"
echo "   - MySQL: localhost:3306"
echo "   - Redis: localhost:6379"
echo "   - Mock服务器: http://192.168.6.118:8080"
echo ""
echo "🔧 运行OneAPI:"
echo "   cd $PROJECT_ROOT && source env.dev && go run main.go"
echo ""
echo "🌐 访问地址:"
echo "   - OneAPI: http://localhost:3000"
echo ""
echo "📝 调试命令:"
echo "   - 停止服务: docker-compose -f $PROJECT_ROOT/docker-compose.dev.yml down"
echo "   - 查看日志: docker-compose -f $PROJECT_ROOT/docker-compose.dev.yml logs -f"
echo "   - 重启服务: docker-compose -f $PROJECT_ROOT/docker-compose.dev.yml restart"
echo ""
echo "💡 提示: 要在当前shell中使用环境变量，请运行:"
echo "   source $PROJECT_ROOT/env.dev"
echo ""
echo "🚀 自动启动OneAPI服务..."
cd "$PROJECT_ROOT"
source env.dev
go run main.go 