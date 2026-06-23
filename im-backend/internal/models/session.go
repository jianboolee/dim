package models

import "time"

type Session struct {
	UserID   string    `bson:"user_id" json:"user_id"`     // 用户ID
	IsOnline bool      `bson:"is_online" json:"is_online"` // 是否在线
	LastSeen time.Time `bson:"last_seen" json:"last_seen"` // 最后在线时间
}
