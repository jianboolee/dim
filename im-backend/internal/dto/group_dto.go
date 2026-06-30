package dto

import (
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupCreateRequest struct {
	Name        string   `json:"name"`
	MemberIDs   []string `json:"member_ids,omitempty"`
	UniqueKey   string   `json:"unique_key,omitempty"`
	ScopeUserID string   `json:"scope_user_id,omitempty"`
}

type GroupUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}

type GroupAddMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

type GroupSetAdminRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type GroupDTO struct {
	ID             primitive.ObjectID `json:"id"`
	ConversationID primitive.ObjectID `json:"conversation_id"`
	Name           string             `json:"name"`
	AvatarURL      string             `json:"avatar_url,omitempty"`
	OwnerID        string             `json:"owner_id"`
	ScopeUserID    string             `json:"scope_user_id,omitempty"`
	UniqueKey      string             `json:"unique_key,omitempty"`
	MemberCount    int                `json:"member_count"`
	Status         models.GroupStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type GroupMemberDTO struct {
	ID            primitive.ObjectID       `json:"id"`
	GroupID       primitive.ObjectID       `json:"group_id"`
	UserID        string                   `json:"user_id"`
	Role          models.GroupMemberRole   `json:"role"`
	Status        models.GroupMemberStatus `json:"status"`
	GroupNickname string                   `json:"group_nickname,omitempty"`
	JoinedAt      time.Time                `json:"joined_at"`
	InvitedBy     string                   `json:"invited_by,omitempty"`
	UserInfo      *UserInfoDto             `json:"user_info,omitempty"`
}

type GroupSummaryDTO struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	AvatarURL   string             `json:"avatar_url,omitempty"`
	MemberCount int                `json:"member_count"`
}

type GroupDetailResponse struct {
	Group   *GroupDTO        `json:"group"`
	Members []GroupMemberDTO `json:"members,omitempty"`
}

func ConvertToGroupDTO(group *models.Group) *GroupDTO {
	if group == nil {
		return nil
	}
	return &GroupDTO{
		ID:             group.ID,
		ConversationID: group.ConversationID,
		Name:           group.Name,
		AvatarURL:      group.AvatarURL,
		OwnerID:        group.OwnerID,
		ScopeUserID:    group.ScopeUserID,
		UniqueKey:      group.UniqueKey,
		MemberCount:    group.MemberCount,
		Status:         group.Status,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
	}
}
