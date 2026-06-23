package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"d-im/internal/config"
	"d-im/internal/contextx"
	jwtpkg "d-im/pkg/jwt"
)

func JWTAuth(cfg *config.Config, jwtValidator *jwtpkg.JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims := &jwtpkg.AuthTokenClaims{}
		if err := jwtValidator.Parse(tokenString, claims); err != nil {
			log.Printf("Token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 将用户ID存储在上下文中
		contextx.SetUserID(c, claims.Subject)
		c.Next()
	}
}
