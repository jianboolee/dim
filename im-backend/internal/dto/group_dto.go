package dto

import (
	"time"

	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupCreateRequest struct {
	Name      string   `json:"name"`
	AvatarURL string   `json:"avatar_url,omitempty"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

type GroupUpdateRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
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

type GroupMemberBriefDTO struct {
	UserID        string                 `json:"user_id"`
	Role          models.GroupMemberRole `json:"role"`
	GroupNickname string                 `json:"group_nickname,omitempty"`
	UserInfo      *UserInfoDto           `json:"user_info,omitempty"`
}

type GroupSummaryDTO struct {
	ID          primitive.ObjectID    `json:"id"`
	Name        string                `json:"name"`
	AvatarURL   string                `json:"avatar_url,omitempty"`
	MemberCount int                   `json:"member_count"`
	Members     []GroupMemberBriefDTO `json:"members,omitempty"`
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
		MemberCount:    group.MemberCount,
		Status:         group.Status,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
	}
}
