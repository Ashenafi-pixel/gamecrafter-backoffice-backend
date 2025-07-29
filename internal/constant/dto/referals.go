package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReferalMultiplierReq struct {
	PointMultiplier int `json:"point_multiplier"`
}

type ReferalData struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	PointMultiplier int       `json:"point_multiplier"`
}

type ReferalMultiplierRes struct {
	Message string      `json:"message"`
	Data    ReferalData `json:"data"`
}

type UserPoint struct {
	UserID uuid.UUID `json:"user_id"`
	Point  int       `json:"point"`
}

type ReferalUpdateResp struct {
	Message         string `json:"message"`
	PointMultiplier int    `json:"point_multiplier"`
}

type UpdateReferralPointReq struct {
	Multiplier int `json:"multiplier"`
}
type MassReferralReq struct {
	UserID uuid.UUID `json:"user_id"`
	Point  int       `json:"point"`
}

type MassReferralResData struct {
	UserID uuid.UUID `json:"user_id"`
	Point  int       `json:"point"`
}

type MassReferralRes struct {
	Message      string                `json:"message"`
	UpdatedUsers []MassReferralResData `json:"updated_users"`
}

type AdminInfo struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Profile   string    `json:"profile"`
}

type GetAdminAssignedPointsData struct {
	Transaction string    `json:"transaction_id"`
	Admin       AdminInfo `json:"admin"`
	Amount      int       `json:"amount"`
	User        User      `json:"player"`
}

type GetAdminAssignedPointsReq struct {
	PerPage int `form:"per-page"`
	Page    int `form:"page"`
}

type GetAdminAssignedPointsRes struct {
	Message             string                       `json:"message"`
	AdminAssignedPoints []GetAdminAssignedPointsData `json:"admin_assigned_points"`
	TotalPages          int32                        `json:"total_pages"`
}

type GetAdminAssignedData struct {
	UserID            uuid.UUID `json:"user_id"`
	AdminID           uuid.UUID `json:"admin_id"`
	AddedPoints       int       `json:"added_points"`
	Timestamp         time.Time `json:"timestamp"`
	PointsAfterUpdate int       `json:"points_after_update"`
	TransactionID     string    `json:"transaction_id"`
}

type GetAdminAssignedResp struct {
	Data      []GetAdminAssignedData `json:"data"`
	TotalPage int                    `json:"total_pages"`
}
