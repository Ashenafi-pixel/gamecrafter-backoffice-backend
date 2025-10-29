package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSquadsReq struct {
	Handle    string    `json:"handle"`
	Type      string    `json:"type"`
	Owner     uuid.UUID `json:"owner" swaggerignore:"true"`
	CreatedAt time.Time `json:"created_at"`
}

type DeleteSquadReq struct {
	ID uuid.UUID `json:"id"`
}
type DeleteSquadMemberReq struct {
	MemberID uuid.UUID `json:"member_id"`
}
type Squad struct {
	ID     uuid.UUID `json:"id"`
	Handle string    `json:"handle"`
	Type   string    `json:"type"`
	Owner  uuid.UUID `json:"owner"`
}

type Owener struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
}
type GetSquadsResp struct {
	ID            uuid.UUID             `json:"id"`
	Handle        string                `json:"handle"`
	Type          string                `json:"type"`
	Owener        Owener                `json:"user"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	SquadMemebers []GetSquadMembersResp `json:"squad_memebers"`
}

type CreateSquadsRes struct {
	Message string `json:"message"`
	Squad   Squad  `json:"squad"`
}

type SquadMember struct {
	ID        uuid.UUID `json:"id"`
	SquadID   uuid.UUID `json:"squad_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SquadMemberResp struct {
	Message string      `json:"message"`
	Data    SquadMember `json:"data"`
}

type CreateSquadReq struct {
	SquadID uuid.UUID `json:"squad_id"`
	UserID  uuid.UUID `json:"user_id"`
}

type CreateSquadMemeberReq struct {
	SquadID uuid.UUID `json:"squad_id"`
	Phone   string    `json:"phone"`
	OwnerID uuid.UUID `json:"owner_id" swaggerignore:"true"`
}

type GetSquadMembersResp struct {
	ID        uuid.UUID `json:"id"`
	SquadID   uuid.UUID `json:"squad_id"`
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SquadEarns struct {
	SquadID   uuid.UUID       `json:"squad_id"`
	UserID    uuid.UUID       `json:"user_id"`
	Currency  string          `json:"currency"`
	Earn      decimal.Decimal `json:"earn"`
	GameID    uuid.UUID       `json:"game_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdateAt  time.Time       `json:"updated_at"`
}

type GetSquadEarnsResp struct {
	TotalPages  int          `json:"total_pages"`
	SquadsEarns []SquadEarns `json:"squad_earns"`
}

type GetSquadEarnsReq struct {
	Page    int       `form:"page"`
	PerPage int       `form:"per_page"`
	SquadID uuid.UUID `form:"squad_id"`
}

type GetSquadMemebersReq struct {
	Page    int       `form:"page"`
	PerPage int       `form:"per_page"`
	SquadID uuid.UUID `form:"squad_id"`
}

type GetSquadMemebersRes struct {
	TotalPages    int                   `json:"total_pages"`
	SquadMemebers []GetSquadMembersResp `json:"squad_memebers"`
}
type GetSquadBalanceReq struct {
	SquadID uuid.UUID `form:"squad_id"`
}

type GetSquadBalanceRes struct {
	Earns decimal.Decimal `json:"earns"`
}

type SquadMemberEarnings struct {
	UserID       uuid.UUID       `json:"user_id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	PhoneNumber  string          `json:"phone_number"`
	TotalEarned  decimal.Decimal `json:"total_earned"`
	TotalGames   int64           `json:"total_games"`
	LastEarnedAt *time.Time      `json:"last_earned_at,omitempty"`
}

type GetSquadMembersEarningsReq struct {
	Page    int       `form:"page"`
	PerPage int       `form:"per_page"`
	SquadID uuid.UUID `form:"squad_id"`
}

type GetSquadMembersEarningsResp struct {
	TotalPages      int                   `json:"total_pages"`
	Members_Earning []SquadMemberEarnings `json:"members_earning"`
}

type GetSquadsByTypeReq struct {
	Type string `form:"type"`
}

type GetSquadsByTypeResp struct {
	Squads []Squad `json:"squads"`
}

// enum Squad type dropdown: Open, request to join, invite-only
type SquadType string

const (
	SquadTypeOpen          SquadType = "Open"
	SquadTypeRequestToJoin SquadType = "Request to Join"
	SquadTypeInviteOnly    SquadType = "Invite Only"
)

// CheckInvalidSquadType validates that input is NOT one of the defined SquadTypes
func CheckInvalidSquadType(input string) error {
	validTypes := map[SquadType]bool{
		SquadTypeOpen:          true,
		SquadTypeRequestToJoin: true,
		SquadTypeInviteOnly:    true,
	}

	if !validTypes[SquadType(input)] {
		return errors.New("invalid squad type: value is not allowed")
	}
	return nil
}

func CheckJoinableSquadType(input string) error {
	validTypes := map[SquadType]bool{
		SquadTypeOpen:          true,
		SquadTypeRequestToJoin: true,
	}
	if !validTypes[SquadType(input)] {

		return errors.New("invalid squad type: value is not allowed for joining")
	}
	return nil
}

type GetTornamentStyleRankingReq struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

type GetTornamentStyleRank struct {
	Rank       int64           `json:"rank"`
	Handle     string          `json:"handle"`
	TotalEarns decimal.Decimal `json:"total_earns"`
}

type GetTornamentStyleRankingResp struct {
	TotalPages int                     `json:"total_pages"`
	Ranking    []GetTornamentStyleRank `json:"ranking"`
}

type Reward struct {
	Type   string          `json:"type"`
	Amount decimal.Decimal `json:"amount"`
}
type CreateTournamentReq struct {
	Rank             string    `json:"rank"`
	Level            int       `json:"level"`
	CumulativePoints int32     `json:"cumulative_points"`
	Rewards          []Reward  `json:"rewards"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Tournament struct {
	ID               uuid.UUID `json:"id"`
	Rank             string    `json:"rank"`
	Level            int       `json:"level"`
	CumulativePoints int32     `json:"cumulative_points"`
	Rewards          []Reward  `json:"rewards"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateTournamentResp struct {
	Message    string     `json:"message"`
	Tournament Tournament `json:"tournament"`
}

var RewardTypes = []string{"bucks"}

type TournamentClaim struct {
	ID           uuid.UUID `json:"id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	SquadID      uuid.UUID `json:"squad_id"`
	ClaimedAt    time.Time `json:"claimed_at"`
}

type GetSquadMemberByIDresp struct {
	ID        uuid.UUID `json:"id"`
	Handle    string    `json:"handle"`
	Owner     uuid.UUID `json:"owner"`
	SquadID   uuid.UUID `json:"squad_id"`
	CreatedAt time.Time `json:"created_at"`
}

type LeaveSquadReq struct {
	UserID  uuid.UUID `json:"user_id"`
	SquadID uuid.UUID `json:"squad_id"`
}

type LeaveSquadResp struct {
	Message string `json:"message"`
}

type WaitingSquadMember struct {
	ID        uuid.UUID `json:"id"`
	SquadID   uuid.UUID `json:"squad_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
}

type GetWaitingSquadMembersReq struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

type GetWaitingSquadMembersResp struct {
	TotalPages     int                  `json:"total_pages"`
	WaitingMembers []WaitingSquadMember `json:"waiting_members"`
}

type WaitingsquadOwner struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	OwnerID   uuid.UUID `json:"owner_id"`
}

type SquadMemberReq struct {
	MemberID uuid.UUID `json:"member_id"`
}
