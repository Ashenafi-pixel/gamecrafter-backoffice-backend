package player

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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

	var phone sql.NullString
	if playerReq.Phone != nil && *playerReq.Phone != "" {
		phone = sql.NullString{String: *playerReq.Phone, Valid: true}
	}

	var firstName sql.NullString
	if playerReq.FirstName != nil && *playerReq.FirstName != "" {
		firstName = sql.NullString{String: *playerReq.FirstName, Valid: true}
	}

	var lastName sql.NullString
	if playerReq.LastName != nil && *playerReq.LastName != "" {
		lastName = sql.NullString{String: *playerReq.LastName, Valid: true}
	}

	var brand sql.NullString
	if playerReq.Brand != nil && *playerReq.Brand != "" {
		brand = sql.NullString{String: *playerReq.Brand, Valid: true}
	}

	var state sql.NullString
	if playerReq.State != nil && *playerReq.State != "" {
		state = sql.NullString{String: *playerReq.State, Valid: true}
	}

	var streetAddress sql.NullString
	if playerReq.StreetAddress != nil && *playerReq.StreetAddress != "" {
		streetAddress = sql.NullString{String: *playerReq.StreetAddress, Valid: true}
	}

	var postalCode sql.NullString
	if playerReq.PostalCode != nil && *playerReq.PostalCode != "" {
		postalCode = sql.NullString{String: *playerReq.PostalCode, Valid: true}
	}

	var brandID sql.NullInt32
	if playerReq.BrandID != nil {
		brandID = sql.NullInt32{Int32: *playerReq.BrandID, Valid: true}
	}

	query := `
		INSERT INTO players (
			email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
	`

	var playerRow playerRow

	err := p.db.GetPool().QueryRow(
		ctx,
		query,
		playerReq.Email,
		playerReq.Username,
		hashedPassword,
		phone,
		firstName,
		lastName,
		playerReq.DefaultCurrency,
		brand,
		playerReq.DateOfBirth,
		playerReq.Country,
		state,
		streetAddress,
		postalCode,
		playerReq.TestAccount,
		playerReq.EnableWithdrawalLimit,
		brandID,
	).Scan(
		&playerRow.ID,
		&playerRow.Email,
		&playerRow.Username,
		&playerRow.Password,
		&playerRow.Phone,
		&playerRow.FirstName,
		&playerRow.LastName,
		&playerRow.DefaultCurrency,
		&playerRow.Brand,
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
		p.log.Error("unable to create player", zap.Error(err), zap.Any("player", playerReq))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create player")
		return dto.Player{}, err
	}

	return p.mapToDTO(playerRow), nil
}

