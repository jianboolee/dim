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
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

func (h *AuthHandler) Exchange(c *gin.Context) {
	accessToken := readBearerToken(c.GetHeader("Authorization"))
	if accessToken == "" {
		response.Unauthorized(c, "Authorization header is required")
		return
	}

	result, err := h.authService.Exchange(c.Request.Context(), accessToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.authService.SetRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	response.Success(c, "success", tokenResponse{
		Token:     result.AccessToken,
		ExpiresIn: result.AccessExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := h.authService.ReadRefreshToken(c)
	if err != nil {
		h.authService.ClearRefreshCookie(c)
		h.handleAuthError(c, err)
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.authService.ClearRefreshCookie(c)
		h.handleAuthError(c, err)
		return
	}

	h.authService.SetRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	response.Success(c, "success", tokenResponse{
		Token:     result.AccessToken,
		ExpiresIn: result.AccessExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := h.authService.ReadRefreshToken(c)
	if err == nil && refreshToken != "" {
		_ = h.authService.Logout(c.Request.Context(), refreshToken)
	}

	h.authService.ClearRefreshCookie(c)
	response.Success(c, "success", gin.H{"success": true})
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
