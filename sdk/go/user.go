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
// 卡片消息支持标题、描述、图片和链接地址，常用于商品推荐、活动推送等场景。
// title 和 description 二选一即可，不填的传空字符串 ""。
//
// 示例：
//
//	msg, err := userClient.SendCardMessage(ctx, conversationID,
//	    "限时特惠", "2021款宝马3系仅售18.8万", "https://img.example.com/car.jpg", "https://example.com/car/123",
//	)
func (c *UserClient) SendCardMessage(ctx context.Context, conversationID string, title, description, imageURL, linkURL string) (*Message, error) {
	return c.SendMessage(ctx, conversationID, SendMessageRequest{
		Type:    "card",
		Content: title,
		Payload: &Payload{
			Title:       title,
			Description: description,
			ImageURL:    imageURL,
			URL:         linkURL,
		},
	})
}
