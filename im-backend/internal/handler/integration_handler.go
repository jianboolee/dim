package handler

import (
	"errors"

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

func (h *IntegrationHandler) EnsureUsers(c *gin.Context) {
	var req dto.IntegrationEnsureUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.integrationService.EnsureUsers(c.Request.Context(), &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "success", gin.H{"success": true})
}

func (h *IntegrationHandler) Login(c *gin.Context) {
	var req dto.IntegrationLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.integrationService.CreateLoginSession(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrIntegrationUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "success", result)
}
