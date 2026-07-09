package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	wxpkg "wuxie-api/pkg/wx"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CronService struct {
	userRepo    *repository.UserRepo
	checkinRepo *repository.CheckinRepo
	rankRepo    *repository.RankRepo
	planRepo    *repository.TrainingRepo
	notifRepo   *repository.NotificationRepo
	wxClient    *wxpkg.Client
	cfg         *config.Config
}

func NewCronService(
	userRepo *repository.UserRepo,
	checkinRepo *repository.CheckinRepo,
	rankRepo *repository.RankRepo,
	planRepo *repository.TrainingRepo,
	notifRepo *repository.NotificationRepo,
	wxClient *wxpkg.Client,
	cfg *config.Config,
) *CronService {
	return &CronService{
		userRepo:    userRepo,
		checkinRepo: checkinRepo,
		rankRepo:    rankRepo,
		planRepo:    planRepo,
		notifRepo:   notifRepo,
		wxClient:    wxClient,
		cfg:         cfg,
	}
}

func (s *CronService) RefreshAllRanks(ctx context.Context) {
	log.Println("[cron] refreshing all ranks...")

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		s.refreshDayRank(ctx)
	}()
	go func() {
		defer wg.Done()
		s.refreshWeekRank(ctx)
	}()
	go func() {
		defer wg.Done()
		s.refreshAllRank(ctx)
	}()

	wg.Wait()
	log.Println("[cron] rank refresh done")
}

func (s *CronService) SendTrainingReminders(ctx context.Context) {
	log.Println("[cron] sending training reminders...")
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	filter := bson.M{
		"status":     model.PlanStatusActive,
		"start_date": bson.M{"$lte": todayEnd},
		"end_date":   bson.M{"$gte": todayStart},
	}

	cursor, err := s.planRepo.FindActive(ctx, filter)
	if err != nil {
		log.Printf("[cron] find active plans failed: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var plans []*model.TrainingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		log.Printf("[cron] decode plans failed: %v", err)
		return
	}

	for _, plan := range plans {
		dayIndex := int(now.Sub(plan.StartDate).Hours() / 24)
		if dayIndex < 0 || dayIndex >= len(plan.Days) {
			continue
		}

		day := plan.Days[dayIndex]
		hasIncomplete := false
		for _, task := range day.Tasks {
			if task.Status == model.TaskStatusPending {
				hasIncomplete = true
				break
			}
		}

		if !hasIncomplete {
			continue
		}

		notif := &model.Notification{
			UserID:     plan.UserID,
			Type:       model.NotifTypePlanRemind,
			Title:      "训练提醒",
			Content:    fmt.Sprintf("「%s」第%d天有未完成的训练任务", plan.Title, dayIndex+1),
			TargetType: "plan",
			TargetID:   plan.ID,
			IsRead:     false,
		}
		if err := s.notifRepo.Create(ctx, notif); err != nil {
			log.Printf("[cron] create notification failed for plan %s: %v", plan.ID.Hex(), err)
		}

		user, err := s.userRepo.FindByID(ctx, plan.UserID)
		if err != nil || user == nil {
			continue
		}

		if s.wxClient != nil && s.cfg.WX.TemplateID != "" && user.OpenID != "" {
			_ = s.wxClient.SendSubscribeMessage(
				user.OpenID,
				s.cfg.WX.TemplateID,
				"/pages/training/today",
				map[string]string{
					"thing1": plan.Title,
					"thing2": fmt.Sprintf("第%d天训练任务", dayIndex+1),
					"time3":  now.Format("2006-01-02 15:04"),
				},
			)
		}
	}

	log.Println("[cron] training reminders done")
}

func (s *CronService) refreshAllRank(ctx context.Context) {
	users, err := s.userRepo.FindTopByScore(ctx, 100)
	if err != nil {
		log.Printf("[cron] refresh all rank failed: %v", err)
		return
	}

	entries := make([]*model.RankEntry, len(users))
	for i, u := range users {
		entries[i] = &model.RankEntry{
			UserID: u.ID,
			Score:  u.Score,
			Rank:   i + 1,
		}
	}

	if err := s.rankRepo.RefreshRank(ctx, model.RankPeriodAll, entries); err != nil {
		log.Printf("[cron] save all rank failed: %v", err)
	}
}

func (s *CronService) refreshWeekRank(ctx context.Context) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == 0 {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-1), 0, 0, 0, 0, now.Location())

	type userScore struct {
		UserID primitive.ObjectID `bson:"user_id"`
		Score  int                `bson:"score"`
	}

	filter := bson.M{
		"status":     model.CheckinStatusDone,
		"created_at": bson.M{"$gte": weekStart},
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":   "$user_id",
			"score": bson.M{"$sum": "$score"},
		}},
		{"$sort": bson.M{"score": -1}},
		{"$limit": 100},
	}

	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("[cron] refresh week rank failed: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var results []userScore
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("[cron] decode week rank failed: %v", err)
		return
	}

	entries := make([]*model.RankEntry, len(results))
	for i, r := range results {
		entries[i] = &model.RankEntry{
			UserID: r.UserID,
			Score:  r.Score,
			Rank:   i + 1,
		}
	}

	if err := s.rankRepo.RefreshRank(ctx, model.RankPeriodWeek, entries); err != nil {
		log.Printf("[cron] save week rank failed: %v", err)
	}
}

func (s *CronService) refreshDayRank(ctx context.Context) {
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	type userScore struct {
		UserID primitive.ObjectID `bson:"user_id"`
		Score  int                `bson:"score"`
	}

	filter := bson.M{
		"status":     model.CheckinStatusDone,
		"created_at": bson.M{"$gte": todayStart},
	}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":   "$user_id",
			"score": bson.M{"$sum": "$score"},
		}},
		{"$sort": bson.M{"score": -1}},
		{"$limit": 100},
	}

	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("[cron] refresh day rank failed: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var results []userScore
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("[cron] decode day rank failed: %v", err)
		return
	}

	entries := make([]*model.RankEntry, len(results))
	for i, r := range results {
		entries[i] = &model.RankEntry{
			UserID: r.UserID,
			Score:  r.Score,
			Rank:   i + 1,
		}
	}

	if err := s.rankRepo.RefreshRank(ctx, model.RankPeriodDay, entries); err != nil {
		log.Printf("[cron] save day rank failed: %v", err)
	}
}
