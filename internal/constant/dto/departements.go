package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Department struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Notifications []string  `json:"notifications"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateDepartementReq struct {
	Name          string    `json:"name"  validate:"required"`
	Notifications []string  `json:"notifications"`
	CreatedAt     time.Time `json:"created_at" swaggerignore:"true"`
}
type CreateDepartementRes struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Notifications []string  `json:"notifications"`
	CreatedAt     time.Time `json:"created_at"`
}

type GetDepartementsReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type Departement struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Notifications []string  `json:"notifications"`
	CreatedAt     time.Time `json:"created_at"`
}
type GetDepartementsRes struct {
	Departements []Departement `json:"departements"`
	TotalPages   int           `json:"total_pages"`
}

func ValidateCreateDepartement(p CreateDepartementReq) error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateDepartment struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Notifications []string  `json:"notifications"`
	CreatedAt     time.Time `json:"created_at"`
}

type AssignDepartmentToUserReq struct {
	UserID       uuid.UUID `json:"user_id"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type AssignDepartmentToUserResp struct {
	Message      string    `json:"message"`
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	DepartmentID uuid.UUID `json:"department_id"`
}

type GetUserDepartmentRes struct {
	Department Department `json:"department"`
	User       User       `json:"user"`
}
