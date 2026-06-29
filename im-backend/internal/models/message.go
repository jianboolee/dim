package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageStatus string

const (
	MessageStatusSent MessageStatus = "sent" // 已发送
)

type MessageType string

const (
	// 系统
	MessageTypeSystem      MessageType = "system"
	MessageTypeSystemEvent MessageType = "system_event"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"

	// 文本类
	MessageTypeText  MessageType = "text"
	MessageTypeEmoji MessageType = "emoji"
	MessageTypeQuote MessageType = "quote"

	// 媒体类
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"

	// 结构化内容
	MessageTypeLink MessageType = "link"
	MessageTypeCard MessageType = "card"
	MessageTypePoll MessageType = "poll"
	MessageTypeForm MessageType = "form"
	MessageTypePost MessageType = "post" // 用于社区动态/帖子

	// 行为类（记录但不展示）
	MessageTypeFavorite MessageType = "favorite"
	MessageTypeLike     MessageType = "like"
)

const (
	SystemEventGroupCreated       = "group_created"
	SystemEventMemberJoined       = "member_joined"
	SystemEventMemberKicked       = "member_kicked"
	SystemEventMemberLeft         = "member_left"
	SystemEventGroupDissolved     = "group_dissolved"
	SystemEventGroupNameUpdated   = "group_name_updated"
	SystemEventGroupAvatarUpdated = "group_avatar_updated"
	SystemEventAdminAdded         = "admin_added"
	SystemEventAdminRemoved       = "admin_removed"
)

// Message 基础消息结构
type Message struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`                                        // 消息 ID
	ClientMessageID string             `bson:"client_message_id,omitempty" json:"client_message_id,omitempty"` // 客户端生成的幂等 ID
	ConversationID  primitive.ObjectID `bson:"conversation_id" json:"conversation_id"`                         // 所属会话 ID
	Seq             int64              `bson:"seq" json:"seq"`                                                 // 会话内单调递增序号
	SenderID        string             `bson:"sender_id" json:"sender_id"`                                     // 发送者 ID
	Type            MessageType        `bson:"type" json:"type"`                                               // 消息类型

	// 内容与扩展字段
	Content    string   `bson:"content,omitempty" json:"content,omitempty"` // 文本内容 / 摘要
	RawPayload bson.Raw `bson:"raw_payload,omitempty" json:"-"`             // 原始结构体，用于多类型消息
	Payload    *Payload `bson:"payload,omitempty" json:"payload"`           // 解码后的动态结构，前端使用

	// 回复/引用关系
	ParentMessageID *primitive.ObjectID `bson:"parent_message_id,omitempty" json:"parent_message_id,omitempty"` // 回复的消息（用于评论、子消息）
	RootMessageID   *primitive.ObjectID `bson:"root_message_id,omitempty" json:"root_message_id,omitempty"`     // 所属的主消息（例如帖子）

	// 互动状态
	Status         MessageStatus       `bson:"status" json:"status"`                                         // 消息状态
	IsRevoked      bool                `bson:"is_revoked" json:"is_revoked"`                                 // 是否被撤回
	MentionedIDs   *[]string           `bson:"mentioned_ids,omitempty" json:"mentioned_ids,omitempty"`       // 被@的人
	QuoteMessageID *primitive.ObjectID `bson:"quote_message_id,omitempty" json:"quote_message_id,omitempty"` // 引用的消息（类似微信回复）

	// 时间戳
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Payload struct {
	Title         string            `bson:"title,omitempty" json:"title,omitempty"`
	Description   string            `bson:"description,omitempty" json:"description,omitempty"`
	URL           string            `bson:"url,omitempty" json:"url,omitempty"`
	ImageURL      string            `bson:"image_url,omitempty" json:"image_url,omitempty"`
	PriceText     string            `bson:"price_text,omitempty" json:"price_text,omitempty"`
	Meta          map[string]string `bson:"meta,omitempty" json:"meta,omitempty"`
	EventType     string            `bson:"event_type,omitempty" json:"event_type,omitempty"`
	OperatorID    string            `bson:"operator_id,omitempty" json:"operator_id,omitempty"`
	TargetUserIDs []string          `bson:"target_user_ids,omitempty" json:"target_user_ids,omitempty"`
	GroupID       string            `bson:"group_id,omitempty" json:"group_id,omitempty"`
	GroupName     string            `bson:"group_name,omitempty" json:"group_name,omitempty"`
	BeforeValue   string            `bson:"before_value,omitempty" json:"before_value,omitempty"`
	AfterValue    string            `bson:"after_value,omitempty" json:"after_value,omitempty"`
}

// 解析消息的Payload

// GenerateDigest 根据消息类型和 Payload 生成展示用摘要，写入 Content 字段。
func (m *Message) GenerateDigest() {
	switch m.Type {
	case MessageTypeText, MessageTypeQuote, MessageTypeSystem, MessageTypeSystemEvent, MessageTypePost:
		// 文本类：保留原始 content
		return
	case MessageTypeImage:
		m.Content = "[图片]"
	case MessageTypeVideo:
		m.Content = "[视频]"
	case MessageTypeAudio:
		m.Content = "[语音]"
	case MessageTypeFile:
		m.Content = "[文件]"
	case MessageTypeEmoji:
		m.Content = "[表情]"
	case MessageTypeCard:
		if m.Payload != nil && m.Payload.Title != "" {
			m.Content = m.Payload.Title
		} else {
			m.Content = "[卡片]"
		}
	case MessageTypeLink:
		if m.Payload != nil && m.Payload.Title != "" {
			m.Content = "[链接] " + m.Payload.Title
		} else {
			m.Content = "[链接]"
		}
	default:
		m.Content = "[消息]"
	}
}

// LastMessageSnapshot 会话列表用的最后一条消息快照，仅保留展示必要字段。
type LastMessageSnapshot struct {
	Content   string    `bson:"content" json:"content"`
	Type      string    `bson:"type" json:"type"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// 解析消息的Payload
func (m *Message) DecodePayload() error {

	if m.RawPayload == nil {
		return nil
	}

	var payload Payload
	if err := bson.Unmarshal(m.RawPayload, &payload); err != nil {
		return err
	}

	m.Payload = &payload
	return nil
}

// 编码消息的Payload
func (m *Message) EncodePayload() error {
	if m.Payload == nil {
		return nil
	}

	var err error
	m.RawPayload, err = bson.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return nil
}
