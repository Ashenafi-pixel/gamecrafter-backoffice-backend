package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type BetRound struct {
	ID         uuid.UUID       `json:"id"`
	Status     string          `json:"status"`
	CrashPoint decimal.Decimal `json:"crash_point"`
	CreatedAt  *time.Time      `json:"created_at"`
	ClosedAt   *time.Time      `json:"closed_at,omitempty"`
	UserID     uuid.UUID       `json:"user_id,omitempty"`
	Currency   string          `json:"currency,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	BetID      uuid.UUID       `json:"bet_id,omitempty"`
}

type BetRoundResp struct {
	ID          uuid.UUID       `json:"id"`
	Multiplayer decimal.Decimal `json:"multiplayer"`
}
type OpenRoundRes struct {
	Message   string    `json:"message"`
	BetRound  uuid.UUID `json:"bet_round,omitempty"`
	StartTime *string   `json:"start_time,omitempty"`
}

type CrashPointRes struct {
	Round   BetRound `json:"round"`
	Message string   `json:"message"`
}

type PlaceBetReq struct {
	UserID   uuid.UUID       `json:"user_id"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	RoundID  uuid.UUID       `json:"round_id"`
}
type PlaceBetReqData struct {
	BetID    uuid.UUID       `json:"bet_id"`
	RoundID  uuid.UUID       `json:"round_id"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}
type PlaceBetRes struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Date    PlaceBetReqData `json:"data"`
}

type Bet struct {
	BetID               uuid.UUID       `json:"bet_id"`
	RoundID             uuid.UUID       `json:"round_id"`
	UserID              uuid.UUID       `json:"user_id"`
	Amount              decimal.Decimal `json:"amount"`
	Currency            string          `json:"currency"`
	ClientTransactionID string          `json:"client_transaction_id"`
	Timestamp           time.Time       `json:"timestamp"`
	Payout              decimal.Decimal `json:"payout"`
	CashOutMultiplier   decimal.Decimal `json:"cash_out_multiplie"`
	Status              string          `json:"status"`
}

type CashOutReq struct {
	UserID  uuid.UUID `json:"user_id"`
	RoundID uuid.UUID `json:"round_id"`
}

type CashOutResData struct {
	BetID             uuid.UUID       `json:"bet_id"`
	CashOutMultiplier decimal.Decimal `json:"cash_out_multiplier"`
	Payout            decimal.Decimal `json:"payout"`
	Currency          string          `json:"currency"`
}

type CashOutRes struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    CashOutResData `json:"data"`
}

type SaveCashoutReq struct {
	ID         uuid.UUID
	Multiplier decimal.Decimal
	Payout     decimal.Decimal
}
type BetRes struct {
	UserID            uuid.UUID       `json:"user_id"`
	BetAmount         decimal.Decimal `json:"bet_amount"`
	CashOutMultiplier decimal.Decimal `json:"cash_out_multiplier"`
	Payout            decimal.Decimal `json:"payout"`
	Currency          string          `json:"currency"`
	Timestamp         time.Time       `json:"timestamp"`
}

type BetHisotryData struct {
	Page       int      `json:"page"`
	TotalPages int      `json:"total_pages"`
	Bets       []BetRes `json:"bets"`
}

type BetHistoryResp struct {
	Status string         `json:"status"`
	Data   BetHisotryData `json:"data"`
}

type GetBetHistoryReq struct {
	Page    int       `json:"page"`
	UserID  uuid.UUID `json:"user_id"`
	PerPage int       `json:"per_page"`
	Offset  int       `json:"offset"`
}

type CancelBetResp struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
	RoundID uuid.UUID `json:"round_id"`
}

type CancelBetReq struct {
	UserID  uuid.UUID `json:"uuid" swaggerignore:"true"`
	RoundID uuid.UUID `json:"round_id"`
}

type Leader struct {
	ProfileURL string          `json:"profile_url"`
	Payout     decimal.Decimal `json:"payout"`
}

type LeadersResp struct {
	TotalPlayers int      `json:"total_player"`
	Leaders      []Leader `json:"leaders"`
}

type SaveFailedBetsLog struct {
	UserID        uuid.UUID `json:"user_id"`
	RoundID       uuid.UUID `json:"round_id"`
	BetID         uuid.UUID `json:"bet_id"`
	AdminID       uuid.UUID `json:"admin_id" swaggerignore:"true"`
	Manual        bool      `json:"manual"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"create_at"`
	TransactionID uuid.UUID `json:"transaction_id"`
}

type GetFailedRoundsReq struct {
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"per_page" form:"per_page"`
	Status  string `json:"statis" form:"status"`
}

