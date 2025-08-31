package agent

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type agent struct {
	storage storage.Agent
	log     *zap.Logger
	cron    *cron.Cron
}

func Init(storage storage.Agent, log *zap.Logger) module.Agent {
	a := &agent{
		storage: storage,
		log:     log,
		cron:    cron.New(cron.WithSeconds()),
	}

	a.StartCallbackProcessor()

	return a
}

func (a *agent) StartCallbackProcessor() {
	_, err := a.cron.AddFunc("@every 1m", func() {
		err := a.ProcessPendingCallbacks(context.Background())
		if err != nil {
			a.log.Error("Error processing agent callbacks", zap.Error(err))
		}
	})

	if err != nil {
		a.log.Error("Failed to schedule agent callback processor", zap.Error(err))
	}
	a.cron.Start()
	a.log.Info("Agent callback cron started")
}

func (a *agent) CreateAgentReferralLink(ctx context.Context, req dto.CreateAgentReferralLinkReq) (dto.CreateAgentReferralLinkRes, error) {
	_, exists, err := a.storage.GetAgentReferralByRequestID(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to check existing referral", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.CreateAgentReferralLinkRes{}, errors.ErrUnableToGet.Wrap(err, "unable to check existing referral")
	}

	authURL := viper.GetString("rise_and_hustle.auth_url")

	if exists {
		a.log.Warn("agent referral link already exists", zap.String("requestID", req.RequestID))
		link := fmt.Sprintf("%s?request_id=%s", authURL, req.RequestID)
		return dto.CreateAgentReferralLinkRes{
			Message: "Agent referral link already exists",
			Link:    link,
		}, nil
	}

	// Fetch provider and get callback URL
	provider, err := a.storage.GetAgentProviderByID(ctx, req.ProviderID)
	if err != nil {
		a.log.Error("unable to fetch agent provider", zap.Error(err), zap.Any("provider_id", req.ProviderID))
		return dto.CreateAgentReferralLinkRes{}, errors.ErrUnableToGet.Wrap(err, "unable to fetch agent provider")
	}
	req.CallbackURL = provider.CallbackUrl.String

	_, err = a.storage.CreateAgentReferralLink(ctx, req)
	if err != nil {
		a.log.Error("unable to create agent referral link", zap.Error(err), zap.Any("request", req))
		return dto.CreateAgentReferralLinkRes{}, err
	}

	link := fmt.Sprintf("%s?request_id=%s", authURL, req.RequestID)

	a.log.Info("agent referral link created successfully",
		zap.String("requestID", req.RequestID),
		zap.String("generatedLink", link))

	return dto.CreateAgentReferralLinkRes{
		Message: "Agent referral link created successfully",
		Link:    link,
	}, nil
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func (a *agent) UpdateAgentReferralWithConversion(ctx context.Context, req dto.UpdateAgentReferralWithConversionReq) (dto.UpdateAgentReferralWithConversionRes, error) {
	existingReferral, exists, err := a.storage.GetAgentReferralByRequestID(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to check existing referral", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.UpdateAgentReferralWithConversionRes{}, errors.ErrUnableToGet.Wrap(err, "unable to check existing referral")
	}

	if !exists {
		a.log.Error("agent referral link not found", zap.String("requestID", req.RequestID))
		return dto.UpdateAgentReferralWithConversionRes{}, errors.ErrResourceNotFound.Wrap(nil, "agent referral link not found")
	}

	if uuidPtr(existingReferral.UserID) != nil && *uuidPtr(existingReferral.UserID) == req.UserID {
		a.log.Warn("user already converted for this request", zap.String("userID", req.UserID.String()), zap.String("requestID", req.RequestID))
		return dto.UpdateAgentReferralWithConversionRes{
			Message:       "User already converted for this request",
			AgentReferral: existingReferral,
		}, nil
	}

	referral, err := a.storage.UpdateAgentReferralWithConversion(ctx, req)
	if err != nil {
		a.log.Error("unable to update agent referral with conversion", zap.Error(err), zap.Any("request", req))
		return dto.UpdateAgentReferralWithConversionRes{}, err
	}

	a.log.Info("agent referral conversion recorded successfully",
		zap.String("requestID", req.RequestID),
		zap.String("userID", req.UserID.String()),
		zap.String("conversionType", req.ConversionType),
		zap.String("amount", req.Amount.String()))

	return dto.UpdateAgentReferralWithConversionRes{
		Message:       "Agent referral conversion recorded successfully",
		AgentReferral: referral,
	}, nil
}

func (a *agent) GetAgentReferralsByRequestID(ctx context.Context, req dto.GetAgentReferralsReq) (dto.GetAgentReferralsRes, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	offset := (req.Page - 1) * req.Limit

	referrals, err := a.storage.GetAgentReferralsByRequestID(ctx, req.RequestID, req.Limit, offset)
	if err != nil {
		a.log.Error("unable to get agent referrals", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.GetAgentReferralsRes{}, err
	}
	total, err := a.storage.CountAgentReferralsByRequestID(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to count agent referrals", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.GetAgentReferralsRes{}, err
	}
	totalPages := (total + req.Limit - 1) / req.Limit

	return dto.GetAgentReferralsRes{
		AgentReferrals: referrals,
		Total:          total,
		Page:           req.Page,
		Limit:          req.Limit,
		TotalPages:     totalPages,
	}, nil
}

func (a *agent) GetReferralsByUserID(ctx context.Context, req dto.GetReferralsByUserReq) (dto.GetReferralsByUserRes, error) {
	offset := (req.Page - 1) * req.PerPage

	referrals, err := a.storage.GetReferralsByUserID(ctx, req.UserID, req.PerPage, offset)
	if err != nil {
		a.log.Error("unable to get referrals by user", zap.Error(err), zap.String("userID", req.UserID.String()))
		return dto.GetReferralsByUserRes{}, err
	}

	return dto.GetReferralsByUserRes{
		AgentReferrals: referrals,
	}, nil
}

func (a *agent) GetReferralStatsByRequestID(ctx context.Context, req dto.GetReferralStatsReq) (dto.GetReferralStatsRes, error) {
	stats, err := a.storage.GetReferralStatsByRequestID(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to get referral stats", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.GetReferralStatsRes{}, err
	}

	conversionTypeStats, err := a.storage.GetReferralStatsByConversionType(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to get conversion type stats", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.GetReferralStatsRes{}, err
	}

	conversionTypes := make(map[string]dto.ConversionTypeStats)
	for _, stat := range conversionTypeStats {
		conversionTypes[stat.ConversionType] = dto.ConversionTypeStats{
			TotalConversions: stat.TotalConversions,
			TotalAmount:      stat.TotalAmount,
		}
	}

	stats.ConversionTypes = conversionTypes

	return dto.GetReferralStatsRes{
		Message: "Referral statistics retrieved successfully",
		Stats:   stats,
	}, nil
}

func (a *agent) ProcessPendingCallbacks(ctx context.Context) error {
	pendingCallbacks, err := a.storage.GetPendingCallbacks(ctx, 100, 0)
	if err != nil {
		a.log.Error("unable to get pending callbacks", zap.Error(err))
		return err
	}

	if len(pendingCallbacks) == 0 {
		a.log.Debug("no pending callbacks to process")
		return nil
	}

	a.log.Info("processing pending callbacks", zap.Int("count", len(pendingCallbacks)))

	for _, callback := range pendingCallbacks {
		referral := dto.AgentReferral{
			ID:               callback.ID,
			RequestID:        callback.RequestID,
			CallbackURL:      callback.CallbackURL,
			UserID:           callback.UserID,
			ConversionType:   callback.ConversionType,
			Amount:           callback.Amount,
			MSISDN:           callback.MSISDN,
			ConvertedAt:      callback.ConvertedAt,
			CallbackSent:     false,
			CallbackAttempts: callback.CallbackAttempts,
		}

		err := a.SendCallback(ctx, referral)
		if err != nil {
			a.log.Error("failed to send callback",
				zap.Error(err),
				zap.String("callbackID", callback.ID.String()),
				zap.String("requestID", callback.RequestID))
		} else {
			a.log.Info("callback processed successfully",
				zap.String("callbackID", callback.ID.String()),
				zap.String("requestID", callback.RequestID),
				zap.String("userID", callback.UserID.String()))
		}
	}

	return nil
}

func (a *agent) GetAgentReferralByRequestID(ctx context.Context, req dto.GetAgentReferralReq) (dto.AgentReferral, bool, error) {
	referral, exists, err := a.storage.GetAgentReferralByRequestID(ctx, req.RequestID)
	if err != nil {
		a.log.Error("unable to get agent referral by request ID", zap.Error(err), zap.String("requestID", req.RequestID))
		return dto.AgentReferral{}, false, errors.ErrUnableToGet.Wrap(err, "unable to get agent referral by request ID")
	}

	return referral, exists, nil
}

func (a *agent) SendCallback(ctx context.Context, referral dto.AgentReferral) error {

	// Prepare callback payload
	callbackPayload := map[string]interface{}{
		"request_id":   referral.RequestID,
		"msisdn":       referral.MSISDN,
		"converted_at": referral.ConvertedAt.Format(time.RFC3339),
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(callbackPayload)
	if err != nil {
		a.log.Error("Failed to marshal callback payload", zap.Error(err))
		return err
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		referral.CallbackURL,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		a.log.Error("Failed to create HTTP request", zap.Error(err))
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TucanBIT-Agent-Referral/1.0")

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		a.log.Error("Failed to send callback", zap.Error(err))
		// Increment callback attempts
		_ = a.storage.IncrementCallbackAttempts(ctx, referral.ID)
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success - mark callback as sent
		err = a.storage.MarkCallbackSent(ctx, referral.ID)
		if err != nil {
			a.log.Error("Failed to mark callback as sent", zap.Error(err))
			return err
		}

		a.log.Info("Callback sent successfully",
			zap.String("requestID", referral.RequestID),
			zap.String("callbackURL", referral.CallbackURL),
			zap.Int("statusCode", resp.StatusCode))
	} else {
		// Failed - increment attempts
		a.log.Error("Callback failed",
			zap.String("requestID", referral.RequestID),
			zap.String("callbackURL", referral.CallbackURL),
			zap.Int("statusCode", resp.StatusCode))

		err = a.storage.IncrementCallbackAttempts(ctx, referral.ID)
		if err != nil {
			a.log.Error("Failed to increment callback attempts", zap.Error(err))
		}
	}

	return nil
}

func (a *agent) CreateAgentProvider(ctx context.Context, req dto.CreateAgentProviderReq) (dto.CreateAgentProviderRes, error) {
	provider, err := a.storage.CreateAgentProvider(ctx, req)
	if err != nil {
		a.log.Error("failed to create agent provider", zap.Error(err))
		return dto.CreateAgentProviderRes{}, err
	}
	return dto.CreateAgentProviderRes{
		Message:  "Agent provider created successfully",
		Provider: provider,
	}, nil
}

func (a *agent) ValidateAgentProviderCredentials(ctx context.Context, clientID string, secret string) (dto.AgentProviderRes, error) {
	provider, err := a.storage.GetAgentProviderByClientID(ctx, clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.AgentProviderRes{}, errors.ErrInvalidUserInput.New("provider not found")
		}
		return dto.AgentProviderRes{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(provider.ClientSecret), []byte(secret)); err != nil {
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
