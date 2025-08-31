package sportsservice

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type sportsservice struct {
	logger               *zap.Logger
	apiKey               string
	apiSecret            string
	sportsStorage        storage.Sports
	balanceLogStorage    storage.BalanceLogs
	operationalGroup     storage.OperationalGroup
	operationalGroupType storage.OperationalGroupType
}

func Init(logger *zap.Logger, apikey, apiSecret string, sportsStorage storage.Sports, balanceLogStorage storage.BalanceLogs, operationalGroup storage.OperationalGroup, operationalGroupType storage.OperationalGroupType) module.SportsService {
	return &sportsservice{
		logger:               logger,
		apiKey:               apikey,
		apiSecret:            apiSecret,
		sportsStorage:        sportsStorage,
		balanceLogStorage:    balanceLogStorage,
		operationalGroup:     operationalGroup,
		operationalGroupType: operationalGroupType,
	}
}

func (s *sportsservice) SignIn(ctx context.Context, req dto.SportsServiceSignInReq) (*dto.SportsServiceSignInRes, error) {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		s.logger.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return nil, err
	}

	if req.ServiceID != s.apiKey {
		s.logger.Error("invalid service ID", zap.String("provided", req.ServiceID))
		err = errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid service ID"), "invalid service ID")
		return nil, err
	}

	if req.ServiceSecret != s.apiSecret {
		s.logger.Error("invalid service secret", zap.String("provided", req.ServiceSecret))
		err = errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid service secret"), "invalid service secret")
		return nil, err
	}

	// Generate JWT token for the sports service
	token, err := utils.GenerateSportsServiceToken(req.ServiceID, "Sports Service")
	if err != nil {
		s.logger.Error("error generating token", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, "error generating token")
		return nil, err
	}

	return &dto.SportsServiceSignInRes{
		Token:   token,
		Message: "Authentication successful",
	}, nil
}

// PlaceBet places a bet on a sports event
func (s *sportsservice) PlaceBet(ctx context.Context, req dto.PlaceBetRequest) (*dto.PlaceBetResponse, error) {
	return s.sportsStorage.PlaceBet(ctx, req)
}

func (s *sportsservice) AwardWinnings(ctx context.Context, req dto.SportsServiceAwardWinningsReq) (*dto.SportsServiceAwardWinningsRes, error) {
	// Call the storage layer to handle the award winnings
	res, err := s.sportsStorage.AwardWinnings(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse the updated balance from the response
	updatedBalance, err := decimal.NewFromString(res.Balance)
	if err != nil {
		s.logger.Error("error parsing updated balance", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "error parsing updated balance")
	}

	// Create or get operational group and type for balance logging
	operationalGroupAndTypeIDs, err := s.CreateOrGetOperationalGroupAndType(ctx, "sports_betting", "sports_winnings")
	if err != nil {
		s.logger.Error("failed to create or get operational group and type", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create or get operational group and type")
	}

	// Generate transaction ID for balance log
	transactionID := utils.GenerateTransactionId()

	// Parse user ID and win amount for balance log
	userID, err := uuid.Parse(req.ExternalUserID)
	if err != nil {
		s.logger.Error("error parsing user id", zap.Error(err))
		return nil, errors.ErrInvalidUserInput.Wrap(err, "invalid user id")
	}

	winAmount, err := decimal.NewFromString(req.WinAmount)
	if err != nil {
		s.logger.Error("error parsing win amount", zap.Error(err))
		return nil, errors.ErrInvalidUserInput.Wrap(err, "invalid win amount")
	}

	// Save balance log for the winnings
	_, err = s.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.NGN_CURRENCY,
		Description:        fmt.Sprintf("Sports bet winnings: %s. Transaction ID: %s", req.Description, req.TransactionID),
		ChangeAmount:       winAmount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &updatedBalance, // This will be the new balance after adding winnings
		TransactionID:      &transactionID,
		Status:             constant.COMPLTE,
	})
	if err != nil {
		err = errors.ErrUnableTocreate.Wrap(err, "unable to save balance log")
		s.logger.Error("failed to save balance log", zap.Error(err))
		return nil, err
	}

	return res, nil
}

// CreateOrGetOperationalGroupAndType creates or gets the operational group and type for sports betting
func (s *sportsservice) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
	// get operational group if not exist create group
	var operationalGroup dto.OperationalGroup
	var exist bool
	var err error
	var operationalGroupTypeID dto.OperationalGroupType

	operationalGroup, exist, err = s.operationalGroup.GetOperationalGroupByName(ctx, operationalGroupName)
	if err != nil {
		return dto.OperationalGroupAndType{}, err
	}
	if !exist {
		// create operational group
		operationalGroup, err = s.operationalGroup.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        operationalGroupName,
			Description: "Sports betting operations",
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = s.operationalGroupType.CreateOperationalType(ctx, dto.OperationalGroupType{
			GroupID:     operationalGroup.ID,
			Name:        operationalType,
			Description: "Sports betting operations",
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
	}
	// create or get operational group type if operational group exist
	if exist {
		// get operational group type
		operationalGroupTypeID, exist, err = s.operationalGroupType.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    operationalType,
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
		if !exist {
			operationalGroupTypeID, err = s.operationalGroupType.CreateOperationalType(ctx, dto.OperationalGroupType{
				GroupID: operationalGroup.ID,
				Name:    operationalType,
			})
			if err != nil {
				return dto.OperationalGroupAndType{}, err
			}
		}
	}
	return dto.OperationalGroupAndType{
		OperationalGroupID: operationalGroup.ID,
		OperationalTypeID:  operationalGroupTypeID.ID,
	}, nil
}
