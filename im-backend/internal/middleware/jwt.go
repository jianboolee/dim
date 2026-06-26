package middleware

import (
	"log"

	"github.com/gin-gonic/gin"

	"d-im/internal/contextx"
	"d-im/internal/response"
	jwtpkg "d-im/pkg/jwt"
)

func JWTAuth(jwtService *jwtpkg.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		tokenString := authHeader[len(prefix):]
		claims := &jwtpkg.AuthTokenClaims{}
		if err := jwtService.ParseAccessToken(tokenString, claims); err != nil {
			log.Printf("Token validation error: %v", err)
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		contextx.SetUserID(c, claims.Subject)
		contextx.SetTokenClaims(c, claims)
		c.Next()
	}
}
