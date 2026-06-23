package service

import (
	"context"
	"time"

	"d-im/internal/models"
	"d-im/internal/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

// UpdateUserStatus 更新用户在线状态
func (s *SessionService) UpdateUserStatus(ctx context.Context, userID string, isOnline bool) error {
	session := &models.Session{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}
	return s.repo.UpsertSession(ctx, session)
}

// GetUserStatus 获取用户在线状态
func (s *SessionService) GetUserStatus(ctx context.Context, userID string) (*models.Session, error) {
	session, err := s.repo.GetSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// IsOnline 检查用户是否在线
func (s *SessionService) IsOnline(userID string) (bool, error) {
	session, err := s.GetUserStatus(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return session.IsOnline, nil
}

// GetUsersStatus 批量获取用户在线状态
func (s *SessionService) GetUsersStatus(ctx context.Context, userIDs []string) (map[string]*models.Session, error) {
	sessionMap, err := s.repo.GetMultiSession(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return sessionMap, nil
}

// UpdateLastSeen 更新用户最后在线时间
func (s *SessionService) UpdateLastSeen(ctx context.Context, userID string) error {
	return s.repo.UpdateLastSeen(ctx, userID)
}

// KeepAlive 保持用户在线状态
func (s *SessionService) KeepAlive(ctx context.Context, userID string) error {
	session := &models.Session{
		UserID:   userID,
		IsOnline: true,
		LastSeen: time.Now(),
	}
	return s.repo.UpsertSession(ctx, session)
}

// GetOnlineUserCount 获取在线用户总数
func (s *SessionService) GetOnlineUserCount(ctx context.Context) (int64, error) {
	return s.repo.GetOnlineUserCount(ctx)
}
