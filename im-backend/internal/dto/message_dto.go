package dto

import (
	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ReceiverID string              `json:"receiver_id" binding:"required"`
	Content    string              `json:"content"`
	Type       *models.MessageType `json:"type,omitempty"`
	Payload    *models.Payload     `json:"payload,omitempty"`
}

type MessageRequest struct {
	ReceiverID     string              `json:"receiver_id"`
	Content        string              `json:"content"`
	ConversationID *primitive.ObjectID `json:"conversation_id,omitempty"`
	MessageType    *string             `json:"message_type,omitempty"`
	Payload        *string             `json:"payload,omitempty"`
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
