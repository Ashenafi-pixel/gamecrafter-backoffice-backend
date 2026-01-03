package game_import

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/game"
	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

type GameImportService struct {
	db                 *persistencedb.PersistenceDB
	gameStorage        game.GameStorage
	houseEdgeStorage   game.HouseEdgeStorage
	systemConfig       *system_config.SystemConfig
	defaultDirectusURL string
	logger             *zap.Logger
}

func NewGameImportService(
	db *persistencedb.PersistenceDB,
	gameStorage game.GameStorage,
	houseEdgeStorage game.HouseEdgeStorage,
	systemConfig *system_config.SystemConfig,
	defaultDirectusURL string,
	logger *zap.Logger,
) *GameImportService {
	return &GameImportService{
		db:                 db,
		gameStorage:        gameStorage,
		houseEdgeStorage:   houseEdgeStorage,
		systemConfig:       systemConfig,
		defaultDirectusURL: defaultDirectusURL,
		logger:             logger,
	}
}

// DirectusGameResponse represents the structure of games from Directus
type DirectusGameResponse struct {
	Data struct {
		NewGames []struct {
			Games []struct {
				GameGameID struct {
					GameID         string `json:"game_id"`
					InternalName   string `json:"internal_name"`
					DefaultRTP     string `json:"default_rtp"`
					GrooveCategory struct {
						InternalName string `json:"internal_name"`
					} `json:"groove_category"`
					Provider struct {
						InternalName string `json:"internal_name"`
					} `json:"provider"`
					DefaultImage struct {
						FilenameDisk string `json:"filename_disk"`
					} `json:"default_image"`
				} `json:"game_game_id"`
			} `json:"games"`
		} `json:"new_games"`
	} `json:"data"`
}

// ImportResult represents the result of a game import operation
type ImportResult struct {
	GamesImported      int       `json:"games_imported"`
	HouseEdgesImported int       `json:"house_edges_imported"`
	TotalFetched       int       `json:"total_fetched"`
	Filtered           int       `json:"filtered"`
	DuplicatesSkipped  int       `json:"duplicates_skipped"`
	Errors             []string  `json:"errors,omitempty"`
	ImportTime         time.Time `json:"import_time"`
}

// FetchGamesFromDirectus fetches games from Directus GraphQL API
func (s *GameImportService) FetchGamesFromDirectus(ctx context.Context, directusURL string) (*DirectusGameResponse, error) {
	s.logger.Info("Fetching games from Directus", zap.String("url", directusURL))

	query := `query NewGamesGet {
		new_games (limit: 1) {
			games (limit: 5000) {
				game_game_id (limit: 5000) {
					game_id
					internal_name
					default_rtp
					groove_category {
						internal_name
					}
					provider {
						internal_name
					}
					default_image {
						filename_disk
					}
				}
			}
		}
	}`

	requestBody := map[string]interface{}{
		"query":     strings.ReplaceAll(query, "\n", " "),
		"variables": map[string]interface{}{},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		s.logger.Error("Failed to marshal request body", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", directusURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip TLS verification for Directus API
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch from Directus", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to fetch from Directus")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.logger.Error("Directus API returned error", zap.Int("status", resp.StatusCode), zap.String("body", string(bodyBytes)))
		return nil, errors.ErrInternalServerError.Wrap(nil, fmt.Sprintf("Directus API returned status %d", resp.StatusCode))
	}

	var response DirectusGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error("Failed to decode Directus response", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to decode Directus response")
	}

	s.logger.Info("Successfully fetched games from Directus",
		zap.Int("new_games_count", len(response.Data.NewGames)),
		zap.Int("total_games", s.countGames(&response)))

	return &response, nil
}

// countGames counts the total number of games in the response
func (s *GameImportService) countGames(response *DirectusGameResponse) int {
	count := 0
	for _, newGame := range response.Data.NewGames {
		count += len(newGame.Games)
	}
	return count
}

