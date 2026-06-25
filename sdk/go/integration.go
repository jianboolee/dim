package dim

import (
	"context"
	"net/http"
)

type IntegrationClient struct {
	client *client
}

func NewIntegrationClient(cfg Config) *IntegrationClient {
	return &IntegrationClient{client: newClient(cfg)}
}

func (c *IntegrationClient) CreateConversation(ctx context.Context, req CreateConversationRequest) (*CreateConversationResponse, error) {
	var out CreateConversationResponse
	if err := c.client.do(ctx, http.MethodPost, "/im/api/integration/conversations", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *IntegrationClient) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var out LoginResponse
	if err := c.client.do(ctx, http.MethodPost, "/im/api/integration/login", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
