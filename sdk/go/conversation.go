package dim

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type ConversationService struct {
	client *client
}

type CreatePrivateConversationOption interface {
	applyCreatePrivateConversationOption(*CreatePrivateConversationRequest)
}

type InitialMemberMutedOption interface {
	CreatePrivateConversationOption
	SendMessageOption
}

type initialMemberMutedOption struct {
	userID string
	muted  bool
}

func WithInitialMemberMuted(userID string, muted bool) InitialMemberMutedOption {
	return initialMemberMutedOption{userID: userID, muted: muted}
}

func (o initialMemberMutedOption) applyCreatePrivateConversationOption(req *CreatePrivateConversationRequest) {
	if req.InitialMemberSettings == nil {
		req.InitialMemberSettings = map[string]ConversationInitialMemberSettings{}
	}
	settings := req.InitialMemberSettings[o.userID]
	settings.Muted = &o.muted
	req.InitialMemberSettings[o.userID] = settings
}

func (o initialMemberMutedOption) applySendMessageOption(options *sendMessageOptions) {
	options.ensureInitialMemberSettings()
	settings := options.initialMemberSettings[o.userID]
	settings.Muted = &o.muted
	options.initialMemberSettings[o.userID] = settings
}

func (s *ConversationService) GetOrCreatePrivate(ctx context.Context, peerID string, options ...CreatePrivateConversationOption) (*Conversation, error) {
	var out Conversation
	req := CreatePrivateConversationRequest{PeerID: peerID}
	for _, option := range options {
		option.applyCreatePrivateConversationOption(&req)
	}
	if err := s.client.do(ctx, http.MethodPost, "/im/api/conversations", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ConversationService) Get(ctx context.Context, conversationID string) (*Conversation, error) {
	var out Conversation
	if err := s.client.do(ctx, http.MethodGet, "/im/api/conversations/"+url.PathEscape(conversationID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ConversationService) Activate(ctx context.Context, conversationID string) (*Conversation, error) {
	var out Conversation
	if err := s.client.do(ctx, http.MethodPost, "/im/api/conversations/"+url.PathEscape(conversationID)+"/activate", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ConversationService) MarkRead(ctx context.Context, conversationID string) error {
	return s.client.do(ctx, http.MethodPut, "/im/api/conversations/"+url.PathEscape(conversationID)+"/read", nil, nil)
}

func (s *ConversationService) UpdateSettings(ctx context.Context, conversationID string, patch ConversationSettingsPatch) (*ConversationMemberState, error) {
	var out ConversationMemberState
	if err := s.client.do(ctx, http.MethodPatch, "/im/api/conversations/"+url.PathEscape(conversationID)+"/settings", patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ConversationService) List(ctx context.Context, params ListConversationsParams) (*ConversationPage, error) {
	values := url.Values{}
	if params.Limit > 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Cursor != "" {
		values.Set("cursor", params.Cursor)
	}
	if params.Query != "" {
		values.Set("q", params.Query)
	}
	if params.ActiveConversationID != "" {
		values.Set("active_conversation_id", params.ActiveConversationID)
	}

	path := "/im/api/conversations"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out ConversationPage
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
