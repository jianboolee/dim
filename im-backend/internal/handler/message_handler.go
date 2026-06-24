package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"d-im/internal/contextx"
	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"
)

type MessageHandler struct {
	messageService      *service.MessageService
	conversationService *service.ConversationService
}

func NewMessageHandler(messageService *service.MessageService, conversationService *service.ConversationService) *MessageHandler {
	return &MessageHandler{
		messageService:      messageService,
		conversationService: conversationService,
	}
}

func (h *MessageHandler) GetMessagesByConversationID(c *gin.Context) {
	conversationIDStr := c.Param("id")
	beforeIDStr := c.Query("before_id")
	afterIDStr := c.Query("after_id")
	limitStr := c.Query("limit")

	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	beforeID := primitive.NilObjectID
	if beforeIDStr != "" {
		beforeID, err = primitive.ObjectIDFromHex(beforeIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid before ID")
			return
		}
	}
	afterID := primitive.NilObjectID
	if afterIDStr != "" {
		afterID, err = primitive.ObjectIDFromHex(afterIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid after ID")
			return
		}
	}

	limit := int64(20)
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid limit")
			return
		}
	}

	currentUserID := contextx.MustGetUserID(c)
	messages, err := h.messageService.FindMessagesByConversationID(c.Request.Context(), conversationID, currentUserID, &beforeID, &afterID, limit)
	if err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Error(c, http.StatusForbidden, http.StatusForbidden, "Forbidden")
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	response.Success(c, "success", messages)

}

// GetUnreadCount 获取未读消息数
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	count, err := h.conversationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkMessageAsRead 标记消息为已读
func (h *MessageHandler) MarkMessageAsRead(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	currentUserID := c.GetString("user_id")
	if err := h.messageService.MarkMessageAsRead(c.Request.Context(), objID, currentUserID); err != nil {
		if err.Error() == "permission denied: only message recipient can mark message as read" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
		return
	}

	c.Status(http.StatusOK)
}

// SendMessageToConversation 通过 HTTP 接口发送会话消息
func (h *MessageHandler) SendMessageToConversation(c *gin.Context) {
	conversationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 从上下文中获取发送者ID
	senderID := contextx.MustGetUserID(c)

	msg, err := h.messageService.SendMessageToConversationHTTP(c.Request.Context(), senderID, conversationID, req.Content, req.Type, req.Payload)
	if err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}
