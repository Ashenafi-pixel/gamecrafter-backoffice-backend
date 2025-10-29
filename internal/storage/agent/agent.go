package agent

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type agent struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Agent {
	return &agent{
		db:  db,
		log: log,
	}
}

func (a *agent) CreateAgentReferralLink(ctx context.Context, req dto.CreateAgentReferralLinkReq) (dto.AgentReferral, error) {
	referral, err := a.db.Queries.CreateAgentReferralLink(ctx, db.CreateAgentReferralLinkParams{
		RequestID:   req.RequestID,
		CallbackUrl: req.CallbackURL,
	})
	if err != nil {
		a.log.Error("unable to create agent referral link", zap.Error(err), zap.Any("request", req))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create agent referral link")
		return dto.AgentReferral{}, err
	}

	return dto.AgentReferral{
		ID:               referral.ID,
		RequestID:        referral.RequestID,
		UserID:           dto.NullToUUID(referral.UserID),
		ConversionType:   dto.GetStringValue(referral.ConversionType),
		Amount:           referral.Amount.Decimal,
		MSISDN:           referral.Msisdn.String,
		ConvertedAt:      referral.ConvertedAt.Time,
		CallbackSent:     referral.CallbackSent.Bool,
		CallbackAttempts: int(referral.CallbackAttempts.Int32),
	}, nil
}

func (a *agent) UpdateAgentReferralWithConversion(ctx context.Context, req dto.UpdateAgentReferralWithConversionReq) (dto.AgentReferral, error) {
	referral, err := a.db.Queries.UpdateAgentReferralWithConversion(ctx, db.UpdateAgentReferralWithConversionParams{
		RequestID:      req.RequestID,
		UserID:         uuid.NullUUID{UUID: req.UserID, Valid: req.UserID != uuid.Nil},
		ConversionType: sql.NullString{String: req.ConversionType, Valid: req.ConversionType != ""},
		Amount:         decimal.NullDecimal{Decimal: req.Amount, Valid: true},
		Msisdn:         sql.NullString{String: req.MSISDN, Valid: req.MSISDN != ""},
	})
	if err != nil {
		a.log.Error("unable to update agent referral with conversion", zap.Error(err), zap.Any("request", req))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update agent referral with conversion")
		return dto.AgentReferral{}, err
	}

	return dto.AgentReferral{
		ID:               referral.ID,
		RequestID:        referral.RequestID,
		UserID:           dto.NullToUUID(referral.UserID),
		ConversionType:   dto.GetStringValue(referral.ConversionType),
		Amount:           referral.Amount.Decimal,
		MSISDN:           referral.Msisdn.String,
		ConvertedAt:      referral.ConvertedAt.Time,
		CallbackSent:     referral.CallbackSent.Bool,
		CallbackAttempts: int(referral.CallbackAttempts.Int32),
	}, nil
}

