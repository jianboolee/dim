# d-im

这是一个基于 Go 语言开发的即时通讯系统，采用微服务架构设计，支持 API 服务和 WebSocket 服务分离部署。

## 项目结构

```
.
├── cmd/                    # 主要的应用程序入口
│   ├── api-server/        # API 服务器入口
│   ├── ws-server/         # WebSocket 服务器入口
│   └── server/            # 统一服务器入口（API + WebSocket）
├── internal/              # 私有应用程序和库代码
│   ├── config/            # 配置相关
│   ├── handler/           # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   ├── repository/        # 数据访问层
│   ├── router/            # 路由配置
│   └── service/           # 业务逻辑服务
├── pkg/                   # 可以被外部应用程序使用的库代码
│   ├── jwt/               # JWT 工具包
│   └── utils/             # 通用工具包
├── scripts/               # 启动脚本
│   ├── start-api.sh       # API 服务器启动脚本
│   ├── start-ws.sh        # WebSocket 服务器启动脚本
│   └── start-unified.sh   # 统一服务器启动脚本
└── build/                 # 构建输出目录
```

## 服务架构

### 1. API 服务器 (api-server)
- **端口**: 8080
- **功能**: 处理 HTTP API 请求
- **路由**: `/api/*`
- **特点**: 无状态服务，可水平扩展

### 2. WebSocket 服务器 (ws-server)
- **端口**: 9000
- **功能**: 处理 WebSocket 连接和实时消息
- **路由**: `/im/ws`
- **特点**: 有状态服务，管理客户端连接

### 3. 统一服务器 (unified-server)
- **端口**: 8080
- **功能**: 同时提供 API 和 WebSocket 服务
- **特点**: 适合小规模部署或开发环境

## 开始使用

### 环境要求
- Go 1.21 或更高版本
- MongoDB 7.0+
- Redis 7.0+
- Docker & Docker Compose (可选)

### 快速开始

1. **克隆项目**
```bash
git clone <repository-url>
cd d-im
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置环境变量**
```bash
# 复制配置文件
cp config.yaml.example config.yaml

# 设置环境变量
export API_SERVER_PORT=8080
export WS_SERVER_PORT=9000
export MONGODB_URI="mongodb://localhost:27017/go_im"
export REDIS_ADDR="localhost:6379"
export JWT_PUBLIC_KEY_PATH="./config/keys/public.pem"
```

4. **启动服务**

**方式一：分离部署（推荐）**
```bash
# 启动 API 服务器
go run cmd/api-server/main.go

# 启动 WebSocket 服务器（新终端）
go run cmd/ws-server/main.go
```

**方式二：统一部署**
```bash
# 启动统一服务器
go run cmd/server/main.go
```

**方式三：生产部署**

生产环境使用仓库根目录 `deploy/` 下的 Portainer Stack 模板，详见 [`deploy/README.md`](../deploy/README.md)。

### 构建

```bash
# 构建所有服务
./scripts/build.sh

# 构建产物在 bin/ 目录
ls bin/
# api-server      # API 服务器二进制
# ws-server       # WebSocket 服务器二进制
# unified-server  # 统一服务器二进制
```

## 许可证

MIT License 



```js
const ws = new WebSocket('ws://localhost:8080/api/ws?token=your_token_here');
ws.onmessage = (event) => {
    console.log('收到消息:', event.data);
};

