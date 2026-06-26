# d-im 架构文档

## 概述

d-im 是一个基于 Go 语言开发的即时通讯系统，采用微服务架构设计，支持 API 服务和 WebSocket 服务分离部署。

## 架构设计

### 1. 服务分离架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Server    │    │  WS Server      │    │   Nginx Proxy   │
│   (Port: 8080)  │    │  (Port: 9000)   │    │   (Port: 80)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Shared DB     │
                    │  (MongoDB +     │
                    │   Redis)        │
                    └─────────────────┘
```

### 2. 服务职责

#### API Server (api-server)
- **职责**: 处理 HTTP API 请求
- **特点**: 无状态服务，可水平扩展
- **路由**: `/api/*`
- **功能**:
  - 消息管理 (CRUD)
  - 会话管理
  - 用户状态查询
  - 认证授权

#### WebSocket Server (ws-server)
- **职责**: 处理 WebSocket 连接和实时消息
- **特点**: 有状态服务，管理客户端连接
- **路由**: `/im/ws`
- **功能**:
  - WebSocket 连接管理
  - 实时消息推送
  - 心跳检测
  - 在线状态管理

#### Unified Server (unified-server)
- **职责**: 同时提供 API 和 WebSocket 服务
- **特点**: 适合小规模部署或开发环境
- **端口**: 8080
- **功能**: 整合了 API 和 WebSocket 的所有功能

## 目录结构

```
d-im/
├── cmd/                    # 应用程序入口
│   ├── api-server/        # API 服务器
│   │   └── main.go        # API 服务器入口
│   ├── ws-server/         # WebSocket 服务器
│   │   └── main.go        # WebSocket 服务器入口
│   └── server/            # 统一服务器
│       └── main.go        # 统一服务器入口
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   │   └── config.go      # 配置结构和方法
│   ├── handler/           # HTTP 处理器
│   │   ├── message_handler.go
│   │   ├── conversation_handler.go
│   │   ├── session_handler.go
│   │   └── ws_handler.go
│   ├── middleware/        # 中间件
│   │   └── jwt.go         # JWT 认证中间件
│   ├── models/            # 数据模型
│   │   ├── message.go
│   │   ├── conversation.go
│   │   ├── session.go
│   │   └── user.go
│   ├── repository/        # 数据访问层
│   │   ├── message_repository.go
│   │   ├── conversation_repository.go
│   │   └── session_repository.go
│   ├── router/            # 路由配置
│   │   └── router.go      # 路由设置
│   └── service/           # 业务逻辑层
│       ├── message_service.go
│       ├── conversation_service.go
│       ├── session_service.go
│       ├── user_service.go
│       └── ws_manager.go  # WebSocket 管理器
├── pkg/                   # 公共包
│   ├── jwt/               # JWT 工具包
│   │   ├── generator.go
│   │   ├── jwt.go
│   │   ├── keyloader.go
│   │   ├── manager.go
│   │   └── validator.go
│   └── utils/             # 通用工具
│       ├── file.go
│       └── image.go
├── scripts/               # 启动脚本
│   ├── start-api.sh       # API 服务器启动脚本
│   ├── start-ws.sh        # WebSocket 服务器启动脚本
│   └── start-unified.sh   # 统一服务器启动脚本
├── docs/                  # 文档
│   ├── api.md             # API 文档
│   └── ARCHITECTURE.md    # 架构文档
├── bin/                   # 构建输出目录
├── config/                # 配置文件
│   ├── config.yaml        # 主配置文件
│   ├── config.yaml.example # 配置文件示例
│   └── keys/              # 密钥文件
├── Makefile               # 构建和部署工具
├── build.sh               # 构建脚本
├── Dockerfile             # Docker 构建文件
├── docker-compose.yml     # 根目录 Docker Compose
├── go.mod                 # Go 模块文件
├── go.sum                 # Go 依赖校验文件
└── README.md              # 项目说明
```

## 数据流

### 1. 消息发送流程

```
Client → API Server → Message Service → Database
                                    ↓
                              WS Manager → WS Server → Target Client
```

### 2. 实时消息流程

```
Client A → WS Server → WS Manager → Message Service → Database
                                                    ↓
Client B ← WS Server ← WS Manager ← Message Service
```

### 3. 在线状态管理

```
Client Connect → WS Server → Session Service → Database
Client Disconnect → WS Server → Session Service → Database
```

## 配置管理

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| API_SERVER_PORT | 8080 | API 服务器端口 |
| WS_SERVER_PORT | 9000 | WebSocket 服务器端口 |
| MONGODB_URI | mongodb://localhost:27017 | MongoDB 连接字符串 |
| MONGODB_DATABASE | go_im | MongoDB 数据库名 |
| REDIS_ADDR | localhost:6379 | Redis 地址 |
| REDIS_PASSWORD | "" | Redis 密码 |
| REDIS_DB | 0 | Redis 数据库编号 |
| JWT_PUBLIC_KEY_PATH | ./config/keys/public.pem | JWT 公钥路径 |

### 配置文件

支持 YAML 格式的配置文件，可以通过环境变量 `CONFIG_FILE` 指定配置文件路径。

## 部署模式

### 1. 分离部署模式（推荐）

```bash
# 启动 API 服务器
make run-api

# 启动 WebSocket 服务器
make run-ws
```

### 2. 统一部署模式

```bash
# 启动统一服务器
make run-unified
```

### 3. 生产部署

使用仓库根目录 `deploy/` 下的 Portainer Stack 模板，详见 [`deploy/README.md`](../../deploy/README.md)。

## 扩展性设计

### 1. 水平扩展

- **API Server**: 无状态设计，支持多实例部署
- **WS Server**: 通过 Redis 共享连接状态，支持多实例部署
- **Database**: MongoDB 支持分片和副本集

### 2. 负载均衡

- **API 请求**: 通过 Nginx 负载均衡到多个 API 实例
- **WebSocket 连接**: 通过 Redis 进行连接分发

### 3. 高可用性

- **服务发现**: 支持服务注册和发现
- **健康检查**: 内置健康检查接口
- **故障转移**: 支持自动故障转移

## 监控和日志

### 1. 健康检查

- API Server: `GET /health`
- WS Server: `GET /health`

### 2. 日志格式

支持结构化日志输出，便于日志分析和监控。

### 3. 指标收集

支持 Prometheus 指标收集，便于监控和告警。

## 安全考虑

### 1. 认证授权

- JWT Token 认证
- 基于角色的访问控制
- API 接口权限验证

### 2. 数据安全

- 敏感数据加密存储
- 传输层 TLS 加密
- 输入验证和过滤

### 3. 网络安全

- CORS 配置
- 请求频率限制
- IP 白名单控制

## 性能优化

### 1. 连接池

- MongoDB 连接池
- Redis 连接池
- HTTP 客户端连接池

### 2. 缓存策略

- Redis 缓存热点数据
- 内存缓存用户会话
- 数据库查询缓存

### 3. 异步处理

- 消息异步存储
- 事件驱动架构
- 批量操作优化 