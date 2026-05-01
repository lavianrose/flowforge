package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/lavianrose/flowforge/internal/execution"
	"github.com/lavianrose/flowforge/internal/repository"
)

type Scheduler struct {
	scheduleRepo *repository.ScheduleRepository
	workflowRepo *repository.WorkflowRepository
	runRepo      *repository.RunRepository
	engine       *execution.Engine
	ticker       *time.Ticker
	done         chan bool
}

func NewScheduler(
	scheduleRepo *repository.ScheduleRepository,
	workflowRepo *repository.WorkflowRepository,
	runRepo *repository.RunRepository,
) *Scheduler {
	engine := execution.NewEngine(runRepo, workflowRepo)

	return &Scheduler{
		scheduleRepo: scheduleRepo,
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		engine:       engine,
		ticker:       time.NewTicker(1 * time.Minute),
		done:         make(chan bool),
	}
}

func (s *Scheduler) Start() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.checkAndRunSchedules()
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
	s.done <- true
}

func (s *Scheduler) checkAndRunSchedules() {
	ctx := context.Background()

	// Get due schedules
	schedules, err := s.scheduleRepo.GetDueSchedules(ctx)
	if err != nil {
		log.Printf("Error getting due schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		// Execute workflow
		_, err := s.engine.Execute(ctx, schedule.WorkflowID, schedule.TenantID, "cron", nil)
		if err != nil {
			log.Printf("Error executing scheduled workflow %s: %v", schedule.WorkflowID, err)
			continue
		}

		// Update last run
		now := time.Now()
		s.scheduleRepo.UpdateLastRun(ctx, schedule.ID, now)

		// Calculate next run time (simple implementation - add 1 hour for now)
		// TODO: Use proper cron library to calculate next run
		nextRun := now.Add(time.Hour)
		s.scheduleRepo.UpdateNextRun(ctx, schedule.ID, nextRun)

		log.Printf("Executed scheduled workflow %s, next run at %v", schedule.WorkflowID, nextRun)
	}
}
