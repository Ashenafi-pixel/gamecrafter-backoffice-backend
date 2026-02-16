package dto

type RiskSettings struct {
	SystemLimitsEnabled               bool  `json:"system_limits_enabled"`
	SystemMaxDailyAirtimeConversion   int64 `json:"system_max_daily_airtime_conversion"`
	SystemMaxWeeklyAirtimeConversion  int64 `json:"system_max_weekly_airtime_conversion"`
	SystemMaxMonthlyAirtimeConversion int64 `json:"system_max_monthly_airtime_conversion"`
	PlayerLimitsEnabled               bool  `json:"player_limits_enabled"`
	PlayerMaxDailyAirtimeConversion   int32 `json:"player_max_daily_airtime_conversion"`
	PlayerMaxWeeklyAirtimeConversion  int32 `json:"player_max_weekly_airtime_conversion"`
	PlayerMaxMonthlyAirtimeConversion int32 `json:"player_max_monthly_airtime_conversion"`
	PlayerMinAirtimeConversionAmount  int32 `json:"player_min_airtime_conversion_amount"`
	PlayerConversionCooldownHours     int32 `json:"player_conversion_cooldown_hours"`
	KycRequiredAboveAmount            int32 `json:"kyc_required_above_amount"`
	KycVerificationTimeoutHours       int32 `json:"kyc_verification_timeout_hours"`
	KycAllowPartial                   bool  `json:"kyc_allow_partial"`
	FraudMaxLoginAttempts             int16 `json:"fraud_max_login_attempts"`
	FraudLoginLockoutDurationMinutes  int32 `json:"fraud_login_lockout_duration_minutes"`
	AlertAdminsOnTrigger              bool  `json:"alert_admins_on_trigger"`
}
