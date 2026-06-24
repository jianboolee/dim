package service

import "encoding/json"

const wsPushChannel = "im:ws:push"

// WSPushEvent API 进程经 Redis 转发给 WS 进程的推送事件
type WSPushEvent struct {
	ReceiverID string          `json:"receiver_id"`
	Message    json.RawMessage `json:"message"`
}
