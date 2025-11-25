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
	ID                       uuid.UUID        `json:"id,omitempty"  swaggerignore:"true"`
	Username                 string           `json:"username,omitempty"`
	PhoneNumber              string           `json:"phone_number" validate:"omitempty,e164,min=8"`
	FirstName                string           `json:"first_name"`
	LastName                 string           `json:"last_name"`
	Email                    string           `json:"email"`
	ReferralCode             string           `json:"referral_code,omitempty" `
	Password                 string           `json:"password,omitempty" validate:"passwordvalidation"`
	DefaultCurrency          string           `json:"default_currency"`
	ProfilePicture           string           `json:"profile_picture"`
	DateOfBirth              string           `json:"date_of_birth"`
	Source                   string           `json:"source,omitempty"  swaggerignore:"true"`
	Roles                    []Role           `gorm:"many2many:user_roles;" json:"user_roles,omitempty" swaggerignore:"true"`
	StreetAddress            string           `json:"street_address"`
	Country                  string           `json:"country"`
	State                    string           `json:"state"`
	City                     string           `json:"city"`
	Status                   string           `json:"status,omitempty" swaggerignore:"true"`
	CreatedBy                uuid.UUID        `json:"created_by" swaggerignore:"true"`
	IsAdmin                  bool             `json:"is_admin,omitempty" swaggerignore:"true"`
	PostalCode               string           `json:"postal_code"`
	KYCStatus                string           `json:"kyc_status"`
	Type                     Type             `json:"type,omitempty"`
	ReferalType              Type             `json:"referal_type,omitempty"`
	ReferedByCode            string           `json:"refered_by_code,omitempty"`
	AgentRequestID           string           `json:"agent_request_id,omitempty"`
	Accounts                 []Balance        `json:"accounts"`
	CreatedAt                *time.Time       `json:"created_at,omitempty"`
	IsEmailVerified          bool             `json:"is_email_verified,omitempty"`
	WalletVerificationStatus string           `json:"wallet_verification_status,omitempty"`
	IsTestAccount            bool             `json:"is_test_account,omitempty"`
	VipLevel                 string           `json:"vip_level,omitempty"`
	CurrentLevel             int              `json:"current_level,omitempty"`
	EffectiveLevel           int              `json:"effective_level,omitempty"`
	LevelManualOverride      *bool            `json:"level_manual_override,omitempty"`
	ManualOverrideLevel      *int             `json:"manual_override_level,omitempty"`
	ManualOverrideSetBy      *uuid.UUID       `json:"manual_override_set_by,omitempty"`
	ManualOverrideSetAt      *time.Time       `json:"manual_override_set_at,omitempty"`
	WithdrawalLimit          *decimal.Decimal `json:"withdrawal_limit,omitempty"`
	WithdrawalLimitEnabled   bool             `json:"withdrawal_limit_enabled,omitempty"`
	WithdrawalAllTimeLimit   *decimal.Decimal `json:"withdrawal_all_time_limit,omitempty"`
	BrandID                  *uuid.UUID       `json:"brand_id,omitempty"` // brand_id from request
}

// profileResponse
type UserProfile struct {
	Username       string    `json:"username"`
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
	Message             string       `json:"message"`
	AccessToken         string       `json:"access_token"`
	UserProfile         *UserProfile `json:"user_profile,omitempty"`
	Requires2FA         bool         `json:"requires_2fa,omitempty"`
	Requires2FASetup    bool         `json:"requires_2fa_setup,omitempty"`
	UserID              string       `json:"user_id,omitempty"`
	Available2FAMethods []string     `json:"available_2fa_methods,omitempty"`
	AllowedPages        []Page       `json:"allowed_pages,omitempty"`
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
	Message string    `json:"message"`
	Email   string    `json:"email"`
	OTPID   uuid.UUID `json:"otp_id"`
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
	EmailOrPhoneOrUserame string    `json:"phone_number"`
	OTP                   string    `json:"otp"`
	OTPID                 uuid.UUID `json:"otp_id"`
}

