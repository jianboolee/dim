package service

import (
	"context"
	"fmt"

	"d-im/internal/models"
	"d-im/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) UpsertUser(ctx context.Context, id, nickname, avatar string) (*models.User, error) {
	user := &models.User{
		ID:       id,
		Nickname: nickname,
		Avatar:   avatar,
	}
	if err := s.repo.Upsert(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) UpsertUsers(ctx context.Context, users ...*models.User) error {
	return s.repo.UpsertMany(ctx, users)
}

func (s *UserService) GetUserInfo(ctx context.Context, userID string) (*models.User, error) {
	return s.repo.GetByID(ctx, userID)
}