func (p *player) GetPlayerByID(ctx context.Context, playerID int32) (dto.Player, bool, error) {
	query := `
		SELECT id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
		FROM players
		WHERE id = $1
	`

	var playerRow playerRow

	err := p.db.GetPool().QueryRow(ctx, query, playerID).Scan(
		&playerRow.ID,
		&playerRow.Email,
		&playerRow.Username,
		&playerRow.Password,
		&playerRow.Phone,
		&playerRow.FirstName,
		&playerRow.LastName,
		&playerRow.DefaultCurrency,
		&playerRow.Brand,
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
		p.log.Error("unable to get player by ID", zap.Error(err), zap.Int32("id", playerID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get player by ID")
		return dto.Player{}, false, err
	}

	return p.mapToDTO(playerRow), true, nil
}

func (p *player) UpdatePlayer(ctx context.Context, playerReq dto.Player) (dto.Player, error) {
	var email sql.NullString
	var username sql.NullString
	var phone sql.NullString
	var firstName sql.NullString
	var lastName sql.NullString
	var defaultCurrency sql.NullString
	var brand sql.NullString
	var dateOfBirth sql.NullTime
	var country sql.NullString
	var state sql.NullString
	var streetAddress sql.NullString
	var postalCode sql.NullString
	var testAccount sql.NullBool
	var enableWithdrawalLimit sql.NullBool
	var brandID sql.NullInt32

	// Set values if provided
	if playerReq.Email != "" {
		email = sql.NullString{String: playerReq.Email, Valid: true}
	}
	if playerReq.Username != "" {
		username = sql.NullString{String: playerReq.Username, Valid: true}
	}
	if playerReq.Phone != nil {
		phone = sql.NullString{String: *playerReq.Phone, Valid: true}
	}
	if playerReq.FirstName != nil {
		firstName = sql.NullString{String: *playerReq.FirstName, Valid: true}
	}
	if playerReq.LastName != nil {
		lastName = sql.NullString{String: *playerReq.LastName, Valid: true}
	}
	if playerReq.DefaultCurrency != "" {
		defaultCurrency = sql.NullString{String: playerReq.DefaultCurrency, Valid: true}
	}
	if playerReq.Brand != nil {
		brand = sql.NullString{String: *playerReq.Brand, Valid: true}
	}
	if !playerReq.DateOfBirth.IsZero() {
		dateOfBirth = sql.NullTime{Time: playerReq.DateOfBirth, Valid: true}
	}
	if playerReq.Country != "" {
		country = sql.NullString{String: playerReq.Country, Valid: true}
	}
	if playerReq.State != nil {
		state = sql.NullString{String: *playerReq.State, Valid: true}
	}
	if playerReq.StreetAddress != nil {
		streetAddress = sql.NullString{String: *playerReq.StreetAddress, Valid: true}
	}
	if playerReq.PostalCode != nil {
		postalCode = sql.NullString{String: *playerReq.PostalCode, Valid: true}
	}
	testAccount = sql.NullBool{Bool: playerReq.TestAccount, Valid: true}
	enableWithdrawalLimit = sql.NullBool{Bool: playerReq.EnableWithdrawalLimit, Valid: true}
	if playerReq.BrandID != nil {
		brandID = sql.NullInt32{Int32: *playerReq.BrandID, Valid: true}
	}

	query := `
		UPDATE players
		SET 
			email = COALESCE($2, email),
			username = COALESCE($3, username),
			phone = COALESCE($4, phone),
			first_name = COALESCE($5, first_name),
			last_name = COALESCE($6, last_name),
			default_currency = COALESCE($7, default_currency),
			brand = COALESCE($8, brand),
			date_of_birth = COALESCE($9, date_of_birth),
			country = COALESCE($10, country),
			state = COALESCE($11, state),
			street_address = COALESCE($12, street_address),
			postal_code = COALESCE($13, postal_code),
			test_account = COALESCE($14, test_account),
			enable_withdrawal_limit = COALESCE($15, enable_withdrawal_limit),
			brand_id = COALESCE($16, brand_id),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
	`

	var playerRow playerRow

	err := p.db.GetPool().QueryRow(
		ctx,
		query,
		playerReq.ID,
		email,
		username,
		phone,
		firstName,
		lastName,
		defaultCurrency,
		brand,
		dateOfBirth,
		country,
		state,
		streetAddress,
		postalCode,
		testAccount,
		enableWithdrawalLimit,
		brandID,
	).Scan(
		&playerRow.ID,
		&playerRow.Email,
		&playerRow.Username,
		&playerRow.Password,
		&playerRow.Phone,
		&playerRow.FirstName,
		&playerRow.LastName,
		&playerRow.DefaultCurrency,
		&playerRow.Brand,
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
		p.log.Error("unable to update player", zap.Error(err), zap.Int32("id", playerReq.ID))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update player")
		return dto.Player{}, err
	}

	return p.mapToDTO(playerRow), nil
}

func (p *player) GetPlayersByIDs(ctx context.Context, playerIDs []int32) ([]dto.Player, error) {
	if len(playerIDs) == 0 {
		return []dto.Player{}, nil
	}

	query := `
		SELECT id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
		FROM players
		WHERE id = ANY($1)
		ORDER BY created_at DESC
	`

	rows, err := p.db.GetPool().Query(ctx, query, playerIDs)
	if err != nil {
		p.log.Error("unable to get players by IDs", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
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
			&playerRow.Password,
			&playerRow.Phone,
			&playerRow.FirstName,
			&playerRow.LastName,
			&playerRow.DefaultCurrency,
			&playerRow.Brand,
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
			err = errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
			return nil, err
		}

		players = append(players, p.mapToDTO(playerRow))
	}

	if err := rows.Err(); err != nil {
		p.log.Error("error iterating player rows", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players by IDs")
		return nil, err
	}

	return players, nil
}

func (p *player) GetAllPlayers(ctx context.Context) ([]dto.Player, error) {
	query := `
		SELECT id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at
		FROM players
		ORDER BY created_at DESC
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
			&playerRow.Password,
			&playerRow.Phone,
			&playerRow.FirstName,
			&playerRow.LastName,
			&playerRow.DefaultCurrency,
			&playerRow.Brand,
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

	var search sql.NullString
	if req.Search != "" {
		search = sql.NullString{String: req.Search, Valid: true}
	}

	var brandID sql.NullInt32
	if req.BrandID != nil {
		parsedID, err := strconv.ParseInt(*req.BrandID, 10, 32)
		if err == nil {
			brandID = sql.NullInt32{Int32: int32(parsedID), Valid: true}
		}
	}

	var country sql.NullString
	if req.Country != nil && *req.Country != "" {
		country = sql.NullString{String: *req.Country, Valid: true}
	}

	var testAccount sql.NullBool
	if req.TestAccount != nil {
		testAccount = sql.NullBool{Bool: *req.TestAccount, Valid: true}
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

	// Validate sort fields to prevent SQL injection
	allowedSortFields := map[string]bool{
		"email":         true,
		"username":      true,
		"created_at":    true,
		"updated_at":    true,
		"date_of_birth": true,
		"country":       true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query := fmt.Sprintf(`
		SELECT 
			id, email, username, password, phone, first_name, last_name,
			default_currency, brand, date_of_birth, country, state,
			street_address, postal_code, test_account, enable_withdrawal_limit,
			brand_id, created_at, updated_at,
			COUNT(*) OVER() AS total
		FROM players
		WHERE 
			($1::text IS NULL OR email ILIKE '%%' || $1 || '%%' OR username ILIKE '%%' || $1 || '%%') AND
			($2::int IS NULL OR brand_id = $2) AND
			($3::text IS NULL OR country ILIKE '%%' || $3 || '%%') AND
			($4::bool IS NULL OR test_account = $4)
		ORDER BY %s %s
		LIMIT $5 OFFSET $6
	`, sortBy, strings.ToUpper(sortOrder))

	rows, err := p.db.GetPool().Query(ctx, query, search, brandID, country, testAccount, req.PerPage, offset)
	if err != nil {
		p.log.Error("unable to get players", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players")
		return dto.GetPlayersRess{}, err
	}
	defer rows.Close()

	type playerRowWithTotal struct {
		playerRow
		Total int64
	}

	var players []playerRowWithTotal
	for rows.Next() {
		var player playerRowWithTotal
		err := rows.Scan(
			&player.ID,
			&player.Email,
			&player.Username,
			&player.Password,
			&player.Phone,
			&player.FirstName,
			&player.LastName,
			&player.DefaultCurrency,
			&player.Brand,
			&player.DateOfBirth,
			&player.Country,
			&player.State,
			&player.StreetAddress,
			&player.PostalCode,
			&player.TestAccount,
			&player.EnableWithdrawalLimit,
			&player.BrandID,
			&player.CreatedAt,
			&player.UpdatedAt,
			&player.Total,
		)
		if err != nil {
			p.log.Error("unable to scan player row", zap.Error(err))
			err = errors.ErrUnableToGet.Wrap(err, "unable to get players")
			return dto.GetPlayersRess{}, err
		}
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		p.log.Error("error iterating player rows", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get players")
		return dto.GetPlayersRess{}, err
	}

	if len(players) == 0 {
		return dto.GetPlayersRess{
			Players:     []dto.Player{},
			TotalCount:  0,
			TotalPages:  0,
			CurrentPage: req.Page,
			PerPage:     req.PerPage,
		}, nil
	}

	var total int
	if len(players) > 0 {
		total = int(players[0].Total)
	}

	result := dto.GetPlayersRess{
		TotalCount:  total,
		CurrentPage: req.Page,
		TotalPages:  int(math.Ceil(float64(total) / float64(req.PerPage))),
		PerPage:     req.PerPage,
		Players:     make([]dto.Player, len(players)),
	}

	for i, player := range players {
		result.Players[i] = p.mapToDTO(player.playerRow)
	}

	return result, nil
}

type playerRow struct {
	ID                    int32
	Email                 string
	Username              string
	Password              string
	Phone                 sql.NullString
	FirstName             sql.NullString
	LastName              sql.NullString
	DefaultCurrency       string
	Brand                 sql.NullString
	DateOfBirth           time.Time
	Country               string
	State                 sql.NullString
	StreetAddress         sql.NullString
	PostalCode            sql.NullString
	TestAccount           bool
	EnableWithdrawalLimit bool
	BrandID               sql.NullInt32
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (p *player) mapToDTO(row playerRow) dto.Player {
	var phone *string
	if row.Phone.Valid {
		phone = &row.Phone.String
	}

	var firstName *string
	if row.FirstName.Valid {
		firstName = &row.FirstName.String
	}

	var lastName *string
	if row.LastName.Valid {
		lastName = &row.LastName.String
	}

	var brand *string
	if row.Brand.Valid {
		brand = &row.Brand.String
	}

	var state *string
	if row.State.Valid {
		state = &row.State.String
	}

	var streetAddress *string
	if row.StreetAddress.Valid {
		streetAddress = &row.StreetAddress.String
	}

	var postalCode *string
	if row.PostalCode.Valid {
		postalCode = &row.PostalCode.String
	}

	var brandID *int32
	if row.BrandID.Valid {
		brandID = &row.BrandID.Int32
	}

	return dto.Player{
		ID:                    row.ID,
		Email:                 row.Email,
		Username:              row.Username,
		Phone:                 phone,
		FirstName:             firstName,
		LastName:              lastName,
		DefaultCurrency:       row.DefaultCurrency,
		Brand:                 brand,
		DateOfBirth:           row.DateOfBirth,
		Country:               row.Country,
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

func (p *player) DeletePlayer(ctx context.Context, playerID int32) error {
	query := `DELETE FROM players WHERE id = $1`
	_, err := p.db.GetPool().Exec(ctx, query, playerID)
	if err != nil {
		p.log.Error("unable to delete player", zap.Error(err), zap.Int32("id", playerID))
		err = errors.ErrUnableToDelete.Wrap(err, "unable to delete player")
		return err
	}
	return nil
}
