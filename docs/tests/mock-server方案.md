# 模拟课件平台API服务器方案

## 文档概述

本文档提供了模拟课件平台API服务器的完整实现方案，用于在课件平台API未提供时进行OneAPI v3.0功能的开发和测试。该模拟服务器提供Web界面，方便开发人员配置和测试各种API响应场景。

## 1. 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                模拟课件平台API服务器                    │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Web管理界面                        │    │
│  │  - 用户信息配置                                │    │
│  │  - API响应配置                                 │    │
│  │  - 异常场景测试                                │    │
│  └─────────────────────────────────────────────────┘    │
│                              │                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │              API服务层                          │    │
│  │  - /api/v1/teacher/{id}/info                   │    │
│  │  - /api/v1/teachers/ids                        │    │
│  │  - /api/v1/teachers/batch                      │    │
│  └─────────────────────────────────────────────────┘    │
│                              │                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │              数据存储层                         │    │
│  │  - 内存存储 (开发模式)                          │    │
│  │  - JSON文件存储 (持久化)                        │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

## 2. 技术栈选择

### 2.1 技术栈
- **后端框架**: Gin (与OneAPI保持一致)
- **前端**: 原生HTML + JavaScript (简单易用)
- **数据存储**: JSON文件 + 内存缓存
- **配置管理**: 环境变量 + 配置文件

### 2.2 目录结构

#### 2.2.1 项目整体scripts目录结构
```
scripts/
├── README.md                 # 脚本工具总说明文档
├── Makefile                  # 统一构建工具
├── config.env.example        # 配置文件模板
├── .env                      # 配置文件（用户创建）
├── main.go                   # 统一入口脚本（主控制台）
├── runner.go                 # 脚本运行器
├── utils/                    # 通用工具库
│   └── config_loader.go      # 配置加载工具
├── migration/                # 数据库迁移脚本
│   └── main.go
├── token/                    # 令牌相关工具
│   └── main.go
├── cmd/                      # 独立功能脚本
│   ├── redis_helper.go       # Redis助手工具
│   └── token_checker.go      # 令牌状态检查
└── mock-server/              # 模拟课件平台API服务器
    ├── main.go               # 主程序入口
    ├── go.mod                # Go模块文件
    ├── go.sum                # 依赖锁定文件
    ├── config/
    │   ├── config.go         # 配置管理
    │   └── config.json       # 配置文件
    ├── handlers/
    │   ├── api.go            # API处理器
    │   └── web.go            # Web界面处理器
    ├── models/
    │   └── models.go         # 数据模型
    ├── storage/
    │   ├── memory.go         # 内存存储
    │   └── file.go           # 文件存储
    ├── static/
    │   ├── index.html        # 主页面
    │   ├── style.css         # 样式文件
    │   └── script.js         # JavaScript文件
    ├── data/
    │   ├── users.json        # 用户数据
    │   └── config.json       # 配置数据
    ├── Dockerfile            # Docker构建文件
    ├── docker-compose.yml    # Docker编排文件
    └── README.md             # 使用说明
```

#### 2.2.2 目录组织说明

**scripts目录重新规划原则：**

1. **工具脚本区** (`cmd/`): 存放各种独立的辅助脚本工具
   - `redis_helper.go`: Redis操作工具
   - `token_checker.go`: 令牌检查工具
   - 其他功能脚本...

2. **服务类工具区** (`mock-server/`): 存放需要以服务形式运行的测试工具
   - 模拟课件平台API服务器
   - 其他模拟服务...

3. **通用工具区** (`utils/`, `migration/`, `token/`): 存放通用工具和基础功能
   - `utils/`: 通用工具库
   - `migration/`: 数据库迁移脚本
   - `token/`: 令牌相关工具

4. **统一入口** (`main.go`, `Makefile`): 提供统一的脚本管理和执行入口

**优势：**
- **清晰分离**: 脚本工具和服务工具明确区分
- **易于维护**: 相关功能集中管理
- **统一入口**: 通过main.go和Makefile统一管理
- **项目简洁**: 保持项目根目录的简洁性

## 3. API接口设计

