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

func (s *ConversationService) GetOrCreatePrivate(ctx context.Context, peerID string) (*Conversation, error) {
	var out Conversation
	req := CreatePrivateConversationRequest{PeerID: peerID}
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
