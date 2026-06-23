package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"d-im/internal/response"
	"d-im/internal/service"
	jwtpkg "d-im/pkg/jwt"
	"d-im/pkg/logger"
)

type WSHandler struct {
	wsManager    *service.WSManager
	jwtValidator *jwtpkg.JWTValidator
	upgrader     websocket.Upgrader
}

func NewWSHandler(wsManager *service.WSManager, jwtValidator *jwtpkg.JWTValidator) *WSHandler {
	return &WSHandler{
		wsManager:    wsManager,
		jwtValidator: jwtValidator,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该更严格
			},
		},
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	// 从 URL 参数获取 token
	token := c.Query("token")
	if token == "" {
		response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "Token is required")
		return
	}

	// 验证 token
	claims := &jwtpkg.AuthTokenClaims{}
	if err := h.jwtValidator.Parse(token, claims); err != nil {
		log.Printf("Token validation error: %v", err)
		response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, err.Error())
		return
	}

	// 升级连接为 WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to upgrade connection")
		return
	}

	// 使用解析出的用户ID
	userID := claims.Subject

	logger.Debug("WebSocket connection established for user", zap.String("user_id", userID))

	client := &service.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.wsManager.Register <- client

	// 启动读写 goroutines
	go client.WritePump()
	go client.ReadPump(h.wsManager)
}