type VerifyResetPasswordRes struct {
	UserID uuid.UUID `json:"user_id"`
	Token  string    `json:"token"`
}

type OTPHolder struct {
	TmpOTP    string    `json:"tmp_OTP"`
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
	UserID                   uuid.UUID        `json:"user_id" swaggerignore:"true"`
	FirstName                string           `json:"first_name"`
	LastName                 string           `json:"last_name"`
	Email                    string           `json:"email"`
	DateOfBirth              string           `json:"date_of_birth"`
	Phone                    string           `json:"phone"`
	Username                 string           `json:"username" swaggerignore:"true"`
	Source                   string           `json:"source" swaggerignore:"true"`
	StreetAddress            string           `json:"street_address"`
	City                     string           `json:"city"`
	PostalCode               string           `json:"postal_code"`
	State                    string           `json:"state"`
	Country                  string           `json:"country"`
	KYCStatus                string           `json:"kyc_status"`
	Status                   string           `json:"status"`
	IsEmailVerified          bool             `json:"is_email_verified"`
	DefaultCurrency          string           `json:"default_currency"`
	WalletVerificationStatus string           `json:"wallet_verification_status"`
	IsTestAccount            *bool            `json:"is_test_account,omitempty"`
	WithdrawalLimit          *decimal.Decimal `json:"withdrawal_limit,omitempty"`
	WithdrawalLimitEnabled   *bool            `json:"withdrawal_limit_enabled,omitempty"`
	WithdrawalAllTimeLimit   *decimal.Decimal `json:"withdrawal_all_time_limit,omitempty"`
}

type UpdateProfileRes struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
	Any    bool      `json:"any"`
}

