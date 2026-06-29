# IM API 快照

当前协议以 `conversation_id` 为消息路由核心。消息本身不携带独立目标用户字段；单聊、群聊的权限、未读数和列表可见性统一由 `conversation_members` 管理。

## 认证

除 integration 与 auth 入口外，HTTP API 使用：

```http
Authorization: Bearer <access-token>
```

WebSocket 使用：

```text
/im/ws?token=<access-token>
```

## Integration

业务服务端入口使用 `X-Integration-Key`。

```http
POST /im/api/integration/users
POST /im/api/integration/login
```

写入用户资料：

```json
{
  "users": [
    {"id": "user_a", "nickname": "买家", "type": "normal"},
    {"id": "system_order", "nickname": "订单消息", "type": "system"}
  ]
}
```

登录请求体：

```json
{
  "user_id": "user_a"
}
```

## Conversations

```http
GET  /im/api/conversations
POST /im/api/conversations
GET  /im/api/conversations/:id
POST /im/api/conversations/:id/activate
PUT  /im/api/conversations/:id/read
GET  /im/api/conversations/:id/messages
POST /im/api/conversations/:id/messages
```

创建或获取单聊：

```json
{
  "peer_id": "user_b"
}
```

发送消息：

```json
{
  "client_message_id": "optional-client-id",
  "type": "text",
  "content": "你好"
}
```

## Groups

```http
POST   /im/api/groups
POST   /im/api/groups/get-or-create
GET    /im/api/groups/:id
PATCH  /im/api/groups/:id
GET    /im/api/groups/:id/members
POST   /im/api/groups/:id/members
DELETE /im/api/groups/:id/members/:user_id
POST   /im/api/groups/:id/leave
POST   /im/api/groups/:id/admins
DELETE /im/api/groups/:id/admins/:user_id
POST   /im/api/groups/:id/dissolve
```

创建群聊：

```json
{
  "name": "内容审核群",
  "member_ids": ["user_a", "user_b"]
}
```

## WebSocket

发送消息：

```json
{
  "client_message_id": "optional-client-id",
  "conversation_id": "conversation_object_id",
  "type": "text",
  "content": "Hello"
}
```

服务端推送的消息包含：

```json
{
  "id": "message_id",
  "conversation_id": "conversation_object_id",
  "seq": 1,
  "sender_id": "user_a",
  "type": "text",
  "content": "Hello",
  "created_at": "2026-06-29T00:00:00Z"
}
```

## 会话状态

会话列表返回 `member_state`，表示当前用户在该会话中的状态：

```json
{
  "member_state": {
    "status": "active",
    "last_read_seq": 12,
    "unread_count": 0,
    "mention_count": 0
  }
}
```

未读数规则：

- 新消息写入后，发送者外的活跃成员未读数增加。
- 当前用户调用 `PUT /im/api/conversations/:id/read` 后，该会话未读数清零。
- 退群或被踢后，`conversation_members.status` 不再是 `active`，会话不会出现在该用户列表。
