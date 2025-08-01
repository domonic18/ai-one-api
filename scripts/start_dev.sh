#!/bin/bash

# OneAPI 开发环境快速启动脚本

set -e

echo "🚀 启动OneAPI开发环境..."

# 1. 启动数据库和Redis
echo "📦 启动数据库和Redis..."
docker-compose -f docker-compose.dev.yml up -d

# 2. 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 3. 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

# 4. 加载环境变量
echo "⚙️  加载环境变量..."
export $(cat env.dev | grep -v '^#' | xargs)

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
echo "   go run main.go"
echo ""
echo "🌐 访问地址:"
echo "   - OneAPI: http://localhost:3000"
echo ""
echo "📝 调试命令:"
echo "   - 停止服务: docker-compose -f docker-compose.dev.yml down"
echo "   - 查看日志: docker-compose -f docker-compose.dev.yml logs -f"
echo "   - 重启服务: docker-compose -f docker-compose.dev.yml restart" 