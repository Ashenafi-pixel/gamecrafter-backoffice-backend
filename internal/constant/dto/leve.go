package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Level struct {
	ID        uuid.UUID `json:"id" swaggerignore:"true"`
	Level     int       `json:"level"`
	CreatedBy uuid.UUID `json:"created_by" swaggerignore:"true"`
	Status    bool      `json:"status"`
	Type      string    `json:"type" swaggerignore:"true"` // "players" or "squads"
}

type GetUserLevelResp struct {
	ID           uuid.UUID       `json:"id" swaggerignore:"true"`
	Level        int             `json:"level"`
	Bucks        decimal.Decimal `json:"bucks"`
	IsFinalLevel bool            `json:"is_final_level"`
}

type GetUserLevelResp2 struct {
	ID                      uuid.UUID       `json:"id" swaggerignore:"true"`
	Level                   int             `json:"level"`
	NextLevel               int             `json:"next_level"`
	AmountSpentToReachLevel decimal.Decimal `json:"amount_spent_to_reach_level"`
	NextLevelRequirement    decimal.Decimal `json:"next_level_requirement"`
	Bucks                   decimal.Decimal `json:"bucks"`
	IsFinalLevel            bool            `json:"is_final_level"`
	SquadID                 uuid.UUID       `json:"squad_id"`
}

type GetUserLevelResp3 struct {
	ID                      uuid.UUID       `json:"id" swaggerignore:"true"`
	Level                   int             `json:"level"`
	NextLevel               int             `json:"next_level"`
	AmountSpentToReachLevel decimal.Decimal `json:"amount_spent_to_reach_level"`
	NextLevelRequirement    decimal.Decimal `json:"next_level_requirement"`
	IsFinalLevel            bool            `json:"is_final_level"`
	SquadID                 uuid.UUID       `json:"squad_id"`
}

type LevelResp struct {
	ID           uuid.UUID          `json:"id" swaggerignore:"true"`
	Level        int                `json:"level"`
	CreatedBy    uuid.UUID          `json:"created_by" swaggerignore:"true"`
	Status       bool               `json:"status"`
	Requirements []LevelRequirement `json:"requirements"`
}
type GetLevelResp struct {
	Message   string      `json:"message"`
	TotalPage int         `json:"total_page"`
	Level     []LevelResp `json:"level"`
}

type GetLevelsResp struct {
	Message    string  `json:"message"`
	Levels     []Level `json:"levels"`
	TotalPages int     `json:"total_pages"`
}

type LevelRequirement struct {
	ID        uuid.UUID `json:"id" swaggerignore:"true"`
	LevelID   uuid.UUID `json:"level_id"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	CreatedBy uuid.UUID `json:"created_by"`
	Status    bool      `json:"status"`
}

type UpdateLevelRequirementReq struct {
	ID    uuid.UUID `json:"id"`
	Type  string    `json:"type"`
	Value string    `json:"value"`
}

type LevelRequirements struct {
	CreatedBy    uuid.UUID          `json:"created_by" swaggerignore:"true"`
	LevelID      uuid.UUID          `json:"level_id"`
	Requirements []LevelRequirement `json:"requirements"`
}

const (
	LevelRequirementTypeBetAmount = "bet_amount"
)

type UserLevelResp struct {
	ID    uuid.UUID `json:"id"`
	Level int       `json:"level"`
}
