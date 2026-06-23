package service

import (
	"context"
	"d-im/internal/models"
	"time"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{
		ID:        userID,
		Nickname:  "test",
		Avatar:    "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return user, nil
}
