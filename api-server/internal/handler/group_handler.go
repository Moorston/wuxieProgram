package handler

import (
	"log"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.groupService.List(c.Request.Context())
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, groups)
}

func (h *GroupHandler) Detail(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid group id")
		return
	}

	group, err := h.groupService.GetDetail(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "group not found")
		return
	}

	response.Success(c, group)
}