package contextx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextKeyUserID ContextKey = "user_id"
)

// SetUserID 设置用户ID
func SetUserID(c *gin.Context, userId string) {
	c.Set(string(ContextKeyUserID), userId)
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
