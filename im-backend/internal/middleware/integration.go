package middleware

import (
	"github.com/gin-gonic/gin"

	"d-im/internal/config"
	"d-im/internal/response"
)

func IntegrationAPIKey(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Integration-Key")
		if apiKey == "" || cfg.Integration.APIKey == "" || apiKey != cfg.Integration.APIKey {
			response.Unauthorized(c, "invalid integration api key")
			c.Abort()
			return
		}
		c.Next()
	}
}
