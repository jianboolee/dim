# Redis 使用说明

## 概述

d-im 中 Redis 只在**一个场景**下使用：**API Server 与 WS Server 分离部署时的跨进程消息推送**。

项目的所有持久化数据均存储在 MongoDB，Redis 仅作为发布/订阅消息通道，不承载任何业务数据。

---

## 为什么需要 Redis

d-im 支持两种部署模式：

| 模式 | 入口 | 进程数 |
|------|------|--------|
| 统一部署 | `cmd/server` | 1 个进程，API + WebSocket 合体 |
| 分离部署 | `cmd/api-server` + `cmd/ws-server` | 2 个进程 |

分离部署时，用户发消息的 HTTP 请求落在 **API Server**，而消息接收方的 WebSocket 连接在 **WS Server**。两者没有共享内存，需要一个中间通道把消息从 API 进程推送到 WS 进程——这个通道就是 Redis Pub/Sub。

---

## 数据流

```
用户通过 HTTP 发消息
        │
        ▼
  API Server: SendMessageToConversationHTTP()
        │
        ├── 1. 保存消息到 MongoDB
        ├── 2. 更新未读计数
        └── 3. FanoutMessage()
              │
              └── pushToUser(userID, msgBytes)
                    │
                    ├── wsManager.TryDeliver()         ← 同进程直连
                    │     (cmd/server 单进程时命中)
                    │
                    └── redisClient.Publish(            ← 跨进程
                          "im:ws:push",
                          WSPushEvent{userID, msgBytes}
                        )
                          │
                          ▼  Redis Pub/Sub
                          │
              WS Server: startRedisSubscriber()
                          │
                          └── SUBSCRIBE "im:ws:push"
                                │
                                └── TryDeliver(userID, msg)
                                      │
                                      └── write 到目标 WebSocket 客户端
```

---

## 涉及的代码

| 文件 | 作用 |
|------|------|
| `internal/config/config.go` | 读取 `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` |
| `cmd/api-server/main.go` | 创建 `redis.Client`，注入 `Dependencies` |
| `cmd/ws-server/main.go` | 创建 `redis.Client`，注入 `Dependencies` |
| `cmd/server/main.go` | 创建 `redis.Client`（单进程模式下通常不实际用到） |
| `internal/app/deps.go` | 将 `redisClient` 传递给 `MessageService` 和 `WSManager` |
| `internal/service/ws_push.go` | 定义 `WSPushEvent` 结构体和 channel 常量 `im:ws:push` |
| `internal/service/message_service.go:pushToUser()` | 推送入口：先尝试同进程直连，失败则 `Publish` 到 Redis |
| `internal/service/ws_manager.go:startRedisSubscriber()` | WS 进程启动时 `Subscribe` Redis channel |

---

## 访问和扩展

### Redis 不是必须的

- **单进程模式**（`cmd/server`）：API 和 WebSocket 在同一进程内，`TryDeliver()` 直接通过内存 channel 投递，不经过 Redis。
- **Redis 连接失败**：`deps.go` 会在启动时打印 warning，进程正常启动。分离部署时没有 Redis 会导致实时推送无法跨进程投递。

### 如何判断是否在用 Redis

```go
// 推送逻辑在 pushToUser 中
func (s *MessageService) pushToUser(ctx context.Context, userID string, messageBytes []byte) error {
    // 优先：同进程直连
    if s.wsManager != nil && s.wsManager.TryDeliver(userID, messageBytes) {
        return nil  // ← 单进程模式在这里结束，不经过 Redis
    }

    // 降级：跨进程 Redis Pub/Sub
    if s.redisClient == nil {
        return nil  // ← Redis 未配置，静默丢弃
    }
    return s.redisClient.Publish(ctx, "im:ws:push", payload).Err()
}
```

### 为什么不存其他数据

- **在线状态**：存在 MongoDB 的 `session` 集合，通过 WebSocket 连接/断连 + keepalive 维护
- **JWT / 认证会话**：存在 MongoDB 的 `auth_session` 集合
- **消息 / 会话 / 群组**：全部在 MongoDB

Redis 只做 Pub/Sub 桥接。如果未来需要消息缓存、在线状态热数据、分布式锁，可以引入，但当前无此需求。

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `REDIS_PASSWORD` | (空) | Redis 密码 |
| `REDIS_DB` | `0` | 数据库编号 |

---

## 结论

| 问题 | 答案 |
|------|------|
| Redis 用于什么 | 分离部署时 API → WS 进程间消息桥接 |
| 存储数据 | 不存储任何业务数据 |
| 单进程是否需要 | 不需要 |
| 去掉后影响 | 单进程无影响；分离部署时实时推送不可用，消息仍能保存但接收方需要刷新才能看到 |
