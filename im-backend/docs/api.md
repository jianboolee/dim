# IM 系统 API 文档

## 目录
- [概述](#概述)
- [认证](#认证)
- [WebSocket](#websocket)
- [REST API](#rest-api)
- [数据模型](#数据模型)
- [SDK 使用](#sdk-使用)

## 概述

本文档描述了 IM 系统的 WebSocket 和 HTTP REST API。系统支持：
- WebSocket 实时消息
- REST API 消息收发
- 会话管理
- 未读消息计数
- 在线状态管理

## 认证

所有的 API 请求都需要通过 JWT Token 认证。Token 需要在 HTTP 请求的 Header 中通过 `Authorization` 字段传递：

```
Authorization: Bearer your-jwt-token
```

## WebSocket

### 连接

```
ws://localhost:8080/api/ws?token=your-jwt-token
```

注意：token 需要通过 URL 参数传递，而不是 Header。

### 消息格式

发送消息：
```json
{
  "to_id": "recipient-user-id",
  "type": "text",  // 消息类型：text/image/video/audio/card
  "content": "消息内容",
  "media_info": {  // 媒体消息必填
    "url": "https://example.com/media.jpg",
    "size": 1024000,
    "duration": 60000,  // 音视频时长（毫秒）
    "width": 1920,      // 图片/视频宽度
    "height": 1080,     // 图片/视频高度
    "format": "jpg"     // 文件格式
  },
  "card_info": {   // 卡片消息必填
    "title": "标题",
    "description": "描述",
    "url": "https://example.com/link",
    "image_url": "https://example.com/cover.jpg"
  }
}
```

接收消息：
```json
{
  "id": "message-id",
  "to_id": "recipient-user-id",
  "from_id": "sender-user-id",
  "type": "text",
  "content": "消息内容",
  "media_info": {
    "url": "https://example.com/media.jpg",
    "size": 1024000,
    "duration": 60000,
    "width": 1920,
    "height": 1080,
    "format": "jpg"
  },
  "card_info": {
    "title": "标题",
    "description": "描述",
    "url": "https://example.com/link",
    "image_url": "https://example.com/cover.jpg"
  },
  "status": "sent",
  "created_at": "2024-01-16T10:00:00Z",
  "updated_at": "2024-01-16T10:00:00Z"
}
```

### 消息类型说明

1. 文本消息 (text)
   - 必填字段：content
   ```json
   {
     "to_id": "user123",
     "type": "text",
     "content": "Hello, world!"
   }
   ```

2. 图片消息 (image)
   - 必填字段：media_info.url
   ```json
   {
     "to_id": "user123",
     "type": "image",
     "content": "这是一张风景照片",
     "media_info": {
       "url": "https://example.com/image.jpg",
       "size": 1024000,
       "width": 1920,
       "height": 1080,
       "format": "jpg"
     }
   }
   ```

3. 视频消息 (video)
   - 必填字段：media_info.url
   ```json
   {
     "to_id": "user123",
     "type": "video",
     "content": "精彩视频",
     "media_info": {
       "url": "https://example.com/video.mp4",
       "size": 10240000,
       "duration": 60000,
       "width": 1920,
       "height": 1080,
       "format": "mp4"
     }
   }
   ```

4. 音频消息 (audio)
   - 必填字段：media_info.url
   ```json
   {
     "to_id": "user123",
     "type": "audio",
     "content": "语音消息",
     "media_info": {
       "url": "https://example.com/audio.mp3",
       "size": 512000,
       "duration": 30000,
       "format": "mp3"
     }
   }
   ```

5. 卡片消息 (card)
   - 必填字段：card_info.url
   ```json
   {
     "to_id": "user123",
     "type": "card",
     "content": "分享一个链接",
     "card_info": {
       "title": "Go语言实战",
       "description": "学习 Go 语言的必备教程",
       "url": "https://example.com/go-tutorial",
       "image_url": "https://example.com/cover.jpg"
     }
   }
   ```

## REST API

### 发送消息

**POST** `/api/im/messages`

请求体：
```json
{
  "to_id": "recipient-user-id",
  "type": "text",  // 消息类型：text/image/video/audio/card
  "content": "Hello, world!",
  "media_info": {  // 媒体消息必填
    "url": "https://example.com/media.jpg",
    "size": 1024000,
    "duration": 60000,
    "width": 1920,
    "height": 1080,
    "format": "jpg"
  },
  "card_info": {   // 卡片消息必填
    "title": "标题",
    "description": "描述",
    "url": "https://example.com/link",
    "image_url": "https://example.com/cover.jpg"
  }
}
```

响应：
```json
{
  "_id": "65a123b789cdef0123456789",
  "from_id": "sender-user-id",
  "to_id": "recipient-user-id",
  "content": "Hello, world!",
  "status": "sent",
  "created_at": "2024-01-16T10:00:00Z",
  "updated_at": "2024-01-16T10:00:00Z"
}
```

### 获取消息列表

**GET** `/api/im/messages`

查询参数：
- `to_id`: 目标用户 ID
- `start_time`: 开始时间 (RFC3339 格式)
- `end_time`: 结束时间 (RFC3339 格式)
- `limit`: 返回消息数量限制 (默认 50)
- `skip`: 跳过消息数量

响应：
```json
[
  {
    "_id": "65a123b789cdef0123456789",
    "from_id": "user1",
    "to_id": "user2",
    "content": "Hello",
    "status": "read",
    "created_at": "2024-01-16T10:00:00Z",
    "updated_at": "2024-01-16T10:01:00Z"
  }
]
```

### 获取会话列表

**GET** `/api/im/conversations`

响应：
```json
[
  {
    "_id": "65a123b789cdef0123456789",
    "type": "private",
    "participants": ["user1", "user2"],
    "last_message": {
      "_id": "65a123b789cdef0123456789",
      "from_id": "user1",
      "to_id": "user2",
      "content": "Hello",
      "status": "sent",
      "created_at": "2024-01-16T10:00:00Z",
      "updated_at": "2024-01-16T10:00:00Z"
    },
    "unread_counts": {
      "user1": 0,
      "user2": 1
    },
    "created_at": "2024-01-16T10:00:00Z",
    "updated_at": "2024-01-16T10:00:00Z",
    "last_activity": "2024-01-16T10:00:00Z"
  }
]
```

### 获取未读消息数

**GET** `/api/im/messages/unread/count`

响应：
```json
{
  "unread_count": 5
}
```

### 标记消息为已读

**PUT** `/api/im/messages/{message_id}/read`

响应：
- 成功: HTTP 200
- 失败: HTTP 4xx/5xx 与错误信息

### 获取用户在线状态

**GET** `/api/im/sessions/{user_id}`

响应：
```json
{
  "user_id": "user1",
  "is_online": true,
  "last_seen": "2024-01-16T10:00:00Z"
}
```

### 批量获取用户在线状态

**POST** `/api/im/sessions/batch`

请求体：
```json
{
  "user_ids": ["user1", "user2", "user3"]
}
```

响应：
```json
{
  "user1": {
    "user_id": "user1",
    "is_online": true,
    "last_seen": "2024-01-16T10:00:00Z"
  },
  "user2": {
    "user_id": "user2",
    "is_online": false,
    "last_seen": "2024-01-16T09:30:00Z"
  }
}
```

### 保持在线状态

**POST** `/api/im/sessions/keepalive`

响应：
- 成功: HTTP 200
- 失败: HTTP 4xx/5xx 与错误信息

## 数据模型

### Message 消息

| 字段 | 类型 | 说明 |
|------|------|------|
| _id | ObjectID | 消息 ID |
| from_id | string | 发送者 ID |
| to_id | string | 接收者 ID |
| type | string | 消息类型：text/image/video/audio/card |
| content | string | 消息内容或描述 |
| status | string | 消息状态：sent/delivered/read |
| media_info | object | 媒体信息（图片/视频/音频） |
| card_info | object | 卡片信息（链接卡片） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### MediaInfo 媒体信息

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 媒体文件的 URL |
| size | int64 | 文件大小（字节） |
| duration | int64 | 音视频时长（毫秒） |
| width | int | 图片/视频宽度 |
| height | int | 图片/视频高度 |
| format | string | 文件格式 |

### CardInfo 卡片信息

| 字段 | 类型 | 说明 |
|------|------|------|
| title | string | 卡片标题 |
| description | string | 卡片描述 |
| url | string | 链接地址 |
| image_url | string | 卡片图片 |

### Conversation 会话

| 字段 | 类型 | 说明 |
|------|------|------|
| _id | ObjectID | 会话 ID |
| type | string | 会话类型：private |
| participants | string[] | 参与者 ID 列表 |
| last_message | Message | 最后一条消息 |
| unread_counts | map | 每个用户的未读消息数 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| last_activity | datetime | 最后活动时间 |

### Session 会话状态

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | string | 用户 ID |
| is_online | boolean | 是否在线 |
| last_seen | datetime | 最后在线时间 |

## SDK 使用

### 初始化

```javascript
import IMSDK from './im.js';

const im = new IMSDK({
  baseURL: 'http://localhost:8080',
  token: 'your-jwt-token'
});
```

### 连接服务器

```javascript
// 监听连接状态
im.onConnection(({ status }) => {
  console.log('Connection status:', status);
});

// 连接服务器
await im.connect();
```

### 发送和接收消息

```javascript
// 监听新消息
im.onMessage((message) => {
  console.log('收到消息:', message);
  // message 格式: { to_id: "xxx", from_id: "xxx", content: "xxx" }
});

// 通过 HTTP API 发送消息
const message = await im.sendMessage('recipient-id', 'Hello!');

// 通过 WebSocket 发送消息
await im.sendMessageWS('recipient-id', 'Hello via WebSocket!');
```

### 获取消息和会话

```javascript
// 获取消息列表
const messages = await im.getMessages({
  toUserID: 'recipient-id',
  limit: 20
});

// 获取会话列表
const conversations = await im.getConversations();

// 获取未读消息数
const { unread_count } = await im.getUnreadCount();
```

### 在线状态管理

```javascript
// 获取用户在线状态
const status = await im.getUserStatus('user-id');

// 保持在线状态（建议每30秒调用一次）
await im.keepAlive();
```

### 标记消息已读

```javascript
await im.markMessageAsRead('message-id');
```

### 断开连接

```javascript
im.disconnect();
```

## 错误处理

所有 API 在发生错误时会返回适当的 HTTP 状态码和错误信息：

- 400 Bad Request: 请求参数错误
- 401 Unauthorized: 认证失败
- 403 Forbidden: 权限不足
- 404 Not Found: 资源不存在
- 500 Internal Server Error: 服务器内部错误

错误响应格式：
```json
{
  "error": "错误信息"
}
```