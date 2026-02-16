package monitoring

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// EnterpriseRegistrationMetrics provides metrics for enterprise registration
type EnterpriseRegistrationMetrics struct {
	// Registration attempts
	registrationAttemptsTotal *prometheus.CounterVec
	registrationSuccessTotal  *prometheus.CounterVec
	registrationFailureTotal  *prometheus.CounterVec

	// OTP verification
	otpVerificationAttemptsTotal *prometheus.CounterVec
	otpVerificationSuccessTotal  *prometheus.CounterVec
	otpVerificationFailureTotal  *prometheus.CounterVec
	otpResendTotal               *prometheus.CounterVec

	// Email operations
	emailSentTotal    *prometheus.CounterVec
	emailFailureTotal *prometheus.CounterVec
	emailDeliveryTime *prometheus.HistogramVec

	// Registration status
	registrationStatusGauge *prometheus.GaugeVec

	// Performance metrics
	registrationDuration *prometheus.HistogramVec
	otpGenerationTime    *prometheus.HistogramVec

	// Business metrics
	userTypeDistribution *prometheus.CounterVec
	referralSourceTotal  *prometheus.CounterVec

	// Error tracking
	errorRateByType *prometheus.CounterVec

	// Database operations
	databaseOperationDuration *prometheus.HistogramVec
	databaseErrorsTotal       *prometheus.CounterVec

	logger *zap.Logger
}

// NewEnterpriseRegistrationMetrics creates new enterprise registration metrics
func NewEnterpriseRegistrationMetrics(logger *zap.Logger) *EnterpriseRegistrationMetrics {
	metrics := &EnterpriseRegistrationMetrics{
		logger: logger,

		// Registration attempts
		registrationAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_attempts_total",
				Help: "Total number of enterprise registration attempts",
			},
			[]string{"user_type", "referral_source", "country"},
		),

		registrationSuccessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_success_total",
				Help: "Total number of successful enterprise registrations",
			},
			[]string{"user_type", "referral_source", "country"},
		),

		registrationFailureTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_failure_total",
				Help: "Total number of failed enterprise registrations",
			},
			[]string{"user_type", "failure_reason", "country"},
		),

		// OTP verification
		otpVerificationAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_otp_verification_attempts_total",
				Help: "Total number of OTP verification attempts",
			},
			[]string{"user_type", "otp_type"},
		),

		otpVerificationSuccessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_otp_verification_success_total",
				Help: "Total number of successful OTP verifications",
			},
			[]string{"user_type", "otp_type"},
		),

		otpVerificationFailureTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_otp_verification_failure_total",
				Help: "Total number of failed OTP verifications",
			},
			[]string{"user_type", "otp_type", "failure_reason"},
		),

		otpResendTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_otp_resend_total",
				Help: "Total number of OTP resend requests",
			},
			[]string{"user_type", "resend_reason"},
		),

		// Email operations
		emailSentTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_email_sent_total",
				Help: "Total number of emails sent for enterprise registration",
			},
			[]string{"email_type", "user_type", "provider"},
		),

		emailFailureTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_email_failure_total",
				Help: "Total number of email failures for enterprise registration",
			},
			[]string{"email_type", "user_type", "provider", "failure_reason"},
		),

		emailDeliveryTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "enterprise_registration_email_delivery_time_seconds",
				Help:    "Time taken to deliver emails for enterprise registration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"email_type", "user_type", "provider"},
		),

		// Registration status
		registrationStatusGauge: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "enterprise_registration_status_current",
				Help: "Current number of registrations by status",
			},
			[]string{"status", "user_type"},
		),

		// Performance metrics
		registrationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "enterprise_registration_duration_seconds",
				Help:    "Time taken to complete enterprise registration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"user_type", "status"},
		),

		otpGenerationTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "enterprise_registration_otp_generation_time_seconds",
				Help:    "Time taken to generate OTP for enterprise registration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"user_type", "otp_type"},
		),

		// Business metrics
		userTypeDistribution: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_user_type_distribution_total",
				Help: "Distribution of user types in enterprise registration",
			},
			[]string{"user_type", "country", "referral_source"},
		),

		referralSourceTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_referral_source_total",
				Help: "Total registrations by referral source",
			},
			[]string{"referral_source", "user_type"},
		),

		// Error tracking
		errorRateByType: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_errors_total",
				Help: "Total errors by type in enterprise registration",
			},
			[]string{"error_type", "user_type", "operation"},
		),

		// Database operations
		databaseOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "enterprise_registration_database_operation_duration_seconds",
				Help:    "Time taken for database operations in enterprise registration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),

		databaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "enterprise_registration_database_errors_total",
				Help: "Total database errors in enterprise registration",
			},
			[]string{"operation", "table", "error_type"},
		),
	}

	logger.Info("Enterprise registration metrics initialized")
	return metrics
}

