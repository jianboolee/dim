package dto

import (
	"d-im/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation 表示一个会话
type ConversationDTO struct {
	ID           primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	Type         models.ConversationType `bson:"type" json:"type"`
	Participants []string                `bson:"participants" json:"participants"` // 参与者的用户ID列表
	LastMessage  *models.Message         `bson:"last_message,omitempty" json:"last_message,omitempty"`
	ToUserInfo   *UserInfoDto            `bson:"-" json:"to_user_info,omitempty"`    // 对方的用户信息
	ImageURL     string                  `bson:"image_url" json:"image_url"`         // 会话图片
	UnreadCounts map[string]int64        `bson:"unread_counts" json:"unread_counts"` // 每个用户的未读数
	LastActivity time.Time               `json:"last_activity" bson:"-"`
	CreatedAt    time.Time               `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time               `bson:"updated_at" json:"updated_at"`
}

// ConversationQuery 会话查询参数
type ConversationQuery struct {
	Limit    int64   `json:"limit,omitempty"`
	BeforeID *string `json:"before_id,omitempty"`
}

// ConversationCreateRequest 创建会话的请求
type ConversationCreateRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
}
