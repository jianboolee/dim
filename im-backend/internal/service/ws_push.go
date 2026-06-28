package service

import (
	"encoding/json"

	"d-im/internal/models"
)

const wsPushChannel = "im:ws:push"

// WSPushEvent API 进程经 Redis 转发给 WS 进程的推送事件
type WSPushEvent struct {
	UserID  string          `json:"user_id"`
	Message json.RawMessage `json:"message"`
}

// WSPushPayload WebSocket 推送给前端客户端的消息信封，
// 包含原始消息和服务端权威的接收者未读数。
type WSPushPayload struct {
	Message     *models.Message `json:"message"`
	UnreadCount int64           `json:"unread_count"`
}