### 3.1 获取用户信息接口

```go
// GET /api/v1/teacher/{teacher_id}/info
type TeacherInfoResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    *TeacherInfo `json:"data"`
}

type TeacherInfo struct {
    TeacherId      string `json:"teacher_id"`
    TeacherName    string `json:"teacher_name"`
    SchoolId       int    `json:"school_id"`
    SchoolName     string `json:"school_name"`
    SubjectId      int    `json:"subject_id"`
    SubjectName    string `json:"subject_name"`
    OneapiGroup    string `json:"oneapi_group"`
    PreferredModel string `json:"preferred_model"`
}
```

### 3.2 获取所有老师ID列表接口

```go
// GET /api/v1/teachers/ids
type TeacherIdsResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        TeacherIds []string `json:"teacher_ids"`
        Total      int      `json:"total"`
    } `json:"data"`
}
```

### 3.3 批量获取用户信息接口

```go
// POST /api/v1/teachers/batch
type BatchTeacherRequest struct {
    TeacherIds []string `json:"teacher_ids"`
}

type BatchTeacherResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        Users        []*TeacherInfo `json:"users"`
        SuccessCount int            `json:"success_count"`
        ErrorCount   int            `json:"error_count"`
    } `json:"data"`
}
```

## 4. Web管理界面设计

### 4.1 主界面布局

```html
<!DOCTYPE html>
<html>
<head>
    <title>模拟课件平台API服务器</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>模拟课件平台API服务器</h1>
            <nav>
                <a href="#users">用户管理</a>
                <a href="#api-test">API测试</a>
                <a href="#config">配置管理</a>
            </nav>
        </header>
        
        <main>
            <!-- 用户管理面板 -->
            <section id="users" class="panel">
                <h2>用户信息管理</h2>
                <div class="user-form">
                    <!-- 用户添加/编辑表单 -->
                </div>
                <div class="user-list">
                    <!-- 用户列表 -->
                </div>
            </section>
            
            <!-- API测试面板 -->
            <section id="api-test" class="panel">
                <h2>API接口测试</h2>
                <div class="api-tester">
                    <!-- API测试工具 -->
                </div>
            </section>
            
            <!-- 配置管理面板 -->
            <section id="config" class="panel">
                <h2>系统配置</h2>
                <div class="config-form">
                    <!-- 配置表单 -->
                </div>
            </section>
        </main>
    </div>
    
    <script src="/static/script.js"></script>
</body>
</html>
```

### 4.2 功能模块

#### 4.2.1 用户管理模块
- **用户列表**: 显示所有配置的用户信息
- **用户添加**: 添加新用户，包含所有必要字段
- **用户编辑**: 修改现有用户信息
- **用户删除**: 删除用户
- **批量导入**: 从JSON文件批量导入用户

#### 4.2.2 API测试模块
- **接口测试**: 测试各个API接口
- **参数配置**: 配置请求参数
- **响应查看**: 查看API响应结果
- **异常测试**: 测试各种异常场景

#### 4.2.3 配置管理模块
- **系统配置**: 配置服务器参数
- **API密钥**: 管理API认证密钥
- **响应延迟**: 配置API响应延迟
- **错误率**: 配置API错误率

## 5. 核心功能实现

### 5.1 数据存储实现

```go
// storage/memory.go
type MemoryStorage struct {
    users    map[string]*TeacherInfo
    config   *ServerConfig
    mu       sync.RWMutex
}

func (m *MemoryStorage) GetUser(teacherId string) (*TeacherInfo, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if user, exists := m.users[teacherId]; exists {
        return user, nil
    }
    return nil, fmt.Errorf("user not found: %s", teacherId)
}

func (m *MemoryStorage) GetAllUserIds() ([]string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    ids := make([]string, 0, len(m.users))
    for id := range m.users {
        ids = append(ids, id)
    }
    return ids, nil
}

func (m *MemoryStorage) BatchGetUsers(teacherIds []string) ([]*TeacherInfo, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    users := make([]*TeacherInfo, 0, len(teacherIds))
    for _, id := range teacherIds {
        if user, exists := m.users[id]; exists {
            users = append(users, user)
        }
    }
    return users, nil
}
```

