package dim

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type GroupService struct {
	client *client
}

func (s *GroupService) Create(ctx context.Context, req CreateGroupRequest) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	if err := s.client.do(ctx, http.MethodPost, "/im/api/groups", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) GetOrCreate(ctx context.Context, req GetOrCreateGroupParams) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	if err := s.client.do(ctx, http.MethodPost, "/im/api/groups/get-or-create", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) Detail(ctx context.Context, groupID string) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	if err := s.client.do(ctx, http.MethodGet, "/im/api/groups/"+url.PathEscape(groupID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) ListMembers(ctx context.Context, groupID string, limit int, cursor string) (*GroupMemberPage, error) {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		values.Set("cursor", cursor)
	}
	path := "/im/api/groups/" + url.PathEscape(groupID) + "/members"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out GroupMemberPage
	if err := s.client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) Invite(ctx context.Context, groupID string, userIDs []string) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	req := struct {
		UserIDs []string `json:"user_ids"`
	}{UserIDs: userIDs}
	if err := s.client.do(ctx, http.MethodPost, "/im/api/groups/"+url.PathEscape(groupID)+"/members", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) Kick(ctx context.Context, groupID string, userID string) (*GroupDetailResponse, error) {
	var out GroupDetailResponse
	if err := s.client.do(ctx, http.MethodDelete, "/im/api/groups/"+url.PathEscape(groupID)+"/members/"+url.PathEscape(userID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *GroupService) Leave(ctx context.Context, groupID string) error {
	return s.client.do(ctx, http.MethodPost, "/im/api/groups/"+url.PathEscape(groupID)+"/leave", nil, nil)
}