// RecordRegistrationAttempt records a registration attempt
func (m *EnterpriseRegistrationMetrics) RecordRegistrationAttempt(userType, referralSource, country string) {
	m.registrationAttemptsTotal.WithLabelValues(userType, referralSource, country).Inc()
	m.logger.Debug("Recorded registration attempt",
		zap.String("user_type", userType),
		zap.String("referral_source", referralSource),
		zap.String("country", country))
}

// RecordRegistrationSuccess records a successful registration
func (m *EnterpriseRegistrationMetrics) RecordRegistrationSuccess(userType, referralSource, country string) {
	m.registrationSuccessTotal.WithLabelValues(userType, referralSource, country).Inc()
	m.logger.Debug("Recorded registration success",
		zap.String("user_type", userType),
		zap.String("referral_source", referralSource),
		zap.String("country", country))
}

// RecordRegistrationFailure records a failed registration
func (m *EnterpriseRegistrationMetrics) RecordRegistrationFailure(userType, failureReason, country string) {
	m.registrationFailureTotal.WithLabelValues(userType, failureReason, country).Inc()
	m.logger.Debug("Recorded registration failure",
		zap.String("user_type", userType),
		zap.String("failure_reason", failureReason),
		zap.String("country", country))
}

// RecordOTPVerificationAttempt records an OTP verification attempt
func (m *EnterpriseRegistrationMetrics) RecordOTPVerificationAttempt(userType, otpType string) {
	m.otpVerificationAttemptsTotal.WithLabelValues(userType, otpType).Inc()
	m.logger.Debug("Recorded OTP verification attempt",
		zap.String("user_type", userType),
		zap.String("otp_type", otpType))
}

// RecordOTPVerificationSuccess records a successful OTP verification
func (m *EnterpriseRegistrationMetrics) RecordOTPVerificationSuccess(userType, otpType string) {
	m.otpVerificationSuccessTotal.WithLabelValues(userType, otpType).Inc()
	m.logger.Debug("Recorded OTP verification success",
		zap.String("user_type", userType),
		zap.String("otp_type", otpType))
}

// RecordOTPVerificationFailure records a failed OTP verification
func (m *EnterpriseRegistrationMetrics) RecordOTPVerificationFailure(userType, otpType, failureReason string) {
	m.otpVerificationFailureTotal.WithLabelValues(userType, otpType, failureReason).Inc()
	m.logger.Debug("Recorded OTP verification failure",
		zap.String("user_type", userType),
		zap.String("otp_type", otpType),
		zap.String("failure_reason", failureReason))
}

// RecordOTPResend records an OTP resend request
func (m *EnterpriseRegistrationMetrics) RecordOTPResend(userType, resendReason string) {
	m.otpResendTotal.WithLabelValues(userType, resendReason).Inc()
	m.logger.Debug("Recorded OTP resend",
		zap.String("user_type", userType),
		zap.String("resend_reason", resendReason))
}

// RecordEmailSent records a sent email
func (m *EnterpriseRegistrationMetrics) RecordEmailSent(emailType, userType, provider string) {
	m.emailSentTotal.WithLabelValues(emailType, userType, provider).Inc()
	m.logger.Debug("Recorded email sent",
		zap.String("email_type", emailType),
		zap.String("user_type", userType),
		zap.String("provider", provider))
}

// RecordEmailFailure records an email failure
func (m *EnterpriseRegistrationMetrics) RecordEmailFailure(emailType, userType, provider, failureReason string) {
	m.emailFailureTotal.WithLabelValues(emailType, userType, provider, failureReason).Inc()
	m.logger.Debug("Recorded email failure",
		zap.String("email_type", emailType),
		zap.String("user_type", userType),
		zap.String("provider", provider),
		zap.String("failure_reason", failureReason))
}

