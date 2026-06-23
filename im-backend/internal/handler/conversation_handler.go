package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"d-im/internal/contextx"
	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"
	"log"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	senderID := contextx.MustGetUserID(c)
	conversation, err := h.conversationService.CreatePrivateConversation(c.Request.Context(), senderID, req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	response.Success(c, "success", conversation)
}

// GetConversation 获取会话详情
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conversation, err := h.conversationService.GetConversation(c.Request.Context(), objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation"})
		return
	}

	response.Success(c, "success", conversation)
}

// GetUserConversations 获取用户的所有会话
func (h *ConversationHandler) GetUserConversations(c *gin.Context) {
	senderID := contextx.MustGetUserID(c)

	query := &dto.ConversationQuery{}
	shoudBind := c.ShouldBindQuery(query)
	if shoudBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": shoudBind.Error()})
		return
	}

	var beforeID *primitive.ObjectID
	if query.BeforeID != nil {
		beforeObjectId, err := primitive.ObjectIDFromHex(*query.BeforeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid before_id"})
			return
		}
		beforeID = &beforeObjectId
	}

	conversations, err := h.conversationService.GetUserConversations(c.Request.Context(), senderID, query.Limit, beforeID)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}

	response.Success(c, "success", conversations)
}