type FailedBetLogs struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	RoundID       uuid.UUID `json:"round_id"`
	BetID         uuid.UUID `json:"bet_id"`
	IsManual      bool      `json:"is_manual"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AdminID       uuid.UUID `json:"admin_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type FailedRoundsResData struct {
	Round         BetRound       `json:"round"`
	Bet           Bet            `json:"user_bet"`
	User          User           `json:"user"`
	FailedBetLogs *FailedBetLogs `json:"failed_bet_logs,omitempty"`
}

type GetFailedRoundsRes struct {
	Message string                `json:"message"`
	Data    []FailedRoundsResData `json:"data"`
}

type ManualRefundFailedRoundsReq struct {
	UserID  uuid.UUID `json:"user_id"`
	AdminID uuid.UUID `json:"admin_id"`
	RoundID uuid.UUID `json:"round_id"`
}

type ManualRefundFailedRoundsData struct {
	UserID        uuid.UUID       `json:"user_id"`
	BetID         uuid.UUID       `json:"bet_id"`
	RefundAmount  decimal.Decimal `json:"refund_amount"`
	TransactionID string          `json:"transaction_id"`
}

type ManualRefundFailedRoundsRes struct {
	Message string                       `json:"message"`
	Data    ManualRefundFailedRoundsData `json:"data"`
}

// plinko bet
type PlinkoBetLimits struct {
	Min decimal.Decimal `json:"min"`
	Max decimal.Decimal `json:"max"`
}
type PlinkoGameConfig struct {
	BetLimits   PlinkoBetLimits   `json:"bet_limits"`
	Rtp         decimal.Decimal   `json:"rtp"`
	Multipliers []decimal.Decimal `json:"multipliers"`
}

type PlacePlinkoGameReq struct {
	UserID   uuid.UUID       `json:"user_id"  swaggerignore:"true"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

type PlacePlinkoGameRes struct {
	ID            uuid.UUID       `json:"id"`
	Timestamp     time.Time       `json:"timestamp"`
	BetAmount     decimal.Decimal `json:"bet_amount"`
	DropPath      []PathEntry     `json:"drop_path"`
	WinAmount     decimal.Decimal `json:"win_amount"`
	Multiplier    decimal.Decimal `json:"multiplier"`
	FinalPosition int             `json:"final_position"`
}

type PlacePlinkoGame struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Timestamp     time.Time       `json:"timestamp"`
	BetAmount     decimal.Decimal `json:"bet_amount"`
	DropPath      map[int]int     `json:"drop_path"`
	Multiplier    decimal.Decimal `json:"multiplier"`
	WinAmount     decimal.Decimal `json:"win_amount"`
	FinalPosition int             `json:"final_position"`
}

type Board struct {
	Rows        int
	Slots       int
	Pegs        [][]int
	Multipliers []decimal.Decimal
}

type PathEntry struct {
	Row int
	Col int
}

type PlinkoBetHistoryReq struct {
	UserID  uuid.UUID `form:"user_id" swaggerignore:"true"`
	Page    int       `form:"page"`
	PerPage int       `form:"per_page"`
}

type PlinkoBetHistoryRes struct {
	TotalPages int               `json:"total_pages"`
	Games      []PlacePlinkoGame `json:"games"`
}

type HighestWin struct {
	ID         uuid.UUID       `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	BetAmount  decimal.Decimal `json:"bet_amount"`
	Multiplier decimal.Decimal `json:"multiplier"`
	WinAmount  decimal.Decimal `json:"win_amount"`
	PinCount   decimal.Decimal `json:"pin_count"`
}

type PlinkoGameStatRes struct {
	TotalGames        int             `json:"total_games"`
	TotalWagered      decimal.Decimal `json:"total_wagered"`
	TotalWon          decimal.Decimal `json:"total_won"`
	NetProfit         decimal.Decimal `json:"net_profit"`
	AverageMultiplier decimal.Decimal `json:"average_multiplier"`
	HighestWin        HighestWin      `json:"highest_win"`
}

