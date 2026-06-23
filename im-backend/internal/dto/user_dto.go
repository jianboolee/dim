package dto

import "d-im/internal/models"

type UserInfoDto struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func ConvertToUserInfoDto(user *models.User) *UserInfoDto {
	return &UserInfoDto{
		ID:       user.ID,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
}