// FilterGamesByProvider filters games by provider names
func (s *GameImportService) FilterGamesByProvider(games []struct {
	GameGameID struct {
		GameID         string `json:"game_id"`
		InternalName   string `json:"internal_name"`
		DefaultRTP     string `json:"default_rtp"`
		GrooveCategory struct {
			InternalName string `json:"internal_name"`
		} `json:"groove_category"`
		Provider struct {
			InternalName string `json:"internal_name"`
		} `json:"provider"`
		DefaultImage struct {
			FilenameDisk string `json:"filename_disk"`
		} `json:"default_image"`
	} `json:"game_game_id"`
}, providers []string) []struct {
	GameGameID struct {
		GameID         string `json:"game_id"`
		InternalName   string `json:"internal_name"`
		DefaultRTP     string `json:"default_rtp"`
		GrooveCategory struct {
			InternalName string `json:"internal_name"`
		} `json:"groove_category"`
		Provider struct {
			InternalName string `json:"internal_name"`
		} `json:"provider"`
		DefaultImage struct {
			FilenameDisk string `json:"filename_disk"`
		} `json:"default_image"`
	} `json:"game_game_id"`
} {
	if providers == nil || len(providers) == 0 {
		return games
	}

	providerMap := make(map[string]bool)
	for _, provider := range providers {
		providerMap[strings.TrimSpace(provider)] = true
	}

	filtered := []struct {
		GameGameID struct {
			GameID         string `json:"game_id"`
			InternalName   string `json:"internal_name"`
			DefaultRTP     string `json:"default_rtp"`
			GrooveCategory struct {
				InternalName string `json:"internal_name"`
			} `json:"groove_category"`
			Provider struct {
				InternalName string `json:"internal_name"`
			} `json:"provider"`
			DefaultImage struct {
				FilenameDisk string `json:"filename_disk"`
			} `json:"default_image"`
		} `json:"game_game_id"`
	}{}

	for _, game := range games {
		providerName := game.GameGameID.Provider.InternalName
		if providerMap[providerName] {
			filtered = append(filtered, game)
		}
	}

	return filtered
}

// CalculateHouseEdge calculates house edge from RTP: house_edge = 100 - rtp
func (s *GameImportService) CalculateHouseEdge(rtpStr string) (decimal.Decimal, error) {
	rtp, err := strconv.ParseFloat(rtpStr, 64)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid RTP value: %s", rtpStr)
	}

	if rtp < 0 || rtp > 100 {
		return decimal.Zero, fmt.Errorf("RTP value out of range: %f", rtp)
	}

	houseEdge := 100 - rtp
	return decimal.NewFromFloat(houseEdge).Round(2), nil
}

// GetExistingGameIDs retrieves all existing game_id values from the database
func (s *GameImportService) GetExistingGameIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := s.db.GetPool().Query(ctx, "SELECT game_id FROM games WHERE game_id IS NOT NULL")
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to query existing game IDs")
	}
	defer rows.Close()

	existingIDs := make(map[string]bool)
	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, errors.ErrInternalServerError.Wrap(err, "failed to scan game ID")
		}
		existingIDs[gameID] = true
	}

	return existingIDs, nil
}