// football match bet
type League struct {
	ID         uuid.UUID `json:"id"`
	LeagueName string    `json:"league_name"`
	Timestamp  time.Time `json:"timestamp"`
	Status     string    `json:"status"`
}

type GetRequest struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Type    string `form:"type"` // "players" or "squads"
}

type GetLeagueRes struct {
	TotalPages int      `json:"total_pages"`
	Leagues    []League `json:"leagues"`
}

type Club struct {
	ID        uuid.UUID `json:"club"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type GetClubRes struct {
	TotalPages int    `json:"total_pages"`
	Clubs      []Club `json:"clubs"`
}

type FootballMatchReq struct {
	ID        uuid.UUID `json:"id" swaggerignore:"true"`
	RoundID   uuid.UUID `json:"round_id"`
	HomeTeam  uuid.UUID `json:"home_team"`
	AwayTeam  uuid.UUID `json:"away_team"`
	LeagueID  uuid.UUID `json:"league_id"`
	Status    string    `json:"status"`
	WinnerID  uuid.UUID `json:"winner_id"`
	MatchDate time.Time `json:"match_date"`
}

type FootballMatch struct {
	ID         uuid.UUID `json:"id" swaggerignore:"true"`
	RoundID    uuid.UUID `json:"round_id"`
	HomeTeam   string    `json:"home_team"`
	AwayTeam   string    `json:"away_team"`
	MatchDate  time.Time `json:"match_date"`
	LeagueID   string    `json:"league_id"`
	LeagueName string    `json:"league_name"`
	Status     string    `json:"status"`
	WinnerID   string    `json:"winner_id"`
	HomeScore  int       `json:"home_score"`
	AwayScore  int       `json:"away_score"`
}

type FootballMatchRound struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type FootballMatchRoundUpdateReq struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

type GetFootballMatchRoundsByStatusReq struct {
	Status  string `json:"status"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

type GetFootballMatchRoundRes struct {
	TotalPages int                  `json:"total_pages"`
	Rounds     []FootballMatchRound `json:"rounds"`
}
type FootballCardMultiplier struct {
	ID             uuid.UUID       `json:"id" swaggerignore:"true"`
	CardMultiplier decimal.Decimal `json:"card_multiplier"`
}

type GetFootballRoundMatchesReq struct {
	RoundID uuid.UUID `json:"round_id"`
	Page    int       `json:"page"`
	PerPage int       `json:"per_page"`
}

type GetFootballRoundMatchesRes struct {
	TotalPages int                `json:"total_pages"`
	Round      FootballMatchRound `json:"round"`
	Matches    []FootballMatch    `json:"matches"`
}

type CloseFootballMatchReq struct {
	ID        uuid.UUID `json:"id"`
	HomeScore int       `json:"home_score"`
	AwayScore int       `json:"away_score"`
	Winner    string    `json:"winner" waggerignore:"true"`
}

type UpdateFootballBetPriceReq struct {
	Price decimal.Decimal `json:"price"`
}

type UpdateFootballBetPriceRes struct {
	Message string `json:"message"`
	Price   string `json:"price"`
}

type Fixture struct {
	ID uuid.UUID `json:"id"`
}

type PlaceFootballMatchBetReq struct {
	ID uuid.UUID `json:"id"`
}

type UserFootballMatcheRound struct {
	ID              uuid.UUID       `json:"id"`
	Status          string          `json:"status"`
	WonStatus       string          `json:"won_status"`
	UserID          uuid.UUID       `json:"user_id"`
	FootballRoundID uuid.UUID       `json:"football_round_id"`
	BetAmount       decimal.Decimal `json:"bet_amount"`
	WonAmount       decimal.Decimal `json:"won_amount"`
	Currency        string          `json:"currency"`
}

type UserFootballMatchSelection struct {
	ID                        uuid.UUID `json:"id"`
	Status                    string    `json:"status"`
	MatchID                   uuid.UUID `json:"match_id"`
	Selection                 string    `json:"selection"`
	UsersFootballMatchRoundID uuid.UUID `json:"users_football_match_round_id"`
}

type FootballMatchSelection struct {
	MatchID   uuid.UUID `json:"match_id"`
	Selection string    `json:"selection"`
}

type UserFootballMatchBetReq struct {
	UserID     uuid.UUID                `json:"user_id" swaggerignore:"true"`
	Currency   string                   `json:"currency"`
	RoundID    uuid.UUID                `json:"round_id"`
	Selections []FootballMatchSelection `json:"selections"`
}

type UserFootballMatchBetRes struct {
	Message string                  `json:"message"`
	Data    UserFootballMatchBetReq `json:"data"`
}
type FootballMatchRes struct {
	ID         uuid.UUID `json:"id"`
	Status     string    `json:"status"`
	Selection  string    `json:"selection"`
	LeagueName string    `json:"league_name"`
	HomeTeam   string    `json:"home_team"`
	AwayTeam   string    `json:"away_team"`
	MatchDate  time.Time `json:"match_date"`
	Winner     string    `json:"winner"`
}

type FootballRoundsRes struct {
	ID            uuid.UUID          `json:"id"`
	BetAmount     decimal.Decimal    `json:"bet_amount"`
	WinningAmount decimal.Decimal    `json:"winning_amount"`
	RoundStatus   string             `json:"round_status"`
	Currency      string             `json:"currency"`
	Matches       []FootballMatchRes `json:"matches"`
}

type GetUserFootballBetRes struct {
	Message    string              `json:"message"`
	TotalPages int                 `json:"total_pages"`
	Rounds     []FootballRoundsRes `json:"matches"`
}

// crash_kings
type CreateCrashKingsReq struct {
	Version   string          `json:"version"`
	BetAmount decimal.Decimal `json:"bet_amount"`
}

type CreateStreetKingsReqData struct {
	CreateCrashKingsReq CreateCrashKingsReq `json:"create_crash_kings_req"`
	UserID              uuid.UUID           `json:"user_id" `
	CrashPoint          decimal.Decimal     `json:"crash_point"`
	Status              string              `json:"status"`
	Timestamp           time.Time           `json:"timestamp" `
}

type CreateStreetKingsRespData struct {
	ID         uuid.UUID       `json:"id"`
	Version    string          `json:"version"`
	BetAmount  decimal.Decimal `json:"bet_amount"`
	CrashPoint decimal.Decimal `json:"crash_point" ignoreswagger:"true"`
	Status     string          `json:"status"`
	Timestamp  time.Time       `json:"timestamp"`
}

type CreateStreetKingsResp struct {
	Message string                `json:"message"`
	Data    CreateStreetKingsData `json:"data"`
}
type CreateStreetKingsData struct {
	RoundID   uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Version   string          `json:"version"`
	BetAmount decimal.Decimal `json:"bet_amount"`
	Timestamp time.Time       `json:"timestamp"`
}

type StreetKingsCashoutReq struct {
	ID           uuid.UUID       `json:"id"`
	Status       string          `json:"status"`
	WonAmount    decimal.Decimal `json:"won_amount"`
	CashoutPoint decimal.Decimal `json:"cashout_point"`
}

type StreetKingsCrashResp struct {
	Message string                   `json:"message"`
	Data    StreetKingsCrashRespData `json:"data"`
}

type StreetKingsCrashRespData struct {
	ID           uuid.UUID       `json:"id"`
	UserID       uuid.UUID       `json:"user_id"`
	Status       string          `json:"status"`
	Version      string          `json:"version"`
	BetAmount    decimal.Decimal `json:"bet_amount"`
	WonAmount    decimal.Decimal `json:"won_amount"`
	CrashPoint   decimal.Decimal `json:"crash_point"`
	CashoutPoint decimal.Decimal `json:"cashout_point"`
	Timestamp    time.Time       `json:"timestamp"`
	WonStatus    string          `json:"won_status,omitempty"`
}

type GetStreetkingHistoryReq struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Version string `form:"version"`
}

