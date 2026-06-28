package dim

import (
	"context"
	"net/http"
)

func (c *UserClient) CreateGroup(ctx context.Context, req CreateGroupRequest) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	if err := c.client.do(ctx, http.MethodPost, "/im/api/groups", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
