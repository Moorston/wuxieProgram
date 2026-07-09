package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingService struct {
	planRepo     *repository.TrainingRepo
	templateRepo *repository.TemplateRepo
	notifService *NotificationService
}

func NewTrainingService(planRepo *repository.TrainingRepo, templateRepo *repository.TemplateRepo, notifService *NotificationService) *TrainingService {
	return &TrainingService{planRepo: planRepo, templateRepo: templateRepo, notifService: notifService}
}

func (s *TrainingService) CreatePlan(ctx context.Context, userID primitive.ObjectID, plan *model.TrainingPlan) error {
	plan.UserID = userID
	plan.Status = model.PlanStatusActive

	totalTasks := 0
	for _, day := range plan.Days {
		totalTasks += len(day.Tasks)
	}
	plan.Stats = model.PlanStats{
		TotalTasks:    totalTasks,
		Completed:     0,
		CompletionRate: 0,
	}

	return s.planRepo.Create(ctx, plan)
}

func (s *TrainingService) GetPlan(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
	return s.planRepo.FindByID(ctx, id)
}

func (s *TrainingService) UpdatePlan(ctx context.Context, id, userID primitive.ObjectID, update map[string]interface{}) error {
	return s.planRepo.Update(ctx, id, update)
}

func (s *TrainingService) DeletePlan(ctx context.Context, id, userID primitive.ObjectID) error {
	return s.planRepo.Delete(ctx, id, userID)
}

func (s *TrainingService) ListPlans(ctx context.Context, userID primitive.ObjectID, status *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error) {
	return s.planRepo.ListByUser(ctx, userID, status, page, pageSize)
}

func (s *TrainingService) GetTodayTasks(ctx context.Context, userID primitive.ObjectID) ([]map[string]interface{}, error) {
	plans, err := s.planRepo.GetTodayTasks(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	var result []map[string]interface{}

	for _, plan := range plans {
		dayIndex := int(todayDate.Sub(plan.StartDate).Hours() / 24)
		if dayIndex < 0 || dayIndex >= len(plan.Days) {
			continue
		}

		day := plan.Days[dayIndex]
		result = append(result, map[string]interface{}{
			"plan_id":   plan.ID,
			"plan_title": plan.Title,
			"day":       day,
		})
	}

	return result, nil
}

func (s *TrainingService) UpdateTaskStatus(ctx context.Context, planID primitive.ObjectID, dayIndex, taskIndex int, status model.TaskStatus, checkinID *primitive.ObjectID) error {
	plan, err := s.planRepo.FindByID(ctx, planID)
	if err != nil {
		return err
	}

	if dayIndex < 0 || dayIndex >= len(plan.Days) {
		return fmt.Errorf("day index %d out of range [0, %d)", dayIndex, len(plan.Days))
	}

	day := &plan.Days[dayIndex]
	if taskIndex < 0 || taskIndex >= len(day.Tasks) {
		return fmt.Errorf("task index %d out of range [0, %d)", taskIndex, len(day.Tasks))
	}

	day.Tasks[taskIndex].Status = status
	if checkinID != nil {
		day.Tasks[taskIndex].CheckinID = *checkinID
	}

	totalTasks := 0
	completed := 0
	for _, d := range plan.Days {
		for _, t := range d.Tasks {
			totalTasks++
			if t.Status == model.TaskStatusDone {
				completed++
			}
		}
	}

	var completionRate float64
	if totalTasks > 0 {
		completionRate = float64(completed) / float64(totalTasks) * 100
	}

	plan.Stats = model.PlanStats{
		TotalTasks:     totalTasks,
		Completed:      completed,
		CompletionRate: completionRate,
	}

	if err := s.planRepo.UpdateTasks(ctx, planID, plan.Days, plan.Stats); err != nil {
		return err
	}

	if completed == totalTasks && s.notifService != nil {
		s.notifService.Send(ctx, plan.UserID, plan.UserID, model.NotifTypePlanComplete,
			"训练计划已完成",
			"恭喜！你已完成「"+plan.Title+"」全部训练任务",
			"plan", planID)
	}

	return nil
}

func (s *TrainingService) GetReport(ctx context.Context, planID primitive.ObjectID) (map[string]interface{}, error) {
	plan, err := s.planRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	totalTasks := 0
	completed := 0
	skipped := 0
	typeStats := map[string]int{
		"basic":  0,
		"taolu":  0,
		"sanda":  0,
		"qigong": 0,
	}
	typeCompleted := map[string]int{
		"basic":  0,
		"taolu":  0,
		"sanda":  0,
		"qigong": 0,
	}

	for _, day := range plan.Days {
		for _, task := range day.Tasks {
			totalTasks++
			typeStats[string(task.Type)]++
			if task.Status == model.TaskStatusDone {
				completed++
				typeCompleted[string(task.Type)]++
			} else if task.Status == model.TaskStatusSkipped {
				skipped++
			}
		}
	}

	daysPassed := int(time.Since(plan.StartDate).Hours() / 24)
	if daysPassed < 0 {
		daysPassed = 0
	}
	totalDays := int(plan.EndDate.Sub(plan.StartDate).Hours() / 24)

	return map[string]interface{}{
		"plan":            plan,
		"total_tasks":     totalTasks,
		"completed":       completed,
		"skipped":         skipped,
		"completion_rate": plan.Stats.CompletionRate,
		"days_passed":     daysPassed,
		"total_days":      totalDays,
		"type_stats":      typeStats,
		"type_completed":  typeCompleted,
	}, nil
}

func (s *TrainingService) ApplyTemplate(ctx context.Context, userID primitive.ObjectID, templateID primitive.ObjectID, startDate time.Time) (*model.TrainingPlan, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if err := s.templateRepo.IncrUsageCount(ctx, templateID); err != nil {
		log.Printf("[WARN] incr template usage count failed: %v", err)
	}

	endDate := startDate.AddDate(0, 0, template.DurationDays)

	days := make([]model.TrainingDay, len(template.Days))
	for i, tDay := range template.Days {
		dayDate := startDate.AddDate(0, 0, i)
		tasks := make([]model.TrainingTask, len(tDay.Tasks))
		for j, tTask := range tDay.Tasks {
			tasks[j] = model.TrainingTask{
				Title:    tTask.Title,
				Type:     tTask.Type,
				Duration: tTask.Duration,
				Reps:     tTask.Reps,
				Note:     tTask.Note,
				Status:   model.TaskStatusPending,
			}
		}
		days[i] = model.TrainingDay{
			Day:   tDay.Day,
			Date:  dayDate,
			Tasks: tasks,
		}
	}

	totalTasks := 0
	for _, day := range days {
		totalTasks += len(day.Tasks)
	}

	plan := &model.TrainingPlan{
		UserID:      userID,
		Title:       template.Name,
		Description: template.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.PlanStatusActive,
		Days:        days,
		Stats: model.PlanStats{
			TotalTasks:    totalTasks,
			Completed:     0,
			CompletionRate: 0,
		},
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *TrainingService) ListTemplates(ctx context.Context, category, style string, page, pageSize int) ([]*model.TrainingTemplate, int64, error) {
	return s.templateRepo.List(ctx, category, style, page, pageSize)
}

func (s *TrainingService) GetTemplate(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error) {
	return s.templateRepo.FindByID(ctx, id)
}
