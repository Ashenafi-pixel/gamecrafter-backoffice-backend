package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LootBox struct {
	ID           uuid.UUID       `json:"id"`
	AssociatedID uuid.UUID       `json:"associated_id,omitempty"`
	PrizeType    string          `json:"prize_type"`
	PrizeValue   decimal.Decimal `json:"prize_value"`
	Probability  decimal.Decimal `json:"probability"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type CreateLootBoxReq struct {
	PrizeType   string          `json:"prize_type"`
	PrizeValue  decimal.Decimal `json:"prize_value"`
	Probability decimal.Decimal `json:"probability"`
}

type CreateLootBoxRes struct {
	Message string  `json:"message"`
	LootBox LootBox `json:"loot_box"`
}

type UpdateLootBoxReq struct {
	ID          uuid.UUID       `json:"id"`
	PrizeType   string          `json:"prize_type"`
	PrizeValue  decimal.Decimal `json:"prize_value"`
	Probability decimal.Decimal `json:"probability"`
}

type UpdateLootBoxRes struct {
	Message string  `json:"message"`
	LootBox LootBox `json:"loot_box"`
}
type DeleteLootBoxRes struct {
	Message string `json:"message"`
}

func CheckValidTypes(ty string) bool {
	validTypes := []string{"bucks"}
	for _, validType := range validTypes {
		if ty == validType {
			return true
		}
	}
	return false
}

type PlaceLootBoxBetReq struct {
	ID            uuid.UUID `json:"id,omitempty"`
	UserID        uuid.UUID `json:"user_id"`
	UserSelection uuid.UUID `json:"user_selection,omitempty"`
	LootBox       []byte    `json:"loot_box"`
	WonStatus     string    `json:"won_status"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PlaceLootBoxBetRes struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
	LootBox []byte    `json:"loot_box"`
}

type GetLootBoxRes struct {
	LootBoxes []LootBox `json:"loot_boxes"`
	Total     int       `json:"total"`
}

type PlaceLootBoxResp struct {
	ID        uuid.UUID `json:"id"`
	LootBoxID uuid.UUID `json:"loot_box_id"`
}

type LootBoxBetResp struct {
	LootBoxID  uuid.UUID       `json:"loot_box_id"`
	PrizeType  string          `json:"prize_type"`
	PrizeValue decimal.Decimal `json:"prize_value"`
	WonStatus  string          `json:"won_status"`
	Status     string          `json:"status"`
}
