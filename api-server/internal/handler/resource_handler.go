package handler

import (
	"regexp"
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var regexSpecialChars = regexp.MustCompile(`[<>"'&]`)

type ResourceHandler struct {
	resourceService *service.ResourceService
	logger          *zap.Logger
}

func NewResourceHandler(resourceService *service.ResourceService, logger *zap.Logger) *ResourceHandler {
	return &ResourceHandler{resourceService: resourceService, logger: logger}
}

func (h *ResourceHandler) Presign(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	ext := regexSpecialChars.ReplaceAllString(c.Query("ext"), "")
	if ext == "" {
		response.BadRequest(c, "ext is required")
		return
	}

	url, err := h.resourceService.Presign(c.Request.Context(), oid, ext)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"url": url})
}

func (h *ResourceHandler) UploadCallback(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req struct {
		FileID   string `json:"file_id" binding:"required"`
		FileName string `json:"file_name" binding:"required"`
		FileSize int64  `json:"file_size"`
		FileType string `json:"file_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.resourceService.UploadCallback(c.Request.Context(), oid, req.FileID, req.FileName, req.FileSize, req.FileType); err != nil {
		h.logger.Error("upload callback failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

type CreateResourceReq struct {
	Title       string   `json:"title" binding:"required"`
	Type        string   `json:"type" binding:"required"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	URL         string   `json:"url" binding:"required"`
	FileSize    int64    `json:"file_size"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	ShareScope  string   `json:"share_scope"`
}

func (h *ResourceHandler) Create(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	res := &model.Resource{
		Title:       req.Title,
		Type:        req.Type,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		URL:         req.URL,
		FileSize:    req.FileSize,
		Tags:        req.Tags,
		Description: req.Description,
		ShareScope:  req.ShareScope,
	}

	if err := h.resourceService.Create(c.Request.Context(), oid, res); err != nil {
		h.logger.Error("create resource failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"id": res.ID.Hex()})
}

func (h *ResourceHandler) GetByID(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	res, err := h.resourceService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "resource not found")
		return
	}

	response.Success(c, res)
}

func (h *ResourceHandler) List(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	resources, total, err := h.resourceService.List(c.Request.Context(), oid, map[string]string{
		"type":       c.Query("type"),
		"category":   c.Query("category"),
		"difficulty": c.Query("difficulty"),
		"tag":        c.Query("tag"),
		"keyword":    c.Query("keyword"),
		"scope":      c.Query("scope"),
		"sort":       c.Query("sort"),
		"group_id":   c.Query("group_id"),
	}, page, pageSize)
	if err != nil {
		h.logger.Error("list resources failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": resources, "total": total})
}

func (h *ResourceHandler) GetTags(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	tags, err := h.resourceService.GetTags(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, tags)
}

func (h *ResourceHandler) Update(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.resourceService.Update(c.Request.Context(), oid, id, req); err != nil {
		h.logger.Error("update resource failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.resourceService.Delete(c.Request.Context(), oid, id); err != nil {
		h.logger.Error("delete resource failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *ResourceHandler) ToggleFavorite(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	favorited, err := h.resourceService.ToggleFavorite(c.Request.Context(), oid, id)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"favorited": favorited})
}

func (h *ResourceHandler) ListFavorites(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	resources, total, err := h.resourceService.ListFavorites(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": resources, "total": total})
}

func (h *ResourceHandler) GetStats(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	stats, err := h.resourceService.GetStats(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, stats)
}