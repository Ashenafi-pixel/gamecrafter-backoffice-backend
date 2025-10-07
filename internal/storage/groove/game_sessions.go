package groove

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
)

type GameSessionStorage interface {
	CreateGameSession(ctx context.Context, userID uuid.UUID, gameID, deviceType, gameMode string) (*dto.GameSession, error)
	GetGameSessionBySessionID(ctx context.Context, sessionID string) (*dto.GameSession, error)
	GetGameSessionByUserID(ctx context.Context, userID uuid.UUID, gameID string) (*dto.GameSession, error)
	UpdateGameSessionURL(ctx context.Context, sessionID, grooveURL string) error
	DeactivateGameSession(ctx context.Context, sessionID string) error
	CleanupExpiredSessions(ctx context.Context) (int, error)
}

type GameSessionStorageImpl struct {
	db *persistencedb.PersistenceDB
}

func NewGameSessionStorage(db *persistencedb.PersistenceDB) GameSessionStorage {
	return &GameSessionStorageImpl{db: db}
}

func (s *GameSessionStorageImpl) CreateGameSession(ctx context.Context, userID uuid.UUID, gameID, deviceType, gameMode string) (*dto.GameSession, error) {
	// First, get the user's test account status
	var isTestAccount bool
	err := s.db.GetPool().QueryRow(ctx,
		"SELECT is_test_account FROM users WHERE id = $1", userID).Scan(&isTestAccount)
	if err != nil {
		// Default to true (test account) if we can't fetch the status
		isTestAccount = true
	}

	query := `
		INSERT INTO game_sessions (user_id, game_id, device_type, game_mode, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, session_id, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval, created_at, expires_at, is_active, last_activity
	`

	var session dto.GameSession
	var homeURL sql.NullString
	var exitURL sql.NullString
	var historyURL sql.NullString

	err = s.db.GetPool().QueryRow(ctx, query,
		userID, gameID, deviceType, gameMode,
		"https://tucanbit.tv",         // home_url
		"https://tucanbit.tv",         // exit_url
		"https://tucanbit.tv/history", // history_url
		"Curacao",                     // license_type
		isTestAccount,                 // is_test_account from database
		0,                             // reality_check_elapsed
		60,                            // reality_check_interval
	).Scan(
		&session.ID,
		&session.SessionID,
		&homeURL,
		&exitURL,
		&historyURL,
		&session.LicenseType,
		&session.IsTestAccount,
		&session.RealityCheckElapsed,
		&session.RealityCheckInterval,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsActive,
		&session.LastActivity,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	// Set nullable fields
	if homeURL.Valid {
		session.HomeURL = homeURL.String
	}
	if exitURL.Valid {
		session.ExitURL = exitURL.String
	}
	if historyURL.Valid {
		session.HistoryURL = historyURL.String
	}
	// grooveURL will be set later when we update the session with the actual GrooveTech URL

	session.UserID = userID.String()
	session.GameID = gameID
	session.DeviceType = deviceType
	session.GameMode = gameMode

	return &session, nil
}

func (s *GameSessionStorageImpl) GetGameSessionBySessionID(ctx context.Context, sessionID string) (*dto.GameSession, error) {
	query := `
		SELECT id, user_id, session_id, game_id, device_type, game_mode, groove_url, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval, created_at, expires_at, is_active, last_activity
		FROM game_sessions
		WHERE session_id = $1 AND is_active = true
	`

	var session dto.GameSession
	var userID uuid.UUID
	var grooveURL sql.NullString
	var homeURL sql.NullString
	var exitURL sql.NullString
	var historyURL sql.NullString

	err := s.db.GetPool().QueryRow(ctx, query, sessionID).Scan(
		&session.ID,
		&userID,
		&session.SessionID,
		&session.GameID,
		&session.DeviceType,
		&session.GameMode,
		&grooveURL,
		&homeURL,
		&exitURL,
		&historyURL,
		&session.LicenseType,
		&session.IsTestAccount,
		&session.RealityCheckElapsed,
		&session.RealityCheckInterval,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsActive,
		&session.LastActivity,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game session not found")
		}
		return nil, fmt.Errorf("failed to get game session: %w", err)
	}

	// Set nullable fields
	if homeURL.Valid {
		session.HomeURL = homeURL.String
	}
	if exitURL.Valid {
		session.ExitURL = exitURL.String
	}
	if historyURL.Valid {
		session.HistoryURL = historyURL.String
	}
	// grooveURL will be set later when we update the session with the actual GrooveTech URL

	session.UserID = userID.String()
	return &session, nil
}

func (s *GameSessionStorageImpl) GetGameSessionByUserID(ctx context.Context, userID uuid.UUID, gameID string) (*dto.GameSession, error) {
	query := `
		SELECT id, user_id, session_id, game_id, device_type, game_mode, groove_url, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval, created_at, expires_at, is_active, last_activity
		FROM game_sessions
		WHERE user_id = $1 AND game_id = $2 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session dto.GameSession
	var dbUserID uuid.UUID
	var grooveURL sql.NullString
	var homeURL sql.NullString
	var exitURL sql.NullString
	var historyURL sql.NullString

	err := s.db.GetPool().QueryRow(ctx, query, userID, gameID).Scan(
		&session.ID,
		&dbUserID,
		&session.SessionID,
		&session.GameID,
		&session.DeviceType,
		&session.GameMode,
		&grooveURL,
		&homeURL,
		&exitURL,
		&historyURL,
		&session.LicenseType,
		&session.IsTestAccount,
		&session.RealityCheckElapsed,
		&session.RealityCheckInterval,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsActive,
		&session.LastActivity,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game session not found")
		}
		return nil, fmt.Errorf("failed to get game session: %w", err)
	}

	// Set nullable fields
	if homeURL.Valid {
		session.HomeURL = homeURL.String
	}
	if exitURL.Valid {
		session.ExitURL = exitURL.String
	}
	if historyURL.Valid {
		session.HistoryURL = historyURL.String
	}
	// grooveURL will be set later when we update the session with the actual GrooveTech URL

	session.UserID = dbUserID.String()
	return &session, nil
}

func (s *GameSessionStorageImpl) UpdateGameSessionURL(ctx context.Context, sessionID, grooveURL string) error {
	query := `
		UPDATE game_sessions 
		SET groove_url = $1, last_activity = NOW()
		WHERE session_id = $2 AND is_active = true
	`

	result, err := s.db.GetPool().Exec(ctx, query, grooveURL, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update game session URL: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("game session not found or already inactive")
	}

	return nil
}

func (s *GameSessionStorageImpl) DeactivateGameSession(ctx context.Context, sessionID string) error {
	query := `
		UPDATE game_sessions 
		SET is_active = false, last_activity = NOW()
		WHERE session_id = $1
	`

	result, err := s.db.GetPool().Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate game session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("game session not found")
	}

	return nil
}

func (s *GameSessionStorageImpl) CleanupExpiredSessions(ctx context.Context) (int, error) {
	query := `
		UPDATE game_sessions 
		SET is_active = false 
		WHERE expires_at < NOW() AND is_active = true
	`

	result, err := s.db.GetPool().Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return int(result.RowsAffected()), nil
}
