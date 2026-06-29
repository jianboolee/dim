package handler

import (
	"errors"

	"d-im/internal/contextx"
	"d-im/internal/dto"
	"d-im/internal/response"
	"d-im/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req dto.GroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.groupService.CreateGroup(c.Request.Context(), contextx.MustGetUserID(c), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) GetOrCreateGroup(c *gin.Context) {
	var req dto.GroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.groupService.GetOrCreateGroup(c.Request.Context(), contextx.MustGetUserID(c), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	result, err := h.groupService.GetGroup(c.Request.Context(), groupID, contextx.MustGetUserID(c))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var req dto.GroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, contextx.MustGetUserID(c), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) GetMembers(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	result, err := h.groupService.ListMembers(c.Request.Context(), groupID, contextx.MustGetUserID(c))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) AddMembers(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var req dto.GroupAddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.groupService.AddMembers(c.Request.Context(), groupID, contextx.MustGetUserID(c), req.UserIDs)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) KickMember(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	result, err := h.groupService.KickMember(c.Request.Context(), groupID, contextx.MustGetUserID(c), c.Param("user_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	if err := h.groupService.LeaveGroup(c.Request.Context(), groupID, contextx.MustGetUserID(c)); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", gin.H{"success": true})
}

func (h *GroupHandler) AddAdmin(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}
	var req dto.GroupSetAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.groupService.AddAdmin(c.Request.Context(), groupID, contextx.MustGetUserID(c), req.UserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) RemoveAdmin(c *gin.Context) {
	groupID, ok := parseGroupID(c)
	if !ok {
		return
	}

	result, err := h.groupService.RemoveAdmin(c.Request.Context(), groupID, contextx.MustGetUserID(c), c.Param("user_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, "success", result)
}

func (h *GroupHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGroupNotFound):
		response.NotFound(c, "Group not found")
	case errors.Is(err, service.ErrGroupDissolved):
		response.Forbidden(c, "Group dissolved")
	case errors.Is(err, service.ErrGroupAccessDenied):
		response.Forbidden(c, "Forbidden")
	case errors.Is(err, service.ErrGroupPermissionDenied), errors.Is(err, service.ErrGroupOwnerRequired):
		response.Forbidden(c, "Permission denied")
	case errors.Is(err, service.ErrGroupMemberLimit):
		response.BadRequest(c, "Group member limit exceeded")
	case errors.Is(err, service.ErrGroupUniqueKeyRequired):
		response.BadRequest(c, "unique_key is required")
	default:
		response.InternalServerError(c, err.Error())
	}
}

func parseGroupID(c *gin.Context) (primitive.ObjectID, bool) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return primitive.NilObjectID, false
	}
	return groupID, true
}
