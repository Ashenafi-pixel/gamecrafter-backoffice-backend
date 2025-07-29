package dto

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rmg/iso4217"
	"github.com/shopspring/decimal"
)

// Player Type Constants
type Type string

const (
	PLAYER     Type = "PLAYER"
	OTG_AGENT  Type = "OTG_AGENT"
	INFLUENCER Type = "INFLUENCER"
)

// User represents the user registration request payload.
type User struct {
	ID              uuid.UUID `json:"id,omitempty"  swaggerignore:"true"`
	PhoneNumber     string    `json:"phone_number" validate:"required,e164,min=8"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	ReferralCode    string    `json:"referral_code,omitempty" `
	Password        string    `json:"password,omitempty" validate:"passwordvalidation"`
	DefaultCurrency string    `json:"default_currency"`
	ProfilePicture  string    `json:"profile_picture"`
	DateOfBirth     string    `json:"date_of_birth"`
	Source          string    `json:"source,omitempty"  swaggerignore:"true"`
	Roles           []Role    `gorm:"many2many:user_roles;" json:"user_roles,omitempty" swaggerignore:"true"`
	StreetAddress   string    `json:"street_address"`
	Country         string    `json:"country"`
	State           string    `json:"state"`
	City            string    `json:"city"`
	Status          string    `json:"status,omitempty" swaggerignore:"true"`
	CreatedBy       uuid.UUID `json:"created_by" swaggerignore:"true"`
	IsAdmin         bool      `json:"is_admin,omitempty" swaggerignore:"true"`
	PostalCode      string    `json:"postal_code"`
	KYCStatus       string    `json:"kyc_status"`
	Type            Type      `json:"type,omitempty"`
	ReferalType     Type      `json:"referal_type,omitempty"`
	ReferedByCode   string    `json:"refered_by_code,omitempty"`
	AgentRequestID  string    `json:"agent_request_id,omitempty"`
	Accounts        []Balance `json:"accounts"`
}

// profileResponse
type UserProfile struct {
	PhoneNumber    string    `json:"phone_number" validate:"required,e164,min=8"`
	Email          string    `json:"email"`
	UserID         uuid.UUID `json:"user_id"`
	ProfilePicture string    `json:"profile_picture"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Type           Type      `json:"type"`
	ReferralCode   string    `json:"referral_code,omitempty"`
	ReferedByCode  string    `json:"refered_by_code,omitempty"`
	ReferalType    Type      `json:"referal_type,omitempty"`
}

// UserRegisterResponse Response to the user who registered to the system
type UserRegisterResponse struct {
	Message      string    `json:"message"`
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// UserLoginReq represents the login request payload.
type UserLoginReq struct {
	LoginID  string `json:"login_id" validate:"required,max=80"`
	Password string `json:"password" validate:"required"`
}

// UserLoginRes hold response to the User login response
type UserLoginRes struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

func IsValidCurrency(currency string) bool {

	// Bucks are represented by 'P'
	if currency == "P" {
		return true
	}

	code, minor := iso4217.ByName(currency)
	if code == 0 && minor == 0 {
		return false
	}
	return true
}

func isValidPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return false
	}

	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return false
	}

	if !strings.ContainsAny(password, "0123456789") {
		return false
	}

	if !strings.ContainsAny(password, "!#$%&'()*+,-./:;<=>?@[\\]^_`{|}~") {
		return false
	}

	return true
}

func ValidateUser(u User) error {
	validate := validator.New()
	validate.RegisterValidation("passwordvalidation", isValidPassword)
	return validate.Struct(u)
}
func ValidateLoginRequest(l UserLoginReq) error {
	validator := validator.New()
	return validator.Struct(l)
}

type ChangePasswordReq struct {
	UserID          uuid.UUID `json:"user_id" swaggerignore:"true"`
	OldPassword     string    `json:"old_password" validate:"required"`
	NewPassword     string    `json:"new_password" validate:"passwordvalidation"`
	ConfirmPassword string    `json:"confirm_password"`
}

func ValidateChangePassword(p ChangePasswordReq) error {
	validate := validator.New()
	validate.RegisterValidation("passwordvalidation", isValidPassword)
	return validate.Struct(p)
}

type ChangePasswordRes struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

type ForgetPasswordRes struct {
	Message string `json:"message"`
}

type ForgetPasswordOTPReq struct {
	UserID uuid.UUID `json:"user_id"`
	OTP    string    `json:"otp"`
}

type EmailReq struct {
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Body    []byte   `json:"body"`
}

type ForgetPasswordReq struct {
	EmailOrPhoneOrUserame string `json:"login_id"`
}

type VerifyResetPasswordReq struct {
	EmailOrPhoneOrUserame string `json:"phone_number"`
	OTP                   string `json:"otp"`
}

type VerifyResetPasswordRes struct {
	UserID uuid.UUID `json:"user_id"`
	Token  string    `json:"token"`
}

type OTPHolder struct {
	TmpOTP    string    `json:"json:"tmp_OTP"`
	CreatedAT time.Time `json:"created_at"`
	Attempts  int       `json:"attempts"`
}

type ResetPasswordReq struct {
	Token           string `json:"token"`
	NewPassword     string `json:"new_password" validate:"validpassword"`
	ConfirmPassword string `json:"confirm_password"`
}

