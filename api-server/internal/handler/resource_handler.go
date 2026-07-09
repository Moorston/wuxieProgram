package handler

import (
	"log"
	"regexp"
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceHandler struct {
	resourceService *service.ResourceService
}

func NewResourceHandler(resourceService *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{resourceService: resourceService}
}

func (h *ResourceHandler) Presign(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	ext := c.DefaultQuery("ext", "mp4")
	// 验证文件扩展名，防止路径遍历
	allowedExt := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !allowedExt.MatchString(ext) || len(ext) > 10 {
		response.BadRequest(c, "invalid file extension")
		return
	}
	objectName := h.resourceService.GenerateObjectName(oid, ext)

	response.Success(c, gin.H{
		"object_name": objectName,
		"bucket":      "resource",
	})
}

type UploadCallbackReq struct {
	ObjectName string `json:"object_name" binding:"required"`
	Bucket     string `json:"bucket" binding:"required"`
	FileSize   int64  `json:"file_size"`
	Title      string `json:"title"`
	CoverURL   string `json:"cover_url"`
	Duration   float64 `json:"duration"`
}

func (h *ResourceHandler) UploadCallback(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req UploadCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	res := &model.Resource{
		Title:    req.Title,
		Type:     model.ResourceTypeVideo,
		FileURL:  req.ObjectName,
		FileSize: req.FileSize,
		CoverURL: req.CoverURL,
		Duration: req.Duration,
	}

	if err := h.resourceService.Create(c.Request.Context(), oid, res); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, res)
}

type CreateResourceReq struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Type        string   `json:"type" binding:"required"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Difficulty  string   `json:"difficulty"`
	FileURL     string   `json:"file_url" binding:"required"`
	FileSize    int64    `json:"file_size"`
	CoverURL    string   `json:"cover_url"`
	Duration    float64  `json:"duration"`
	ShareScope  string   `json:"share_scope"`
	GroupID     string   `json:"group_id"`
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
		Description: req.Description,
		Type:        model.ResourceType(req.Type),
		Category:    req.Category,
		Tags:        req.Tags,
		Difficulty:  req.Difficulty,
		FileURL:     req.FileURL,
		FileSize:    req.FileSize,
		CoverURL:    req.CoverURL,
		Duration:    req.Duration,
		ShareScope:  model.ShareScope(req.ShareScope),
	}

	if req.GroupID != "" {
		if id, err := primitive.ObjectIDFromHex(req.GroupID); err == nil {
			res.GroupID = id
		}
	}

	if err := h.resourceService.Create(c.Request.Context(), oid, res); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, res)
}

func (h *ResourceHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	oid, ok := getUserID(c)
	if !ok {
		return
	}

	res, err := h.resourceService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "resource not found")
		return
	}

	if res.ShareScope != model.ShareScopePublic && res.UserID != oid {
		response.Forbidden(c, "no access")
		return
	}

	h.resourceService.IncrViewCount(c.Request.Context(), id)

	response.Success(c, res)
}

func (h *ResourceHandler) List(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	resType := c.Query("type")
	category := c.Query("category")
	difficulty := c.Query("difficulty")
	tag := c.Query("tag")
	keyword := c.Query("keyword")
	shareScope := c.DefaultQuery("scope", "own")
	sortBy := c.DefaultQuery("sort", "time")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var groupID *primitive.ObjectID
	if gid := c.Query("group_id"); gid != "" {
		if id, err := primitive.ObjectIDFromHex(gid); err == nil {
			groupID = &id
		}
	}

	resources, total, err := h.resourceService.List(c.Request.Context(), oid, resType, category, difficulty, tag, keyword, shareScope, sortBy, groupID, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  resources,
		"total": total,
	})
}

func (h *ResourceHandler) Update(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	allowed := map[string]bool{"title": true, "description": true, "category": true, "tags": true, "difficulty": true, "share_scope": true, "group_id": true}
	update := map[string]interface{}{}
	for k, v := range req {
		if allowed[k] {
			update[k] = v
		}
	}

	if err := h.resourceService.Update(c.Request.Context(), id, uid, update); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	// 验证资源所有权
	res, err := h.resourceService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "resource not found")
		return
	}

	// 检查是否是资源所有者
	if res.UserID != uid {
		response.Forbidden(c, "permission denied: you can only delete your own resources")
		return
	}

	if err := h.resourceService.Delete(c.Request.Context(), id, uid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *ResourceHandler) ToggleFavorite(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	favorited, err := h.resourceService.ToggleFavorite(c.Request.Context(), id, uid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"is_favorite": favorited})
}

func (h *ResourceHandler) ListFavorites(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resources, total, err := h.resourceService.ListFavorites(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  resources,
		"total": total,
	})
}

func (h *ResourceHandler) GetTags(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	tags, err := h.resourceService.GetTags(c.Request.Context(), uid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, tags)
}

func (h *ResourceHandler) GetStats(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	stats, err := h.resourceService.GetStats(c.Request.Context(), uid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, stats)
}