type GetStreetkingHistoryRes struct {
	Message    string                     `json:"message"`
	TotalPages int                        `json:"total_pages"`
	Data       []StreetKingsCrashRespData `json:"data"`
}

// crypto kings
// also for potential win m = 2n-1
type UpdateCryptoKingsConfigReq struct {
	CryptoKingsRangeMaxValue decimal.Decimal `json:"crypto_kings_range_max_value"` // for r = v/n  / where n is the range
	CryptoKingsTimeMaxValue  decimal.Decimal `json:"crypto_kings_time_max_value"`  // for 1 second  r = floor(v/n)
}

type UpdateCryptokingsConfigRes struct {
	Message string                         `json:"message"`
	Data    UpdateCryptoKingsConfigResData `json:"data"`
}

type UpdateCryptoKingsConfigResData struct {
	CryptoKingsRangeMaxValue decimal.Decimal `json:"crypto_kings_range_max_value"` // for r = v/n  / where n is the range
	CryptoKingsTimeMaxValue  decimal.Decimal `json:"crypto_kings_time_max_value"`  // for 1 second  r = floor(v/n)
}

type PlaceCryptoKingsBetReq struct {
	Type      string          `json:"type"`
	BetAmount int64           `json:"bet_amount"`
	MinValue  decimal.Decimal `json:"min_value"` // set only if type is range
	MaxValue  decimal.Decimal `json:"max_value"` // set only if type is range
	Second    int             `json:"second"`    // set only if type is time
	Timestamp time.Time       `json:"timestamp" swaggerignore:"true"`
}

