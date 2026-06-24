package dto

import (
	"d-im/internal/models"
)

// SendMessageRequest 发送会话消息请求；会话 ID 来自 URL，接收方由后端根据会话参与者推导。
type SendMessageRequest struct {
	Content string              `json:"content"`
	Type    *models.MessageType `json:"type,omitempty"`
	Payload *models.Payload     `json:"payload,omitempty"`
}

type ConversationMessageQuery struct {
	ConversationID string `form:"conversation_id"`
	BeforeID       string `form:"before_id,omitempty"`
	Limit          int64  `form:"limit,omitempty"`
}

// GetLimit 返回查询限制，如果未设置则返回默认值
func (q ConversationMessageQuery) GetLimit() int64 {
	if q.Limit <= 0 {
		return 50 // 默认限制
	}
	return q.Limit
}
