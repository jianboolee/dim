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
			Nickname: req.FromUser.Nickname,
			Avatar:   req.FromUser.Avatar,
		},
		{
			ID:       req.ToUser.ID,
			Nickname: req.ToUser.Nickname,
			Avatar:   req.ToUser.Avatar,
		},
	}
	if err := s.userService.UpsertUsers(ctx, users...); err != nil {
		return nil, fmt.Errorf("failed to upsert users: %w", err)
	}

	conversation, err := s.conversationService.GetOrCreatePrivateConversation(ctx, req.FromUser.ID, req.ToUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	token, err := s.jwtService.SignUserToken(req.FromUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	baseURL := strings.TrimRight(s.frontendBaseURL, "/")
	redirectURL := fmt.Sprintf(
		"%s/im/enter?token=%s&conversation_id=%s",
		baseURL,
		url.QueryEscape(token),
		url.QueryEscape(conversation.ID.Hex()),
	)

	return &dto.IntegrationCreateConversationResponse{
		Token:          token,
		ConversationID: conversation.ID.Hex(),
		RedirectURL:    redirectURL,
	}, nil
}