type ResetPasswordRes struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}

func ValidateResetPassword(rp ResetPasswordReq) error {
	validate := validator.New()
	validate.RegisterValidation("validpassword", isValidPassword)
	return validate.Struct(rp)
}

type UpdateProfileReq struct {
	UserID        uuid.UUID `json:"user_id" swaggerignore:"true"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	DateOfBirth   string    `json:"date_of_birth"`
	Phone         string    `json:"phone"`
	Username      string    `json:"username" swaggerignore:"true"`
	Source        string    `json:"source" swaggerignore:"true"`
	StreetAddress string    `json:"street_address"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	State         string    `json:"state"`
	Country       string    `json:"country"`
	KYCStatus     string    `json:"kyc_status"`
}

type UpdateProfileRes struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
	Any    bool      `json:"any"`
}

type UpdateProfileTmpHolder struct {
	TmpOTP           string           `json:"json:"tmp_OTP"`
	CreatedAT        time.Time        `json:"created_at"`
	Attempts         int              `json:"attempts"`
	Any              bool             `json:"any"`
	UpdateProfileReq UpdateProfileReq `json:"update_profile_req"`
}

type ConfirmUpdateProfile struct {
	Token string `json:"token"`
	OTP   string `json:"otp"`
}

type FacebookOauthRes struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Name      string `json:"name"`
}

type GetUsersForNotificationRes struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Phone  string    `json:"phone"`
}
type NotifyDepartmentsReq struct {
	BlockerUser   User `json:"blocker_user"`
	BlockedUser   User `json:"blocked_user"`
	BlockReq      AccountBlockReq
	UsersToNotify []GetUsersForNotificationRes
}

type EditProfileAdminReq struct {
	UserID        uuid.UUID `json:"user_id"`
	AdminID       uuid.UUID `json:"admin_id" swaggerignore:"true"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	UserName      string    `json:"username"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number"`
	StreetAddress string    `json:"street_address"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	State         string    `json:"state"`
	Country       string    `json:"country"`
	KYCStatus     string    `json:"kyc_status"`
}

type AdminResetPasswordReq struct {
	UserID          uuid.UUID `json:"user_id"`
	AdminID         uuid.UUID `json:"admin_id" swaggerignore:"true"`
	NewPassword     string    `json:"new_password" validate:"passwordvalidation"`
	ConfirmPassword string    `json:"confirm_password"`
}

type EditProfileAdminRes struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

type AdminResetPasswordRes struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

func ValidateAdminResetPassword(u AdminResetPasswordReq) error {
	validate := validator.New()
	validate.RegisterValidation("passwordvalidation", isValidPassword)
	return validate.Struct(u)
}

type GetPlayersFilter struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
}
type GetPlayersReq struct {
	Page    int              `json:"page"`
	PerPage int              `json:"per_page"`
	Filter  GetPlayersFilter `json:"filter"`
}

type GetPlayersRes struct {
	TotalCount int64  `json:"total_count"`
	Message    string `json:"message"`
	TotalPages int    `json:"total_pages"`
	Users      []User `json:"users"`
}

type RefferedUsers struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	Amount    int       `json:"amount"`
}

type MyRefferedUsers struct {
	Amount int             `json:"amount"`
	Users  []RefferedUsers `json:"reffered_users"`
}

type GetPointsResp struct {
	Message string    `json:"message"`
	Data    UserPoint `json:"data"`
}

type AdminRoleRes struct {
	RoleID uuid.UUID `json:"role_id"`
	Name   string    `json:"name"`
}
type Admin struct {
	ID          uuid.UUID      `json:"id"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phone_number"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Status      string         `json:"status"`
	Roles       []AdminRoleRes `json:"roles"`
}

type SignUpBonusReq struct {
	Amount decimal.Decimal `json:"amount"`
}

type SignUpBonusRes struct {
	Message string          `json:"message"`
	Amount  decimal.Decimal `json:"amount"`
}

type ReferralBonusReq struct {
	Amount decimal.Decimal `json:"amount" validate:"required,gte=0"`
}

type ReferralBonusRes struct {
	Message string          `json:"message"`
	Amount  decimal.Decimal `json:"amount"`
}

type UserBalanceResp struct {
	UserID   uuid.UUID       `json:"user_id"`
	Balance  decimal.Decimal `json:"balance"`
	Currency string          `json:"currency"`
}

// UserSession represents a user session for session management
type UserSession struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	Token              string    `json:"token"`
	ExpiresAt          time.Time `json:"expires_at"`
	IpAddress          string    `json:"ip_address,omitempty"`
	UserAgent          string    `json:"user_agent,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	RefreshToken       string    `json:"refresh_token"`
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry"`
}

type ReferralData struct {
	ReferedByCode  string `json:"refered_by_code,omitempty"`
	AgentRequestID string `json:"agent_request_id,omitempty"`
	ReferralCode   string `json:"referral_code,omitempty"`
}
type UserReferals struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Data   []byte    `json:"data"`
}

type GetUserReferals struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	ReferralData ReferralData `json:"referral_data"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type VerifyPhoneNumberReq struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=8"`
	OTP         string `json:"otp" validate:"required"`
}

type ResendVerificationOTPReq struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=8"`
}

type TestOtp struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164,min=8"`
}
