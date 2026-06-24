package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"d-im/internal/dto"
	"d-im/internal/models"
	jwtpkg "d-im/pkg/jwt"
)

type IntegrationService struct {
	userService         *UserService
	conversationService *ConversationService
	jwtService          *jwtpkg.Service
	frontendBaseURL     string
}

func NewIntegrationService(
	userService *UserService,
	conversationService *ConversationService,
	jwtService *jwtpkg.Service,
	frontendBaseURL string,
) *IntegrationService {
	return &IntegrationService{
		userService:         userService,
		conversationService: conversationService,
		jwtService:          jwtService,
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
	}
	return s.userService.UpsertUsers(ctx, user)
}

func (s *IntegrationService) signTokenForUser(userID string) (string, error) {
	return s.jwtService.SignUserToken(userID)
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

	token, err := s.signTokenForUser(req.User.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &dto.IntegrationLoginResponse{
		Token:       token,
		ExpiresIn:   s.jwtService.ExpiresInSeconds(),
		RedirectURL: s.buildEnterRedirectURL(token, ""),
	}, nil
}

func (s *IntegrationService) CreateConversationSession(
	ctx context.Context,
	req *dto.IntegrationCreateConversationRequest,
) (*dto.IntegrationCreateConversationResponse, error) {
	if req.FromUser.ID == req.ToUser.ID {
		return nil, fmt.Errorf("from_user and to_user must be different")
	}

	users := []*models.User{
		{
			ID:       req.FromUser.ID,
			Nickname: req.FromUser.ResolveNickname(),
			Avatar:   req.FromUser.ResolveAvatar(),
		},
		{
			ID:       req.ToUser.ID,
			Nickname: req.ToUser.ResolveNickname(),
			Avatar:   req.ToUser.ResolveAvatar(),
		},
	}
	if err := s.userService.UpsertUsers(ctx, users...); err != nil {
		return nil, fmt.Errorf("failed to upsert users: %w", err)
	}

	conversation, err := s.conversationService.GetOrCreatePrivateConversation(ctx, req.FromUser.ID, req.ToUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	token, err := s.signTokenForUser(req.FromUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &dto.IntegrationCreateConversationResponse{
		Token:          token,
		ExpiresIn:      s.jwtService.ExpiresInSeconds(),
		ConversationID: conversation.ID.Hex(),
		RedirectURL:    s.buildEnterRedirectURL(token, conversation.ID.Hex()),
	}, nil
}
