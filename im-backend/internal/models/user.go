package models

import (
	"time"
)

const (
	SystemUserID = "system_user_id"
)

type User struct {
	ID        string    `bson:"id" json:"id"`
	Nickname  string    `bson:"nickname" json:"nickname"` // 昵称
	Avatar    string    `bson:"avatar" json:"avatar"`     // 头像
	Bio       string    `bson:"bio" json:"bio"`           // 个人简介
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
