package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"d-im/internal/contextx"
	"d-im/internal/response"
	jwtpkg "d-im/pkg/jwt"
)

type AuthHandler struct {
	jwtService *jwtpkg.Service
}

func NewAuthHandler(jwtService *jwtpkg.Service) *AuthHandler {
	return &AuthHandler{jwtService: jwtService}
}

type refreshTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// Refresh 滑动续期 access token
func (h *AuthHandler) Refresh(c *gin.Context) {
	claims := contextx.GetTokenClaims(c)
	if claims == nil {
		response.Unauthorized(c, "invalid token")
		return
	}

	token, expiresIn, err := h.jwtService.RefreshToken(claims)
	if err != nil {
		if errors.Is(err, jwtpkg.ErrSessionExpired) {
			response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "session expired")
			return
		}
		response.InternalServerError(c, "failed to refresh token")
		return
	}

	response.Success(c, "success", refreshTokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	})
}
