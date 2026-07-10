package handler

import (
	"fmt"
	"time"

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

	h.logger.Info("admin login success", zap.String("username", req.Username), zap.String("ip", c.ClientIP()))
	response.Success(c, gin.H{"token": token})
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("admin get stats failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, stats)
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

func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	detail, err := h.adminService.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, detail)
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.adminService.BanUser(c.Request.Context(), id, "admin", c.ClientIP()); err != nil {
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

	if err := h.adminService.UnbanUser(c.Request.Context(), id, "admin", c.ClientIP()); err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) GetCheckins(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)
	checkins, total, err := h.adminService.GetCheckins(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("admin get checkins failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"list": checkins, "total": total})
}

func (h *AdminHandler) GetInsights(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)
	insights, total, err := h.adminService.GetInsights(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("admin get insights failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"list": insights, "total": total})
}

func (h *AdminHandler) DeleteCheckin(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}
	if err := h.adminService.DeleteCheckin(c.Request.Context(), id, "admin", c.ClientIP()); err != nil {
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
	if err := h.adminService.DeleteInsight(c.Request.Context(), id, "admin", c.ClientIP()); err != nil {
		h.logger.Error("admin delete insight failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, nil)
}

// --- 批量操作 ---

type BatchIDsReq struct {
	IDs []string `json:"ids" binding:"required"`
}

func parseBatchIDs(c *gin.Context) ([]primitive.ObjectID, bool) {
	var req BatchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params: ids array required")
		return nil, false
	}
	if len(req.IDs) > 100 {
		response.BadRequest(c, "too many ids (max 100)")
		return nil, false
	}
	ids := make([]primitive.ObjectID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			response.BadRequest(c, fmt.Sprintf("invalid id: %s", idStr))
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

func (h *AdminHandler) BatchBanUsers(c *gin.Context) {
	ids, ok := parseBatchIDs(c)
	if !ok {
		return
	}
	count, err := h.adminService.BatchBanUsers(c.Request.Context(), ids, "admin", c.ClientIP())
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"affected": count})
}

func (h *AdminHandler) BatchDeleteCheckins(c *gin.Context) {
	ids, ok := parseBatchIDs(c)
	if !ok {
		return
	}
	count, err := h.adminService.BatchDeleteCheckins(c.Request.Context(), ids, "admin", c.ClientIP())
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"affected": count})
}

func (h *AdminHandler) BatchDeleteInsights(c *gin.Context) {
	ids, ok := parseBatchIDs(c)
	if !ok {
		return
	}
	count, err := h.adminService.BatchDeleteInsights(c.Request.Context(), ids, "admin", c.ClientIP())
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"affected": count})
}

// --- 操作日志 ---

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)
	logs, total, err := h.adminService.GetAuditLogs(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("admin get audit logs failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

// --- 系统配置 ---

func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	cfg := h.adminService.GetSystemConfig()
	response.Success(c, cfg)
}

// --- CSV 导出 ---

func (h *AdminHandler) ExportUsers(c *gin.Context) {
	users, _, err := h.adminService.GetUsers(c.Request.Context(), 1, 10000, "")
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=users_%s.csv", time.Now().Format("20060102")))
	c.Writer.Write([]byte("\xEF\xBB\xBF"))
	c.Writer.Write([]byte("ID,昵称,积分,打卡天数,状态,注册时间\n"))
	for _, u := range users {
		status := "正常"
		if u.Status == 1 {
			status = "封禁"
		}
		c.Writer.Write([]byte(fmt.Sprintf("%s,%s,%d,%d,%s,%s\n",
			u.ID.Hex(), u.Nickname, u.Score, u.CheckDays, status, u.CreatedAt.Format("2006-01-02"))))
	}
}

func (h *AdminHandler) ExportCheckins(c *gin.Context) {
	checkins, _, err := h.adminService.GetCheckins(c.Request.Context(), 1, 10000)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=checkins_%s.csv", time.Now().Format("20060102")))
	c.Writer.Write([]byte("\xEF\xBB\xBF"))
	c.Writer.Write([]byte("ID,用户ID,描述,评分,点赞数,评论数,创建时间\n"))
	for _, ci := range checkins {
		c.Writer.Write([]byte(fmt.Sprintf("%s,%s,%s,%d,%d,%d,%s\n",
			ci.ID.Hex(), ci.UserID.Hex(), ci.Description, ci.Score, ci.LikeCount, ci.CommentCount, ci.CreatedAt.Format("2006-01-02"))))
	}
}

func (h *AdminHandler) ExportInsights(c *gin.Context) {
	insights, _, err := h.adminService.GetInsights(c.Request.Context(), 1, 10000)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=insights_%s.csv", time.Now().Format("20060102")))
	c.Writer.Write([]byte("\xEF\xBB\xBF"))
	c.Writer.Write([]byte("ID,用户ID,内容,心情,点赞数,公开,创建时间\n"))
	for _, in := range insights {
		public := "否"
		if in.IsPublic {
			public = "是"
		}
		content := in.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		c.Writer.Write([]byte(fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s\n",
			in.ID.Hex(), in.UserID.Hex(), content, in.Mood, in.LikeCount, public, in.CreatedAt.Format("2006-01-02"))))
	}
}