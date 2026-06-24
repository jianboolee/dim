package middleware

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"d-im/internal/contextx"
	jwtpkg "d-im/pkg/jwt"
)

// JWTRefreshAuth 允许已过期的 access token（签名有效且未超过绝对会话上限）用于续期
func JWTRefreshAuth(jwtService *jwtpkg.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			c.JSON(401, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := authHeader[len(prefix):]
		claims := &jwtpkg.AuthTokenClaims{}
		if err := jwtService.ParseForRefresh(tokenString, claims); err != nil {
			log.Printf("Refresh token validation error: %v", err)
			message := err.Error()
			if errors.Is(err, jwtpkg.ErrSessionExpired) {
				message = "session expired"
			}
			c.JSON(401, gin.H{"error": message})
			c.Abort()
			return
		}

		contextx.SetUserID(c, claims.Subject)
		contextx.SetTokenClaims(c, claims)
		c.Next()
	}
}