type UpdateProfileTmpHolder struct {
	TmpOTP           string           `json:"tmp_OTP"`
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
	UserID                   uuid.UUID        `json:"user_id"`
	AdminID                  uuid.UUID        `json:"admin_id" swaggerignore:"true"`
	FirstName                string           `json:"first_name"`
	LastName                 string           `json:"last_name"`
	UserName                 string           `json:"username"`
	Email                    string           `json:"email"`
	PhoneNumber              string           `json:"phone_number"`
	StreetAddress            string           `json:"street_address"`
	City                     string           `json:"city"`
	PostalCode               string           `json:"postal_code"`
	State                    string           `json:"state"`
	Country                  string           `json:"country"`
	KYCStatus                string           `json:"kyc_status"`
	DateOfBirth              string           `json:"date_of_birth"`
	Status                   string           `json:"status"`
	IsEmailVerified          bool             `json:"is_email_verified"`
	DefaultCurrency          string           `json:"default_currency"`
	WalletVerificationStatus string           `json:"wallet_verification_status"`
	IsTestAccount            *bool            `json:"is_test_account,omitempty"`
	WithdrawalLimit          *decimal.Decimal `json:"withdrawal_limit,omitempty"`
	WithdrawalLimitEnabled   *bool            `json:"withdrawal_limit_enabled,omitempty"`
	WithdrawalAllTimeLimit   *decimal.Decimal `json:"withdrawal_all_time_limit,omitempty"`
	LevelManualOverride      *bool            `json:"level_manual_override,omitempty"`
	ManualOverrideLevel      *int             `json:"manual_override_level,omitempty"`
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

type AdminAutoResetPasswordReq struct {
	UserID  uuid.UUID `json:"user_id"`
	AdminID uuid.UUID `json:"admin_id" swaggerignore:"true"`
}

type AdminAutoResetPasswordRes struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

func ValidateAdminResetPassword(u AdminResetPasswordReq) error {
	validate := validator.New()
	validate.RegisterValidation("passwordvalidation", isValidPassword)
	return validate.Struct(u)
}

type GetPlayersFilter struct {
	UserID        string   `json:"user_id"`
	SearchTerm    string   `json:"searchterm"`
	Status        []string `json:"status"`
	KycStatus     []string `json:"kyc_status"`
	VipLevel      []string `json:"vip_level"`
	IsTestAccount *bool    `json:"is_test_account,omitempty"`
	BrandID       []string `json:"brand_id,omitempty"`
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
	ID                       uuid.UUID      `json:"id"`
	Username                 string         `json:"username"`
	Email                    string         `json:"email"`
	PhoneNumber              string         `json:"phone_number"`
	FirstName                string         `json:"first_name"`
	LastName                 string         `json:"last_name"`
	DateOfBirth              string         `json:"date_of_birth"`
	StreetAddress            string         `json:"street_address"`
	City                     string         `json:"city"`
	PostalCode               string         `json:"postal_code"`
	State                    string         `json:"state"`
	Country                  string         `json:"country"`
	KycStatus                string         `json:"kyc_status"`
	IsEmailVerified          bool           `json:"is_email_verified"`
	DefaultCurrency          string         `json:"default_currency"`
	WalletVerificationStatus string         `json:"wallet_verification_status"`
	Status                   string         `json:"status"`
	IsAdmin                  bool           `json:"is_admin"`
	UserType                 string         `json:"user_type"`
	Roles                    []AdminRoleRes `json:"roles"`
	CreatedAt                time.Time      `json:"created_at"`
}

type CreateAdminUserReq struct {
	Username                 string           `json:"username" validate:"required,min=3,max=50"`
	Email                    string           `json:"email" validate:"required,email"`
	Password                 string           `json:"password" validate:"required,min=6"`
	FirstName                string           `json:"first_name" validate:"required,min=2,max=50"`
	LastName                 string           `json:"last_name" validate:"required,min=2,max=50"`
	Phone                    string           `json:"phone,omitempty"`
	DateOfBirth              string           `json:"date_of_birth,omitempty"`
	StreetAddress            string           `json:"street_address,omitempty"`
	City                     string           `json:"city,omitempty"`
	PostalCode               string           `json:"postal_code,omitempty"`
	State                    string           `json:"state,omitempty"`
	Country                  string           `json:"country,omitempty"`
	KycStatus                string           `json:"kyc_status,omitempty"`
	IsEmailVerified          bool             `json:"is_email_verified,omitempty"`
	DefaultCurrency          string           `json:"default_currency,omitempty"`
	WalletVerificationStatus string           `json:"wallet_verification_status,omitempty"`
	Status                   string           `json:"status,omitempty"`
	IsAdmin                  bool             `json:"is_admin,omitempty"`
	UserType                 string           `json:"user_type,omitempty"`
	WithdrawalLimit          *decimal.Decimal `json:"withdrawal_limit,omitempty"`
	WithdrawalLimitEnabled   bool             `json:"withdrawal_limit_enabled,omitempty"`
}

type UpdateAdminUserReq struct {
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Password  *string `json:"password,omitempty" validate:"omitempty,min=6"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	Phone     *string `json:"phone,omitempty"`
	Status    *string `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE"`
	IsAdmin   *bool   `json:"is_admin,omitempty"`
	UserType  *string `json:"user_type,omitempty"`
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
	UserID           uuid.UUID       `json:"user_id"`
	Balance          decimal.Decimal `json:"balance"`
	BalanceFormatted string          `json:"balance_formatted"`
	Currency         string          `json:"currency"`
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

// DetailedAccount represents the account structure in detailed registration
type DetailedAccount struct {
	BonusMoney int    `json:"bonus_money"`
	Currency   string `json:"currency"`
	ID         string `json:"id"`
	RealMoney  int    `json:"real_money"`
	UpdatedAt  string `json:"updated_at"`
	UserID     string `json:"user_id"`
}

// DetailedUserRegistration represents the comprehensive user registration request payload
type DetailedUserRegistration struct {
	Accounts        []DetailedAccount `json:"accounts"`
	AgentRequestID  string            `json:"agent_request_id"`
	City            string            `json:"city"`
	Country         string            `json:"country"`
	DateOfBirth     string            `json:"date_of_birth"`
	DefaultCurrency string            `json:"default_currency"`
	Email           string            `json:"email"`
	FirstName       string            `json:"first_name"`
	KYCStatus       string            `json:"kyc_status"`
	LastName        string            `json:"last_name"`
	Password        string            `json:"password"`
	PhoneNumber     string            `json:"phone_number"`
	PostalCode      string            `json:"postal_code"`
	ProfilePicture  string            `json:"profile_picture"`
	ReferalType     string            `json:"referal_type"`
	ReferedByCode   string            `json:"refered_by_code"`
	ReferralCode    string            `json:"referral_code"`
	State           string            `json:"state"`
	StreetAddress   string            `json:"street_address"`
	Type            string            `json:"type"`
	Username        string            `json:"username"`
}

// SuspensionHistory represents a player's suspension history
type SuspensionHistory struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	BlockedBy   uuid.UUID  `json:"blocked_by"`
	Duration    string     `json:"duration"`
	Type        string     `json:"type"`
	BlockedFrom *time.Time `json:"blocked_from"`
	BlockedTo   *time.Time `json:"blocked_to"`
	UnblockedAt *time.Time `json:"unblocked_at"`
	Reason      string     `json:"reason"`
	Note        string     `json:"note"`
	CreatedAt   time.Time  `json:"created_at"`
	// Admin details
	BlockedByUsername string `json:"blocked_by_username"`
	BlockedByEmail    string `json:"blocked_by_email"`
}

// BalanceLog represents a balance transaction log
type BalanceLog struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	Component           string          `json:"component"`
	Currency            string          `json:"currency"`
	ChangeAmount        decimal.Decimal `json:"change_amount"`
	OperationalGroupID  uuid.UUID       `json:"operational_group_id"`
	OperationalTypeID   uuid.UUID       `json:"operational_type_id"`
	Description         string          `json:"description"`
	Timestamp           time.Time       `json:"timestamp"`
	BalanceAfterUpdate  decimal.Decimal `json:"balance_after_update"`
	TransactionID       string          `json:"transaction_id"`
	Status              string          `json:"status"`
	Type                string          `json:"type"`
	OperationalTypeName string          `json:"operational_type_name"`
}