// ImportGames imports games from Directus into the database
func (s *GameImportService) ImportGames(ctx context.Context, brandID uuid.UUID) (*ImportResult, error) {
	s.logger.Info("Starting game import", zap.String("brand_id", brandID.String()))

	// Get configuration
	config, err := s.systemConfig.GetGameImportConfig(ctx, brandID)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game import config")
	}

	// Determine Directus URL (use config's URL or fallback to default)
	directusURL := s.defaultDirectusURL
	if config.DirectusURL != nil && *config.DirectusURL != "" {
		directusURL = *config.DirectusURL
	}
	if directusURL == "" {
		directusURL = "https://tucanbit-prod.directus.app/graphql"
	}

	// Fetch games from Directus
	directusResponse, err := s.FetchGamesFromDirectus(ctx, directusURL)
	if err != nil {
		return nil, err
	}

	// Extract all games
	allGames := []struct {
		GameGameID struct {
			GameID         string `json:"game_id"`
			InternalName   string `json:"internal_name"`
			DefaultRTP     string `json:"default_rtp"`
			GrooveCategory struct {
				InternalName string `json:"internal_name"`
			} `json:"groove_category"`
			Provider struct {
				InternalName string `json:"internal_name"`
			} `json:"provider"`
			DefaultImage struct {
				FilenameDisk string `json:"filename_disk"`
			} `json:"default_image"`
		} `json:"game_game_id"`
	}{}

	for _, newGame := range directusResponse.Data.NewGames {
		allGames = append(allGames, newGame.Games...)
	}

	s.logger.Info("Fetched games from Directus", zap.Int("count", len(allGames)))

	// Filter by provider
	filteredGames := s.FilterGamesByProvider(allGames, config.Providers)
	s.logger.Info("Filtered games by provider", zap.Int("count", len(filteredGames)))

	// Get existing game IDs
	existingGameIDs, err := s.GetExistingGameIDs(ctx)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Found existing games", zap.Int("count", len(existingGameIDs)))

	// Import games
	result := &ImportResult{
		TotalFetched: len(allGames),
		Filtered:     len(filteredGames),
		ImportTime:   time.Now(),
		Errors:       []string{},
	}

	for _, gameWrapper := range filteredGames {
		gameData := gameWrapper.GameGameID

		// Skip if game_id is missing
		if gameData.GameID == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Game missing game_id: %s", gameData.InternalName))
			continue
		}

		// Skip if already exists
		if existingGameIDs[gameData.GameID] {
			result.DuplicatesSkipped++
			continue
		}

		// Calculate house edge
		houseEdge, err := s.CalculateHouseEdge(gameData.DefaultRTP)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid RTP for game %s: %s", gameData.GameID, gameData.DefaultRTP))
			s.logger.Warn("Invalid RTP", zap.String("game_id", gameData.GameID), zap.String("rtp", gameData.DefaultRTP), zap.Error(err))
			continue
		}

		// Create game
		providerName := gameData.Provider.InternalName
		if providerName == "" {
			providerName = "Unknown"
		}

		gameModel := &game.Game{
			Name:               gameData.InternalName,
			Status:             "ACTIVE",
			Timestamp:          time.Now(),
			Photo:              nil,
			Price:              nil,
			Enabled:            true,
			GameID:             &gameData.GameID,
			InternalName:       &gameData.InternalName,
			IntegrationPartner: stringPtr("GrooveTech"),
			Provider:           &providerName,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		// Insert game (with duplicate check)
		gameIDStr := ""
		if gameModel.GameID != nil {
			gameIDStr = *gameModel.GameID
		}
		internalNameStr := ""
		if gameModel.InternalName != nil {
			internalNameStr = *gameModel.InternalName
		}
		providerStr := ""
		if gameModel.Provider != nil {
			providerStr = *gameModel.Provider
		}
		integrationPartnerStr := ""
		if gameModel.IntegrationPartner != nil {
			integrationPartnerStr = *gameModel.IntegrationPartner
		}

		var exists bool
		err = s.db.GetPool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM games WHERE game_id = $1)`, gameIDStr).Scan(&exists)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check if game exists %s: %v", gameData.GameID, err))
			s.logger.Error("Failed to check if game exists", zap.String("game_id", gameData.GameID), zap.Error(err))
			continue
		}

		if exists {
			result.DuplicatesSkipped++
			continue
		}

		// Insert game
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO games (name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, gameModel.Name, gameModel.Status, gameModel.Timestamp, gameModel.Photo, gameModel.Price, gameModel.Enabled,
			gameIDStr, internalNameStr, integrationPartnerStr, providerStr,
			gameModel.CreatedAt, gameModel.UpdatedAt)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert game %s: %v", gameData.GameID, err))
			s.logger.Error("Failed to insert game", zap.String("game_id", gameData.GameID), zap.Error(err))
			continue
		}

		result.GamesImported++

		// Map groove category to game type
		gameType := s.MapGrooveCategoryToGameType(gameData.GrooveCategory.InternalName)
		if gameType == "" {
			// Default to "slot" if category is unknown
			gameType = "slot"
			s.logger.Warn("Unknown groove category, defaulting to slot",
				zap.String("game_id", gameData.GameID),
				zap.String("category", gameData.GrooveCategory.InternalName))
		}

		// Create house edge
		effectiveUntil := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
		houseEdgeModel := &game.GameHouseEdge{
			GameID:         &gameData.GameID,
			GameType:       gameType,
			GameVariant:    stringPtr("v1"),
			HouseEdge:      houseEdge,
			MinBet:         decimal.NewFromFloat(0.10),
			MaxBet:         decimalPtr(decimal.NewFromFloat(100.00)),
			IsActive:       true,
			EffectiveFrom:  time.Now(),
			EffectiveUntil: &effectiveUntil,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		houseEdgeGameIDStr := ""
		if houseEdgeModel.GameID != nil {
			houseEdgeGameIDStr = *houseEdgeModel.GameID
		}
		houseEdgeGameVariantStr := ""
		if houseEdgeModel.GameVariant != nil {
			houseEdgeGameVariantStr = *houseEdgeModel.GameVariant
		}

		// Check if house edge already exists
		var houseEdgeExists bool
		err = s.db.GetPool().QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM game_house_edges 
				WHERE game_id = $1 AND game_type = $2 AND game_variant = $3
			)
		`, houseEdgeGameIDStr, houseEdgeModel.GameType, houseEdgeGameVariantStr).Scan(&houseEdgeExists)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check if house edge exists for game %s: %v", gameData.GameID, err))
			s.logger.Error("Failed to check if house edge exists", zap.String("game_id", gameData.GameID), zap.Error(err))
			continue
		}

		if houseEdgeExists {
			// Skip if already exists
			continue
		}

		// Insert house edge
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO game_house_edges (game_id, game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, houseEdgeGameIDStr, houseEdgeModel.GameType, houseEdgeGameVariantStr,
			houseEdgeModel.HouseEdge, houseEdgeModel.MinBet, houseEdgeModel.MaxBet,
			houseEdgeModel.IsActive, houseEdgeModel.EffectiveFrom, houseEdgeModel.EffectiveUntil,
			houseEdgeModel.CreatedAt, houseEdgeModel.UpdatedAt)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert house edge for game %s: %v", gameData.GameID, err))
			s.logger.Error("Failed to insert house edge", zap.String("game_id", gameData.GameID), zap.Error(err))
			continue
		}

		result.HouseEdgesImported++
	}

	// Update last run time
	nextRunAt := s.calculateNextRunTime(config)
	err = s.systemConfig.UpdateGameImportLastRun(ctx, brandID, result.ImportTime, nextRunAt)
	if err != nil {
		s.logger.Warn("Failed to update last run time", zap.Error(err))
	}

	s.logger.Info("Game import completed",
		zap.Int("games_imported", result.GamesImported),
		zap.Int("house_edges_imported", result.HouseEdgesImported),
		zap.Int("duplicates_skipped", result.DuplicatesSkipped),
		zap.Int("errors", len(result.Errors)))

	return result, nil
}

