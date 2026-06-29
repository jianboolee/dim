package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"d-im/internal/dto"
	"d-im/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
)

var ErrIntegrationUserNotFound = errors.New("integration user not found")

type IntegrationService struct {
	userService     *UserService
	authService     *AuthService
	frontendBaseURL string
}

func NewIntegrationService(
	userService *UserService,
	authService *AuthService,
	frontendBaseURL string,
) *IntegrationService {
	return &IntegrationService{
		userService:     userService,
		authService:     authService,
		frontendBaseURL: frontendBaseURL,
	}
}

func (s *IntegrationService) buildEnterRedirectURL(token, conversationID string) string {
	baseURL := strings.TrimRight(s.frontendBaseURL, "/")
	redirectURL := fmt.Sprintf("%s/im/enter?token=%s", baseURL, url.QueryEscape(token))
	if conversationID != "" {
		redirectURL += "&conversation_id=" + url.QueryEscape(conversationID)
	}
	return redirectURL
}

func integrationUsersFromInputs(inputs []dto.IntegrationUserInput) ([]*models.User, error) {
	merged := map[string]*models.User{}
	order := make([]string, 0, len(inputs))

	for _, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if id == "" {
			return nil, fmt.Errorf("user.id is required")
		}
		userType, err := resolveUserType(input.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid user.type for %s: %w", id, err)
		}

		user, exists := merged[id]
		if !exists {
			user = &models.User{ID: id}
			merged[id] = user
			order = append(order, id)
		}
		if nickname := input.ResolveNickname(); nickname != "" {
			user.Nickname = nickname
		}
		if avatar := input.ResolveAvatar(); avatar != "" {
			user.Avatar = avatar
		}
		if userType != "" {
			user.Type = userType
		}
	}

	users := make([]*models.User, 0, len(order))
	for _, id := range order {
		users = append(users, merged[id])
	}
	return users, nil
}

func resolveUserType(explicitType string) (models.UserType, error) {
	switch t := models.UserType(strings.TrimSpace(explicitType)); t {
	case "", models.UserTypeNormal:
		return models.UserTypeNormal, nil
	case models.UserTypeSystem, models.UserTypeBot:
		return t, nil
	default:
		return "", fmt.Errorf("must be normal, system, or bot")
	}
}

func (s *IntegrationService) EnsureUsers(ctx context.Context, req *dto.IntegrationEnsureUsersRequest) error {
	if req == nil || len(req.Users) == 0 {
		return fmt.Errorf("users is required")
	}
	users, err := integrationUsersFromInputs(req.Users)
	if err != nil {
		return err
	}
	return s.userService.UpsertUsers(ctx, users...)
}

func deviceMetaFromInput(input dto.DeviceInput) DeviceMeta {
	return DeviceMeta{
		Platform:   input.Platform,
		DeviceID:   input.DeviceID,
		DeviceName: input.DeviceName,
		AppVersion: input.AppVersion,
		PushToken:  input.PushToken,
	}
}

// CreateLoginSession 业务用户 SSO 进入 IM 会话列表
func (s *IntegrationService) CreateLoginSession(
	ctx context.Context,
	req *dto.IntegrationLoginRequest,
) (*dto.IntegrationLoginResponse, error) {
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if _, err := s.userService.GetUserInfo(ctx, userID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrIntegrationUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	session, err := s.authService.CreateSession(ctx, userID, deviceMetaFromInput(req.Device))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}

	return &dto.IntegrationLoginResponse{
		Token:        session.AccessToken,
		ExpiresIn:    session.AccessExpiresIn,
		RefreshToken: session.RefreshToken,
		SessionID:    session.SessionID,
		RedirectURL:  s.buildEnterRedirectURL(session.AccessToken, ""),
	}, nil
}
