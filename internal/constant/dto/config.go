package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Config struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Value string    `json:"value"`
}

type SpinningWheelTypes string // enum

const (
	Point               SpinningWheelTypes = "point"
	InternetPackageInGB SpinningWheelTypes = "internet_package_in_gb"
	Better              SpinningWheelTypes = "better"
	Mystery             SpinningWheelTypes = "mystery"
	Spin                SpinningWheelTypes = "free spin"
)

type SpinningWheelMysteryTypes string

const (
	SpinningWheelMysteryTypesPoint               SpinningWheelMysteryTypes = "point"
	SpinningWheelMysteryTypesInternetPackageInGB SpinningWheelMysteryTypes = "internet_package_in_gb"
	SpinningWheelMysteryTypesBetter              SpinningWheelMysteryTypes = "better"
	SpinningWheelMysteryTypesSpin                SpinningWheelMysteryTypes = "free spin"
)

type SpinningWheelConfigData struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Status    string             `json:"status"`
	Amount    decimal.Decimal    `json:"amount"`
	Type      SpinningWheelTypes `json:"type"`
	Frequency int                `json:"frequency"`
	Icon      string             `json:"icon"`
	Color     string             `json:"color"`
}

type CreateSpinningWheelConfigReq struct {
	Name      string             `json:"name"`
	Amount    decimal.Decimal    `json:"amount"`
	Type      SpinningWheelTypes `json:"type"`
	Frequency int                `json:"frequency"`
	CreatedBy uuid.UUID          `json:"created_by" swaggerignore:"true"`
	Icon      string             `json:"icon"`
	Color     string             `json:"color"`
}

type CreateSpinningWheelConfigRes struct {
	Message string                  `json:"message"`
	Data    SpinningWheelConfigData `json:"data"`
}

type GetSpinningWheelConfigResp struct {
	Message                 string
	SpinningWheelConfigData []SpinningWheelConfigData `json:"configs"` // max 6 items
}

type CreateSpinningWheelMysteryReq struct {
	Name      string                    `json:"name"`
	Amount    decimal.Decimal           `json:"amount"`
	Type      SpinningWheelMysteryTypes `json:"type"`
	Frequency int                       `json:"frequency"`
	Status    string                    `json:"status"`
	Icon      string                    `json:"icon"`
	CreatedBy uuid.UUID                 `json:"created_by" swaggerignore:"true"`
}

type SpinningWheelMysteryResData struct {
	ID        uuid.UUID                 `json:"id"`
	Amount    decimal.Decimal           `json:"amount"`
	Frequency int                       `json:"frequency"`
	Type      SpinningWheelMysteryTypes `json:"type"`
	Status    string                    `json:"status"`
	Icon      string                    `json:"icon"`
}
type CreateSpinningWheelMysteryRes struct {
	Message string                      `json:"message"`
	Data    SpinningWheelMysteryResData `json:"data"`
}

type GetSpinningWheelMysteryRes struct {
	TotalPage int                           `json:"total_page"`
	Message   string                        `json:"message"`
	Data      []SpinningWheelMysteryResData `json:"data"`
}

type UpdateSpinningWheelMysteryReq struct {
	ID        uuid.UUID                 `json:"id"`
	Amount    decimal.Decimal           `json:"amount"`
	Type      SpinningWheelMysteryTypes `json:"type"`
	Status    string                    `json:"status"`
	Frequency int                       `json:"frequency"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

type UpdateSpinningWheelMysteryRes struct {
	Message string                      `json:"message"`
	Data    SpinningWheelMysteryResData `json:"data"`
}

type DeleteReq struct {
	ID uuid.UUID `json:"id"`
}

type GetSpinningWheelConfigRes struct {
	Message                 string
	SpinningWheelConfigData []SpinningWheelConfigData `json:"configs"`
	TotalPage               int                       `json:"total_page"`
}

type UpdateSpinningWheelConfigReq struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Amount    decimal.Decimal    `json:"amount"`
	Type      SpinningWheelTypes `json:"type"`
	Status    string             `json:"status"`
	Frequency int                `json:"frequency"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type UpdateSpinningWheelConfigRes struct {
	Message string                  `json:"message"`
	Data    SpinningWheelConfigData `json:"data"`
}

type DeleteSpinningWheelConfigRes struct {
	Message string `json:"message"`
}

type UploadIconsResp struct {
	Status string `json:"status"`
	Url    string `json:"url"`
}