// calculates the next run time based on schedule configuration
func (s *GameImportService) calculateNextRunTime(config *system_config.GameImportConfig) *time.Time {
	now := time.Now()
	var nextRun time.Time

	switch config.ScheduleType {
	case "daily":
		nextRun = now.Add(24 * time.Hour)
	case "weekly":
		nextRun = now.Add(7 * 24 * time.Hour)
	case "monthly":
		nextRun = now.AddDate(0, 1, 0)
	case "custom":
		if config.ScheduleCron == nil || *config.ScheduleCron == "" {
			s.logger.Warn("Custom schedule type requires schedule_cron, returning nil")
			return nil
		}

		// Parse cron expression
		// Using standard cron format: "minute hour day month weekday"
		// robfig/cron/v3 supports both standard (5 fields) and extended (6 fields with seconds) formats
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		schedule, err := parser.Parse(*config.ScheduleCron)
		if err != nil {
			s.logger.Error("Failed to parse cron expression",
				zap.String("cron", *config.ScheduleCron),
				zap.Error(err))
			return nil
		}

		// Calculate next run time from now
		nextRun = schedule.Next(now)
		if nextRun.IsZero() {
			s.logger.Warn("Cron schedule returned zero time, returning nil",
				zap.String("cron", *config.ScheduleCron))
			return nil
		}

		s.logger.Info("Calculated next run time from cron expression",
			zap.String("cron", *config.ScheduleCron),
			zap.Time("next_run", nextRun))
	default:
		return nil
	}

	return &nextRun
}

// MapGrooveCategoryToGameType maps Groove category names to normalized game_type values
func (s *GameImportService) MapGrooveCategoryToGameType(categoryName string) string {
	if categoryName == "" {
		return "slot" // Default fallback
	}

	// Normalize category name (case-insensitive matching)
	categoryLower := strings.ToLower(strings.TrimSpace(categoryName))

	// Map Groove categories to game_type values
	switch categoryLower {
	case "slots":
		return "slot"
	case "video bingo & keno", "video bingo", "keno", "bingo":
		return "bingo"
	case "live dealer", "live casino":
		return "live"
	case "crash":
		return "crash"
	case "instant win", "instant win games":
		return "instant_win"
	default:
		// For unknown categories, try to normalize by converting to lowercase and replacing spaces with underscores
		normalized := strings.ToLower(strings.ReplaceAll(categoryName, " ", "_"))
		// Remove special characters
		normalized = strings.ReplaceAll(normalized, "&", "")
		normalized = strings.ReplaceAll(normalized, "'", "")
		normalized = strings.TrimSpace(normalized)
		return normalized
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
