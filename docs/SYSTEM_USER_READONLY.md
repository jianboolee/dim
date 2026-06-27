# 系统用户只读保护

## 概述

系统用户（如 `system_notice`、`system_order` 等）用于向普通用户推送系统级消息。本次改造实现了对系统用户的**只读保护**：普通用户不能回复系统用户的消息。

## 用户类型

在 `User` 模型中新增 `Type` 字段，定义三种用户类型：

| Type | 含义 | 能否回复 | WebSocket |
|------|------|---------|-----------|
| `normal` | 普通用户 | ✅ | ✅ |
| `system` | 系统用户（只读） | ❌ | ❌ |
| `bot` | 机器人/客服（可回复） | ✅ | ❌ |

## 改动要点

### 后端

1. **User 模型** (`models/user.go`)
   - 新增 `UserType` 类型和 `normal`/`system`/`bot` 三种常量
   - `User` 结构体新增 `Type` 字段

2. **Repository** (`repository/user_repository.go`)
   - `Upsert` 方法：`Type` 非空时才 `$set`，空则保留原值，避免覆盖

3. **DTO** (`dto/user_dto.go`, `dto/integration_dto.go`)
   - `UserInfoDto` 返回 `Type` 字段，含 `system_` 前缀兜底逻辑
   - `IntegrationUserInput` 支持显式传入 `Type`

4. **Integration** (`service/integration_service.go`)
   - `resolveUserType()` 优先使用显式传入的 `Type`，否则根据 `system_` 前缀自动推断

5. **消息发送拦截** (`service/message_service.go`)
   - `SendMessageToConversation` 中检查接收方类型，若为 `system` 返回 `ErrCannotReplyToSystemUser`
   - HTTP 路径（`/im/api/conversations/:id/messages`）和 WebSocket 路径均覆盖

6. **WebSocket 拦截** (`handler/ws_handler.go`)
   - 系统用户（`system` / `bot`）不允许建立 WebSocket 连接

### SDK

7. **Go SDK** (`sdk/go/types.go`)
   - `User` 结构体新增 `Type` 字段

8. **Demo 系统用户** (`sdk/go/examples/demo/system_user.go`)
   - 所有系统用户显式设置 `Type`
   - `system_service` 设为 `"bot"`（用户可回复）

### 前端

9. **类型定义** (`types/user.ts`, `sdk/im.ts`)
   - `UserInfo` 和 `to_user_info` 接口新增 `type` 字段

10. **聊天页** (`views/im/chat.vue`)
    - 新增 `peerUserType` 计算属性
    - 当对方为 `system` 类型时，隐藏消息输入框，展示「系统消息，暂不支持回复」提示条

## 涉及文件

```
im-backend/internal/models/user.go
im-backend/internal/repository/user_repository.go
im-backend/internal/dto/user_dto.go
im-backend/internal/dto/integration_dto.go
im-backend/internal/service/integration_service.go
im-backend/internal/service/message_service.go
im-backend/internal/handler/message_handler.go
im-backend/internal/handler/ws_handler.go
im-backend/internal/app/deps.go
sdk/go/types.go
sdk/go/examples/demo/system_user.go
im-frontend/src/types/user.ts
im-frontend/src/sdk/im.ts
im-frontend/src/views/im/chat.vue
```
