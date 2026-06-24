package handler

import (
	"github.com/gin-gonic/gin"

	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"
)

type IntegrationHandler struct {
	integrationService *service.IntegrationService
}

func NewIntegrationHandler(integrationService *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{integrationService: integrationService}
}

func (h *IntegrationHandler) CreateConversation(c *gin.Context) {
	var req dto.IntegrationCreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.integrationService.CreateConversationSession(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "success", result)
}
