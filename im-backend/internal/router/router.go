package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"d-im/internal/config"
	"d-im/internal/handler"
	"d-im/internal/upload"
)

// SetupAPI 设置API路由
func SetupAPI(
	cfg *config.Config,
	messageHandler *handler.MessageHandler,
	conversationHandler *handler.ConversationHandler,
	groupHandler *handler.GroupHandler,
	sessionHandler *handler.SessionHandler,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	integrationHandler *handler.IntegrationHandler,
	uploadHandler *upload.Handler,
	groupAvatarDirectory string,
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
	if groupAvatarDirectory != "" {
		im.StaticFS("/assets/group-avatars", gin.Dir(groupAvatarDirectory, false))
	}

	integration := im.Group("/api/integration")
	integration.Use(integrationAPIKeyMiddleware)
	{
		integration.POST("/users", integrationHandler.EnsureUsers)
		integration.POST("/login", integrationHandler.Login)
	}

	api := im.Group("/api")
	api.POST("/auth/exchange", authHandler.Exchange)
	api.POST("/auth/refresh", authHandler.Refresh)
	api.POST("/auth/logout", authHandler.Logout)
	api.Use(jwtAuthMiddleware)
	{
		api.GET("/users/me", userHandler.GetMe)
		api.GET("/users/:id", userHandler.GetUser)

		api.GET("/messages/unread/count", messageHandler.GetUnreadCount)
		api.POST("/uploads/image", uploadHandler.UploadImage)
		api.POST("/uploads/images", uploadHandler.UploadImages)

		api.GET("/conversations", conversationHandler.GetUserConversations)
		api.GET("/conversations/:id/messages", messageHandler.GetMessagesByConversationID)
		api.GET("/conversations/:id/messages/search", messageHandler.SearchMessages)
		api.POST("/conversations/:id/messages", messageHandler.SendMessageToConversation)
		api.POST("/conversations", conversationHandler.CreateConversation)
		api.POST("/conversations/:id/activate", conversationHandler.ActivateConversation)
		api.PUT("/conversations/:id/read", conversationHandler.MarkConversationRead)
		api.PATCH("/conversations/:id/settings", conversationHandler.UpdateSettings)
		api.GET("/conversations/:id", conversationHandler.GetConversation)

		api.POST("/groups/conversations", groupHandler.CreateGroupConversation)
		api.POST("/groups/conversations/get-or-create", groupHandler.GetOrCreateGroupConversation)
		api.GET("/groups/:id", groupHandler.GetGroup)
		api.PATCH("/groups/:id", groupHandler.UpdateGroup)
		api.GET("/groups/:id/members", groupHandler.GetMembers)
		api.POST("/groups/:id/members", groupHandler.AddMembers)
		api.DELETE("/groups/:id/members/:user_id", groupHandler.KickMember)
		api.POST("/groups/:id/leave", groupHandler.LeaveGroup)
		api.POST("/groups/:id/admins", groupHandler.AddAdmin)
		api.DELETE("/groups/:id/admins/:user_id", groupHandler.RemoveAdmin)

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
