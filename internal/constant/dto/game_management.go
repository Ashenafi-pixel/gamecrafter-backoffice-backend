package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Enhanced Game DTO with all database fields
type GameManagement struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Name               string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Status             string    `json:"status" db:"status" validate:"required,oneof=ACTIVE INACTIVE MAINTENANCE"`
	Timestamp          time.Time `json:"timestamp" db:"timestamp"`
	Photo              *string   `json:"photo,omitempty" db:"photo"`
	Price              *string   `json:"price,omitempty" db:"price"`
	Enabled            bool      `json:"enabled" db:"enabled"`
	GameID             *string   `json:"game_id,omitempty" db:"game_id" validate:"omitempty,max=50"`
	InternalName       *string   `json:"internal_name,omitempty" db:"internal_name" validate:"omitempty,max=255"`
	IntegrationPartner *string   `json:"integration_partner,omitempty" db:"integration_partner" validate:"omitempty,max=100"`
	Provider           *string   `json:"provider,omitempty" db:"provider" validate:"omitempty,max=100"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Game creation request
type CreateGameRequest struct {
	Name               string  `json:"name" validate:"required,min=1,max=255"`
	Status             string  `json:"status" validate:"required,oneof=ACTIVE INACTIVE MAINTENANCE"`
	Photo              *string `json:"photo,omitempty"`
	Price              *string `json:"price,omitempty"`
	Enabled            bool    `json:"enabled"`
	GameID             *string `json:"game_id,omitempty" validate:"omitempty,max=50"`
	InternalName       *string `json:"internal_name,omitempty" validate:"omitempty,max=255"`
	IntegrationPartner *string `json:"integration_partner,omitempty" validate:"omitempty,max=100"`
	Provider           *string `json:"provider,omitempty" validate:"omitempty,max=100"`
}

// Game update request
type UpdateGameRequest struct {
	Name               *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Status             *string `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE MAINTENANCE"`
	Photo              *string `json:"photo,omitempty"`
	Price              *string `json:"price,omitempty"`
	Enabled            *bool   `json:"enabled,omitempty"`
	GameID             *string `json:"game_id,omitempty" validate:"omitempty,max=50"`
	InternalName       *string `json:"internal_name,omitempty" validate:"omitempty,max=255"`
	IntegrationPartner *string `json:"integration_partner,omitempty" validate:"omitempty,max=100"`
	Provider           *string `json:"provider,omitempty" validate:"omitempty,max=100"`
}

// Game response
type GameResponse struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	Timestamp          time.Time `json:"timestamp"`
	Photo              *string   `json:"photo,omitempty"`
	Price              *string   `json:"price,omitempty"`
	Enabled            bool      `json:"enabled"`
	GameID             *string   `json:"game_id,omitempty"`
	InternalName       *string   `json:"internal_name,omitempty"`
	IntegrationPartner *string   `json:"integration_partner,omitempty"`
	Provider           *string   `json:"provider,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	RTPPercent         *string   `json:"rtp_percent,omitempty"`
}

// Game list response
type GameListResponse struct {
	Games      []GameResponse `json:"games"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// Game query parameters
type GameQueryParams struct {
	Page      int    `form:"page" validate:"min=1"`
	PerPage   int    `form:"per_page" validate:"min=1,max=100"`
	Search    string `form:"search"`
	Status    string `form:"status" validate:"omitempty,oneof=ACTIVE INACTIVE MAINTENANCE"`
	Provider  string `form:"provider"`
	GameID    string `form:"game_id"` // Filter by game_id
	Enabled   *bool  `form:"enabled"`
	SortBy    string `form:"sort_by" validate:"omitempty,oneof=name status created_at updated_at"`
	SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// GameHouseEdge CRUD DTOs
type GameHouseEdgeRequest struct {
	GameID         *string          `json:"game_id,omitempty" validate:"omitempty,max=50"`
	GameType       string           `json:"game_type" validate:"required,oneof=slot sports table live crash plinko wheel"`
	GameVariant    *string          `json:"game_variant,omitempty" validate:"omitempty,oneof=classic v1 v2 demo real mobile desktop"`
	HouseEdge      decimal.Decimal  `json:"house_edge" validate:"required,min=0,max=1"`
	MinBet         decimal.Decimal  `json:"min_bet" validate:"required,min=0"`
	MaxBet         *decimal.Decimal `json:"max_bet,omitempty" validate:"omitempty,min=0"`
	IsActive       bool             `json:"is_active"`
	EffectiveFrom  *FlexibleTime    `json:"effective_from,omitempty"`
	EffectiveUntil *FlexibleTime    `json:"effective_until,omitempty"`
}

type GameHouseEdgeResponse struct {
	ID               uuid.UUID        `json:"id"`
	GameID           *string          `json:"game_id,omitempty"`
	GameName         *string          `json:"game_name,omitempty"`
	GameType         string           `json:"game_type"`
	GameVariant      *string          `json:"game_variant,omitempty"`
	HouseEdge        decimal.Decimal  `json:"house_edge"`
	HouseEdgePercent string           `json:"house_edge_percent"`
	MinBet           decimal.Decimal  `json:"min_bet"`
	MaxBet           *decimal.Decimal `json:"max_bet,omitempty"`
	IsActive         bool             `json:"is_active"`
	EffectiveFrom    time.Time        `json:"effective_from"`
	EffectiveUntil   *time.Time       `json:"effective_until,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type GameHouseEdgeListResponse struct {
	HouseEdges []GameHouseEdgeResponse `json:"house_edges"`
	TotalCount int                     `json:"total_count"`
	Page       int                     `json:"page"`
	PerPage    int                     `json:"per_page"`
	TotalPages int                     `json:"total_pages"`
}

type GameHouseEdgeQueryParams struct {
	Page        int    `form:"page" validate:"min=1"`
	PerPage     int    `form:"per_page" validate:"min=1,max=100"`
	Search      string `form:"search"`  // Search by game_id and game name
	GameID      string `form:"game_id"` // Specific game ID search
	GameType    string `form:"game_type"`
	GameVariant string `form:"game_variant"`
	IsActive    *bool  `form:"is_active"`
	SortBy      string `form:"sort_by" validate:"omitempty,oneof=game_type game_variant house_edge created_at updated_at"`
	SortOrder   string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Game with house edge information
type GameWithHouseEdge struct {
	Game       GameResponse            `json:"game"`
	HouseEdges []GameHouseEdgeResponse `json:"house_edges"`
}

// Bulk operations
type BulkUpdateGameStatusRequest struct {
	GameIDs []uuid.UUID `json:"game_ids" validate:"required,min=1"`
	Status  string      `json:"status" validate:"required,oneof=ACTIVE INACTIVE MAINTENANCE"`
}

type BulkUpdateHouseEdgeRequest struct {
	HouseEdgeIDs []uuid.UUID `json:"house_edge_ids" validate:"required,min=1"`
	IsActive     bool        `json:"is_active"`
}

// Game statistics
type GameManagementStats struct {
	TotalGames       int `json:"total_games"`
	ActiveGames      int `json:"active_games"`
	InactiveGames    int `json:"inactive_games"`
	MaintenanceGames int `json:"maintenance_games"`
	EnabledGames     int `json:"enabled_games"`
	DisabledGames    int `json:"disabled_games"`
}

type HouseEdgeStats struct {
	TotalHouseEdges    int `json:"total_house_edges"`
	ActiveHouseEdges   int `json:"active_house_edges"`
	InactiveHouseEdges int `json:"inactive_house_edges"`
	UniqueGameTypes    int `json:"unique_game_types"`
	UniqueGameVariants int `json:"unique_game_variants"`
}
