package dto

import (
	"d-im/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation 表示一个会话
type ConversationDTO struct {
	ID              primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	Type            models.ConversationType     `bson:"type" json:"type"`
	Participants    []string                    `bson:"participants" json:"participants"` // 参与者的用户ID列表
	LastMessage     *models.LastMessageSnapshot `bson:"last_message,omitempty" json:"last_message,omitempty"`
	DisplayName     string                      `bson:"-" json:"display_name"`
	DisplayAvatar   string                      `bson:"-" json:"display_avatar,omitempty"`
	GroupID         *primitive.ObjectID         `bson:"group_id,omitempty" json:"group_id,omitempty"`
	GroupInfo       *GroupSummaryDTO            `bson:"-" json:"group_info,omitempty"`
	PeerUserInfo    *UserInfoDto                `bson:"-" json:"peer_user_info,omitempty"`
	PreviewImageURL string                      `bson:"preview_image_url,omitempty" json:"preview_image_url,omitempty"` // 会话业务预览图
	MemberState     *ConversationMemberStateDTO `bson:"-" json:"member_state,omitempty"`
	LastActivity    time.Time                   `json:"last_activity" bson:"-"`
	CreatedAt       time.Time                   `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time                   `bson:"updated_at" json:"updated_at"`
}

type ConversationMemberStateDTO struct {
	Status          models.ConversationMemberStatus `json:"status"`
	LastReadSeq     int64                           `json:"last_read_seq"`
	LastReadAt      time.Time                       `json:"last_read_at,omitempty"`
	LastActivatedAt time.Time                       `json:"last_activated_at,omitempty"`
	UnreadCount     int64                           `json:"unread_count"`
	MentionCount    int64                           `json:"mention_count"`
	Muted           bool                            `json:"muted"`
	Pinned          bool                            `json:"pinned"`
}

// ConversationQuery 会话查询参数
type ConversationQuery struct {
	Limit                int64  `form:"limit,omitempty"`
	Cursor               string `form:"cursor,omitempty"`
	Q                    string `form:"q,omitempty"`
	ActiveConversationID string `form:"active_conversation_id,omitempty"`
}

type ConversationListResponse struct {
	Items      []*ConversationDTO `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

// ConversationCreateRequest 创建会话的请求
type ConversationCreateRequest struct {
	PeerID string `json:"peer_id" binding:"required"`
}
