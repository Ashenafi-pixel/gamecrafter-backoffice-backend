package operator

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type operatorStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func NewOperatorStorage(db *persistencedb.PersistenceDB, log *zap.Logger) *operatorStorage {
	return &operatorStorage{
		db:  db,
		log: log,
	}
}

// CreateOperator inserts a new operator row.
func (s *operatorStorage) CreateOperator(ctx context.Context, req dto.CreateOperatorReq) (dto.Operator, error) {
	var op dto.Operator
	var allowedDomainsJSON any
	if len(req.AllowedEmbedDomains) > 0 {
		// Marshal slice to JSON so Postgres jsonb gets valid syntax.
		data, err := json.Marshal(req.AllowedEmbedDomains)
		if err != nil {
			return dto.Operator{}, errors.ErrUnableTocreate.Wrap(err, "unable to marshal allowed_embed_domains")
		}
		allowedDomainsJSON = string(data)
	} else {
		allowedDomainsJSON = nil
	}

	query := `
		INSERT INTO operators (
			operator_id, name, code, domain, logo_url, is_active,
			allowed_embed_domains, embed_referer_required, transaction_url
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6,
		        $7::jsonb, $8, NULLIF($9, ''))
		RETURNING operator_id, name, code,
		          COALESCE(domain, ''), COALESCE(logo_url, ''), is_active,
		          COALESCE(allowed_embed_domains, '[]'::jsonb), embed_referer_required,
		          COALESCE(transaction_url, ''), created_at, updated_at
	`

	row := s.db.GetPool().QueryRow(ctx, query,
		req.OperatorID,
		req.Name,
		req.Code,
		req.Domain,
		req.LogoURL,
		req.IsActive,
		allowedDomainsJSON,
		req.EmbedRefererRequired,
		req.TransactionURL,
	)

	var allowedDomains []string
	if err := row.Scan(
		&op.OperatorID,
		&op.Name,
		&op.Code,
		&op.Domain,
		&op.LogoURL,
		&op.IsActive,
		&allowedDomains,
		&op.EmbedRefererRequired,
		&op.TransactionURL,
		&op.CreatedAt,
		&op.UpdatedAt,
	); err != nil {
		s.log.Error("unable to create operator", zap.Error(err), zap.Int32("operator_id", req.OperatorID))
		return dto.Operator{}, errors.ErrUnableTocreate.Wrap(err, "unable to create operator")
	}

	op.AllowedEmbedDomains = allowedDomains
	return op, nil
}

// GetOperatorByID fetches an operator by operator_id.
func (s *operatorStorage) GetOperatorByID(ctx context.Context, operatorID int32) (dto.Operator, bool, error) {
	query := `
		SELECT operator_id, name, code,
		       COALESCE(domain, ''), COALESCE(logo_url, ''), is_active,
		       COALESCE(allowed_embed_domains, '[]'::jsonb), embed_referer_required,
		       COALESCE(transaction_url, ''), created_at, updated_at
		FROM operators
		WHERE operator_id = $1
	`

	row := s.db.GetPool().QueryRow(ctx, query, operatorID)
	var op dto.Operator
	var allowedDomains []string

	err := row.Scan(
		&op.OperatorID,
		&op.Name,
		&op.Code,
		&op.Domain,
		&op.LogoURL,
		&op.IsActive,
		&allowedDomains,
		&op.EmbedRefererRequired,
		&op.TransactionURL,
		&op.CreatedAt,
		&op.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return dto.Operator{}, false, nil
		}
		s.log.Error("unable to get operator by id", zap.Error(err), zap.Int32("operator_id", operatorID))
		return dto.Operator{}, false, errors.ErrReadError.Wrap(err, "unable to get operator by id")
	}

	op.AllowedEmbedDomains = allowedDomains
	return op, true, nil
}

