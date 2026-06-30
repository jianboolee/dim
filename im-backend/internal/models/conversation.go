package models

import (
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationType string

const (
	ConversationTypePrivate ConversationType = "private" // 单聊
	ConversationTypeGroup   ConversationType = "group"   // 群聊
	ConversationTypeSystem  ConversationType = "system"  // 系统会话
	ConversationTypeChannel ConversationType = "channel" // 频道会话
)

// Conversation 表示一个会话
type Conversation struct {
	ID              primitive.ObjectID   `bson:"_id" json:"id"`
	HashID          string               `bson:"hash_id" json:"hash_id"` // 会话ID的哈希值, 根据参与者生成唯一、稳定的 ObjectID（顺序无关 + 去重）
	Type            ConversationType     `bson:"type" json:"type"`       // 会话类型
	GroupID         *primitive.ObjectID  `bson:"group_id,omitempty" json:"group_id,omitempty"`
	MessageSeq      int64                `bson:"message_seq" json:"message_seq"`
	Participants    []string             `bson:"participants" json:"participants"` // 参与者的用户ID列表
	LastMessage     *LastMessageSnapshot `bson:"last_message,omitempty" json:"last_message"`
	PreviewImageURL string               `bson:"preview_image_url,omitempty" json:"preview_image_url,omitempty"` // 会话业务预览图
	DisplayName     string               `bson:"display_name,omitempty" json:"display_name,omitempty"`
	LastActivity    time.Time            `json:"last_activity" bson:"-"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`

	// 新增字段 👇
	RefType     string            `bson:"ref_type,omitempty" json:"ref_type,omitempty"`         // 引用类型，如 "used_car"、"listing"
	RefID       string            `bson:"ref_id,omitempty" json:"ref_id,omitempty"`             // 引用实体ID（例如二手车ID）
	RefTitle    string            `bson:"ref_title,omitempty" json:"ref_title,omitempty"`       // 引用标题（例如车名或商品名，冗余存储方便展示）
	RefSnapshot map[string]string `bson:"ref_snapshot,omitempty" json:"ref_snapshot,omitempty"` // 冗余字段，如价格、车型、封面图
}

func (c *Conversation) GetLastActivityWithMember(member *ConversationMember) time.Time {
	lastActivity := c.UpdatedAt
	if c.LastMessage != nil {
		lastActivity = c.LastMessage.CreatedAt
	}
	if member != nil && member.LastActivatedAt.After(lastActivity) {
		lastActivity = member.LastActivatedAt
	}
	return lastActivity
}

func (c *Conversation) HasParticipant(userID string) bool {
	for _, participantID := range c.Participants {
		if participantID == userID {
			return true
		}
	}
	return false
}

// NormalizeParticipants 规范化参与者列表, 确保参与者列表按字典序排序
func (c *Conversation) NormalizeParticipants() {
	mp := map[string]struct{}{}
	for _, p := range c.Participants {
		mp[p] = struct{}{}
	}
	result := make([]string, 0, len(mp))
	for p := range mp {
		result = append(result, p)
	}
	sort.Strings(result)
	c.Participants = result
}
