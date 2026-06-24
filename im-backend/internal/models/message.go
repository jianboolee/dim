package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"      // 已发送
	MessageStatusDelivered MessageStatus = "delivered" // 已送达
	MessageStatusRead      MessageStatus = "read"      // 已读
)

type MessageType string

const (
	// 系统
	MessageTypeSystem MessageType = "system"
	MessageTypePing   MessageType = "ping"
	MessageTypePong   MessageType = "pong"

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

// Message 基础消息结构
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`                // 消息 ID
	ConversationID primitive.ObjectID `bson:"conversation_id" json:"conversation_id"` // 所属会话 ID
	SenderID       string             `bson:"sender_id" json:"sender_id"`             // 发送者 ID
	ReceiverID     string             `bson:"receiver_id" json:"receiver_id"`         // 接收者 ID
	Type           MessageType        `bson:"type" json:"type"`                       // 消息类型

	// 内容与扩展字段
	Content    string   `bson:"content,omitempty" json:"content,omitempty"` // 文本内容 / 摘要
	RawPayload bson.Raw `bson:"raw_payload,omitempty" json:"-"`             // 原始结构体，用于多类型消息
	Payload    *Payload `bson:"payload,omitempty" json:"payload"`           // 解码后的动态结构，前端使用

	// 回复/引用关系
	ParentMessageID *primitive.ObjectID `bson:"parent_message_id,omitempty" json:"parent_message_id,omitempty"` // 回复的消息（用于评论、子消息）
	RootMessageID   *primitive.ObjectID `bson:"root_message_id,omitempty" json:"root_message_id,omitempty"`     // 所属的主消息（例如帖子）

	// 互动状态
	Status         MessageStatus       `bson:"status" json:"status"`                                         // 消息状态：已发送、已送达、已读
	IsRevoked      bool                `bson:"is_revoked" json:"is_revoked"`                                 // 是否被撤回
	MentionedIDs   *[]string           `bson:"mentioned_ids,omitempty" json:"mentioned_ids,omitempty"`       // 被@的人
	QuoteMessageID *primitive.ObjectID `bson:"quote_message_id,omitempty" json:"quote_message_id,omitempty"` // 引用的消息（类似微信回复）

	// 时间戳
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Payload struct {
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	URL         string             `bson:"url" json:"url"`
	ImageURL    *string            `bson:"image_url" json:"image_url"`
	Width       *int               `bson:"width" json:"width"`
	Height      *int               `bson:"height" json:"height"`
	Duration    *int               `bson:"duration" json:"duration"`
	Size        *int               `bson:"size" json:"size"`
	Price       *float64           `bson:"price" json:"price"`
	Currency    *string            `bson:"currency" json:"currency"`
	ExtInt      *int               `bson:"ext_int" json:"ext_int"`
	ExtFloat    *float64           `bson:"ext_float" json:"ext_float"`
	ExtString   *string            `bson:"ext_string" json:"ext_string"`
	ExtArray    *[]string          `bson:"ext_array" json:"ext_array"`
	ExtMap      *map[string]string `bson:"ext_map" json:"ext_map"`
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