// GetOperators returns a paginated list of operators.
func (s *operatorStorage) GetOperators(ctx context.Context, req dto.GetOperatorsReq) (dto.GetOperatorsRes, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	offset := (req.Page - 1) * req.PerPage

	baseQuery := `
		FROM operators
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR code ILIKE '%' || $1 || '%')
		  AND ($2::bool IS NULL OR is_active = $2)
	`

	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int
	if err := s.db.GetPool().QueryRow(ctx, countQuery, req.Search, req.IsActive).Scan(&total); err != nil {
		s.log.Error("unable to count operators", zap.Error(err))
		return dto.GetOperatorsRes{}, errors.ErrReadError.Wrap(err, "unable to count operators")
	}

	listQuery := `
		SELECT operator_id, name, code,
		       COALESCE(domain, ''), COALESCE(logo_url, ''), is_active,
		       COALESCE(allowed_embed_domains, '[]'::jsonb), embed_referer_required,
		       COALESCE(transaction_url, ''), created_at, updated_at
	` + baseQuery + `
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.GetPool().Query(ctx, listQuery, req.Search, req.IsActive, req.PerPage, offset)
	if err != nil {
		s.log.Error("unable to list operators", zap.Error(err))
		return dto.GetOperatorsRes{}, errors.ErrReadError.Wrap(err, "unable to list operators")
	}
	defer rows.Close()

	var operators []dto.Operator
	for rows.Next() {
		var op dto.Operator
		var allowedDomains []string
		if err := rows.Scan(
			&op.OperatorID,
			&op.Name,
			&op.Code,
			&op.Domain,
			&op.LogoURL,
			&op.IsActive,
			&allowedDomains,
			&op.EmbedRefererRequired,
			&op.TransactionURL,
			&op.CreatedAt,
			&op.UpdatedAt,
		); err != nil {
			s.log.Error("unable to scan operator", zap.Error(err))
			return dto.GetOperatorsRes{}, errors.ErrReadError.Wrap(err, "unable to scan operator")
		}
		op.AllowedEmbedDomains = allowedDomains
		operators = append(operators, op)
	}

	return dto.GetOperatorsRes{
		Operators: operators,
		Total:     total,
		Page:      req.Page,
		PerPage:   req.PerPage,
	}, nil
}

// UpdateOperator updates mutable fields of an operator.
func (s *operatorStorage) UpdateOperator(ctx context.Context, req dto.UpdateOperatorReq) (dto.Operator, error) {
	current, exists, err := s.GetOperatorByID(ctx, req.OperatorID)
	if err != nil {
		return dto.Operator{}, err
	}
	if !exists {
		return dto.Operator{}, errors.ErrResourceNotFound.New("operator not found")
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	code := current.Code
	if req.Code != nil {
		code = *req.Code
	}
	domain := current.Domain
	if req.Domain != nil {
		domain = *req.Domain
	}
	logoURL := current.LogoURL
	if req.LogoURL != nil {
		logoURL = *req.LogoURL
	}
	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	embedRefRequired := current.EmbedRefererRequired
	if req.EmbedRefererRequired != nil {
		embedRefRequired = *req.EmbedRefererRequired
	}
	transactionURL := current.TransactionURL
	if req.TransactionURL != nil {
		transactionURL = *req.TransactionURL
	}
	allowedDomains := current.AllowedEmbedDomains
	if req.AllowedEmbedDomains != nil {
		allowedDomains = req.AllowedEmbedDomains
	}

	var allowedDomainsJSON any
	if len(allowedDomains) > 0 {
		data, err := json.Marshal(allowedDomains)
		if err != nil {
			return dto.Operator{}, errors.ErrUnableToUpdate.Wrap(err, "unable to marshal allowed_embed_domains")
		}
		allowedDomainsJSON = string(data)
	} else {
		allowedDomainsJSON = nil
	}

	query := `
		UPDATE operators
		SET name = $2,
		    code = $3,
		    domain = NULLIF($4, ''),
		    logo_url = NULLIF($5, ''),
		    is_active = $6,
		    allowed_embed_domains = $7::jsonb,
		    embed_referer_required = $8,
		    transaction_url = NULLIF($9, ''),
		    updated_at = now()
		WHERE operator_id = $1
		RETURNING operator_id, name, code,
		          COALESCE(domain, ''), COALESCE(logo_url, ''), is_active,
		          COALESCE(allowed_embed_domains, '[]'::jsonb), embed_referer_required,
		          COALESCE(transaction_url, ''), created_at, updated_at
	`

	row := s.db.GetPool().QueryRow(ctx, query,
		req.OperatorID,
		name,
		code,
		domain,
		logoURL,
		isActive,
		allowedDomainsJSON,
		embedRefRequired,
		transactionURL,
	)

	var op dto.Operator
	var allowedDomainsOut []string
	if err := row.Scan(
		&op.OperatorID,
		&op.Name,
		&op.Code,
		&op.Domain,
		&op.LogoURL,
		&op.IsActive,
		&allowedDomainsOut,
		&op.EmbedRefererRequired,
		&op.TransactionURL,
		&op.CreatedAt,
		&op.UpdatedAt,
	); err != nil {
		s.log.Error("unable to update operator", zap.Error(err), zap.Int32("operator_id", req.OperatorID))
		return dto.Operator{}, errors.ErrUnableToUpdate.Wrap(err, "unable to update operator")
	}

	op.AllowedEmbedDomains = allowedDomainsOut
	return op, nil
}

// DeleteOperator deletes an operator row.
func (s *operatorStorage) DeleteOperator(ctx context.Context, operatorID int32) error {
	_, err := s.db.GetPool().Exec(ctx, `DELETE FROM operators WHERE operator_id = $1`, operatorID)
	if err != nil {
		s.log.Error("unable to delete operator", zap.Error(err), zap.Int32("operator_id", operatorID))
		return errors.ErrDBDelError.Wrap(err, "unable to delete operator")
	}
	return nil
}

// UpdateOperatorStatus updates only the is_active flag.
func (s *operatorStorage) UpdateOperatorStatus(ctx context.Context, operatorID int32, isActive bool) error {
	_, err := s.db.GetPool().Exec(ctx, `
		UPDATE operators
		SET is_active = $2, updated_at = now()
		WHERE operator_id = $1
	`, operatorID, isActive)
	if err != nil {
		s.log.Error("unable to update operator status", zap.Error(err), zap.Int32("operator_id", operatorID))
		return errors.ErrUnableToUpdate.Wrap(err, "unable to update operator status")
	}
	return nil
}

// AssignGamesToOperator creates operator_games rows for the given game IDs.
func (s *operatorStorage) AssignGamesToOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if len(gameIDs) == 0 {
		return nil
	}

	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		return errors.ErrUnableTocreate.Wrap(err, "unable to start transaction for assigning games to operator")
	}
	defer tx.Rollback(ctx)

	for _, gid := range gameIDs {
		if gid == "" {
			continue
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO operator_games (operator_id, game_id)
			VALUES ($1, $2::uuid)
			ON CONFLICT (operator_id, game_id) DO NOTHING
		`, operatorID, gid)
		if err != nil {
			s.log.Error("unable to assign game to operator", zap.Error(err),
				zap.Int32("operator_id", operatorID), zap.String("game_id", gid))
			return errors.ErrUnableTocreate.Wrap(err, "unable to assign game to operator")
		}
	}

	return tx.Commit(ctx)
}

