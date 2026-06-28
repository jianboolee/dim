package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupMemberRole string

const (
	GroupMemberRoleOwner  GroupMemberRole = "owner"
	GroupMemberRoleAdmin  GroupMemberRole = "admin"
	GroupMemberRoleMember GroupMemberRole = "member"
	GroupMemberRoleBot    GroupMemberRole = "bot"
)

type GroupMemberStatus string

const (
	GroupMemberStatusActive GroupMemberStatus = "active"
	GroupMemberStatusLeft   GroupMemberStatus = "left"
	GroupMemberStatusKicked GroupMemberStatus = "kicked"
)

type GroupMember struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID       primitive.ObjectID `bson:"group_id" json:"group_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	Role          GroupMemberRole    `bson:"role" json:"role"`
	Status        GroupMemberStatus  `bson:"status" json:"status"`
	GroupNickname string             `bson:"group_nickname,omitempty" json:"group_nickname,omitempty"`
	JoinedAt      time.Time          `bson:"joined_at" json:"joined_at"`
	InvitedBy     string             `bson:"invited_by,omitempty" json:"invited_by,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

func (m *GroupMember) IsActive() bool {
	return m != nil && m.Status == GroupMemberStatusActive
}

func (m *GroupMember) CanKick(target *GroupMember) bool {
	if !m.IsActive() || !target.IsActive() {
		return false
	}
	if m.Role == GroupMemberRoleOwner {
		return target.Role != GroupMemberRoleOwner
	}
	if m.Role == GroupMemberRoleAdmin {
		return target.Role == GroupMemberRoleMember || target.Role == GroupMemberRoleBot
	}
	return false
}