// 发送消息
ws.send(JSON.stringify({
    conversation_id: "会话ID",
    type: "text",
    content: "你好！"
}));
```

## 当前 IM API 约定

- `POST /im/api/integration/login`：业务系统登录 IM，进入会话列表。
- `POST /im/api/integration/conversations`：业务系统创建或进入单聊，请求体使用 `user` 和 `peer_user`。
- `POST /im/api/conversations`：登录用户创建或获取单聊，请求体使用 `peer_id`。
- `POST /im/api/groups`：创建群聊，可传 `member_ids`。
- `GET /im/api/conversations`：获取当前用户会话列表。
- `GET /im/api/conversations/:id/messages`：按 `conversation_id` 获取消息。
- `POST /im/api/conversations/:id/messages`：按 `conversation_id` 发送消息。
- `PUT /im/api/conversations/:id/read`：会话级已读，清空当前用户在该会话的未读数。
- `GET /im/api/messages/unread/count`：获取当前用户总未读数。
- `/im/ws?token=your_token_here`：WebSocket 连接。


## 会话（Conversation）

我来解释一下会话（Conversation）中每个字段的含义和用途：

1. `id`: 会话的唯一标识符
   - MongoDB 生成的 ObjectID
   - 用于在需要获取特定会话详情时作为标识

2. `type`: 会话类型
   - 当前值为 "private"，表示这是一个私聊会话
   - 未来可以扩展支持 "group"（群聊）等其他类型

3. `participants`: 会话参与者数组
   - 包含所有参与会话的用户 ID
   - 在私聊中只有两个参与者：`["67879eed5c03b9b6123b3c1d", "678888ff530e60a73934013e"]`
   - 用于确定谁可以访问这个会话

4. `last_message`: 最后一条消息的完整信息
   - `id`: 消息的唯一标识符
   - `from_id`: 发送者的用户 ID
   - `content`: 消息内容
   - `status`: 消息状态（sent=已发送，delivered=已送达，read=已读）
   - `created_at`: 消息创建时间
   - `updated_at`: 消息最后更新时间

5. `member_state.unread_count`: 当前用户在该会话的未读消息数量
   - 收到非当前会话的新消息时增加
   - 调用会话级已读接口后清零

6. `last_activity`: 最后活动时间
   - 基于最后一条消息的时间
   - 用于对会话进行排序（最近活动的会话排在前面）
   - 注意：这里显示 "0001-01-01T00:00:00Z" 可能是一个 bug，应该和 last_message.created_at 保持一致

7. `created_at`: 会话创建时间
   - 记录会话首次创建的时间戳
   - 这里是 "2025-01-16T04:20:30.612Z"

8. `updated_at`: 会话最后更新时间
   - 记录会话最后一次更新的时间戳
   - 当有新消息或其他更新时会更新这个时间
   - 这里是 "2025-01-16T04:21:24.698Z"

使用场景示例：
1. 获取会话列表时，可以根据 `last_activity` 或 `updated_at` 排序，显示最近的对话
2. 使用 `unread_count` 在 UI 上显示未读消息数量的角标
3. 通过 `participants` 确定当前用户是否有权限访问该会话
4. 使用 `last_message` 在会话列表中预览最后一条消息的内容
5. 根据 `type` 区分不同类型的会话，以便使用不同的 UI 展示

我注意到 `last_activity` 字段显示的是零值，这是一个 bug，我来修复一下：

## 在线状态

WebSocket 管理器:
   - 连接时自动更新用户为在线状态
   - 断开连接时自动更新用户为离线状态

路由配置:
   - `/api/im/sessions/:user_id`: 获取用户在线状态
   - `/api/im/sessions/batch`: 批量获取用户在线状态
   - `/api/im/sessions/keepalive`: 保持在线状态


使用方法：

1. 获取用户在线状态：
```bash
curl -H "Authorization: Bearer your-token" \
  http://localhost:8080/api/im/sessions/user123
```

2. 批量获取用户在线状态：
```bash
curl -X POST -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"user_ids": ["user1", "user2", "user3"]}' \
  http://localhost:8080/api/im/sessions/batch
```

3. 保持在线状态：
```bash
curl -X POST -H "Authorization: Bearer your-token" \
  http://localhost:8080/api/im/sessions/keepalive
```

系统会自动处理：
1. WebSocket 连接时将用户标记为在线
2. WebSocket 断开时将用户标记为离线
3. 记录用户最后在线时间
4. 通过 keepalive 接口保持在线状态

建议客户端：
1. 建立 WebSocket 连接后定期调用 keepalive 接口（如每 30 秒）
2. 获取会话列表时同时获取对应用户的在线状态
3. 在 UI 上显示用户的在线状态和最后在线时间
