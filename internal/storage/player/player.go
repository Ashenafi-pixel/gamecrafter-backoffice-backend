package player

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type player struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Player {
	return &player{
		db:  db,
		log: log,
	}
}

func (p *player) CreatePlayer(ctx context.Context, playerReq dto.Player) (dto.Player, error) {
	// Hash password if provided
	hashedPassword := playerReq.Password
	if hashedPassword != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(hashedPassword), bcrypt.DefaultCost)
		if err != nil {
			p.log.Error("unable to hash password", zap.Error(err))
			err = errors.ErrUnableTocreate.Wrap(err, "unable to hash password")
			return dto.Player{}, err
		}
		hashedPassword = string(hashedBytes)
	}

	// Map optional user fields
	var phone sql.NullString
	if playerReq.Phone != nil && *playerReq.Phone != "" {
		phone = sql.NullString{String: *playerReq.Phone, Valid: true}
	}
	firstName := sql.NullString{String: "", Valid: false}
	if playerReq.FirstName != nil && *playerReq.FirstName != "" {
		firstName = sql.NullString{String: *playerReq.FirstName, Valid: true}
	}
	lastName := sql.NullString{String: "", Valid: false}
	if playerReq.LastName != nil && *playerReq.LastName != "" {
		lastName = sql.NullString{String: *playerReq.LastName, Valid: true}
	}

	state := ""
	if playerReq.State != nil {
		state = *playerReq.State
	}
	streetAddress := ""
	if playerReq.StreetAddress != nil {
		streetAddress = *playerReq.StreetAddress
	}
	postalCode := ""
	if playerReq.PostalCode != nil {
		postalCode = *playerReq.PostalCode
	}

	// users.brand_id is NOT NULL in migrations, so ensure a fallback default.
	brandID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if playerReq.BrandID != nil {
		brandID = *playerReq.BrandID
	}

	// users.date_of_birth is stored as VARCHAR in migrations (YYYY-MM-DD).
	dateOfBirthStr := ""
	if !playerReq.DateOfBirth.IsZero() {
		dateOfBirthStr = playerReq.DateOfBirth.Format("2006-01-02")
	}

	// NOTE: we keep referral_code simple; the system can refine it later if needed.
	referalCode := playerReq.Username

	query := `
		INSERT INTO users (
			username, phone_number, password, default_currency, email, source, referal_code,
			date_of_birth, created_by, is_admin, first_name, last_name, referal_type, refered_by_code,
			user_type, status, street_address, country, state, city, postal_code, kyc_status, profile,
			brand_id, is_test_account
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23,
			$24, $25
		)
		RETURNING
			id, email, username, phone_number, first_name, last_name,
			default_currency, date_of_birth, country, state, street_address,
			postal_code, is_test_account, brand_id, created_at
	`

	var playerRow playerRow
	err := p.db.GetPool().QueryRow(
		ctx,
		query,
		playerReq.Username,
		phone,
		hashedPassword,
		playerReq.DefaultCurrency,
		playerReq.Email,
		constant.SOURCE_PHONE,
		referalCode,
		dateOfBirthStr,
		nil,              // created_by
		false,            // is_admin
		firstName,
		lastName,
		"",               // referal_type
		"",               // refered_by_code
		"PLAYER",         // user_type
		"ACTIVE",         // status
		streetAddress,    // street_address
		playerReq.Country,
		state,            // state
		"" /* city */,
		postalCode,       // postal_code
		"PENDING",        // kyc_status
		"" /* profile */,
		brandID,
		playerReq.TestAccount,
	).Scan(
		&playerRow.ID,
		&playerRow.Email,
		&playerRow.Username,
		&playerRow.Phone,
		&playerRow.FirstName,
		&playerRow.LastName,
		&playerRow.DefaultCurrency,
		&playerRow.DateOfBirth,
		&playerRow.Country,
		&playerRow.State,
		&playerRow.StreetAddress,
		&playerRow.PostalCode,
		&playerRow.TestAccount,
		&playerRow.BrandID,
		&playerRow.CreatedAt,
	)

	if err != nil {
		p.log.Error("unable to create player", zap.Error(err), zap.Any("player", playerReq))
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key value") {
			err = errors.ErrDataAlredyExist.Wrap(err, errStr)
			return dto.Player{}, err
		}
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create player")
		return dto.Player{}, err
	}

	// Upsert withdrawal limit enablement
	_, _ = p.db.GetPool().Exec(ctx, `
		INSERT INTO user_limits (user_id, limit_type, withdrawal_limit_enabled)
		VALUES ($1, 'withdrawal', $2)
		ON CONFLICT (user_id, limit_type) DO UPDATE
		SET withdrawal_limit_enabled = EXCLUDED.withdrawal_limit_enabled
	`, playerRow.ID, playerReq.EnableWithdrawalLimit)

	// Create an initial balance row (amount defaults to 0)
	_, _ = p.db.GetPool().Exec(ctx, `
		INSERT INTO balances (user_id, currency_code)
		VALUES ($1, $2)
		ON CONFLICT (user_id, currency_code) DO NOTHING
	`, playerRow.ID, playerReq.DefaultCurrency)

	playerRow.EnableWithdrawalLimit = playerReq.EnableWithdrawalLimit
	playerRow.UpdatedAt = playerRow.CreatedAt
	return p.mapToDTO(playerRow), nil
}