// RevokeGamesFromOperator removes operator_games rows for the given game IDs.
func (s *operatorStorage) RevokeGamesFromOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if len(gameIDs) == 0 {
		return nil
	}

	_, err := s.db.GetPool().Exec(ctx, `
		DELETE FROM operator_games
		WHERE operator_id = $1 AND game_id = ANY($2::uuid[])
	`, operatorID, gameIDs)
	if err != nil {
		s.log.Error("unable to revoke games from operator", zap.Error(err),
			zap.Int32("operator_id", operatorID))
		return errors.ErrDBDelError.Wrap(err, "unable to revoke games from operator")
	}

	return nil
}

// AssignAllGamesToOperator assigns all existing games to the operator.
func (s *operatorStorage) AssignAllGamesToOperator(ctx context.Context, operatorID int32) error {
	// Insert all game IDs from games table into operator_games, ignoring duplicates.
	query := `
		INSERT INTO operator_games (operator_id, game_id, created_at)
		SELECT $1, g.id, NOW()
		FROM games g
		ON CONFLICT (operator_id, game_id) DO NOTHING
	`
	if _, err := s.db.GetPool().Exec(ctx, query, operatorID); err != nil {
		s.log.Error("unable to assign all games to operator", zap.Error(err), zap.Int32("operator_id", operatorID))
		return errors.ErrUnableTocreate.Wrap(err, "unable to assign all games to operator")
	}
	return nil
}