// GameActivity represents a player's game activity
type GameActivity struct {
	Game         string          `json:"game"`
	Provider     string          `json:"provider"`
	Sessions     int             `json:"sessions"`
	TotalWagered decimal.Decimal `json:"total_wagered"`
	NetResult    decimal.Decimal `json:"net_result"`
	LastPlayed   time.Time       `json:"last_played"`
	FavoriteGame bool            `json:"favorite_game"`
}

// PlayerStatistics represents player statistics calculated from database
type PlayerStatistics struct {
	TotalWagered decimal.Decimal `json:"total_wagered"`
	NetPL        decimal.Decimal `json:"net_pl"`
	Sessions     int             `json:"sessions"`
	TotalBets    int             `json:"total_bets"`
	TotalWins    int             `json:"total_wins"`
	TotalLosses  int             `json:"total_losses"`
	WinRate      decimal.Decimal `json:"win_rate"`
	AvgBetSize   decimal.Decimal `json:"avg_bet_size"`
	LastActivity time.Time       `json:"last_activity"`
}

// PlayerDetailsResponse represents the complete player details response
type PlayerDetailsResponse struct {
	Player            User                `json:"player"`
	SuspensionHistory []SuspensionHistory `json:"suspension_history"`
	BalanceLogs       []BalanceLog        `json:"balance_logs"`
	Balances          []Balance           `json:"balances"`
	GameActivity      []GameActivity      `json:"game_activity"`
	Statistics        PlayerStatistics    `json:"statistics"`
}
