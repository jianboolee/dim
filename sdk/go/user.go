package dim

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type UserClient struct {
	client *client
}

type ConversationListOptions struct {
	Limit                int
	Cursor               string
	Query                string
	ActiveConversationID string
}

type MessageListOptions struct {
	Limit    int
	BeforeID string
	AfterID  string
}

func NewUserClient(cfg Config) *UserClient {
	return &UserClient{client: newClient(cfg)}
}

func (c *UserClient) GetConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	var out Conversation
	if err := c.client.do(ctx, http.MethodGet, "/im/api/conversations/"+url.PathEscape(conversationID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *UserClient) ActivateConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	var out Conversation
	if err := c.client.do(ctx, http.MethodPost, "/im/api/conversations/"+url.PathEscape(conversationID)+"/activate", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *UserClient) ListConversations(ctx context.Context, options ConversationListOptions) (*ConversationPage, error) {
	values := url.Values{}
	if options.Limit > 0 {
		values.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Cursor != "" {
		values.Set("cursor", options.Cursor)
	}
	if options.Query != "" {
		values.Set("q", options.Query)
	}
	if options.ActiveConversationID != "" {
		values.Set("active_conversation_id", options.ActiveConversationID)
	}

	path := "/im/api/conversations"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out ConversationPage
	if err := c.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *UserClient) ListMessages(ctx context.Context, conversationID string, options MessageListOptions) ([]Message, error) {
	values := url.Values{}
	if options.Limit > 0 {
		values.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.BeforeID != "" {
		values.Set("before_id", options.BeforeID)
	}
	if options.AfterID != "" {
		values.Set("after_id", options.AfterID)
	}

	path := fmt.Sprintf("/im/api/conversations/%s/messages", url.PathEscape(conversationID))
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []Message
	if err := c.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *UserClient) SendMessage(ctx context.Context, conversationID string, req SendMessageRequest) (*Message, error) {
	var out Message
	if err := c.client.do(ctx, http.MethodPost, "/im/api/conversations/"+url.PathEscape(conversationID)+"/messages", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *UserClient) SendTextMessage(ctx context.Context, conversationID string, content string) (*Message, error) {
	return c.SendMessage(ctx, conversationID, SendMessageRequest{
		Type:    "text",
		Content: content,
	})
}

// SendCardMessage 发送一张卡片消息。
//
// 卡片消息常用于商品推荐、活动推送等场景。
//
// 示例：
//
//	msg, err := userClient.SendCardMessage(ctx, conversationID, dim.Payload{
//	    Title:       "限时特惠",
//	    Description: "2021款宝马3系仅售18.8万",
//	    ImageURL:    "https://img.example.com/car.jpg",
//	    URL:         "https://example.com/car/123",
//	})
func (c *UserClient) SendCardMessage(ctx context.Context, conversationID string, payload Payload) (*Message, error) {
	return c.SendMessage(ctx, conversationID, SendMessageRequest{
		Type:    "card",
		Content: payload.Title,
		Payload: &payload,
	})
}