### 5.2 API处理器实现

```go
// handlers/api.go
func GetTeacherInfo(c *gin.Context) {
    teacherId := c.Param("teacher_id")
    
    // 模拟网络延迟
    if config.GetConfig().ResponseDelay > 0 {
        time.Sleep(config.GetConfig().ResponseDelay)
    }
    
    // 模拟错误率
    if rand.Float64() < config.GetConfig().ErrorRate {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "internal server error",
            "data":    nil,
        })
        return
    }
    
    user, err := storage.GetUser(teacherId)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "code":    404,
            "message": "teacher not found",
            "data":    nil,
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code":    200,
        "message": "success",
        "data":    user,
    })
}
```

### 5.3 Web界面处理器

```go
// handlers/web.go
func ServeWebInterface(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "模拟课件平台API服务器",
    })
}

func GetUsers(c *gin.Context) {
    users, err := storage.GetAllUsers()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, users)
}

func UpdateUser(c *gin.Context) {
    var user TeacherInfo
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := storage.UpdateUser(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}
```

## 6. 配置管理

### 6.1 配置文件结构

```json
{
    "server": {
        "port": 8080,
        "host": "0.0.0.0"
    },
    "api": {
        "response_delay": "0ms",
        "error_rate": 0.0,
        "api_key": "mock_api_key_123"
    },
    "default_users": [
        {
            "teacher_id": "teacher_001",
            "teacher_name": "张老师",
            "school_id": 1,
            "school_name": "北京中学",
            "subject_id": 10,
            "subject_name": "数学组",
            "oneapi_group": "beijing_math_group",
            "preferred_model": "gpt-4"
        },
        {
            "teacher_id": "teacher_002",
            "teacher_name": "李老师",
            "school_id": 1,
            "school_name": "北京中学",
            "subject_id": 2,
            "subject_name": "语文组",
            "oneapi_group": "beijing_chinese_group",
            "preferred_model": "gemini-pro"
        }
    ]
}
```

### 6.2 环境变量配置

```bash
# 服务器配置
MOCK_SERVER_PORT=8080
MOCK_SERVER_HOST=0.0.0.0

# API配置
MOCK_API_KEY=mock_api_key_123
MOCK_RESPONSE_DELAY=0ms
MOCK_ERROR_RATE=0.0

# 数据文件路径
MOCK_DATA_FILE=./data/users.json
MOCK_CONFIG_FILE=./data/config.json
```

## 7. Docker部署

### 7.1 Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mock-courseware-platform .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/mock-courseware-platform .
COPY --from=builder /app/static ./static
COPY --from=builder /app/data ./data

EXPOSE 8080
CMD ["./mock-courseware-platform"]
```

### 7.2 Docker Compose

```yaml
version: '3.8'

services:
  mock-courseware-platform:
    build: .
    container_name: mock-courseware-platform
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
    environment:
      - MOCK_SERVER_PORT=8080
      - MOCK_API_KEY=mock_api_key_123
      - MOCK_RESPONSE_DELAY=0ms
      - MOCK_ERROR_RATE=0.0
    restart: unless-stopped
```

## 8. 使用说明

### 8.1 启动服务

```bash
# 方式1: 直接运行
cd scripts/mock-server
go run main.go

# 方式2: Docker运行
cd scripts/mock-server
docker-compose up -d

# 方式3: 构建运行
cd scripts/mock-server
go build -o mock-courseware-platform .
./mock-courseware-platform

# 方式4: 通过scripts统一入口（推荐）
cd scripts
make mock-server
```

### 8.2 访问管理界面

```
Web管理界面: http://localhost:8080
API文档: http://localhost:8080/api/docs
健康检查: http://localhost:8080/health
```

### 8.3 配置OneAPI连接

在OneAPI的docker-compose.yml中配置：

```yaml
environment:
  - COURSEWARE_ENABLED=true
  - COURSEWARE_BASE_URL=http://localhost:8080/api/v1
  - COURSEWARE_API_KEY=mock_api_key_123
```