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
	jwtAuthMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// 添加中间件
	r.Use(gin.Recovery())

	// 健康检查路由 - 不需要认证
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API服务不包含WebSocket路由

	// IM 路由组
	im := r.Group("/im")

	// IM 相关路由
	api := im.Group("/api")
	api.Use(jwtAuthMiddleware)
	{
		// 消息相关路由
		api.GET("/messages", messageHandler.GetMessages)                 // 获取消息列表
		api.POST("/messages", messageHandler.SendMessageHTTP)            // 发送消息
		api.GET("/messages/unread/count", messageHandler.GetUnreadCount) // 获取未读消息数
		api.PUT("/messages/:id/read", messageHandler.MarkMessageAsRead)  // 标记消息为已读

		// 会话相关路由
		api.GET("/conversations", conversationHandler.GetUserConversations)                // 获取会话列表
		api.GET("/conversations/:id/messages", messageHandler.GetMessagesByConversationID) // 获取会话消息
		api.POST("/conversations", conversationHandler.CreateConversation)                 // 创建会话
		api.GET("/conversations/:id", conversationHandler.GetConversation)                 // 获取会话详情

		// 会话状态相关路由
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

	// 添加中间件
	r.Use(gin.Recovery())

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// WebSocket 连接
	r.GET("/im/ws", wsHandler.HandleWebSocket)

	return r
}
