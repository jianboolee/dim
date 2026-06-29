package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"d-im/internal/response"
	"d-im/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type tokenResponse struct {
	Token        string `json:"token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
}

type authTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	Device       struct {
		Platform   string `json:"platform"`
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
		AppVersion string `json:"app_version"`
		PushToken  string `json:"push_token"`
	} `json:"device"`
}

func (h *AuthHandler) Exchange(c *gin.Context) {
	accessToken := readBearerToken(c.GetHeader("Authorization"))
	if accessToken == "" {
		response.Unauthorized(c, "Authorization header is required")
		return
	}

	req := readAuthTokenRequest(c)
	result, err := h.authService.Exchange(c.Request.Context(), accessToken, deviceMetaFromAuthRequest(req))
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.authService.SetRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	response.Success(c, "success", tokenResponse{
		Token:        result.AccessToken,
		ExpiresIn:    result.AccessExpiresIn,
		RefreshToken: result.RefreshToken,
		SessionID:    result.SessionID,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	req := readAuthTokenRequest(c)
	refreshToken, err := h.readRefreshToken(c, req)
	if err != nil {
		h.authService.ClearRefreshCookie(c)
		h.handleAuthError(c, err)
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), refreshToken, deviceMetaFromAuthRequest(req))
	if err != nil {
		h.authService.ClearRefreshCookie(c)
		h.handleAuthError(c, err)
		return
	}

	h.authService.SetRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	response.Success(c, "success", tokenResponse{
		Token:        result.AccessToken,
		ExpiresIn:    result.AccessExpiresIn,
		RefreshToken: result.RefreshToken,
		SessionID:    result.SessionID,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	req := readAuthTokenRequest(c)
	refreshToken, err := h.readRefreshToken(c, req)
	if err == nil && refreshToken != "" {
		_ = h.authService.Logout(c.Request.Context(), refreshToken)
	}

	h.authService.ClearRefreshCookie(c)
	response.Success(c, "success", gin.H{"success": true})
}

func (h *AuthHandler) readRefreshToken(c *gin.Context, req authTokenRequest) (string, error) {
	if token := strings.TrimSpace(req.RefreshToken); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(c.GetHeader("X-Refresh-Token")); token != "" {
		return token, nil
	}
	return h.authService.ReadRefreshToken(c)
}

func readAuthTokenRequest(c *gin.Context) authTokenRequest {
	var req authTokenRequest
	_ = c.ShouldBindJSON(&req)
	return req
}

func deviceMetaFromAuthRequest(req authTokenRequest) service.DeviceMeta {
	return service.DeviceMeta{
		Platform:   req.Device.Platform,
		DeviceID:   req.Device.DeviceID,
		DeviceName: req.Device.DeviceName,
		AppVersion: req.Device.AppVersion,
		PushToken:  req.Device.PushToken,
	}
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAuthTokenMissing):
		response.Unauthorized(c, "refresh token is required")
	case errors.Is(err, service.ErrAuthSessionExpired):
		response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "session expired")
	case errors.Is(err, service.ErrInvalidAuthToken), errors.Is(err, service.ErrAuthSessionNotFound):
		response.Unauthorized(c, "invalid token")
	default:
		response.InternalServerError(c, "authentication failed")
	}
}

func readBearerToken(authHeader string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return ""
	}
	return strings.TrimSpace(authHeader[len(prefix):])
}
