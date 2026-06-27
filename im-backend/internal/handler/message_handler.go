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
	userID := contextx.MustGetUserID(c)

	count, err := h.conversationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to get unread count")
		return
	}

	response.Success(c, "success", gin.H{"unread_count": count})
}

// MarkMessageAsRead 标记消息为已读
func (h *MessageHandler) MarkMessageAsRead(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	currentUserID := contextx.MustGetUserID(c)
	if err := h.messageService.MarkMessageAsRead(c.Request.Context(), objID, currentUserID); err != nil {
		if err.Error() == "permission denied: only message recipient can mark message as read" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to mark message as read")
		return
	}

	response.Success(c, "success", gin.H{"success": true})
}

// SendMessageToConversation 通过 HTTP 接口发送会话消息
func (h *MessageHandler) SendMessageToConversation(c *gin.Context) {
	conversationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 从上下文中获取发送者ID
	senderID := contextx.MustGetUserID(c)

	msg, err := h.messageService.SendMessageToConversationHTTP(c.Request.Context(), senderID, conversationID, req.ClientMessageID, req.Content, req.Type, req.Payload)
	if err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Forbidden(c, "Forbidden")
			return
		}
		if errors.Is(err, service.ErrCannotReplyToSystemUser) {
			response.Forbidden(c, "不能回复系统消息")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, "success", msg)
}
