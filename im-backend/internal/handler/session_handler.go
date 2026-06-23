package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"d-im/internal/contextx"
	"d-im/internal/response"
	"d-im/internal/service"
)

type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// GetUserStatus 获取用户在线状态
func (h *SessionHandler) GetUserStatus(c *gin.Context) {
	userID := c.Param("user_id")

	session, err := h.sessionService.GetUserStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user status"})
		return
	}

	response.Success(c, "success", session)
}

// GetUsersStatus 批量获取用户在线状态
func (h *SessionHandler) GetUsersStatus(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	sessions, err := h.sessionService.GetUsersStatus(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users status"})
		return
	}

	response.Success(c, "success", sessions)
}

// KeepAlive 保持用户在线状态
func (h *SessionHandler) KeepAlive(c *gin.Context) {
	userID := contextx.MustGetUserID(c)

	if err := h.sessionService.KeepAlive(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to keep alive"})
		return
	}

	response.Success(c, "success", gin.H{"is_online": true})
}

// GetOnlineUserCount 获取在线用户总数
func (h *SessionHandler) GetOnlineUserCount(c *gin.Context) {
	count, err := h.sessionService.GetOnlineUserCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get online user count"})
		return
	}
	response.Success(c, "success", count)
}
