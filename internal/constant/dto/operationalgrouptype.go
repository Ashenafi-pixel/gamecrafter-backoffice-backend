package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OperationalGroupType struct {
	ID          uuid.UUID `json:"id"`
	GroupID     uuid.UUID `json:"group_id" `
	Name        string    `json:"name" validate:"required,max=50"`
	Description string    `json:"description" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

type OperationalTypesRes struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func ValidateOperationalGroupType(opg OperationalGroupType) error {
	validate := validator.New()
	return validate.Struct(opg)
}
