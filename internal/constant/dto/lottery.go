package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateLotteryServiceReq struct {
	ID              uuid.UUID `json:"id"`
	ServiceName     string    `json:"service_name" validate:"required"`
	ServiceSecret   string    `json:"service_secret" validate:"required"`
	ServiceClientID string    `json:"service_client_id" validate:"required"`
	Description     string    `json:"description" validate:"omitempty"`
	CallbackURL     string    `json:"callback_url"`
}

type CreateLotteryServiceRes struct {
	Message   string    `json:"message"`
	ServiceID uuid.UUID `json:"service_id"`
}

type GetLotteryServiceRes struct {
	ServiceID     string `json:"service_id"`
	ServiceName   string `json:"service_name"`
	ServiceSecret string `json:"service"`
	Description   string `json:"description"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CreateLocalLotteryReq struct {
	Name          string  `json:"name" validate:"required"`
	Price         float64 `json:"price" validate:"required"`
	MinSelectable int     `json:"min_selectable" validate:"required"`
	MaxSelectable int     `json:"max_selectable" validate:"required"`
	DrawFrequency string  `json:"draw_frequency" validate:"required"`
	NumberOfBalls int     `json:"number_of_balls" validate:"required"`
	Description   string  `json:"description" validate:"required"`
	Status        string  `json:"status" validate:"required"`
	ServiceID     string  `json:"service_id" validate:"required"`
}

type Lottery struct {
	ID             uuid.UUID       `json:"id,omitempty"`
	Name           string          `json:"name" validate:"required"`
	Description    string          `json:"description"`
	DrawHour       int             `json:"draw_hour"`
	DrawMinute     int             `json:"draw_minute"`
	DrawDayOfWeek  int             `json:"draw_day_of_week"`
	DrawDayOfMonth int             `json:"draw_day_of_month"`
	Active         bool            `json:"active"`
	PamID          string          `json:"pam_id"`
	Price          decimal.Decimal `json:"price"`
	MinSelectable  int             `json:"min_selectable"`
	MaxSelectable  int             `json:"max_selectable"`
	DrawFrequency  string          `json:"draw_frequency" validate:"required"`
	NumberOfBalls  int             `json:"number_of_balls" validate:"required"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type LotteryReward struct {
	ID              string          `json:"id"`
	LotteryID       string          `json:"lottery_id"`
	Prize           decimal.Decimal `json:"prize"`
	NumberOfWinners int             `json:"number_of_winners"`
	Type            string          `json:"type"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type LotteryRequestCreate struct {
	Lottery        Lottery         `json:"lottery"`
	LotteryRewards []LotteryReward `json:"lottery_rewards"`
}

type KafkaLotteryWins struct {
	RewardID        uuid.UUID       `json:"reward_id"`
	Prize           decimal.Decimal `json:"prize"`
	TicketNumber    string          `json:"ticket_number"`
	NumberOfTickets int             `json:"number_of_tickets"`
	PrizeType       string          `json:"prize_type"`
	Currency        string          `json:"currency"`
	PrizeCurrency   string          `json:"prize_currency"`
	UserID          uuid.UUID       `json:"user_id"`
}

type KafkaLotteryEvent struct {
	UniqueID  uuid.UUID            `json:"unique_id" validate:"required,uuid4"`
	LotteryID uuid.UUID            `json:"lottery_id"`
	Rewards   []KafkaLotteryReward `json:"rewards"`
	Winners   []KafkaLotteryWins   `json:"winners"`
}

type KafkaLotteryReward struct {
	ID              uuid.UUID       `json:"id" validate:"omitempty,uuid4"`
	LotteryID       uuid.UUID       `json:"lottery_id" validate:"required,uuid4"`
	Prize           decimal.Decimal `json:"prize" validate:"required,gt=0"`
	DrawedNumbers   [][]int32       `json:"drawed_numbers"`
	NumberOfWinners int             `json:"number_of_winners" validate:"required,gte=1"`
	Type            string          `json:"type" validate:"required,oneof=cash gift voucher"`
	Currency        string          `json:"currency" validate:"required,currency"`
	CreatedAt       time.Time       `json:"created_at" validate:"required"`
	UpdatedAt       time.Time       `json:"updated_at" validate:"required"`
}

type LotteryLog struct {
	ID              uuid.UUID       `json:"id"`
	LotteryID       uuid.UUID       `json:"lottery_id"`
	UserID          uuid.UUID       `json:"user_id"`
	RewardID        uuid.UUID       `json:"reward_id"`
	WonAmount       decimal.Decimal `json:"won_amount"`
	Currency        string          `json:"currency"`
	NumberOfTickets int             `json:"number_of_tickets"`
	TicketNumber    string          `json:"ticket_number"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type LotteryKafkaLog struct {
	ID              uuid.UUID       `json:"id" validate:"required,uuid4"`
	LotteryID       uuid.UUID       `json:"lottery_id" validate:"required,uuid4"`
	LotteryRewardID uuid.UUID       `json:"lottery_reward_id" validate:"required,uuid4"`
	DrawNumbers     [][]int32       `json:"draw_numbers" validate:"required"`
	Prize           decimal.Decimal `json:"prize" validate:"required,gt=0"`
	CreatedAt       time.Time       `json:"created_at" validate:"required"`
	UpdatedAt       time.Time       `json:"updated_at" validate:"required"`
	UniqIdentifier  uuid.UUID       `json:"unique_id" validate:"required,uuid4"`
}

type LotteryVerifyAndDeductBalanceReq struct {
	UserID    uuid.UUID       `json:"user_id" validate:"required,uuid4"`
	Currency  string          `json:"currency" validate:"required,currency"`
	Amount    decimal.Decimal `json:"amount" validate:"required,gt=0"`
	Component string          `json:"component" validate:"required,oneof=real_money bonus_money"`
}

type LotteryVerifyAndDeductBalanceRes struct {
	UserID    uuid.UUID       `json:"user_id"`
	Currency  string          `json:"currency"`
	Amount    decimal.Decimal `json:"amount"`
	Component string          `json:"component"`
}

type LotteryServiceLoginReq struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

type LotteryServiceLoginRes struct {
	Token string `json:"token"`
}
