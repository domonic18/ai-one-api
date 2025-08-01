# 模拟课件平台API服务器

## 概述

这是一个模拟课件平台API的服务器，用于在课件平台API未提供时进行OneAPI v3.0功能的开发和测试。

## 功能特性

- ✅ 模拟课件平台API接口
- ✅ Web管理界面
- ✅ 用户信息管理
- ✅ API响应配置
- ✅ 异常场景测试
- ✅ Docker部署支持

## 快速开始

### 1. 启动服务

```bash
# 方式1: 直接运行
go run main.go

# 方式2: Docker运行
docker-compose up -d

# 方式3: 通过scripts统一入口
cd ../
make mock-server
```

### 2. 访问管理界面

```
Web管理界面: http://localhost:8080
API文档: http://localhost:8080/api/docs
健康检查: http://localhost:8080/health
```

### 3. 测试API接口

```bash
# 获取用户信息
curl -X GET "http://localhost:8080/api/v1/teacher/teacher_001/info" \
  -H "Authorization: Bearer mock_api_key_123"

# 获取所有用户ID
curl -X GET "http://localhost:8080/api/v1/teachers/ids" \
  -H "Authorization: Bearer mock_api_key_123"

# 批量获取用户信息
curl -X POST "http://localhost:8080/api/v1/teachers/batch" \
  -H "Authorization: Bearer mock_api_key_123" \
  -H "Content-Type: application/json" \
  -d '{"teacher_ids": ["teacher_001", "teacher_002"]}'
```

## 配置OneAPI连接

在OneAPI的docker-compose.yml中配置：

```yaml
environment:
  - COURSEWARE_ENABLED=true
  - COURSEWARE_BASE_URL=http://localhost:8080/api/v1
  - COURSEWARE_API_KEY=mock_api_key_123
```

## 详细文档

请参考 `docs/tests/mock-server方案.md` 获取完整的实现方案和使用说明。

## 开发指南

### 添加新的API接口

1. 在 `handlers/api.go` 中添加新的处理器函数
2. 在 `main.go` 中注册新的路由
3. 在Web界面中添加对应的测试功能
4. 更新API文档

### 修改数据模型

1. 更新 `models/models.go` 中的结构体定义
2. 修改存储层的相关方法
3. 更新Web界面的表单和显示
4. 更新默认数据文件

## 故障排除

### 常见问题

1. **端口冲突**: 修改配置文件中的端口号
2. **权限问题**: 确保数据目录有读写权限
3. **API密钥错误**: 检查OneAPI配置中的API密钥
4. **网络连接**: 确保OneAPI能够访问模拟服务器

### 日志查看

```bash
# 查看服务日志
docker-compose logs -f mock-courseware-platform

# 查看API访问日志
tail -f logs/api.log
``` 