type PlaceCryptoKingsBetRes struct {
	Message string                     `json:"message"`
	Data    PlaceCryptoKingsBetResData `json:"data"`
}

type PlaceCryptoKingsBetResData struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	BetAmount          int64     `json:"bet_amount"`
	Type               string    `json:"type"`
	PotentialWinAmount int64     `json:"potential_win_amount"`
	Timestamp          time.Time `json:"timestamp"`
}

type CreateCryptoKingData struct {
	ID                 uuid.UUID       `json:"id"`
	UserID             uuid.UUID       `json:"user_id"`
	Status             string          `json:"status"`
	BetAmount          int64           `json:"bet_amount"`
	WonAmount          decimal.Decimal `json:"won_amount"`
	StartCryptoValue   decimal.Decimal `json:"start_crypto_value"`
	EndCryptoValue     decimal.Decimal `json:"end_crypto_value"`
	SelectedEndSecond  int             `json:"selected_end_second"`
	SelectedStartValue decimal.Decimal `json:"selected_start_value"`
	SelectedEndValue   decimal.Decimal `json:"selected_end_value"`
	WonStatus          string          `json:"won_status"`
	Type               string          `json:"type"`
	Timestamp          time.Time       `json:"timestamp"`
}

type TradingState struct {
	CurrentValue decimal.Decimal   `json:"current_value"`
	History      []decimal.Decimal `json:"history"`
	StartTime    time.Time         `json:"start_time"`
}

type StreamTradingReq struct {
	ID          uuid.UUID                `json:"id"`
	UserID      uuid.UUID                `json:"user_id"`
	StartValue  decimal.Decimal          `json:"start_value"`
	TargetValue decimal.Decimal          `json:"traget_value"`
	TypeSecond  bool                     `json:"type_second"`
	Second      int                      `json:"second"`
	UserConns   map[*websocket.Conn]bool `json:"user_conns"`
	WonStatus   string                   `json:"won_status"`
	WonAmount   int                      `json:"won_amount"`
}

type StreamTradingRes struct {
	ID           uuid.UUID       `json:"id"`
	CurrentValue decimal.Decimal `json:"current_value"`
	WonStatus    string          `json:"won_status"`
	WonAmount    int             `json:"won_amount,omitempty"`
}

type GetCryptoKingsUserBetHistoryResData struct {
	UserID     uuid.UUID              `json:"user_id"`
	TotalPages int                    `json:"total_pages"`
	Histories  []CreateCryptoKingData `json:"histories"`
}

type GetCryptoKingsUserBetHistoryRes struct {
	Message string                              `json:"message"`
	Data    GetCryptoKingsUserBetHistoryResData `json:"data"`
}

type GetCryptoCurrencyPriceResp struct {
	Message string          `json:"message"`
	Price   decimal.Decimal `json:"price"`
}

// quick hustle
type CreateQuickHustleBetReq struct {
	ID         uuid.UUID       `json:"id" swaggerignore:"true"`
	UserID     uuid.UUID       `json:"user_id" swaggerignore:"true"`
	BetAmount  decimal.Decimal `json:"bet_amount"`
	FirstCard  string          `json:"first_card" swaggerignore:"true"`
	SecondCard string          `json:"second_card" swaggerignore:"true"`
	Timestamp  time.Time       `json:"timestamp" swaggerignore:"true"`
}

// quick hustle response
type CreateQuickHustelBetResData struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	BetAmount  decimal.Decimal `json:"bet_amount"`
	FirstCard  string          `json:"first_card"`
	SecondCard string          `json:"second_card,omitempty" swaggerignore:"true"`
	Status     string          `json:"status"`
	Timestamp  time.Time       `json:"timestamp"`
}

