# 会话基础架构约定

当前项目不保留旧的消息级目标用户与逐条已读机制。

## 核心规则

- 消息只归属 `conversation_id`，不携带单独的目标用户字段。
- 单聊和群聊都通过 `conversation_members` 管理成员可见性、排序状态和未读数。
- 会话列表只返回当前用户的 active 成员关系。
- 已读是会话级动作：`PUT /im/api/conversations/:id/read`。
- WebSocket 发送消息必须携带 `conversation_id`。

## 主要职责

- `MessageService`：保存消息、分配会话内 `seq`、更新最后消息、按会话成员扇出推送。
- `ConversationService`：创建单聊、获取会话列表、激活会话、会话级已读。
- `GroupService`：维护群组生命周期，并同步 `conversation_members`。
- `ConversationMemberRepository`：维护每个用户在会话内的状态，包括未读数、排序时间、成员状态。
