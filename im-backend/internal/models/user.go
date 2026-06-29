package models

import (
	"time"
)

const (
	SystemUserID = "system_user_id"
)

// UserType 用户类型
type UserType string

const (
	UserTypeNormal UserType = "normal" // 普通用户
	UserTypeSystem UserType = "system" // 系统用户（只读，用户不能回复）
	UserTypeBot    UserType = "bot"    // 机器人/客服（可回复）
)

// IsSystemLike 判断是否为服务端托管账号（system 或 bot），用于限制实时客户端连接。
func (t UserType) IsSystemLike() bool {
	return t == UserTypeSystem || t == UserTypeBot
}

type User struct {
	ID        string    `bson:"id" json:"id"`
	Nickname  string    `bson:"nickname" json:"nickname"` // 昵称
	Avatar    string    `bson:"avatar" json:"avatar"`     // 头像
	Bio       string    `bson:"bio" json:"bio"`           // 个人简介
	Type      UserType  `bson:"type,omitempty" json:"type,omitempty"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