type CreateQuickHustelBetRes struct {
	Message string                      `json:"message"`
	Data    CreateQuickHustelBetResData `json:"data"`
}

type SelectQuickHustlePossibilityReq struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	UserGuess string    `json:"user_guess"`
}

type CloseQuickHustleBetData struct {
	ID         uuid.UUID       `json:"id"`
	UserGuess  string          `json:"user_guess"`
	WonStatus  string          `json:"won_status"`
	SecondCard string          `json:"second_card"`
	WonAmount  decimal.Decimal `json:"won_amount"`
}

type QuickHustelBetResData struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Status      string          `json:"status"`
	BetAmount   decimal.Decimal `json:"bet_amount"`
	WonStatus   string          `json:"won_status"`
	UserGuessed string          `json:"user_guessed"`
	FirstCard   string          `json:"first_card"`
	SecondCard  string          `json:"second_card"`
	Timestamp   time.Time       `json:"timestamp"`
	WonAmount   decimal.Decimal `json:"won_amount"`
}

type CloseQuickHustleResp struct {
	Message string                `json:"message"`
	Data    QuickHustelBetResData `json:"data"`
}

type QuickHustleCard struct {
	Name  string
	Value int
}

type QuickHustleGetHistoryData struct {
	UserID     uuid.UUID               `json:"user_id"`
	TotalPages int                     `json:"total_pages"`
	Histories  []QuickHustelBetResData `json:""`
}
type GetQuickHustleResp struct {
	Message string                    `json:"message"`
	Data    QuickHustleGetHistoryData `json:"data"`
}

// roll da dice
type CreateRollDaDiceReq struct {
	UserID              uuid.UUID       `json:"user_id" swaggerignore:"true"`
	BetAmount           decimal.Decimal `json:"bet_amount"`
	WonAmount           decimal.Decimal `json:"won_amount" swaggerignore:"true"`
	CrashPoint          decimal.Decimal `json:"crash_point" swaggerignore:"true"`
	Timestamp           time.Time       `json:"timestamp"`
	UserGuessStartPoint decimal.Decimal `json:"user_guess_start_point"`
	UserGuessEndPoint   decimal.Decimal `json:"user_guess_end_point"`
	WonStatus           string          `json:"won_status"`
	Multiplier          decimal.Decimal `json:"multiplier" swaggerignore:"true"`
}
type RollDaDiceData struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	BetAmount           decimal.Decimal `json:"bet_amount"`
	Status              string          `json:"status"`
	CrashPoint          decimal.Decimal `json:"crash_point,omitempty"`
	Timestamp           time.Time       `json:"timestamp"`
	UserGuessStartPoint decimal.Decimal `json:"user_guess_start_point"`
	UserGuessEndPoint   decimal.Decimal `json:"user_guess_end_point"`
	WonStatus           string          `json:"won_status,omitempty"`
	WonAmount           decimal.Decimal `json:"won_amount,omitempty"`
	Multiplier          decimal.Decimal `json:"multiplier,omitempty"`
}

type RollDaDiceRespData struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	BetAmount           decimal.Decimal `json:"bet_amount"`
	Timestamp           time.Time       `json:"timestamp"`
	UserGuessStartPoint decimal.Decimal `json:"user_guess_start_point"`
	UserGuessEndPoint   decimal.Decimal `json:"user_guess_end_point"`
	Multiplier          decimal.Decimal `json:"multiplier,omitempty"`
}
type CreateRollDaDiceResp struct {
	Message string             `json:"message"`
	Data    RollDaDiceRespData `json:"data"`
}

type StreamRollDaDiceData struct {
	ID           uuid.UUID       `json:"id"`
	CurrentValue decimal.Decimal `json:"current_value"`
	WonStatus    string          `json:"won_status"`
	WonAmount    decimal.Decimal `json:"won_amount,omitempty"`
}

type GetRollDaDiceRespData struct {
	TotalPages int              `json:"total_pages"`
	UserID     uuid.UUID        `json:"user_id"`
	Histories  []RollDaDiceData `json:"histories"`
}

type GetRollDaDiceResp struct {
	Message string                `json:"message"`
	Data    GetRollDaDiceRespData `json:"data"`
}

