package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminService *service.AdminService
	logger       *zap.Logger
}

func NewAdminHandler(adminService *service.AdminService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{adminService: adminService, logger: logger}
}

type AdminLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	token, err := h.adminService.Login(req.Username, req.Password)
	if err != nil {
		h.logger.Warn("admin login failed", zap.Error(err))
		response.Unauthorized(c, "invalid credentials")
		return
	}

	response.Success(c, gin.H{"token": token})
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)
	keyword := c.Query("q")
	users, total, err := h.adminService.GetUsers(c.Request.Context(), page, pageSize, keyword)
	if err != nil {
		h.logger.Error("admin get users failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": users, "total": total})
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.adminService.BanUser(c.Request.Context(), id); err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *AdminHandler) UnbanUser(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.adminService.UnbanUser(c.Request.Context(), id); err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *AdminHandler) DeleteCheckin(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.adminService.DeleteCheckin(c.Request.Context(), id); err != nil {
		h.logger.Error("admin delete checkin failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *AdminHandler) DeleteInsight(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.adminService.DeleteInsight(c.Request.Context(), id); err != nil {
		h.logger.Error("admin delete insight failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}