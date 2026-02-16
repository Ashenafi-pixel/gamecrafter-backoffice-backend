package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OperationalGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name" validate:"required,max=50"`
	Description string    `json:"description" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

func ValidateOperationalGroup(op OperationalGroup) error {
	validate := validator.New()
	return validate.Struct(op)
}
