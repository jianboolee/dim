package dto

import (
	"strings"

	"d-im/internal/models"
)

type UserInfoDto struct {
	ID       string           `json:"id"`
	Nickname string           `json:"nickname"`
	Avatar   string           `json:"avatar"`
	Type     models.UserType  `json:"type,omitempty"`
}

func ConvertToUserInfoDto(user *models.User) *UserInfoDto {
	t := user.Type
	if t == "" && strings.HasPrefix(user.ID, "system_") {
		t = models.UserTypeSystem
	}
	return &UserInfoDto{
		ID:       user.ID,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Type:     t,
	}
}
