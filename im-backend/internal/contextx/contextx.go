package contextx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	jwtpkg "d-im/pkg/jwt"
)

type ContextKey string

const (
	ContextKeyUserID      ContextKey = "user_id"
	ContextKeyTokenClaims ContextKey = "token_claims"
)

// SetUserID 设置用户ID
func SetUserID(c *gin.Context, userId string) {
	c.Set(string(ContextKeyUserID), userId)
}

// SetTokenClaims 设置 JWT claims（续期等场景使用）
func SetTokenClaims(c *gin.Context, claims *jwtpkg.AuthTokenClaims) {
	c.Set(string(ContextKeyTokenClaims), claims)
}

// GetTokenClaims 获取 JWT claims
func GetTokenClaims(c *gin.Context) *jwtpkg.AuthTokenClaims {
	value, exists := c.Get(string(ContextKeyTokenClaims))
	if !exists {
		return nil
	}

	claims, ok := value.(*jwtpkg.AuthTokenClaims)
	if !ok {
		return nil
	}

	return claims
}

// GetUserId 获取用户ID
func GetUserId(c *gin.Context) string {
	return c.GetString(string(ContextKeyUserID))
}

// MustGetUserID 获取上下文中的用户ID, 如果用户ID不存在, 中断请求
func MustGetUserID(c *gin.Context) string {
	userID, exists := c.Get(string(ContextKeyUserID))
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		c.Abort()
		return ""
	}

	id, ok := userID.(string)
	if !ok || id == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		c.Abort()
		return ""
	}

	return id
}
