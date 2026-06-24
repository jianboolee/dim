package middleware

import (
	"log"

	"github.com/gin-gonic/gin"

	"d-im/internal/contextx"
	jwtpkg "d-im/pkg/jwt"
)

func JWTAuth(jwtService *jwtpkg.Service) gin.HandlerFunc {
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
		if err := jwtService.Parse(tokenString, claims); err != nil {
			log.Printf("Token validation error: %v", err)
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		contextx.SetUserID(c, claims.Subject)
		c.Next()
	}
}