// AssignProviderToOperator creates an operator_providers row.
func (s *operatorStorage) AssignProviderToOperator(ctx context.Context, operatorID int32, providerID string) error {
	_, err := s.db.GetPool().Exec(ctx, `
		INSERT INTO operator_providers (operator_id, provider_id)
		VALUES ($1, $2::uuid)
		ON CONFLICT (operator_id, provider_id) DO NOTHING
	`, operatorID, providerID)
	if err != nil {
		s.log.Error("unable to assign provider to operator", zap.Error(err),
			zap.Int32("operator_id", operatorID), zap.String("provider_id", providerID))
		return errors.ErrUnableTocreate.Wrap(err, "unable to assign provider to operator")
	}
	return nil
}

// RevokeProviderFromOperator removes an operator_providers row.
func (s *operatorStorage) RevokeProviderFromOperator(ctx context.Context, operatorID int32, providerID string) error {
	_, err := s.db.GetPool().Exec(ctx, `
		DELETE FROM operator_providers
		WHERE operator_id = $1 AND provider_id = $2::uuid
	`, operatorID, providerID)
	if err != nil {
		s.log.Error("unable to revoke provider from operator", zap.Error(err),
			zap.Int32("operator_id", operatorID), zap.String("provider_id", providerID))
		return errors.ErrDBDelError.Wrap(err, "unable to revoke provider from operator")
	}
	return nil
}

