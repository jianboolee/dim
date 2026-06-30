package dto

import (
	"d-im/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendMessageRequest 发送会话消息请求；会话 ID 来自 URL，成员权限由后端根据 conversation_members 判断。
type SendMessageRequest struct {
	ClientMessageID string              `json:"client_message_id,omitempty"`
	Content         string              `json:"content"`
	Type            *models.MessageType `json:"type,omitempty"`
	Payload         *models.Payload     `json:"payload,omitempty"`
}

type ConversationMessageQuery struct {
	ConversationID string `form:"conversation_id"`
	BeforeID       string `form:"before_id,omitempty"`
	Limit          int64  `form:"limit,omitempty"`
}

type MessageDTO struct {
	ID              primitive.ObjectID   `json:"id,omitempty"`
	ClientMessageID string               `json:"client_message_id,omitempty"`
	ConversationID  primitive.ObjectID   `json:"conversation_id,omitempty"`
	Seq             int64                `json:"seq,omitempty"`
	SenderID        string               `json:"sender_id,omitempty"`
	SenderProfile   *UserInfoDto         `json:"sender_profile,omitempty"`
	Type            models.MessageType   `json:"type"`
	Content         string               `json:"content,omitempty"`
	PreviewText     string               `json:"preview_text,omitempty"`
	Payload         *models.Payload      `json:"payload,omitempty"`
	Status          models.MessageStatus `json:"status,omitempty"`
	CreatedAt       time.Time            `json:"created_at,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at,omitempty"`
}

type LastMessageDTO struct {
	ID             primitive.ObjectID `json:"id,omitempty"`
	ConversationID primitive.ObjectID `json:"conversation_id,omitempty"`
	Seq            int64              `json:"seq,omitempty"`
	SenderID       string             `json:"sender_id,omitempty"`
	SenderProfile  *UserInfoDto       `json:"sender_profile,omitempty"`
	Type           string             `json:"type"`
	Content        string             `json:"content"`
	PreviewText    string             `json:"preview_text,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

func ConvertToMessageDTO(message *models.Message, senderProfile *UserInfoDto) *MessageDTO {
	if message == nil {
		return nil
	}
	return &MessageDTO{
		ID:              message.ID,
		ClientMessageID: message.ClientMessageID,
		ConversationID:  message.ConversationID,
		Seq:             message.Seq,
		SenderID:        message.SenderID,
		SenderProfile:   senderProfile,
		Type:            message.Type,
		Content:         message.Content,
		PreviewText:     message.PreviewText,
		Payload:         message.Payload,
		Status:          message.Status,
		CreatedAt:       message.CreatedAt,
		UpdatedAt:       message.UpdatedAt,
	}
}

func ConvertToLastMessageDTO(snapshot *models.LastMessageSnapshot, senderProfile *UserInfoDto) *LastMessageDTO {
	if snapshot == nil {
		return nil
	}
	return &LastMessageDTO{
		ID:             snapshot.ID,
		ConversationID: snapshot.ConversationID,
		Seq:            snapshot.Seq,
		SenderID:       snapshot.SenderID,
		SenderProfile:  senderProfile,
		Type:           snapshot.Type,
		Content:        snapshot.Content,
		PreviewText:    snapshot.PreviewText,
		CreatedAt:      snapshot.CreatedAt,
	}
}

// GetLimit 返回查询限制，如果未设置则返回默认值
func (q ConversationMessageQuery) GetLimit() int64 {
	if q.Limit <= 0 {
		return 50 // 默认限制
	}
	return q.Limit
}
