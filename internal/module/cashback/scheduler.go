package cashback

import (
	"context"
	"time"

	"github.com/tucanbit/internal/storage/cashback"
	"go.uber.org/zap"
)

// RakebackScheduler handles automatic activation/de activation of scheduled rakeback events
type RakebackScheduler struct {
	storage cashback.CashbackStorage
	logger  *zap.Logger
	ticker  *time.Ticker
	done    chan bool
}

// NewRakebackScheduler creates a new rakeback scheduler
func NewRakebackScheduler(storage cashback.CashbackStorage, logger *zap.Logger) *RakebackScheduler {
	return &RakebackScheduler{
		storage: storage,
		logger:  logger,
		ticker:  time.NewTicker(1 * time.Minute),
		done:    make(chan bool),
	}
}

// Start begins the scheduler loop
func (s *RakebackScheduler) Start(ctx context.Context) {
	s.logger.Info("Starting rakeback scheduler - checking every 1 minute")
	
	// Run once immediately on start
	go s.ProcessSchedules(ctx)
	
	// Then run every minute
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.ProcessSchedules(ctx)
			case <-s.done:
				s.logger.Info("Rakeback scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("Rakeback scheduler context cancelled")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *RakebackScheduler) Stop() {
	s.ticker.Stop()
	s.done <- true
}

// ProcessSchedules checks and processes schedules that need to be activated or deactivated
func (s *RakebackScheduler) ProcessSchedules(ctx context.Context) {
	s.logger.Debug("Processing rakeback schedules")
	
	// Activate schedules that should start
	s.activateSchedules(ctx)
	
	// Deactivate schedules that should end
	s.deactivateSchedules(ctx)
}

// activateSchedules finds and activates schedules that should be running now
func (s *RakebackScheduler) activateSchedules(ctx context.Context) {
	schedules, err := s.storage.GetSchedulesToActivate(ctx)
	if err != nil {
		s.logger.Error("Failed to get schedules to activate", zap.Error(err))
		return
	}
	
	if len(schedules) == 0 {
		return
	}
	
	s.logger.Info("Found schedules to activate", zap.Int("count", len(schedules)))
	
	for _, schedule := range schedules {
		err := s.storage.ActivateSchedule(ctx, schedule.ID)
		if err != nil {
			s.logger.Error("Failed to activate schedule",
				zap.String("schedule_id", schedule.ID.String()),
				zap.String("schedule_name", schedule.Name),
				zap.Error(err))
			continue
		}
		
		s.logger.Info("Activated rakeback schedule",
			zap.String("schedule_id", schedule.ID.String()),
			zap.String("schedule_name", schedule.Name),
			zap.String("percentage", schedule.Percentage.String()),
			zap.String("scope_type", schedule.ScopeType),
			zap.Time("start_time", schedule.StartTime),
			zap.Time("end_time", schedule.EndTime))
	}
}

// deactivateSchedules finds and deactivates schedules that have ended
func (s *RakebackScheduler) deactivateSchedules(ctx context.Context) {
	schedules, err := s.storage.GetSchedulesToDeactivate(ctx)
	if err != nil {
		s.logger.Error("Failed to get schedules to deactivate", zap.Error(err))
		return
	}
	
	if len(schedules) == 0 {
		return
	}
	
	s.logger.Info("Found schedules to deactivate", zap.Int("count", len(schedules)))
	
	for _, schedule := range schedules {
		err := s.storage.DeactivateSchedule(ctx, schedule.ID)
		if err != nil {
			s.logger.Error("Failed to deactivate schedule",
				zap.String("schedule_id", schedule.ID.String()),
				zap.String("schedule_name", schedule.Name),
				zap.Error(err))
			continue
		}
		
		s.logger.Info("Deactivated rakeback schedule",
			zap.String("schedule_id", schedule.ID.String()),
			zap.String("schedule_name", schedule.Name),
			zap.String("percentage", schedule.Percentage.String()),
			zap.Time("end_time", schedule.EndTime))
	}
}

