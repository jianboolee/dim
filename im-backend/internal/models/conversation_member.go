package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationMemberStatus string

const (
	ConversationMemberStatusActive ConversationMemberStatus = "active"
	ConversationMemberStatusLeft   ConversationMemberStatus = "left"
	ConversationMemberStatusKicked ConversationMemberStatus = "kicked"
)

type ConversationMember struct {
	ID                primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	ConversationID    primitive.ObjectID       `bson:"conversation_id" json:"conversation_id"`
	UserID            string                   `bson:"user_id" json:"user_id"`
	Status            ConversationMemberStatus `bson:"status" json:"status"`
	RoleSnapshot      string                   `bson:"role_snapshot,omitempty" json:"role_snapshot,omitempty"`
	LastReadSeq       int64                    `bson:"last_read_seq" json:"last_read_seq"`
	LastReadMessageID *primitive.ObjectID      `bson:"last_read_message_id,omitempty" json:"last_read_message_id,omitempty"`
	LastReadAt        time.Time                `bson:"last_read_at,omitempty" json:"last_read_at,omitempty"`
	LastActivatedAt   time.Time                `bson:"last_activated_at,omitempty" json:"last_activated_at,omitempty"`
	SortAt            time.Time                `bson:"sort_at" json:"sort_at"`
	UnreadCount       int64                    `bson:"unread_count" json:"unread_count"`
	MentionCount      int64                    `bson:"mention_count" json:"mention_count"`
	Muted             bool                     `bson:"muted" json:"muted"`
	Pinned            bool                     `bson:"pinned" json:"pinned"`
	JoinedAt          time.Time                `bson:"joined_at" json:"joined_at"`
	CreatedAt         time.Time                `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time                `bson:"updated_at" json:"updated_at"`
}

func (m *ConversationMember) IsActive() bool {
	return m != nil && m.Status == ConversationMemberStatusActive
}
