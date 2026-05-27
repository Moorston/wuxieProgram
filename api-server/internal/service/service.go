package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	"wuxie-api/pkg/jwt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	userRepo *repository.UserRepo
	jwtMgr   *jwt.JWTManager
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepo, jwtMgr *jwt.JWTManager, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, jwtMgr: jwtMgr, cfg: cfg}
}

func (s *AuthService) WXLogin(ctx context.Context, code string, nickname, avatar string, gender int) (string, *model.User, error) {
	// 调用微信接口获取openid
	openid, err := s.getOpenID(code)
	if err != nil {
		return "", nil, fmt.Errorf("get openid failed: %w", err)
	}

	// 查找或创建用户
	user, err := s.userRepo.FindByOpenID(ctx, openid)
	if err != nil {
		// 新用户
		user = &model.User{
			OpenID:   openid,
			Nickname: nickname,
			Avatar:   avatar,
			Gender:   gender,
			Score:    0,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return "", nil, fmt.Errorf("create user failed: %w", err)
		}
	}

	// 生成JWT
	token, err := s.jwtMgr.Generate(user.ID.Hex())
	if err != nil {
		return "", nil, fmt.Errorf("generate token failed: %w", err)
	}

	return token, user, nil
}

func (s *AuthService) getOpenID(code string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.cfg.WX.AppID, s.cfg.WX.Secret, code)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OpenID string `json:"openid"`
		ErrCode int   `json:"errcode"`
		ErrMsg string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wx api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return result.OpenID, nil
}

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService(userRepo *repository.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, nickname, avatar string) error {
	update := map[string]interface{}{}
	if nickname != "" {
		update["nickname"] = nickname
	}
	if avatar != "" {
		update["avatar"] = avatar
	}
	return s.userRepo.Update(ctx, userID, update)
}

type CheckinService struct {
	checkinRepo *repository.CheckinRepo
	userRepo    *repository.UserRepo
	mediaURL    string
}

func NewCheckinService(checkinRepo *repository.CheckinRepo, userRepo *repository.UserRepo, mediaURL string) *CheckinService {
	return &CheckinService{checkinRepo: checkinRepo, userRepo: userRepo, mediaURL: mediaURL}
}

func (s *CheckinService) Prepare(ctx context.Context, userID primitive.ObjectID, description string) (*model.Checkin, error) {
	checkin := &model.Checkin{
		UserID:      userID,
		Description: description,
		Status:      model.CheckinStatusPending,
		Score:       10, // 默认打卡积分
	}
	if err := s.checkinRepo.Create(ctx, checkin); err != nil {
		return nil, err
	}
	return checkin, nil
}

func (s *CheckinService) Callback(ctx context.Context, checkinID primitive.ObjectID, videoURL, coverURL string, duration float64) error {
	return s.checkinRepo.UpdateStatus(ctx, checkinID, model.CheckinStatusDone, videoURL, coverURL, duration)
}

