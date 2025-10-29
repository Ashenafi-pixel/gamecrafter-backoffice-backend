package dto

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("ipaddr", ValidateIpAddress); err != nil {
		fmt.Printf("failed to register ipaddr validation: %v\n", err)
	}
	if err := validate.RegisterValidation("iplist", ValidateIpList); err != nil {
		fmt.Printf("failed to register iplist validation: %v\n", err)
	}
}

type Company struct {
	ID                             uuid.UUID  `json:"id"`
	SiteName                       string     `json:"site_name" validate:"required,min=1,max=255"`
	SupportEmail                   string     `json:"support_email" validate:"required,email"`
	SupportPhone                   string     `json:"support_phone" validate:"required,e164,min=8,max=20"`
	MaintenanceMode                bool       `json:"maintenance_mode"`
	MaximumLoginAttempt            int        `json:"maximum_login_attempt" validate:"requiredgte=1,lte=10"`
	PasswordExpiry                 int        `json:"password_expiry" validate:"required,gte=0,lte=365"`
	LockoutDuration                int        `json:"lockout_duration" validate:"required,gte=0,lte=86400"`
	RequireTwoFactorAuthentication bool       `json:"require_two_factor_authentication"`
	IpList                         []string   `json:"ip_list" validate:"iplist,dive,ip"`
	CreatedBy                      string     `json:"created_by"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`
	DeletedAt                      *time.Time `json:"deleted_at,omitempty"`
}

type CreateCompanyReq struct {
	SiteName                       string    `json:"site_name" validate:"required,min=1,max=255"`
	SupportEmail                   string    `json:"support_email" validate:"required,email"`
	SupportPhone                   string    `json:"support_phone" validate:"required,e164,min=8,max=20"`
	MaintenanceMode                bool      `json:"maintenance_mode,omitempty"`
	MaximumLoginAttempt            int       `json:"maximum_login_attempt" validate:"omitempty,gte=1,lte=10"`
	PasswordExpiry                 int       `json:"password_expiry" validate:"omitempty,gte=0,lte=365"`
	LockoutDuration                int       `json:"lockout_duration" validate:"omitempty,gte=0,lte=86400"`
	RequireTwoFactorAuthentication bool      `json:"require_two_factor_authentication,omitempty"`
	IpList                         []string  `json:"ip_list" validate:"iplist,dive,ip"`
	CreatedBy                      uuid.UUID `json:"created_by"`
}

type CreateCompanyRes struct {
	ID                             uuid.UUID  `json:"id"`
	SiteName                       string     `json:"site_name"`
	SupportEmail                   string     `json:"support_email"`
	SupportPhone                   string     `json:"support_phone"`
	MaintenanceMode                bool       `json:"maintenance_mode"`
	MaximumLoginAttempt            int        `json:"maximum_login_attempt"`
	PasswordExpiry                 int        `json:"password_expiry"`
	LockoutDuration                int        `json:"lockout_duration"`
	RequireTwoFactorAuthentication bool       `json:"require_two_factor_authentication"`
	IpList                         []string   `json:"ip_list"`
	CreatedBy                      string     `json:"created_by"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`
	DeletedAt                      *time.Time `json:"deleted_at,omitempty"`
}

type UpdateCompanyReq struct {
	ID                             uuid.UUID `json:"id"`
	SiteName                       string    `json:"site_name" validate:"omitempty,min=1,max=255"`
	SupportEmail                   string    `json:"support_email" validate:"omitempty,email"`
	SupportPhone                   string    `json:"support_phone" validate:"omitempty,e164,min=8,max=20"`
	MaintenanceMode                bool      `json:"maintenance_mode,omitempty"`
	MaximumLoginAttempt            int       `json:"maximum_login_attempt" validate:"omitempty,gte=0,lte=10"`
	PasswordExpiry                 int       `json:"password_expiry" validate:"omitempty,gte=0,lte=365"`
	LockoutDuration                int       `json:"lockout_duration" validate:"omitempty,gte=0,lte=86400"`
	RequireTwoFactorAuthentication bool      `json:"require_two_factor_authentication,omitempty"`
}

type UpdateCompanyRes struct {
	ID                             uuid.UUID  `json:"id"`
	SiteName                       string     `json:"site_name"`
	SupportEmail                   string     `json:"support_email"`
	SupportPhone                   string     `json:"support_phone"`
	MaintenanceMode                bool       `json:"maintenance_mode"`
	MaximumLoginAttempt            int        `json:"maximum_login_attempt"`
	PasswordExpiry                 int        `json:"password_expiry"`
	LockoutDuration                int        `json:"lockout_duration"`
	RequireTwoFactorAuthentication bool       `json:"require_two_factor_authentication"`
	IpList                         []string   `json:"ip_list"`
	CreatedBy                      string     `json:"created_by"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`
	DeletedAt                      *time.Time `json:"deleted_at,omitempty"`
}

type GetCompaniesReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type GetCompaniesRes struct {
	Companies  []Company `json:"companies"`
	TotalPages int       `json:"total_pages"`
}

type AddCompanyIPReq struct {
	IpAddr string `json:"ip_addr" validate:"required,ipaddr"`
}

func ValidateCreateCompany(req CreateCompanyReq) error {
	return validate.Struct(req)
}

func ValidateUpdateCompany(req UpdateCompanyReq) error {
	return validate.Struct(req)
}

func ValidateAddCompanyIP(req AddCompanyIPReq) error {
	return validate.Struct(req)
}

func ValidateIpAddress(fl validator.FieldLevel) bool {
	ipAddr := fl.Field().String()
	return net.ParseIP(ipAddr) != nil
}

func ValidateIpList(fl validator.FieldLevel) bool {
	ipList, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}
	for _, ip := range ipList {
		if net.ParseIP(ip) == nil {
			return false
		}
	}
	return true
}

func ToPgtypeInetArray(ipList []string) ([]pgtype.Inet, error) {
	var result []pgtype.Inet
	for _, ip := range ipList {
		var inet pgtype.Inet
		if err := inet.Set(ip); err != nil {
			return nil, err
		}
		result = append(result, inet)
	}
	return result, nil
}

func GetStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func GetBoolValue(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

func GetIntValue(ni sql.NullInt32) int {
	if ni.Valid {
		return int(ni.Int32)
	}
	return 0
}

func GetTimePtrFromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
