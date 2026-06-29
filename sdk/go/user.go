package dim

import (
	"context"
	"net/http"
	"net/url"
)

type UserService struct {
	client *client
}

func (s *UserService) Me(ctx context.Context) (*User, error) {
	var out User
	if err := s.client.do(ctx, http.MethodGet, "/im/api/users/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *UserService) Get(ctx context.Context, userID string) (*User, error) {
	var out User
	if err := s.client.do(ctx, http.MethodGet, "/im/api/users/"+url.PathEscape(userID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
