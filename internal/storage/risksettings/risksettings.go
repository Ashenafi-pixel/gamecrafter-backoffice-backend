package risksettings

import (
	"context"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type riskSettings struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.RiskSettings {
	return &riskSettings{
		db:  db,
		log: log,
	}
}

func (r *riskSettings) GetRiskSettings(ctx context.Context) (dto.RiskSettings, error) {

	settings, err := r.db.Queries.GetRiskSettings(context.Background())

	if err != nil {
		r.log.Warn("unable to get risk settings", zap.Error(err))
		err = errors.ErrResourceNotFound.Wrap(err, "No risk settings found")
		return dto.RiskSettings{}, err
	}

	return dto.RiskSettings{
		SystemLimitsEnabled:               settings.SystemLimitsEnabled,
		SystemMaxDailyAirtimeConversion:   settings.SystemMaxDailyAirtimeConversion,
		SystemMaxWeeklyAirtimeConversion:  settings.SystemMaxWeeklyAirtimeConversion,
		SystemMaxMonthlyAirtimeConversion: settings.SystemMaxMonthlyAirtimeConversion,
		PlayerLimitsEnabled:               settings.PlayerLimitsEnabled,
		PlayerMaxDailyAirtimeConversion:   settings.PlayerMaxDailyAirtimeConversion,
		PlayerMaxWeeklyAirtimeConversion:  settings.PlayerMaxWeeklyAirtimeConversion,
		PlayerMaxMonthlyAirtimeConversion: settings.PlayerMaxMonthlyAirtimeConversion,
		PlayerMinAirtimeConversionAmount:  settings.PlayerMinAirtimeConversionAmount,
		PlayerConversionCooldownHours:     settings.PlayerConversionCooldownHours,
		KycRequiredAboveAmount:            settings.KycRequiredAboveAmount,
		KycVerificationTimeoutHours:       settings.KycVerificationTimeoutHours,
		KycAllowPartial:                   settings.KycAllowPartial,
		FraudMaxLoginAttempts:             settings.FraudMaxLoginAttempts,
		FraudLoginLockoutDurationMinutes:  settings.FraudLoginLockoutDurationMinutes,
		AlertAdminsOnTrigger:              settings.AlertAdminsOnTrigger,
	}, nil

}

func (r *riskSettings) SetRiskSettings(ctx context.Context, settings dto.RiskSettings) (dto.RiskSettings, error) {

	updatedSettings, err := r.db.Queries.CreateOrUpdateRiskSettings(context.Background(), db.CreateOrUpdateRiskSettingsParams{
		SystemLimitsEnabled:               settings.SystemLimitsEnabled,
		SystemMaxDailyAirtimeConversion:   settings.SystemMaxDailyAirtimeConversion,
		SystemMaxWeeklyAirtimeConversion:  settings.SystemMaxWeeklyAirtimeConversion,
		SystemMaxMonthlyAirtimeConversion: settings.SystemMaxMonthlyAirtimeConversion,
		PlayerLimitsEnabled:               settings.PlayerLimitsEnabled,
		PlayerMaxDailyAirtimeConversion:   settings.PlayerMaxDailyAirtimeConversion,
		PlayerMaxWeeklyAirtimeConversion:  settings.PlayerMaxWeeklyAirtimeConversion,
		PlayerMaxMonthlyAirtimeConversion: settings.PlayerMaxMonthlyAirtimeConversion,
		PlayerMinAirtimeConversionAmount:  settings.PlayerMinAirtimeConversionAmount,
		PlayerConversionCooldownHours:     settings.PlayerConversionCooldownHours,
		KycRequiredAboveAmount:            settings.KycRequiredAboveAmount,
		KycVerificationTimeoutHours:       settings.KycVerificationTimeoutHours,
		KycAllowPartial:                   settings.KycAllowPartial,
		FraudMaxLoginAttempts:             settings.FraudMaxLoginAttempts,
		FraudLoginLockoutDurationMinutes:  settings.FraudLoginLockoutDurationMinutes,
		AlertAdminsOnTrigger:              settings.AlertAdminsOnTrigger,
	})

	if err != nil {
		r.log.Error("unable to update risk settings", zap.Error(err))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update risk settings")
		return dto.RiskSettings{}, err
	}

	return dto.RiskSettings{
		SystemLimitsEnabled:               updatedSettings.SystemLimitsEnabled,
		SystemMaxDailyAirtimeConversion:   updatedSettings.SystemMaxDailyAirtimeConversion,
		SystemMaxWeeklyAirtimeConversion:  updatedSettings.SystemMaxWeeklyAirtimeConversion,
		SystemMaxMonthlyAirtimeConversion: updatedSettings.SystemMaxMonthlyAirtimeConversion,
		PlayerLimitsEnabled:               updatedSettings.PlayerLimitsEnabled,
		PlayerMaxDailyAirtimeConversion:   updatedSettings.PlayerMaxDailyAirtimeConversion,
		PlayerMaxWeeklyAirtimeConversion:  updatedSettings.PlayerMaxWeeklyAirtimeConversion,
		PlayerMaxMonthlyAirtimeConversion: updatedSettings.PlayerMaxMonthlyAirtimeConversion,
		PlayerMinAirtimeConversionAmount:  updatedSettings.PlayerMinAirtimeConversionAmount,
		PlayerConversionCooldownHours:     updatedSettings.PlayerConversionCooldownHours,
		KycRequiredAboveAmount:            updatedSettings.KycRequiredAboveAmount,
		KycVerificationTimeoutHours:       updatedSettings.KycVerificationTimeoutHours,
		KycAllowPartial:                   updatedSettings.KycAllowPartial,
		FraudMaxLoginAttempts:             updatedSettings.FraudMaxLoginAttempts,
		FraudLoginLockoutDurationMinutes:  updatedSettings.FraudLoginLockoutDurationMinutes,
		AlertAdminsOnTrigger:              updatedSettings.AlertAdminsOnTrigger,
	}, nil
}
