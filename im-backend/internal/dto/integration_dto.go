package dto

import "strings"

type IntegrationUserInput struct {
	ID        string `json:"id" binding:"required"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"` // 可选：normal / system / bot，不传时默认为 normal
}

type IntegrationEnsureUsersRequest struct {
	Users []IntegrationUserInput `json:"users" binding:"required"`
}

type DeviceInput struct {
	Platform   string `json:"platform"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	AppVersion string `json:"app_version"`
	PushToken  string `json:"push_token"`
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

type IntegrationLoginRequest struct {
	UserID string      `json:"user_id" binding:"required"`
	Device DeviceInput `json:"device"`
}

type IntegrationLoginResponse struct {
	Token        string `json:"token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	RedirectURL  string `json:"redirect_url"`
}
