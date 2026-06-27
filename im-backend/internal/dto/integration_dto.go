package dto

import "strings"

type IntegrationUserInput struct {
	ID        string `json:"id" binding:"required"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"` // 可选：normal / system / bot，不传时根据 system_ 前缀自动推断
}

// ResolveAvatar 兼容 avatar / avatar_url 两种字段名
func (u IntegrationUserInput) ResolveAvatar() string {
	if v := strings.TrimSpace(u.Avatar); v != "" {
		return v
	}
	return strings.TrimSpace(u.AvatarURL)
}

func (u IntegrationUserInput) ResolveNickname() string {
	return strings.TrimSpace(u.Nickname)
}

type IntegrationCreateConversationRequest struct {
	FromUser IntegrationUserInput `json:"from_user" binding:"required"`
	ToUser   IntegrationUserInput `json:"to_user" binding:"required"`
}

type IntegrationCreateConversationResponse struct {
	Token          string `json:"token"`
	ExpiresIn      int    `json:"expires_in"`
	ConversationID string `json:"conversation_id"`
	RedirectURL    string `json:"redirect_url"`
}

type IntegrationLoginRequest struct {
	User IntegrationUserInput `json:"user" binding:"required"`
}

type IntegrationLoginResponse struct {
	Token       string `json:"token"`
	ExpiresIn   int    `json:"expires_in"`
	RedirectURL string `json:"redirect_url"`
}
