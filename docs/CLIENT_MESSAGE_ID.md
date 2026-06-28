# Client Message ID 规范

## 概述

`client_message_id` 是客户端生成的全局唯一标识，用于**幂等去重**和**发送/回显一致性**。每条消息在客户端创建时分配，贯穿发送、存储、回显全链路。

## 生成规则

```
格式: cmid_{userID}_{uuid}
示例: cmid_user123_550e8400-e29b-41d4-a716-446655440000
```

- 前缀 `cmid_` 标识来源为客户端
- `userID` 发送者 ID，用于多用户场景下隔离
- `uuid` 使用 `crypto.randomUUID()`（降级方案 `Date.now()-random` 仅保留作为兜底）

```typescript
const createClientMessageId = (userId: string) => {
  const random =
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(36).slice(2)}`
  return `cmid_${userId}_${random}`
}
```

## 生命周期

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 前端 gen cid                                              │
│    createClientMessageId() → "cmid_user123_xxx"             │
├─────────────────────────────────────────────────────────────┤
│ 2. 创建本地临时消息                                           │
│    { id: "temp-cmid_...", client_message_id: "cmid_...",    │
│      status: "sending" }                                    │
├─────────────────────────────────────────────────────────────┤
│ 3. HTTP POST /messages                                       │
│    body: { client_message_id, type, content, payload }      │
├─────────────────────────────────────────────────────────────┤
│ 4. 后端处理                                                   │
│    a. 检查 MongoDB 唯一索引 (conversation_id+sender_id+cid)  │
│    b. 已存在 → 返回已有消息（幂等）                              │
│    c. 不存在 → 插入新消息 → 生成 ObjectID                    │
│    d. 异步 WS push 给 sender + receiver                      │
├─────────────────────────────────────────────────────────────┤
│ 5. 前端 HTTP 响应                                             │
│    confirmPendingMessage() → 用服务端返回的消息替换本地临时消息   │
├─────────────────────────────────────────────────────────────┤
│ 6. 前端 WS 回显                                               │
│    handleNewMessage() → mergeMessages()                      │
│    按 client_message_id 去重，不重复插入                       │
└─────────────────────────────────────────────────────────────┘
```

## 去重机制

### 前端 `mergeMessages()`

```typescript
// 按 client_message_id 或 id 匹配已存在的消息
const existingIndex = next.findIndex((item) => {
  if (msg.client_message_id && item.client_message_id === msg.client_message_id) return true
  if (msg.id && item.id === msg.id) return true
  return false
})
```

合并策略：`{ ...existing, ...msg }`。已存在则覆盖更新，不存在则追加。

### 后端唯一索引

```
MongoDB Index: unique_message_client_id
  Fields: { conversation_id: 1, sender_id: 1, client_message_id: 1 }
  Unique:  true
  Partial: client_message_id 非空
```

## 关键约束

| 规则 | 说明 |
|------|------|
| 每条消息必须带 cid | 发送时必须透传，不允许丢失或截断 |
| 同一用户在同一会话内 cid 唯一 | 后端唯一索引保证，重复发送返回已有消息 |
| cid 不替代 ObjectID | `id` 是服务端主键，`cid` 是客户端幂等键 |
| temp id 格式 `temp-{cid}` | 本地临时消息的 `id` 前缀，便于过滤 |

## 常见问题

### Q: WS 回显和 HTTP 响应谁先到？

先到先得，后到的被 `mergeMessages` 去重合并。不会出现重复。

### Q: HTTP 发送失败怎么办？

本地临时消息标记 `status: "failed"`，用户可点击重试。重试复用同一个 cid，后端幂等返回已有消息（如果之前其实成功了）。

### Q: cid 在不同会话间会冲突吗？

不会。唯一索引约束了 `conversation_id + sender_id + cid`，不同会话完全独立。并且 cid 前缀包含了 `userID`。

### Q: 后端需要做哪些验证？

- 检查 cid 格式不为空
- 唯一索引自动兜底重复插入
- `FindByClientMessageID` 方法用于幂等查询