func (a *agent) GetAgentReferralByRequestID(ctx context.Context, requestID string) (dto.AgentReferral, bool, error) {
	referral, err := a.db.Queries.GetAgentReferralByRequestID(ctx, requestID)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error("unable to get agent referral by request ID", zap.Error(err), zap.String("requestID", requestID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get agent referral by request ID")
		return dto.AgentReferral{}, false, err
	}
	if err != nil && err.Error() == dto.ErrNoRows {
		return dto.AgentReferral{}, false, nil
	}

	return dto.AgentReferral{
		ID:               referral.ID,
		RequestID:        referral.RequestID,
		UserID:           dto.NullToUUID(referral.UserID),
		ConversionType:   dto.GetStringValue(referral.ConversionType),
		Amount:           referral.Amount.Decimal,
		MSISDN:           referral.Msisdn.String,
		ConvertedAt:      referral.ConvertedAt.Time,
		CallbackSent:     referral.CallbackSent.Bool,
		CallbackAttempts: int(referral.CallbackAttempts.Int32),
	}, true, nil
}

func (a *agent) GetAgentReferralsByRequestID(ctx context.Context, requestID string, limit, offset int) ([]dto.AgentReferral, error) {
	referrals, err := a.db.Queries.GetAgentReferralsByRequestID(ctx, db.GetAgentReferralsByRequestIDParams{
		RequestID: requestID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		a.log.Error("unable to get agent referrals by request ID", zap.Error(err), zap.String("requestID", requestID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get agent referrals by request ID")
		return nil, err
	}

	var result []dto.AgentReferral
	for _, ref := range referrals {
		result = append(result, dto.AgentReferral{
			ID:               ref.ID,
			RequestID:        ref.RequestID,
			UserID:           dto.NullToUUID(ref.UserID),
			ConversionType:   dto.GetStringValue(ref.ConversionType),
			Amount:           ref.Amount.Decimal,
			MSISDN:           ref.Msisdn.String,
			ConvertedAt:      ref.ConvertedAt.Time,
			CallbackSent:     ref.CallbackSent.Bool,
			CallbackAttempts: int(ref.CallbackAttempts.Int32),
		})
	}

	return result, nil
}

func (a *agent) CountAgentReferralsByRequestID(ctx context.Context, requestID string) (int, error) {
	count, err := a.db.Queries.CountAgentReferralsByRequestID(ctx, requestID)
	if err != nil {
		a.log.Error("unable to count agent referrals by request ID", zap.Error(err), zap.String("requestID", requestID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to count agent referrals by request ID")
		return 0, err
	}

	return int(count), nil
}

func (a *agent) GetReferralsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.AgentReferral, error) {
	referrals, err := a.db.Queries.GetReferralsByUserID(ctx, db.GetReferralsByUserIDParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		a.log.Error("unable to get referrals by user ID", zap.Error(err), zap.String("userID", userID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get referrals by user ID")
		return nil, err
	}

	var result []dto.AgentReferral
	for _, ref := range referrals {
		result = append(result, dto.AgentReferral{
			ID:               ref.ID,
			RequestID:        ref.RequestID,
			UserID:           dto.NullToUUID(ref.UserID),
			ConversionType:   dto.GetStringValue(ref.ConversionType),
			Amount:           ref.Amount.Decimal,
			MSISDN:           ref.Msisdn.String,
			ConvertedAt:      ref.ConvertedAt.Time,
			CallbackSent:     ref.CallbackSent.Bool,
			CallbackAttempts: int(ref.CallbackAttempts.Int32),
		})
	}

	return result, nil
}

func (a *agent) CountReferralsByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := a.db.Queries.CountReferralsByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil})
	if err != nil {
		a.log.Error("unable to count referrals by user ID", zap.Error(err), zap.String("userID", userID.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to count referrals by user ID")
		return 0, err
	}

	return int(count), nil
}

func (a *agent) GetPendingCallbacks(ctx context.Context, limit, offset int) ([]dto.PendingCallback, error) {
	callbacks, err := a.db.Queries.GetPendingCallbacks(ctx, db.GetPendingCallbacksParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		a.log.Error("unable to get pending callbacks", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get pending callbacks")
		return nil, err
	}

	var result []dto.PendingCallback
	for _, cb := range callbacks {
		result = append(result, dto.PendingCallback{
			ID:               cb.ID,
			RequestID:        cb.RequestID,
			CallbackURL:      cb.CallbackUrl,
			UserID:           dto.NullToUUID(cb.UserID),
			ConversionType:   dto.GetStringValue(cb.ConversionType),
			Amount:           cb.Amount.Decimal,
			MSISDN:           cb.Msisdn.String,
			ConvertedAt:      cb.ConvertedAt.Time,
			CallbackAttempts: int(cb.CallbackAttempts.Int32),
		})
	}

	return result, nil
}

func (a *agent) MarkCallbackSent(ctx context.Context, referralID uuid.UUID) error {
	err := a.db.Queries.MarkCallbackSent(ctx, referralID)
	if err != nil {
		a.log.Error("unable to mark callback as sent", zap.Error(err), zap.String("referralID", referralID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to mark callback as sent")
		return err
	}

	return nil
}

func (a *agent) IncrementCallbackAttempts(ctx context.Context, referralID uuid.UUID) error {
	err := a.db.Queries.IncrementCallbackAttempts(ctx, referralID)
	if err != nil {
		a.log.Error("unable to increment callback attempts", zap.Error(err), zap.String("referralID", referralID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to increment callback attempts")
		return err
	}

	return nil
}

func (a *agent) GetReferralStatsByRequestID(ctx context.Context, requestID string) (dto.ReferralStats, error) {
	stats, err := a.db.Queries.GetReferralStatsByRequestID(ctx, requestID)
	if err != nil {
		a.log.Error("unable to get referral stats by request ID", zap.Error(err), zap.String("requestID", requestID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get referral stats by request ID")
		return dto.ReferralStats{}, err
	}

	return dto.ReferralStats{
		TotalConversions: int(stats.TotalConversions),
		TotalAmount:      decimal.NewFromInt(stats.TotalAmount),
		UniqueUsers:      int(stats.UniqueUsers),
	}, nil
}

func (a *agent) GetReferralStatsByConversionType(ctx context.Context, requestID string) ([]dto.ConversionTypeStats, error) {
	stats, err := a.db.Queries.GetReferralStatsByConversionType(ctx, requestID)
	if err != nil {
		a.log.Error("unable to get referral stats by conversion type", zap.Error(err), zap.String("requestID", requestID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get referral stats by conversion type")
		return nil, err
	}

	var result []dto.ConversionTypeStats
	for _, stat := range stats {
		result = append(result, dto.ConversionTypeStats{
			ConversionType:   dto.GetStringValue(stat.ConversionType),
			TotalConversions: int(stat.TotalConversions),
			TotalAmount:      decimal.NewFromInt(stat.TotalAmount),
		})
	}

	return result, nil
}

func (a *agent) CreateAgentProvider(ctx context.Context, req dto.CreateAgentProviderReq) (dto.AgentProviderRes, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.ClientSecret), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash client secret", zap.Error(err))
		return dto.AgentProviderRes{}, err
	}
	provider, err := a.db.Queries.CreateAgentProvider(ctx, db.CreateAgentProviderParams{
		ClientID:     req.ClientID,
		ClientSecret: string(hash),
		Status:       "active",
		Name:         req.Name,
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
		CallbackUrl:  sql.NullString{String: strings.TrimSpace(req.CallbackURL), Valid: req.CallbackURL != ""},
	})
	if err != nil {
		a.log.Error("failed to create agent provider", zap.Error(err))
		return dto.AgentProviderRes{}, err
	}
	return dto.AgentProviderRes{
		ID:          provider.ID,
		Name:        provider.Name,
		ClientID:    provider.ClientID,
		Description: provider.Description.String,
		CallbackURL: provider.CallbackUrl.String,
		Status:      provider.Status,
		CreatedAt:   provider.CreatedAt,
		UpdatedAt:   provider.UpdatedAt,
	}, nil
}

func (a *agent) GetAgentProviderByClientID(ctx context.Context, clientID string) (db.AgentProvider, error) {
	return a.db.Queries.GetAgentProviderByClientID(ctx, clientID)
}

func (a *agent) GetAgentProviderByID(ctx context.Context, id uuid.UUID) (db.AgentProvider, error) {
	return a.db.Queries.GetAgentProviderByID(ctx, id)
}

func (a *agent) ListAgentProviders(ctx context.Context, limit, offset int) ([]dto.AgentProviderRes, error) {
	providers, err := a.db.Queries.ListAgentProviders(ctx, db.ListAgentProvidersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	var res []dto.AgentProviderRes
	for _, p := range providers {
		res = append(res, dto.AgentProviderRes{
			ID:          p.ID,
			Name:        p.Name,
			ClientID:    p.ClientID,
			Description: p.Description.String,
			CallbackURL: p.CallbackUrl.String,
			Status:      p.Status,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}
	return res, nil
}

func (a *agent) UpdateAgentProviderStatus(ctx context.Context, id uuid.UUID, status string) error {
	return a.db.Queries.UpdateAgentProviderStatus(ctx, db.UpdateAgentProviderStatusParams{
		ID:     id,
		Status: status,
	})
}

func (a *agent) ValidateAgentProviderCredentials(ctx context.Context, clientID, clientSecret string) (dto.AgentProviderRes, error) {
	provider, err := a.GetAgentProviderByClientID(ctx, clientID)
	if err != nil {
		return dto.AgentProviderRes{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(provider.ClientSecret), []byte(clientSecret)); err != nil {
		return dto.AgentProviderRes{}, errors.ErrInvalidUserInput.New("invalid credentials")
	}
	return dto.AgentProviderRes{
		ID:          provider.ID,
		Name:        provider.Name,
		ClientID:    provider.ClientID,
		Description: provider.Description.String,
		CallbackURL: provider.CallbackUrl.String,
		Status:      provider.Status,
		CreatedAt:   provider.CreatedAt,
		UpdatedAt:   provider.UpdatedAt,
	}, nil
}