func (p *player) GetPlayerByID(ctx context.Context, playerID uuid.UUID) (dto.Player, bool, error) {
	query := `
		SELECT
			u.id,
			u.email,
			u.username,
			u.phone_number,
			u.first_name,
			u.last_name,
			u.default_currency,
			u.date_of_birth,
			u.country,
			u.state,
			u.street_address,
			u.postal_code,
			u.is_test_account,
			COALESCE((
				SELECT ul.withdrawal_limit_enabled
				FROM user_limits ul
				WHERE ul.user_id = u.id AND ul.limit_type = 'withdrawal'
				LIMIT 1
			), false) AS enable_withdrawal_limit,
			u.brand_id,
			u.created_at,
			u.created_at AS updated_at
		FROM users u
		WHERE u.id = $1 AND u.user_type = 'PLAYER' AND u.is_admin = false
	`

	var playerRow playerRow
	err := p.db.GetPool().QueryRow(ctx, query, playerID).Scan(
		&playerRow.ID,
		&playerRow.Email,
		&playerRow.Username,
		&playerRow.Phone,
		&playerRow.FirstName,
		&playerRow.LastName,
		&playerRow.DefaultCurrency,
		&playerRow.DateOfBirth,
		&playerRow.Country,
		&playerRow.State,
		&playerRow.StreetAddress,
		&playerRow.PostalCode,
		&playerRow.TestAccount,
		&playerRow.EnableWithdrawalLimit,
		&playerRow.BrandID,
		&playerRow.CreatedAt,
		&playerRow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Player{}, false, nil
		}
		p.log.Error("unable to get player by ID", zap.Error(err), zap.String("id", playerID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get player by ID")
		return dto.Player{}, false, err
	}

	return p.mapToDTO(playerRow), true, nil
}

func (p *player) UpdatePlayer(ctx context.Context, playerReq dto.Player) (dto.Player, error) {
	if playerReq.ID == uuid.Nil {
		return dto.Player{}, errors.ErrInvalidUserInput.New("invalid player id")
	}

	// Convert optional fields. Some DB columns are NOT NULL, so we always provide fallback empty strings.
	phone := sql.NullString{Valid: false}
	if playerReq.Phone != nil {
		phone = sql.NullString{String: *playerReq.Phone, Valid: true}
	}

	firstName := sql.NullString{Valid: false}
	if playerReq.FirstName != nil {
		firstName = sql.NullString{String: *playerReq.FirstName, Valid: true}
	}

	lastName := sql.NullString{Valid: false}
	if playerReq.LastName != nil {
		lastName = sql.NullString{String: *playerReq.LastName, Valid: true}
	}

	streetAddress := sql.NullString{Valid: false}
	if playerReq.StreetAddress != nil {
		streetAddress = sql.NullString{String: *playerReq.StreetAddress, Valid: true}
	}

	postalCode := sql.NullString{Valid: false}
	if playerReq.PostalCode != nil {
		postalCode = sql.NullString{String: *playerReq.PostalCode, Valid: true}
	}

	brandID := playerReq.BrandID
	if brandID == nil {
		// users.brand_id is NOT NULL; fallback to default brand.
		defaultBrand := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		brandID = &defaultBrand
	}

	dateOfBirthStr := ""
	if !playerReq.DateOfBirth.IsZero() {
		dateOfBirthStr = playerReq.DateOfBirth.Format("2006-01-02")
	}

	query := `
		UPDATE users
		SET
			email = COALESCE($2, email),
			username = COALESCE($3, username),
			phone_number = COALESCE($4, phone_number),
			first_name = COALESCE($5, first_name),
			last_name = COALESCE($6, last_name),
			default_currency = COALESCE($7, default_currency),
			date_of_birth = COALESCE($8, date_of_birth),
			country = COALESCE($9, country),
			state = COALESCE($10, state),
			street_address = COALESCE($11, street_address),
			postal_code = COALESCE($12, postal_code),
			is_test_account = COALESCE($13, is_test_account),
			brand_id = COALESCE($14, brand_id)
		WHERE id = $1
		RETURNING
			id, email, username, phone_number, first_name, last_name,
			default_currency, date_of_birth, country, state,
			street_address, postal_code, is_test_account, brand_id, created_at
	`

	var row playerRow
	err := p.db.GetPool().QueryRow(
		ctx,
		query,
		playerReq.ID,
		sql.NullString{String: playerReq.Email, Valid: playerReq.Email != ""},
		sql.NullString{String: playerReq.Username, Valid: playerReq.Username != ""},
		phone,
		firstName,
		lastName,
		sql.NullString{String: playerReq.DefaultCurrency, Valid: playerReq.DefaultCurrency != ""},
		sql.NullString{String: dateOfBirthStr, Valid: dateOfBirthStr != ""},
		sql.NullString{String: playerReq.Country, Valid: playerReq.Country != ""},
		sql.NullString{String: func() string { if playerReq.State == nil { return "" }; return *playerReq.State }(), Valid: playerReq.State != nil},
		streetAddress,
		postalCode,
		sql.NullBool{Bool: playerReq.TestAccount, Valid: true},
		brandIDToNullUUID(playerReq.BrandID),
	).Scan(
		&row.ID,
		&row.Email,
		&row.Username,
		&row.Phone,
		&row.FirstName,
		&row.LastName,
		&row.DefaultCurrency,
		&row.DateOfBirth,
		&row.Country,
		&row.State,
		&row.StreetAddress,
		&row.PostalCode,
		&row.TestAccount,
		&row.BrandID,
		&row.CreatedAt,
	)
	if err != nil {
		p.log.Error("unable to update player", zap.Error(err), zap.String("id", playerReq.ID.String()))
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") {
			return dto.Player{}, errors.ErrDataAlredyExist.Wrap(err, errStr)
		}
		return dto.Player{}, errors.ErrUnableToUpdate.Wrap(err, "unable to update player")
	}

	// Upsert withdrawal limit enablement
	_, _ = p.db.GetPool().Exec(ctx, `
		INSERT INTO user_limits (user_id, limit_type, withdrawal_limit_enabled)
		VALUES ($1, 'withdrawal', $2)
		ON CONFLICT (user_id, limit_type)
		DO UPDATE SET withdrawal_limit_enabled = EXCLUDED.withdrawal_limit_enabled
	`, row.ID, playerReq.EnableWithdrawalLimit)

	row.EnableWithdrawalLimit = playerReq.EnableWithdrawalLimit
	row.UpdatedAt = row.CreatedAt
	return p.mapToDTO(row), nil
}

func brandIDToNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func (p *player) GetPlayersByIDs(ctx context.Context, playerIDs []uuid.UUID) ([]dto.Player, error) {
	if len(playerIDs) == 0 {
		return []dto.Player{}, nil
	}

	query := `
		SELECT
			u.id,
			u.email,
			u.username,
			u.phone_number,
			u.first_name,
			u.last_name,
			u.default_currency,
			u.date_of_birth,
			u.country,
			u.state,
			u.street_address,
			u.postal_code,
			u.is_test_account,
			COALESCE((
				SELECT ul.withdrawal_limit_enabled
				FROM user_limits ul
				WHERE ul.user_id = u.id AND ul.limit_type = 'withdrawal'
				LIMIT 1
			), false) AS enable_withdrawal_limit,
			u.brand_id,
			u.created_at,
			u.created_at AS updated_at
		FROM users u
		WHERE u.id = ANY($1::uuid[]) AND u.user_type = 'PLAYER' AND u.is_admin = false
		ORDER BY u.created_at DESC
	`

	rows, err := p.db.GetPool().Query(ctx, query, pq.Array(playerIDs))
	if err != nil {
		p.log.Error("unable to get players by IDs", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
		return nil, err
	}
	defer rows.Close()

	var players []dto.Player
	for rows.Next() {
		var row playerRow
		if err := rows.Scan(
			&row.ID,
			&row.Email,
			&row.Username,
			&row.Phone,
			&row.FirstName,
			&row.LastName,
			&row.DefaultCurrency,
			&row.DateOfBirth,
			&row.Country,
			&row.State,
			&row.StreetAddress,
			&row.PostalCode,
			&row.TestAccount,
			&row.EnableWithdrawalLimit,
			&row.BrandID,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			p.log.Error("unable to scan player row", zap.Error(err))
			return nil, errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
		}
		players = append(players, p.mapToDTO(row))
	}

	if err := rows.Err(); err != nil {
		p.log.Error("error iterating player rows", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
	}

	return players, nil
}

func (p *player) GetAllPlayers(ctx context.Context) ([]dto.Player, error) {
	query := `
		SELECT
			u.id,
			u.email,
			u.username,
			u.phone_number,
			u.first_name,
			u.last_name,
			u.default_currency,
			u.date_of_birth,
			u.country,
			u.state,
			u.street_address,
			u.postal_code,
			u.is_test_account,
			COALESCE((
				SELECT ul.withdrawal_limit_enabled
				FROM user_limits ul
				WHERE ul.user_id = u.id AND ul.limit_type = 'withdrawal'
				LIMIT 1
			), false) AS enable_withdrawal_limit,
			u.brand_id,
			u.created_at,
			u.created_at AS updated_at
		FROM users u
		WHERE u.user_type = 'PLAYER' AND u.is_admin = false
		ORDER BY u.created_at DESC
	`

	rows, err := p.db.GetPool().Query(ctx, query)
	if err != nil {
		p.log.Error("unable to get all players", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get all players")
		return nil, err
	}
	defer rows.Close()

	var players []dto.Player
	for rows.Next() {
		var playerRow playerRow

		err := rows.Scan(
			&playerRow.ID,
			&playerRow.Email,
			&playerRow.Username,
			&playerRow.Phone,
			&playerRow.FirstName,
			&playerRow.LastName,
			&playerRow.DefaultCurrency,
			&playerRow.DateOfBirth,
			&playerRow.Country,
			&playerRow.State,
			&playerRow.StreetAddress,
			&playerRow.PostalCode,
			&playerRow.TestAccount,
			&playerRow.EnableWithdrawalLimit,
			&playerRow.BrandID,
			&playerRow.CreatedAt,
			&playerRow.UpdatedAt,
		)
		if err != nil {
			p.log.Error("unable to scan player row", zap.Error(err))
			err = errors.ErrUnableToGet.Wrap(err, "unable to get all players")
			return nil, err
		}

		players = append(players, p.mapToDTO(playerRow))
	}

	if err := rows.Err(); err != nil {
		p.log.Error("error iterating player rows", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get all players")
		return nil, err
	}

	return players, nil
}

func (p *player) GetPlayers(ctx context.Context, req dto.GetPlayersReqs) (dto.GetPlayersRess, error) {
	offset := (req.Page - 1) * req.PerPage

	searchStr := strings.TrimSpace(req.Search)
	countryStr := ""
	if req.Country != nil {
		countryStr = strings.TrimSpace(*req.Country)
	}

	// brand_id is UUID in users table.
	var brandID *uuid.UUID
	if req.BrandID != nil && strings.TrimSpace(*req.BrandID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.BrandID))
		if err != nil {
			return dto.GetPlayersRess{}, errors.ErrInvalidUserInput.Wrap(err, "invalid brand_id (expected UUID)")
		}
		brandID = &parsed
	}

	var testAccount *bool
	if req.TestAccount != nil {
		testAccount = req.TestAccount
	}

	// Set default sort if not provided
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Validate sort fields to prevent SQL injection
	// NOTE: users table migration doesn't clearly expose updated_at, so we map "updated_at" to created_at.
	sortExprMap := map[string]string{
		"email":          "u.email",
		"username":       "u.username",
		"created_at":     "u.created_at",
		"updated_at":     "u.created_at",
		// user.date_of_birth is stored as 'YYYY-MM-DD' (varchar), so lexicographic ordering matches date ordering.
		// Keep it simple here to avoid SQL syntax issues from inline casts/functions.
		"date_of_birth":  "u.date_of_birth",
		"country":        "u.country",
	}
	sortExpr := sortExprMap[sortBy]
	if sortExpr == "" {
		sortExpr = "u.created_at"
	}
	sortDir := strings.ToUpper(sortOrder)

	limit := req.PerPage
	if limit <= 0 {
		limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	query := fmt.Sprintf(`
		SELECT
			u.id,
			u.email,
			u.username,
			u.phone_number,
			u.first_name,
			u.last_name,
			u.default_currency,
			u.date_of_birth,
			u.country,
			u.state,
			u.street_address,
			u.postal_code,
			u.is_test_account,
			COALESCE((
				SELECT ul.withdrawal_limit_enabled
				FROM user_limits ul
				WHERE ul.user_id = u.id AND ul.limit_type = 'withdrawal'
				LIMIT 1
			), false) AS enable_withdrawal_limit,
			u.brand_id,
			u.created_at,
			u.created_at AS updated_at,
			COUNT(*) OVER() AS total
		FROM users u
		WHERE
			u.user_type = 'PLAYER' AND u.is_admin = false
			AND ($1::text = '' OR u.email ILIKE '%%' || $1 || '%%' OR u.username ILIKE '%%' || $1 || '%%')
			AND ($2::uuid IS NULL OR u.brand_id = $2)
			AND ($3::text = '' OR u.country ILIKE '%%' || $3 || '%%')
			AND ($4::bool IS NULL OR u.is_test_account = $4)
		ORDER BY %s %s
		LIMIT $5 OFFSET $6
	`, sortExpr, sortDir)

	rows, err := p.db.GetPool().Query(ctx, query, searchStr, brandID, countryStr, testAccount, limit, offset)
	if err != nil {
		// Include the built SQL to quickly pinpoint syntax errors in dynamic ORDER BY / filters.
		p.log.Error("unable to get players",
			zap.Error(err),
			zap.String("built_query", query),
			zap.Any("params", map[string]any{
				"search":     searchStr,
				"brand_id":   brandID,
				"country":    countryStr,
				"testAccount": testAccount,
				"limit":      limit,
				"offset":     offset,
				"sort_by":    sortBy,
				"sort_order": sortOrder,
			}),
		)
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players")
		return dto.GetPlayersRess{}, err
	}
	defer rows.Close()

	type playerRowWithTotal struct {
		playerRow
		Total int64
	}

	var resultRows []playerRowWithTotal
	for rows.Next() {
		var row playerRowWithTotal
		if err := rows.Scan(
			&row.ID,
			&row.Email,
			&row.Username,
			&row.Phone,
			&row.FirstName,
			&row.LastName,
			&row.DefaultCurrency,
			&row.DateOfBirth,
			&row.Country,
			&row.State,
			&row.StreetAddress,
			&row.PostalCode,
			&row.TestAccount,
			&row.EnableWithdrawalLimit,
			&row.BrandID,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Total,
		); err != nil {
			p.log.Error("unable to scan player row", zap.Error(err))
			return dto.GetPlayersRess{}, errors.ErrUnableToGet.Wrap(err, "unable to scan player rows")
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		p.log.Error("error iterating player rows", zap.Error(err))
		return dto.GetPlayersRess{}, errors.ErrUnableToGet.Wrap(err, "unable to get players")
	}

	if len(resultRows) == 0 {
		return dto.GetPlayersRess{
			Players:     []dto.Player{},
			TotalCount:  0,
			TotalPages:  0,
			CurrentPage: req.Page,
			PerPage:     req.PerPage,
		}, nil
	}

	total := int(resultRows[0].Total)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	result := dto.GetPlayersRess{
		TotalCount:  total,
		CurrentPage: req.Page,
		TotalPages:  totalPages,
		PerPage:     limit,
		Players:     make([]dto.Player, len(resultRows)),
	}

	for i, row := range resultRows {
		result.Players[i] = p.mapToDTO(row.playerRow)
	}

	return result, nil
}

type playerRow struct {
	ID                    uuid.UUID
	Email                 sql.NullString
	Username              sql.NullString
	Phone                 sql.NullString
	FirstName             sql.NullString
	LastName              sql.NullString
	DefaultCurrency       sql.NullString
	DateOfBirth           sql.NullString
	Country               sql.NullString
	State                 sql.NullString
	StreetAddress         sql.NullString
	PostalCode            sql.NullString
	TestAccount           bool
	EnableWithdrawalLimit bool
	BrandID               uuid.UUID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (p *player) mapToDTO(row playerRow) dto.Player {
	var phone *string
	if row.Phone.Valid && strings.TrimSpace(row.Phone.String) != "" {
		phone = &row.Phone.String
	}

	var firstName *string
	if row.FirstName.Valid && strings.TrimSpace(row.FirstName.String) != "" {
		firstName = &row.FirstName.String
	}

	var lastName *string
	if row.LastName.Valid && strings.TrimSpace(row.LastName.String) != "" {
		lastName = &row.LastName.String
	}

	var state *string
	if row.State.Valid && strings.TrimSpace(row.State.String) != "" {
		state = &row.State.String
	}

	var streetAddress *string
	if row.StreetAddress.Valid && strings.TrimSpace(row.StreetAddress.String) != "" {
		streetAddress = &row.StreetAddress.String
	}

	var postalCode *string
	if row.PostalCode.Valid && strings.TrimSpace(row.PostalCode.String) != "" {
		postalCode = &row.PostalCode.String
	}

	brandID := &row.BrandID

	email := ""
	if row.Email.Valid {
		email = row.Email.String
	}
	username := ""
	if row.Username.Valid {
		username = row.Username.String
	}
	defaultCurrency := ""
	if row.DefaultCurrency.Valid {
		defaultCurrency = row.DefaultCurrency.String
	}
	country := ""
	if row.Country.Valid {
		country = row.Country.String
	}

	dobStr := ""
	if row.DateOfBirth.Valid {
		dobStr = strings.TrimSpace(row.DateOfBirth.String)
	}

	var dob time.Time
	if dobStr != "" {
		if t, err := time.Parse("2006-01-02", dobStr); err == nil {
			dob = t
		}
	}

	return dto.Player{
		ID:                    row.ID,
		Email:                 email,
		Username:              username,
		Phone:                 phone,
		FirstName:             firstName,
		LastName:              lastName,
		DefaultCurrency:       defaultCurrency,
		Brand:                 nil,
		DateOfBirth:           dob,
		Country:               country,
		State:                 state,
		StreetAddress:         streetAddress,
		PostalCode:            postalCode,
		TestAccount:           row.TestAccount,
		EnableWithdrawalLimit: row.EnableWithdrawalLimit,
		BrandID:               brandID,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func (p *player) DeletePlayer(ctx context.Context, playerID uuid.UUID) error {
	tx, err := p.db.GetPool().Begin(ctx)
	if err != nil {
		p.log.Error("unable to start transaction for delete player", zap.Error(err))
		return errors.ErrUnableToDelete.Wrap(err, "unable to delete player")
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `DELETE FROM user_limits WHERE user_id = $1 AND limit_type = 'withdrawal'`, playerID)
	_, _ = tx.Exec(ctx, `DELETE FROM balances WHERE user_id = $1`, playerID)

	res, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, playerID)
	if err != nil {
		p.log.Error("unable to delete player", zap.Error(err), zap.String("id", playerID.String()))
		return errors.ErrUnableToDelete.Wrap(err, "unable to delete player")
	}

	_ = res // module already checks existence, keep delete idempotent
	if err := tx.Commit(ctx); err != nil {
		p.log.Error("unable to commit transaction for delete player", zap.Error(err))
		return errors.ErrUnableToDelete.Wrap(err, "unable to delete player")
	}
	return nil
}