// GetOperatorGameIDs returns game IDs directly assigned to an operator via operator_games.
func (s *operatorStorage) GetOperatorGameIDs(ctx context.Context, operatorID int32) ([]string, error) {
	query := `
		SELECT game_id::text
		FROM operator_games
		WHERE operator_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.GetPool().Query(ctx, query, operatorID)
	if err != nil {
		s.log.Error("unable to get operator game ids", zap.Error(err),
			zap.Int32("operator_id", operatorID))
		return nil, errors.ErrUnableToGet.Wrap(err, "unable to get operator game ids")
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.ErrUnableToGet.Wrap(err, "unable to scan operator game id")
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.ErrUnableToGet.Wrap(err, "error reading operator game ids")
	}
	return ids, nil
}

// GetOperatorGames returns full game info for games directly assigned to an operator.
func (s *operatorStorage) GetOperatorGames(ctx context.Context, operatorID int32) ([]dto.GameResponse, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.status,
			g.timestamp,
			g.photo,
			g.price,
			g.enabled,
			g.game_id,
			g.internal_name,
			g.integration_partner,
			g.provider,
			g.created_at,
			g.updated_at
		FROM operator_games og
		JOIN games g ON g.id = og.game_id
		WHERE og.operator_id = $1
		ORDER BY og.created_at DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, operatorID)
	if err != nil {
		s.log.Error("unable to get operator games", zap.Error(err),
			zap.Int32("operator_id", operatorID))
		return nil, errors.ErrUnableToGet.Wrap(err, "unable to get operator games")
	}
	defer rows.Close()

	var games []dto.GameResponse
	for rows.Next() {
		var g dto.GameResponse
		// dto.GameResponse fields
		var (
			id                 uuid.UUID
			name               string
			status             string
			timestamp          time.Time
			photo              *string
			price              *string
			enabled            bool
			gameID             *string
			internalName       *string
			integrationPartner *string
			provider           *string
			createdAt          time.Time
			updatedAt          time.Time
		)

		if err := rows.Scan(
			&id,
			&name,
			&status,
			&timestamp,
			&photo,
			&price,
			&enabled,
			&gameID,
			&internalName,
			&integrationPartner,
			&provider,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, errors.ErrUnableToGet.Wrap(err, "unable to scan operator game")
		}

		g = dto.GameResponse{
			ID:                 id,
			Name:               name,
			Status:             status,
			Timestamp:          timestamp,
			Photo:              photo,
			Price:              price,
			Enabled:            enabled,
			GameID:             gameID,
			InternalName:       internalName,
			IntegrationPartner: integrationPartner,
			Provider:           provider,
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.ErrUnableToGet.Wrap(err, "error reading operator games")
	}
	if games == nil {
		games = []dto.GameResponse{}
	}
	return games, nil
}

// AddOperatorAllowedOrigin adds an allowed origin for operator embeds.
func (s *operatorStorage) AddOperatorAllowedOrigin(ctx context.Context, operatorID int32, origin string) (dto.OperatorAllowedOriginRes, error) {
	query := `
		INSERT INTO operator_allowed_origins (operator_id, origin, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (operator_id, origin) DO UPDATE SET origin = EXCLUDED.origin
		RETURNING id, operator_id, origin, created_at
	`

	var res dto.OperatorAllowedOriginRes
	if err := s.db.GetPool().QueryRow(ctx, query, operatorID, origin).Scan(
		&res.ID,
		&res.OperatorID,
		&res.Origin,
		&res.CreatedAt,
	); err != nil {
		s.log.Error("unable to add operator allowed origin", zap.Error(err),
			zap.Int32("operator_id", operatorID),
			zap.String("origin", origin),
		)
		return dto.OperatorAllowedOriginRes{}, errors.ErrUnableTocreate.Wrap(err, "unable to add operator allowed origin")
	}

	return res, nil
}

// RemoveOperatorAllowedOrigin removes an allowed origin by ID.
func (s *operatorStorage) RemoveOperatorAllowedOrigin(ctx context.Context, operatorID int32, originID int32) error {
	commandTag, err := s.db.GetPool().Exec(ctx, `
		DELETE FROM operator_allowed_origins
		WHERE operator_id = $1 AND id = $2
	`, operatorID, originID)
	if err != nil {
		s.log.Error("unable to remove operator allowed origin", zap.Error(err),
			zap.Int32("operator_id", operatorID),
			zap.Int32("origin_id", originID),
		)
		return errors.ErrDBDelError.Wrap(err, "unable to remove operator allowed origin")
	}
	if commandTag.RowsAffected() == 0 {
		return errors.ErrResourceNotFound.New("operator allowed origin not found")
	}
	return nil
}

// ListOperatorAllowedOrigins lists allowed origins for an operator.
func (s *operatorStorage) ListOperatorAllowedOrigins(ctx context.Context, operatorID int32) ([]dto.OperatorAllowedOriginRes, error) {
	rows, err := s.db.GetPool().Query(ctx, `
		SELECT id, operator_id, origin, created_at
		FROM operator_allowed_origins
		WHERE operator_id = $1
		ORDER BY created_at DESC
	`, operatorID)
	if err != nil {
		s.log.Error("unable to list operator allowed origins", zap.Error(err), zap.Int32("operator_id", operatorID))
		return nil, errors.ErrReadError.Wrap(err, "unable to list operator allowed origins")
	}
	defer rows.Close()

	var origins []dto.OperatorAllowedOriginRes
	for rows.Next() {
		var origin dto.OperatorAllowedOriginRes
		if err := rows.Scan(&origin.ID, &origin.OperatorID, &origin.Origin, &origin.CreatedAt); err != nil {
			s.log.Error("unable to scan operator allowed origin", zap.Error(err))
			return nil, errors.ErrReadError.Wrap(err, "unable to scan operator allowed origin")
		}
		origins = append(origins, origin)
	}
	return origins, nil
}

// GetOperatorFeatureFlags returns operator feature flags (jsonb) as a map.
func (s *operatorStorage) GetOperatorFeatureFlags(ctx context.Context, operatorID int32) (map[string]bool, error) {
	var flagsJSON []byte
	query := `
		SELECT COALESCE(flags, '{}'::jsonb)
		FROM operator_feature_flags
		WHERE operator_id = $1
	`

	err := s.db.GetPool().QueryRow(ctx, query, operatorID).Scan(&flagsJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return map[string]bool{}, nil
		}
		s.log.Error("unable to get operator feature flags", zap.Error(err), zap.Int32("operator_id", operatorID))
		return nil, errors.ErrReadError.Wrap(err, "unable to get operator feature flags")
	}

	var flags map[string]bool
	if len(flagsJSON) == 0 {
		return map[string]bool{}, nil
	}
	if err := json.Unmarshal(flagsJSON, &flags); err != nil {
		s.log.Error("unable to unmarshal operator feature flags", zap.Error(err), zap.Int32("operator_id", operatorID))
		return nil, errors.ErrReadError.Wrap(err, "unable to decode operator feature flags")
	}
	if flags == nil {
		flags = map[string]bool{}
	}
	return flags, nil
}

// UpdateOperatorFeatureFlags upserts operator feature flags (jsonb).
func (s *operatorStorage) UpdateOperatorFeatureFlags(ctx context.Context, operatorID int32, flags map[string]bool) error {
	flagsBytes, err := json.Marshal(flags)
	if err != nil {
		return errors.ErrUnableToUpdate.Wrap(err, "unable to encode operator feature flags")
	}

	_, err = s.db.GetPool().Exec(ctx, `
		INSERT INTO operator_feature_flags (operator_id, flags, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
		ON CONFLICT (operator_id)
		DO UPDATE SET flags = EXCLUDED.flags, updated_at = NOW()
	`, operatorID, string(flagsBytes))
	if err != nil {
		s.log.Error("unable to update operator feature flags", zap.Error(err), zap.Int32("operator_id", operatorID))
		return errors.ErrUnableToUpdate.Wrap(err, "unable to update operator feature flags")
	}
	return nil
}

// generateSecureSecret creates a base64-like random string of given length.
func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// CreateOperatorCredential creates API credentials (api_key + signing_key) for an operator.
func (s *operatorStorage) CreateOperatorCredential(ctx context.Context, operatorID int32) (dto.OperatorCredentialRes, error) {
	apiKey, err := generateSecureSecret(32)
	if err != nil {
		return dto.OperatorCredentialRes{}, errors.ErrUnableTocreate.Wrap(err, "unable to generate api_key")
	}
	signingKey, err := generateSecureSecret(64)
	if err != nil {
		return dto.OperatorCredentialRes{}, errors.ErrUnableTocreate.Wrap(err, "unable to generate signing_key")
	}

	query := `
		INSERT INTO operator_credentials (
			operator_id, api_key, signing_key, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id, operator_id, api_key, signing_key, is_active, created_at, updated_at
	`

	var res dto.OperatorCredentialRes
	if err := s.db.GetPool().QueryRow(ctx, query, operatorID, apiKey, signingKey).Scan(
		&res.ID,
		&res.OperatorID,
		&res.APIKey,
		&res.SigningKey,
		&res.IsActive,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		s.log.Error("unable to create operator credential", zap.Error(err), zap.Int32("operator_id", operatorID))
		return dto.OperatorCredentialRes{}, errors.ErrUnableTocreate.Wrap(err, "unable to create operator credential")
	}

	return res, nil
}

// RotateOperatorCredential rotates signing_key for a given credential.
func (s *operatorStorage) RotateOperatorCredential(ctx context.Context, operatorID int32, credentialID int32) (dto.RotateOperatorCredentialRes, error) {
	newSigningKey, err := generateSecureSecret(64)
	if err != nil {
		return dto.RotateOperatorCredentialRes{}, errors.ErrUnableToUpdate.Wrap(err, "unable to generate signing_key")
	}

	query := `
		UPDATE operator_credentials
		SET signing_key = $3, updated_at = NOW()
		WHERE id = $2 AND operator_id = $1
		RETURNING api_key, signing_key, updated_at
	`

	var res dto.RotateOperatorCredentialRes
	if err := s.db.GetPool().QueryRow(ctx, query, operatorID, credentialID, newSigningKey).Scan(
		&res.APIKey,
		&res.SigningKey,
		&res.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return dto.RotateOperatorCredentialRes{}, errors.ErrResourceNotFound.New("operator credential not found")
		}
		s.log.Error("unable to rotate operator credential", zap.Error(err), zap.Int32("credential_id", credentialID), zap.Int32("operator_id", operatorID))
		return dto.RotateOperatorCredentialRes{}, errors.ErrUnableToUpdate.Wrap(err, "unable to rotate operator credential")
	}

	return res, nil
}

// GetActiveSigningKeyByOperatorID returns signing_key for any active credential of the operator.
func (s *operatorStorage) GetActiveSigningKeyByOperatorID(ctx context.Context, operatorID int32) (string, error) {
	var signingKey string
	query := `
		SELECT signing_key
		FROM operator_credentials
		WHERE operator_id = $1 AND is_active = true
		ORDER BY updated_at DESC
		LIMIT 1
	`
	err := s.db.GetPool().QueryRow(ctx, query, operatorID).Scan(&signingKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", errors.ErrResourceNotFound.New("no active operator credentials found")
		}
		s.log.Error("unable to get active signing key for operator", zap.Error(err), zap.Int32("operator_id", operatorID))
		return "", errors.ErrReadError.Wrap(err, "unable to get active signing key for operator")
	}
	return signingKey, nil
}

// Ensure interface compliance at compile time.
var _ storage.Operator = (*operatorStorage)(nil)

