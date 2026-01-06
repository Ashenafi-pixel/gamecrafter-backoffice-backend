package game_import

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

// GameImportCronjobService handles scheduled game imports
type GameImportCronjobService interface {
	StartScheduler(ctx context.Context) error
	StopScheduler()
	IsRunning() bool
}

type gameImportCronjobServiceImpl struct {
	gameImportService *GameImportService
	systemConfig      *system_config.SystemConfig
	cron              *cron.Cron
	isRunning         bool
	mu                sync.RWMutex
	logger            *zap.Logger
}

// NewGameImportCronjobService creates a new game import cronjob service
func NewGameImportCronjobService(
	gameImportService *GameImportService,
	systemConfig *system_config.SystemConfig,
	logger *zap.Logger,
) GameImportCronjobService {
	return &gameImportCronjobServiceImpl{
		gameImportService: gameImportService,
		systemConfig:      systemConfig,
		logger:            logger,
	}
}

// StartScheduler starts the base scheduler that runs every 5 days
func (s *gameImportCronjobServiceImpl) StartScheduler(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		s.logger.Warn("Game import scheduler is already running")
		return nil
	}

	// Create cron with seconds support
	s.cron = cron.New(cron.WithSeconds())

	// Base scheduler runs every 5 days (@every 120h)
	_, err := s.cron.AddFunc("@every 120h", func() {
		s.checkAndRunImports(ctx)
	})
	if err != nil {
		s.logger.Error("Failed to add base scheduler cron job", zap.Error(err))
		return err
	}

	s.cron.Start()
	s.isRunning = true

	s.logger.Info("Game import scheduler started successfully", zap.String("schedule", "@every 120h"))

	// Run initial check
	go s.checkAndRunImports(ctx)

	return nil
}

// StopScheduler stops the scheduler
func (s *gameImportCronjobServiceImpl) StopScheduler() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
	}
	s.isRunning = false
	s.logger.Info("Game import scheduler stopped")
}

// IsRunning returns whether the scheduler is running
func (s *gameImportCronjobServiceImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// checkAndRunImports checks all active brands and runs imports if due
func (s *gameImportCronjobServiceImpl) checkAndRunImports(ctx context.Context) {
	s.logger.Info("Checking for due game imports")

	// Get all active game import configs
	configs, err := s.systemConfig.GetAllActiveGameImportConfigs(ctx)
	if err != nil {
		s.logger.Error("Failed to get active game import configs", zap.Error(err))
		return
	}

	if len(configs) == 0 {
		s.logger.Info("No active game import configs found")
		return
	}

	s.logger.Info("Found active game import configs", zap.Int("count", len(configs)))

	// Process all brands in parallel using WaitGroup
	var wg sync.WaitGroup
	for _, config := range configs {
		wg.Add(1)
		go func(brandID uuid.UUID) {
			defer wg.Done()
			s.processBrandImport(ctx, brandID)
		}(config.BrandID)
	}

	// Wait for all imports to complete
	wg.Wait()

	s.logger.Info("Game import check completed", zap.Int("brands_processed", len(configs)))
}

// processBrandImport processes import for a single brand
func (s *gameImportCronjobServiceImpl) processBrandImport(ctx context.Context, brandID uuid.UUID) {
	s.logger.Info("Processing game import for brand", zap.String("brand_id", brandID.String()))

	// Get config for this brand
	config, err := s.systemConfig.GetGameImportConfig(ctx, brandID)
	if err != nil {
		s.logger.Error("Failed to get game import config", zap.String("brand_id", brandID.String()), zap.Error(err))
		return
	}

	if !config.IsActive {
		s.logger.Debug("Game import is not active for brand", zap.String("brand_id", brandID.String()))
		return
	}

	now := time.Now()

	// Check if it's time to check (based on check_frequency_minutes)
	checkFreq := 15 // default
	if config.CheckFrequencyMinutes != nil {
		checkFreq = *config.CheckFrequencyMinutes
	}

	shouldCheck := false
	if config.LastCheckAt == nil {
		shouldCheck = true
	} else {
		nextCheckTime := config.LastCheckAt.Add(time.Duration(checkFreq) * time.Minute)
		if now.After(nextCheckTime) || now.Equal(nextCheckTime) {
			shouldCheck = true
		}
	}

	if !shouldCheck {
		s.logger.Debug("Not time to check yet for brand", zap.String("brand_id", brandID.String()))
		return
	}

	// Update last_check_at BEFORE checking if import is due
	// This prevents multiple checks in quick succession
	if err := s.systemConfig.UpdateGameImportLastCheck(ctx, brandID, now); err != nil {
		s.logger.Warn("Failed to update last_check_at", zap.String("brand_id", brandID.String()), zap.Error(err))
	}

	// Check if import is due
	isDue := false
	if config.NextRunAt == nil {
		// First run - always due
		isDue = true
		s.logger.Info("First import run for brand - importing immediately", zap.String("brand_id", brandID.String()))
	} else if now.After(*config.NextRunAt) || now.Equal(*config.NextRunAt) {
		isDue = true
	}

	if !isDue {
		s.logger.Debug("Import not due yet for brand", zap.String("brand_id", brandID.String()), zap.Time("next_run_at", *config.NextRunAt))
		return
	}

	// Run import in goroutine (parallel execution)
	go func() {
		s.logger.Info("Starting game import for brand", zap.String("brand_id", brandID.String()))
		result, err := s.gameImportService.ImportGames(ctx, brandID)
		if err != nil {
			s.logger.Error("Game import failed", zap.String("brand_id", brandID.String()), zap.Error(err))
			return
		}

		s.logger.Info("Game import completed",
			zap.String("brand_id", brandID.String()),
			zap.Int("games_imported", result.GamesImported),
			zap.Int("house_edges_imported", result.HouseEdgesImported),
			zap.Int("duplicates_skipped", result.DuplicatesSkipped))
	}()
}
