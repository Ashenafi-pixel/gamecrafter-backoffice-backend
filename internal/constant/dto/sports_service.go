package dto

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// SportsServiceClaim represents the JWT claims for sports service authentication
type SportsServiceClaim struct {
	ServiceID   uuid.UUID `json:"service_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceName string    `json:"service_name" example:"Sports Service"`
	jwt.StandardClaims
}

// SportsServiceSignInReq represents the sign-in request for sports service
type SportsServiceSignInReq struct {
	ServiceID     string `json:"service_id" validate:"required" example:"sports-service-001"`
	ServiceSecret string `json:"service_secret" validate:"required" example:"sports-secret-key"`
}

// SportsServiceSignInRes represents the sign-in response for sports service
type SportsServiceSignInRes struct {
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Message string `json:"message" example:"Authentication successful"`
}

// PlaceBetResponse represents the response structure for placing a bet in the sports service
type PlaceBetResponse struct {
	Balance          string `json:"balance"`
	ExtTransactionID string `json:"ext_transaction_id"`
	AlreadyProcessed string `json:"already_processed"`
	CustomerId       string `json:"customer_id,omitempty"`
	BonusAmount      string `json:"bonus_amount,omitempty"`
}

// PlaceBetRequest represents the request structure for placing a bet in the sports service
type PlaceBetRequest struct {
	TransactionID   string      `json:"transaction_id"`
	BetAmount       string      `json:"bet_amount"`
	BetReferenceNum string      `json:"bet_reference_num"`
	GameReference   string      `json:"game_reference"`
	BetMode         string      `json:"bet_mode"`
	Description     string      `json:"description"`
	UserID          uuid.UUID   `json:"user_id"`
	FrontendType    string      `json:"frontend_type"`
	BetStatus       string      `json:"bet_status"`
	SportIDs        string      `json:"sport_ids"`
	SiteId          string      `json:"site_id"`
	ClientIP        string      `json:"client_ip,omitempty"`
	AffiliateUserID string      `json:"affiliate_user_id,omitempty"`
	Autorecharge    string      `json:"autorecharge,omitempty"`
	BetDetails      *BetDetails `json:"bet_details,omitempty"`
}

// BetDetails represents the complete bet structure from XML
type BetDetails struct {
	IsSystem    string    `json:"is_system"`
	EventCount  string    `json:"event_count"`
	BankerCount string    `json:"banker_count"`
	Events      string    `json:"events"`
	EventList   []Event   `json:"event_list"`
	BetStake    *BetStake `json:"bet_stake"`
}

// Event represents individual event in the bet
type Event struct {
	EventName      string `json:"event_name"`
	EventID        string `json:"event_id"`
	CategoryID     string `json:"category_id"`
	ChampionshipID string `json:"championship_id"`
	SportID        string `json:"sport_id"`
	ExtEventID     string `json:"ext_event_id"`
	EventDate      string `json:"event_date"`
	Market         Market `json:"market"`
}

// Market represents market information in an event
type Market struct {
	MarketType string `json:"market_type"`
	MarketID   string `json:"market_id"`
	ExtType    string `json:"ext_type"`
	Outcome    string `json:"outcome"`
	Odds       string `json:"odds"`
}

// BetStake represents the bet stake information
type BetStake struct {
	BetAmount     string `json:"bet_amount"`
	CombLength    string `json:"comb_length"`
	Winnings      string `json:"winnings"`
	MultipleBonus string `json:"multiple_bonus"`
	Odds          string `json:"odds"`
}

type SportsServiceAwardWinningsReq struct {
	TransactionID   string `json:"transaction_id" validate:"required"`
	WinAmount       string `json:"win_amount" validate:"required"`
	WinReferenceNum string `json:"win_reference_num" validate:"required"`
	BetMode         string `json:"bet_mode" validate:"required"`
	Description     string `json:"description" validate:"required"`
	ExternalUserID  string `json:"external_user_id" validate:"required"`
	FrontendType    string `json:"frontend_type" validate:"required"`
}

type SportsServiceAwardWinningsRes struct {
	Balance          string `json:"balance"`
	ExtTransactionID string `json:"ext_transaction_id"`
	AlreadyProcessed string `json:"already_processed"`
}