// RecordEmailDeliveryTime records email delivery time
func (m *EnterpriseRegistrationMetrics) RecordEmailDeliveryTime(emailType, userType, provider string, duration time.Duration) {
	m.emailDeliveryTime.WithLabelValues(emailType, userType, provider).Observe(duration.Seconds())
	m.logger.Debug("Recorded email delivery time",
		zap.String("email_type", emailType),
		zap.String("user_type", userType),
		zap.String("provider", provider),
		zap.Duration("duration", duration))
}

// UpdateRegistrationStatus updates the registration status gauge
func (m *EnterpriseRegistrationMetrics) UpdateRegistrationStatus(status, userType string, count float64) {
	m.registrationStatusGauge.WithLabelValues(status, userType).Set(count)
	m.logger.Debug("Updated registration status gauge",
		zap.String("status", status),
		zap.String("user_type", userType),
		zap.Float64("count", count))
}

// RecordRegistrationDuration records registration completion time
func (m *EnterpriseRegistrationMetrics) RecordRegistrationDuration(userType, status string, duration time.Duration) {
	m.registrationDuration.WithLabelValues(userType, status).Observe(duration.Seconds())
	m.logger.Debug("Recorded registration duration",
		zap.String("user_type", userType),
		zap.String("status", status),
		zap.Duration("duration", duration))
}

// RecordOTPGenerationTime records OTP generation time
func (m *EnterpriseRegistrationMetrics) RecordOTPGenerationTime(userType, otpType string, duration time.Duration) {
	m.otpGenerationTime.WithLabelValues(userType, otpType).Observe(duration.Seconds())
	m.logger.Debug("Recorded OTP generation time",
		zap.String("user_type", userType),
		zap.String("otp_type", otpType),
		zap.Duration("duration", duration))
}

// RecordUserTypeDistribution records user type distribution
func (m *EnterpriseRegistrationMetrics) RecordUserTypeDistribution(userType, country, referralSource string) {
	m.userTypeDistribution.WithLabelValues(userType, country, referralSource).Inc()
	m.logger.Debug("Recorded user type distribution",
		zap.String("user_type", userType),
		zap.String("country", country),
		zap.String("referral_source", referralSource))
}

// RecordReferralSource records referral source
func (m *EnterpriseRegistrationMetrics) RecordReferralSource(referralSource, userType string) {
	m.referralSourceTotal.WithLabelValues(referralSource, userType).Inc()
	m.logger.Debug("Recorded referral source",
		zap.String("referral_source", referralSource),
		zap.String("user_type", userType))
}

// RecordError records an error by type
func (m *EnterpriseRegistrationMetrics) RecordError(errorType, userType, operation string) {
	m.errorRateByType.WithLabelValues(errorType, userType, operation).Inc()
	m.logger.Debug("Recorded error",
		zap.String("error_type", errorType),
		zap.String("user_type", userType),
		zap.String("operation", operation))
}

// RecordDatabaseOperationDuration records database operation duration
func (m *EnterpriseRegistrationMetrics) RecordDatabaseOperationDuration(operation, table string, duration time.Duration) {
	m.databaseOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	m.logger.Debug("Recorded database operation duration",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration))
}

// RecordDatabaseError records a database error
func (m *EnterpriseRegistrationMetrics) RecordDatabaseError(operation, table, errorType string) {
	m.databaseErrorsTotal.WithLabelValues(operation, table, errorType).Inc()
	m.logger.Debug("Recorded database error",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.String("error_type", errorType))
}

// GetMetricsSummary returns a summary of current metrics
func (m *EnterpriseRegistrationMetrics) GetMetricsSummary(ctx context.Context) map[string]interface{} {
	// This would typically query Prometheus or other metrics storage
	// For now, return a placeholder summary
	return map[string]interface{}{
		"total_registrations": "Available via Prometheus metrics",
		"success_rate":        "Available via Prometheus metrics",
		"average_duration":    "Available via Prometheus metrics",
		"error_rate":          "Available via Prometheus metrics",
		"timestamp":           time.Now().UTC(),
	}
}

// ResetMetrics resets all metrics (useful for testing)
func (m *EnterpriseRegistrationMetrics) ResetMetrics() {
	// Reset all metrics to zero
	// Note: This is typically not used in production
	m.logger.Info("Enterprise registration metrics reset")
}