// scratch card
type GetScratchCardRes struct {
	Message  string          `json:"message"`
	Price    decimal.Decimal `json:"price"`
	MaxPrize decimal.Decimal `json:"max_prize"`
}

type ScratchCard struct {
	Board         [3][3]string    `json:"board"`
	Symbols       []string        `json:"symbols"`
	WinningSymbol string          `json:"winning_symbol"`
	Prize         decimal.Decimal `json:"prize"`
	MatchCells    [][2]int        `json:"match_cells"`
	BetAmount     decimal.Decimal `json:"bet_amount"`
}

type ScratchCardsBetData struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Status    string          `json:"status"`
	BetAmount decimal.Decimal `json:"bet_amount"`
	WonStatus string          `json:"won_status"`
	Timestamp time.Time       `json:"timestamp"`
	WonAmount decimal.Decimal `json:"won_amount"`
}

type ScratchCardsBetRespData struct {
	TotalPages int                   `json:"total_pages"`
	UserID     uuid.UUID             `json:"user_id"`
	Histories  []ScratchCardsBetData `json:"histories"`
}

type GetScratchBetHistoriesResp struct {
	Message string                  `json:"message"`
	Data    ScratchCardsBetRespData `json:"data"`
}

// spinning wheel
type GetSpinningWheelPrice struct {
	Message string          `json:"message"`
	Price   decimal.Decimal `json:"price"`
}
type PlaceSpinningWheelData struct {
	ID        uuid.UUID `json:"id"`
	BetAmount string    `json:"bet_amount"`
	Prize     string    `json:"prize"`
}
type PlaceSpinningWheelResp struct {
	Message string                 `json:"message"`
	Data    PlaceSpinningWheelData `json:"data"`
}

type SpinningWheelData struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	BetAmount string    `json:"bet_amount"`
	WonStatus string    `json:"won_status"`
	WonAmount string    `json:"won_amount"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}
type GetSpinningWheelData struct {
	TotalPages int                 `json:"total_pages"`
	Histories  []SpinningWheelData `json:"histories"`
}
type GetSpinningWheelHistoryResp struct {
	Message string               `json:"message"`
	Data    GetSpinningWheelData `json:"data"`
}

type Game struct {
	ID      uuid.UUID `json:"id"`
	Status  string    `json:"status"`
	Name    string    `json:"name"`
	Photo   string    `json:"photo"`
	Enabled bool      `json:"enabled"`
}

type GetGamesData struct {
	TotalPages int    `json:"total_pages"`
	Games      []Game `json:"games"`
}
type GetGamesResp struct {
	Message string       `json:"message"`
	Data    GetGamesData `json:"data"`
}

type BlockGamesResp struct {
	Message string `json:"message"`
	Data    []Game `json:"data"`
}

type GetAdminsReq struct {
	RoleID  uuid.UUID `form:"role_id"`
	Status  string    `form:"status"`
	PerPage int       `form:"per_page"`
	Page    int       `form:"page"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

type ScratchCardConfig struct {
	Name  string          `json:"name"`
	Prize decimal.Decimal `json:"prize"`
	Id    uuid.UUID       `json:"id"`
}

type GetScratchCardConfigs struct {
	Message string              `json:"message"`
	Data    []ScratchCardConfig `json:"data"`
}

type UpdateScratchGameConfigRequest struct {
	Name  string          `json:"name"`
	Prize decimal.Decimal `json:"prize"`
	Id    uuid.UUID       `json:"id"`
}

type UpdateScratchGameConfigResponse struct {
	Message string `json:"message"`
}

