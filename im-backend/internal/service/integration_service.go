package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"d-im/internal/dto"
	"d-im/internal/models"
)

type IntegrationService struct {
	userService         *UserService
	conversationService *ConversationService
	authService         *AuthService
	frontendBaseURL     string
}

func NewIntegrationService(
	userService *UserService,
	conversationService *ConversationService,
	authService *AuthService,
	frontendBaseURL string,
) *IntegrationService {
	return &IntegrationService{
		userService:         userService,
		conversationService: conversationService,
		authService:         authService,
		frontendBaseURL:     frontendBaseURL,
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

func (s *IntegrationService) upsertIntegrationUser(ctx context.Context, input dto.IntegrationUserInput) error {
	user := &models.User{
		ID:       input.ID,
		Nickname: input.ResolveNickname(),
		Avatar:   input.ResolveAvatar(),
		Type:     resolveUserType(input.ID, input.Type),
	}
	return s.userService.UpsertUsers(ctx, user)
}

// resolveUserType 根据显式传入或 ID 前缀推断用户类型
func resolveUserType(userID, explicitType string) models.UserType {
	if t := models.UserType(strings.TrimSpace(explicitType)); t != "" {
		return t
	}
	if strings.HasPrefix(userID, "system_") {
		return models.UserTypeSystem
	}
	return ""
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
	if strings.TrimSpace(req.User.ID) == "" {
		return nil, fmt.Errorf("user.id is required")
	}

	if err := s.upsertIntegrationUser(ctx, req.User); err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	session, err := s.authService.CreateSession(ctx, req.User.ID, deviceMetaFromInput(req.Device))
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

func (s *IntegrationService) CreateConversationSession(
	ctx context.Context,
	req *dto.IntegrationPrivateConversationRequest,
) (*dto.IntegrationCreateConversationResponse, error) {
	if req.User.ID == req.PeerUser.ID {
		return nil, fmt.Errorf("user and peer_user must be different")
	}

	users := []*models.User{
		{
			ID:       req.User.ID,
			Nickname: req.User.ResolveNickname(),
			Avatar:   req.User.ResolveAvatar(),
			Type:     resolveUserType(req.User.ID, req.User.Type),
		},
		{
			ID:       req.PeerUser.ID,
			Nickname: req.PeerUser.ResolveNickname(),
			Avatar:   req.PeerUser.ResolveAvatar(),
			Type:     resolveUserType(req.PeerUser.ID, req.PeerUser.Type),
		},
	}
	if err := s.userService.UpsertUsers(ctx, users...); err != nil {
		return nil, fmt.Errorf("failed to upsert users: %w", err)
	}

	conversation, err := s.conversationService.GetOrCreatePrivateConversation(ctx, req.User.ID, req.PeerUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	session, err := s.authService.CreateSession(ctx, req.User.ID, deviceMetaFromInput(req.Device))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}

	return &dto.IntegrationCreateConversationResponse{
		Token:          session.AccessToken,
		ExpiresIn:      session.AccessExpiresIn,
		RefreshToken:   session.RefreshToken,
		SessionID:      session.SessionID,
		ConversationID: conversation.ID.Hex(),
		RedirectURL:    s.buildEnterRedirectURL(session.AccessToken, conversation.ID.Hex()),
	}, nil
}