func (s *CheckinService) GetList(ctx context.Context, userID primitive.ObjectID, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	var groupUserIDs []primitive.ObjectID
	if groupID != nil {
		users, err := s.userRepo.FindByGroupID(ctx, *groupID)
		if err != nil {
			return nil, 0, err
		}
		for _, u := range users {
			groupUserIDs = append(groupUserIDs, u.ID)
		}
		if len(groupUserIDs) == 0 {
			return nil, 0, nil
		}
	}

	checkins, total, err := s.checkinRepo.List(ctx, userID, groupUserIDs, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 填充用户信息和点赞状态
	userIDs := make([]primitive.ObjectID, 0, len(checkins))
	for _, c := range checkins {
		userIDs = append(userIDs, c.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, c := range checkins {
		c.User = userMap[c.UserID]
	}

	return checkins, total, nil
}

func (s *CheckinService) GetMine(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	return s.checkinRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *CheckinService) Delete(ctx context.Context, checkinID, userID primitive.ObjectID) error {
	return s.checkinRepo.Delete(ctx, checkinID, userID)
}

func (s *CheckinService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
	checkin, err := s.checkinRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, checkin.UserID)
	if err == nil {
		checkin.User = user
	}

	return checkin, nil
}

func (s *CheckinService) Search(ctx context.Context, userID primitive.ObjectID, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
	checkins, total, err := s.checkinRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(checkins))
	for _, c := range checkins {
		userIDs = append(userIDs, c.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, c := range checkins {
		c.User = userMap[c.UserID]
	}

	return checkins, total, nil
}

type SocialService struct {
	commentRepo  *repository.CommentRepo
	likeRepo     *repository.LikeRepo
	checkinRepo  *repository.CheckinRepo
	userRepo     *repository.UserRepo
	notifService *NotificationService
}

func NewSocialService(commentRepo *repository.CommentRepo, likeRepo *repository.LikeRepo, checkinRepo *repository.CheckinRepo, userRepo *repository.UserRepo, notifService *NotificationService) *SocialService {
	return &SocialService{commentRepo: commentRepo, likeRepo: likeRepo, checkinRepo: checkinRepo, userRepo: userRepo, notifService: notifService}
}

func (s *SocialService) ToggleLike(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	liked, err := s.likeRepo.Toggle(ctx, checkinID, userID)
	if err != nil {
		return false, err
	}

	delta := -1
	if liked {
		delta = 1
	}
	if err := s.checkinRepo.IncrLikeCount(ctx, checkinID, delta); err != nil {
		return false, err
	}

	if liked && s.notifService != nil {
		checkin, err := s.checkinRepo.FindByID(ctx, checkinID)
		if err == nil && checkin.UserID != userID {
			sender, _ := s.userRepo.FindByID(ctx, userID)
			senderName := "有人"
			if sender != nil {
				senderName = sender.Nickname
			}
			s.notifService.Send(ctx, checkin.UserID, userID, model.NotifTypeLike,
				senderName+" 赞了你的打卡",
				"", "checkin", checkinID)
		}
	}

	return liked, nil
}

func (s *SocialService) AddComment(ctx context.Context, checkinID, userID primitive.ObjectID, content string) (*model.Comment, error) {
	comment := &model.Comment{
		CheckinID: checkinID,
		UserID:    userID,
		Content:   content,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	if err := s.checkinRepo.IncrCommentCount(ctx, checkinID); err != nil {
		return nil, err
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	comment.User = user

	if s.notifService != nil {
		checkin, err := s.checkinRepo.FindByID(ctx, checkinID)
		if err == nil && checkin.UserID != userID {
			senderName := "有人"
			if user != nil {
				senderName = user.Nickname
			}
			s.notifService.Send(ctx, checkin.UserID, userID, model.NotifTypeComment,
				senderName+" 评论了你的打卡",
				content, "checkin", checkinID)
		}
	}

	return comment, nil
}

func (s *SocialService) GetComments(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error) {
	comments, total, err := s.commentRepo.ListByCheckin(ctx, checkinID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, c := range comments {
		c.User = userMap[c.UserID]
	}

	return comments, total, nil
}

func (s *SocialService) BatchIsLiked(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	return s.likeRepo.BatchIsLiked(ctx, checkinIDs, userID)
}

type RankService struct {
	rankRepo *repository.RankRepo
}

func NewRankService(rankRepo *repository.RankRepo) *RankService {
	return &RankService{rankRepo: rankRepo}
}

func (s *RankService) GetRankList(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
	return s.rankRepo.GetRankList(ctx, period, page, pageSize)
}

func (s *RankService) GetUserRank(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
	return s.rankRepo.GetUserRank(ctx, userID, period)
}

type GroupService struct {
	groupRepo *repository.GroupRepo
	userRepo  *repository.UserRepo
}

func NewGroupService(groupRepo *repository.GroupRepo, userRepo *repository.UserRepo) *GroupService {
	return &GroupService{groupRepo: groupRepo, userRepo: userRepo}
}

func (s *GroupService) List(ctx context.Context) ([]*model.Group, error) {
	groups, err := s.groupRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, g := range groups {
		users, _ := s.userRepo.FindByIDs(ctx, g.MemberIDs)
		g.Members = users
		if len(users) > 0 {
			for _, u := range users {
				if u.ID == g.LeaderID {
					g.Leader = u
					break
				}
			}
		}
	}

	return groups, nil
}

func (s *GroupService) GetDetail(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	users, _ := s.userRepo.FindByIDs(ctx, group.MemberIDs)
	group.Members = users
	for _, u := range users {
		if u.ID == group.LeaderID {
			group.Leader = u
			break
		}
	}

	return group, nil
}

// MediaClient 用于调用 media-server
type MediaClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMediaClient(mediaURL string) *MediaClient {
	return &MediaClient{
		baseURL: mediaURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MediaClient) NotifyTranscodeDone(checkinID, videoURL, coverURL string, duration float64) error {
	payload := map[string]interface{}{
		"checkin_id": checkinID,
		"video_url":  videoURL,
		"cover_url":  coverURL,
		"duration":   duration,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/internal/transcode/done", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("media server returned %d", resp.StatusCode)
	}
	return nil
}