var QuickHustelCards = []struct {
	Name   string
	Weight int
}{
	{"A", 1}, {"B", 2}, {"2", 2}, {"C", 3}, {"3", 3}, {"D", 4}, {"4", 4},
	{"E", 5}, {"5", 5}, {"F", 6}, {"6", 6}, {"G", 7}, {"7", 7}, {"H", 8},
	{"8", 8}, {"I", 9}, {"9", 9}, {"J", 10}, {"10", 10}, {"K", 11}, {"L", 12},
	{"M", 13}, {"N", 14}, {"O", 15}, {"P", 16}, {"Q", 17}, {"R", 18}, {"S", 19},
	{"T", 20}, {"U", 21}, {"V", 22}, {"W", 23}, {"X", 24}, {"Y", 25}, {"Z", 26},
}
var QuickHustelLowerMultiplier = []struct {
	Name       string
	Multiplier decimal.Decimal
}{
	{"A", decimal.Zero}, {"B", decimal.NewFromFloat(32.3)}, {"2", decimal.NewFromFloat(32.3)},
	{"C", decimal.NewFromFloat(10.77)}, {"3", decimal.NewFromFloat(10.77)}, {"D", decimal.NewFromFloat(6.46)}, {"4", decimal.NewFromFloat(6.46)},
	{"E", decimal.NewFromFloat(4.61)}, {"5", decimal.NewFromFloat(4.61)}, {"F", decimal.NewFromFloat(3.59)}, {"6", decimal.NewFromFloat(3.59)},
	{"G", decimal.NewFromFloat(2.94)}, {"7", decimal.NewFromFloat(2.94)}, {"H", decimal.NewFromFloat(2.48)}, {"8", decimal.NewFromFloat(2.48)},
	{"I", decimal.NewFromFloat(2.15)}, {"9", decimal.NewFromFloat(2.15)}, {"J", decimal.NewFromFloat(1.9)}, {"10", decimal.NewFromFloat(1.9)},
	{"K", decimal.NewFromFloat(1.7)}, {"L", decimal.NewFromFloat(1.61)}, {"M", decimal.NewFromFloat(1.54)}, {"N", decimal.NewFromFloat(1.47)},
	{"O", decimal.NewFromFloat(1.4)}, {"P", decimal.NewFromFloat(1.35)}, {"Q", decimal.NewFromFloat(1.29)}, {"R", decimal.NewFromFloat(1.24)},
	{"S", decimal.NewFromFloat(1.2)}, {"T", decimal.NewFromFloat(1.15)}, {"U", decimal.NewFromFloat(1.11)}, {"V", decimal.NewFromFloat(1.08)},
	{"W", decimal.NewFromFloat(1.04)}, {"X", decimal.NewFromFloat(1.01)}, {"Y", decimal.NewFromFloat(0.98)}, {"Z", decimal.NewFromFloat(0.95)},
}

var QuickHustelHigherMultiplier = []struct {
	Name       string
	Multiplier decimal.Decimal
}{
	{"A", decimal.NewFromFloat(0.95)}, {"B", decimal.NewFromFloat(1.01)}, {"2", decimal.NewFromFloat(1.01)},
	{"C", decimal.NewFromFloat(1.08)}, {"3", decimal.NewFromFloat(1.08)}, {"D", decimal.NewFromFloat(1.15)}, {"4", decimal.NewFromFloat(1.15)},
	{"E", decimal.NewFromFloat(1.24)}, {"5", decimal.NewFromFloat(1.24)}, {"F", decimal.NewFromFloat(1.35)}, {"6", decimal.NewFromFloat(1.35)},
	{"G", decimal.NewFromFloat(1.47)}, {"7", decimal.NewFromFloat(1.47)}, {"H", decimal.NewFromFloat(1.61)}, {"8", decimal.NewFromFloat(1.61)},
	{"I", decimal.NewFromFloat(1.79)}, {"9", decimal.NewFromFloat(1.79)}, {"J", decimal.NewFromFloat(2.02)}, {"10", decimal.NewFromFloat(2.02)},
	{"K", decimal.NewFromFloat(2.15)}, {"L", decimal.NewFromFloat(2.31)}, {"M", decimal.NewFromFloat(2.48)}, {"N", decimal.NewFromFloat(2.69)},
	{"O", decimal.NewFromFloat(2.94)}, {"P", decimal.NewFromFloat(3.23)}, {"Q", decimal.NewFromFloat(3.59)}, {"R", decimal.NewFromFloat(4.04)},
	{"S", decimal.NewFromFloat(4.61)}, {"T", decimal.NewFromFloat(5.38)}, {"U", decimal.NewFromFloat(6.46)}, {"V", decimal.NewFromFloat(8.07)},
	{"W", decimal.NewFromFloat(10.77)}, {"X", decimal.NewFromFloat(16.15)}, {"Y", decimal.NewFromFloat(32.3)}, {"Z", decimal.Zero},
}

type AddFakeBalanceLogReq struct {
	UserID uuid.UUID       `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
}

type FakeBalanceLogResp struct {
	Message string `json:"message"`
}
