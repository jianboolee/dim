package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"d-im/internal/contextx"
	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"
)

type ConversationHandler struct {
	conversationService *service.ConversationService
}

func NewConversationHandler(conversationService *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

// CreateConversation 创建会话
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req dto.ConversationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	senderID := contextx.MustGetUserID(c)
	conversation, err := h.conversationService.CreatePrivateConversation(c.Request.Context(), senderID, req.PeerID)
	if err != nil {
		if errors.Is(err, service.ErrCannotStartConversationWithSystemUser) {
			response.Forbidden(c, "Cannot start conversation with system user")
			return
		}
		response.InternalServerError(c, "Failed to create conversation")
		return
	}

	response.Success(c, "success", conversation)
}

// GetConversation 获取会话详情
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	currentUserID := contextx.MustGetUserID(c)
	conversation, err := h.conversationService.GetConversation(c.Request.Context(), objID, currentUserID)
	if err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Forbidden(c, "Forbidden")
			return
		}
		response.InternalServerError(c, "Failed to get conversation")
		return
	}

	response.Success(c, "success", conversation)
}

func (h *ConversationHandler) ActivateConversation(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	currentUserID := contextx.MustGetUserID(c)
	conversation, err := h.conversationService.ActivateConversation(c.Request.Context(), objID, currentUserID)
	if err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Forbidden(c, "Forbidden")
			return
		}
		response.InternalServerError(c, "Failed to activate conversation")
		return
	}

	response.Success(c, "success", conversation)
}

func (h *ConversationHandler) MarkConversationRead(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	currentUserID := contextx.MustGetUserID(c)
	if err := h.conversationService.MarkConversationRead(c.Request.Context(), objID, currentUserID); err != nil {
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Forbidden(c, "Forbidden")
			return
		}
		response.InternalServerError(c, "Failed to mark conversation as read")
		return
	}

	response.Success(c, "success", gin.H{"success": true})
}

// GetUserConversations 获取用户的所有会话
func (h *ConversationHandler) GetUserConversations(c *gin.Context) {
	senderID := contextx.MustGetUserID(c)

	query := &dto.ConversationQuery{}
	shoudBind := c.ShouldBindQuery(query)
	if shoudBind != nil {
		response.BadRequest(c, shoudBind.Error())
		return
	}

	conversations, err := h.conversationService.GetUserConversations(c.Request.Context(), senderID, query.Limit, query.Cursor, query.Q, query.ActiveConversationID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidConversationCursor) {
			response.BadRequest(c, "Invalid cursor")
			return
		}
		if errors.Is(err, service.ErrInvalidConversationID) {
			response.BadRequest(c, "Invalid conversation ID")
			return
		}
		if errors.Is(err, service.ErrConversationAccessDenied) {
			response.Forbidden(c, "Forbidden")
			return
		}
		log.Printf("Error getting conversations: %v", err)
		response.InternalServerError(c, "Failed to get conversations")
		return
	}

	response.Success(c, "success", conversations)
}
