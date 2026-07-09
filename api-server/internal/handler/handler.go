package handler

import (
	"log"
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginReq struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int    `json:"gender"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	token, user, err := h.authService.WXLogin(c.Request.Context(), req.Code, req.Nickname, req.Avatar, req.Gender)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), oid)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}

type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), oid, req.Nickname, req.Avatar); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

type CheckinHandler struct {
	checkinService *service.CheckinService
	socialService  *service.SocialService
}

func NewCheckinHandler(checkinService *service.CheckinService, socialService *service.SocialService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService, socialService: socialService}
}

type PrepareReq struct {
	Description string `json:"description"`
}

func (h *CheckinHandler) Prepare(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req PrepareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	checkin, err := h.checkinService.Prepare(c.Request.Context(), oid, req.Description)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, checkin)
}

func (h *CheckinHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkin, err := h.checkinService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "checkin not found")
		return
	}

	likedMap, _ := h.socialService.BatchIsLiked(c.Request.Context(), []primitive.ObjectID{id}, oid)
	checkin.IsLiked = likedMap[id]

	response.Success(c, checkin)
}

func (h *CheckinHandler) GetList(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var groupID *primitive.ObjectID
	if gid := c.Query("group_id"); gid != "" {
		if id, err := primitive.ObjectIDFromHex(gid); err == nil {
			groupID = &id
		}
	}

	checkins, total, err := h.checkinService.GetList(c.Request.Context(), oid, groupID, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	// 填充点赞状态
	checkinIDs := make([]primitive.ObjectID, len(checkins))
	for i, ci := range checkins {
		checkinIDs[i] = ci.ID
	}
	likedMap, _ := h.socialService.BatchIsLiked(c.Request.Context(), checkinIDs, oid)
	for i := range checkins {
		checkins[i].IsLiked = likedMap[checkins[i].ID]
	}

	response.Success(c, gin.H{
		"list":      checkins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CheckinHandler) GetMine(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	checkins, total, err := h.checkinService.GetMine(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  checkins,
		"total": total,
	})
}

func (h *CheckinHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	if err := h.checkinService.Delete(c.Request.Context(), checkinID, oid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *CheckinHandler) Search(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	keyword := c.Query("q")
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	checkins, total, err := h.checkinService.Search(c.Request.Context(), oid, keyword, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	checkinIDs := make([]primitive.ObjectID, len(checkins))
	for i, ci := range checkins {
		checkinIDs[i] = ci.ID
	}
	likedMap, _ := h.socialService.BatchIsLiked(c.Request.Context(), checkinIDs, oid)
	for i := range checkins {
		checkins[i].IsLiked = likedMap[checkins[i].ID]
	}

	response.Success(c, gin.H{
		"list":      checkins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type SocialHandler struct {
	socialService *service.SocialService
}

func NewSocialHandler(socialService *service.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

func (h *SocialHandler) ToggleLike(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	liked, err := h.socialService.ToggleLike(c.Request.Context(), checkinID, oid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"liked": liked})
}

type CommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

func (h *SocialHandler) AddComment(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	var req CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "content is required")
		return
	}

	comment, err := h.socialService.AddComment(c.Request.Context(), checkinID, oid, req.Content)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, comment)
}

func (h *SocialHandler) GetComments(c *gin.Context) {
	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := h.socialService.GetComments(c.Request.Context(), checkinID, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  comments,
		"total": total,
	})
}

type RankHandler struct {
	rankService *service.RankService
}

func NewRankHandler(rankService *service.RankService) *RankHandler {
	return &RankHandler{rankService: rankService}
}

func (h *RankHandler) GetRankList(c *gin.Context) {
	period := model.RankPeriod(c.DefaultQuery("period", "all"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	entries, err := h.rankService.GetRankList(c.Request.Context(), period, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, entries)
}

func (h *RankHandler) GetMyRank(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	period := model.RankPeriod(c.DefaultQuery("period", "all"))

	entry, err := h.rankService.GetUserRank(c.Request.Context(), oid, period)
	if err != nil {
		response.NotFound(c, "rank not found")
		return
	}

	response.Success(c, entry)
}

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

// 回调接口：media-server 转码完成后调用
type TranscodeCallbackReq struct {
	CheckinID string  `json:"checkin_id" binding:"required"`
	VideoURL  string  `json:"video_url" binding:"required"`
	CoverURL  string  `json:"cover_url" binding:"required"`
	Duration  float64 `json:"duration"`
}

func (h *CheckinHandler) TranscodeCallback(c *gin.Context) {
	var req TranscodeCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(req.CheckinID)
	if err != nil {
		response.BadRequest(c, "invalid checkin_id")
		return
	}

	if err := h.checkinService.Callback(c.Request.Context(), checkinID, req.VideoURL, req.CoverURL, req.Duration); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}
