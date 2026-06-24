package handler

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"d-im/internal/contextx"
	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := contextx.MustGetUserID(c)
	user, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "failed to get user")
		return
	}

	response.Success(c, "success", dto.ConvertToUserInfoDto(user))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.userService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "failed to get user")
		return
	}

	response.Success(c, "success", dto.ConvertToUserInfoDto(user))
}
