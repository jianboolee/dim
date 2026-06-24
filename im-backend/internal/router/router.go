package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"d-im/internal/config"
	"d-im/internal/handler"
)

// SetupAPI 设置API路由
func SetupAPI(
	cfg *config.Config,
	messageHandler *handler.MessageHandler,
	conversationHandler *handler.ConversationHandler,
	sessionHandler *handler.SessionHandler,
	userHandler *handler.UserHandler,
	integrationHandler *handler.IntegrationHandler,
	jwtAuthMiddleware gin.HandlerFunc,
	integrationAPIKeyMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	im := r.Group("/im")

	integration := im.Group("/api/integration")
	integration.Use(integrationAPIKeyMiddleware)
	{
		integration.POST("/conversations", integrationHandler.CreateConversation)
	}

	api := im.Group("/api")
	api.Use(jwtAuthMiddleware)
	{
		api.GET("/users/me", userHandler.GetMe)
		api.GET("/users/:id", userHandler.GetUser)

		api.GET("/messages", messageHandler.GetMessages)
		api.POST("/messages", messageHandler.SendMessageHTTP)
		api.GET("/messages/unread/count", messageHandler.GetUnreadCount)
		api.PUT("/messages/:id/read", messageHandler.MarkMessageAsRead)

		api.GET("/conversations", conversationHandler.GetUserConversations)
		api.GET("/conversations/:id/messages", messageHandler.GetMessagesByConversationID)
		api.POST("/conversations", conversationHandler.CreateConversation)
		api.GET("/conversations/:id", conversationHandler.GetConversation)

		api.GET("/sessions/:user_id", sessionHandler.GetUserStatus)
		api.POST("/sessions/keepalive", sessionHandler.KeepAlive)
		api.POST("/sessions/batch", sessionHandler.GetUsersStatus)
		api.GET("/sessions/online/count", sessionHandler.GetOnlineUserCount)
	}

	return r
}

// SetupWS 设置WebSocket路由
func SetupWS(
	cfg *config.Config,
	wsHandler *handler.WSHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	r.GET("/im/ws", wsHandler.HandleWebSocket)

	return r
}
