package dto

import (
	"encoding/json"
	"time"
)

// FalconLiquidityConfig represents the configuration for Falcon Liquidity integration
type FalconLiquidityConfig struct {
	Enabled        bool   `json:"enabled"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	VirtualHost    string `json:"virtual_host"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	ExchangeName   string `json:"exchange_name"`
	QueueName      string `json:"queue_name"`
	RoutingKey     string `json:"routing_key"`
	ClientName     string `json:"client_name"`
	ManagementPort int    `json:"management_port"`
}

// FalconBetData represents the base structure for all bet data sent to Falcon Liquidity
type FalconBetData struct {
	BetID                  string          `json:"bet_id"`
	UserID                 string          `json:"user_id"`
	Username               string          `json:"username,omitempty"`
	Active                 bool            `json:"active"`
	Amount                 float64         `json:"amount"`
	Payout                 float64         `json:"payout"`
	Status                 string          `json:"status"`
	Currency               string          `json:"currency"`
	BetPlacedUnixTimestamp int64           `json:"bet_placed_unix_timestamp"`
	Type                   string          `json:"type"` // "sport" or "casino"
	Metadata               json.RawMessage `json:"metadata,omitempty"`
	PayoutMultiplier       float64         `json:"payout_multiplier"`
}

// FalconSportBetData represents sport-specific bet data
type FalconSportBetData struct {
	FalconBetData
	NumberOfBets   int     `json:"number_of_bets"`
	OddsTaken      float64 `json:"odds_taken"`
	FairPrice      float64 `json:"fair_price"`
	MarketName     string  `json:"market_name,omitempty"`
	FixtureSlug    string  `json:"fixture_slug,omitempty"`
	TournamentSlug string  `json:"tournament_slug,omitempty"`
	Sport          string  `json:"sport,omitempty"`
}

// FalconCasinoBetData represents casino-specific bet data
type FalconCasinoBetData struct {
	FalconBetData
	GameName     string  `json:"game_name"`
	GameID       string  `json:"game_id,omitempty"`
	Edge         float64 `json:"edge"`
	Provider     string  `json:"provider"`
	ProviderType string  `json:"provider_type"`
}

// FalconMetadata represents additional metadata for bet tracking
type FalconMetadata struct {
	ClientName string                 `json:"client_name"`
	DeviceType string                 `json:"device_type,omitempty"`
	TableID    string                 `json:"table_id,omitempty"`
	RoundID    string                 `json:"round_id,omitempty"`
	ClientRef  string                 `json:"client_reference,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty"`
}

// NewFalconCasinoBet creates a new casino bet data structure for Falcon Liquidity
func NewFalconCasinoBet(betID, userID, username, gameName, provider, providerType, currency, status string, amount, payout, edge float64, active bool) *FalconCasinoBetData {
	payoutMultiplier := 0.0
	if payout > amount {
		payoutMultiplier = payout / amount
	}

	metadata := FalconMetadata{
		ClientName: "tucanbit",
	}

	metadataJSON, _ := json.Marshal(metadata)

	return &FalconCasinoBetData{
		FalconBetData: FalconBetData{
			BetID:                  betID,
			UserID:                 userID,
			Username:               username,
			Active:                 active,
			Amount:                 amount,
			Payout:                 payout,
			Status:                 status,
			Currency:               currency,
			BetPlacedUnixTimestamp: time.Now().Unix(),
			Type:                   "casino",
			Metadata:               metadataJSON,
			PayoutMultiplier:       payoutMultiplier,
		},
		GameName:     gameName,
		Edge:         edge,
		Provider:     provider,
		ProviderType: providerType,
	}
}

// NewFalconSportBet creates a new sport bet data structure for Falcon Liquidity
func NewFalconSportBet(betID, userID, username, currency, status string, amount, payout, oddsTaken, fairPrice float64, numberOfBets int, active bool) *FalconSportBetData {
	payoutMultiplier := 0.0
	if payout > amount {
		payoutMultiplier = payout / amount
	}

	metadata := FalconMetadata{
		ClientName: "tucanbit",
	}

	metadataJSON, _ := json.Marshal(metadata)

	return &FalconSportBetData{
		FalconBetData: FalconBetData{
			BetID:                  betID,
			UserID:                 userID,
			Username:               username,
			Active:                 active,
			Amount:                 amount,
			Payout:                 payout,
			Status:                 status,
			Currency:               currency,
			BetPlacedUnixTimestamp: time.Now().Unix(),
			Type:                   "sport",
			Metadata:               metadataJSON,
			PayoutMultiplier:       payoutMultiplier,
		},
		NumberOfBets: numberOfBets,
		OddsTaken:    oddsTaken,
		FairPrice:    fairPrice,
	}
}
