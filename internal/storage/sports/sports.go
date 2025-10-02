package sports

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type sports struct {
	log *zap.Logger
	db  *persistencedb.PersistenceDB
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Sports {
	return &sports{
		log: log,
		db:  db,
	}
}

// PlaceBet places a bet on a sports event
func (s *sports) PlaceBet(ctx context.Context, req dto.PlaceBetRequest) (*dto.PlaceBetResponse, error) {
	requestedAmount, err := decimal.NewFromString(req.BetAmount)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid amount format")
		s.log.Error("error parsing requested amount", zap.Error(err))
		return nil, err
	}
	// get user balance by user id
	balance, err := s.db.Queries.GetUserBalanaceByUserIDAndCurrency(ctx, db.GetUserBalanaceByUserIDAndCurrencyParams{
		UserID:       req.UserID,
		Currency: constant.NGN_CURRENCY,
	})
	if err != nil {
		err = errors.ErrUnableToGet.Wrap(err, "error fetching user balance")
		s.log.Error("error fetching user balance", zap.Error(err))
		return nil, err
	}

	if balance.RealMoney.Decimal.LessThan(requestedAmount) {
		err = errors.ErrInvalidUserInput.Wrap(err, "insufficient balance")
		s.log.Error("insufficient balance", zap.Error(err))
		return nil, err
	}

	// update user balance by the requested amount
	updatedBalance, err := s.db.Queries.UpdateBalance(ctx, db.UpdateBalanceParams{
		UserID:     req.UserID,
		Currency:   constant.NGN_CURRENCY,
		RealMoney:  balance.RealMoney.Decimal.Sub(requestedAmount),
		BonusMoney: balance.BonusMoney.Decimal,
		Points:     balance.Points.Int32,
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		err = errors.ErrUnableToUpdate.Wrap(err, "error updating user balance")
		s.log.Error("error updating user balance", zap.Error(err))
		return nil, err
	}

	betDetails, err := json.Marshal(req.BetDetails)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "error marshalling bet details")
		s.log.Error("error marshalling bet details", zap.Error(err))
		return nil, err
	}

	status := req.BetStatus
	if req.BetStatus == "" {
		status = "PENDING"
	}

	// Parse client IP if provided
	var clientIp pgtype.Inet
	if req.ClientIP != "" {
		if err := clientIp.Set(req.ClientIP); err != nil {
			s.log.Warn("failed to parse client IP", zap.String("client_ip", req.ClientIP), zap.Error(err))
			// Continue with empty client IP if parsing fails
		}
	}

	// create a new sport bet record in the database
	bet, err := s.db.Queries.CreateSportBet(ctx, db.CreateSportBetParams{
		TransactionID:   req.TransactionID,
		BetAmount:       requestedAmount,
		BetReferenceNum: req.BetReferenceNum,
		GameReference:   req.GameReference,
		BetMode:         req.BetMode,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		UserID:          req.UserID,
		FrontendType:    sql.NullString{String: req.FrontendType, Valid: req.FrontendType != ""},
		SportIds:        sql.NullString{String: req.SportIDs, Valid: req.SportIDs != ""},
		SiteID:          req.SiteId,
		ClientIp:        sql.NullString{String: req.ClientIP, Valid: true},
		Autorecharge:    sql.NullString{String: req.Autorecharge, Valid: req.Autorecharge != ""},
		BetDetails:      pgtype.JSONB{Bytes: betDetails, Status: pgtype.Present},
		Currency:        constant.NGN_CURRENCY,
		PotentialWin:    decimal.NullDecimal{Decimal: requestedAmount, Valid: true},
		ActualWin:       decimal.NullDecimal{Decimal: decimal.Zero, Valid: false},
		Odds:            decimal.NullDecimal{Decimal: decimal.Zero, Valid: false},
		Status:          sql.NullString{String: status, Valid: true},
	})
	if err != nil {
		err = errors.ErrUnableToUpdate.Wrap(err, "error creating sport bet")
		s.log.Error("error creating sport bet", zap.Error(err))
		return nil, err
	}

	return &dto.PlaceBetResponse{
		Balance:          updatedBalance.RealMoney.Decimal.String(),
		ExtTransactionID: bet.TransactionID,
		CustomerId:       updatedBalance.UserID.String(),
		BonusAmount:      updatedBalance.BonusMoney.Decimal.String(),
	}, nil
}

func (s *sports) AwardWinnings(ctx context.Context, req dto.SportsServiceAwardWinningsReq) (*dto.SportsServiceAwardWinningsRes, error) {

	requestedAmount, err := decimal.NewFromString(req.WinAmount)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid amount format")
		s.log.Error("error parsing requested amount", zap.Error(err))
		return nil, err
	}
	userID, err := uuid.Parse(req.ExternalUserID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid user id")
		s.log.Error("error parsing user id", zap.Error(err))
		return nil, err
	}

	// First, get the existing sport bet record to verify it exists
	_, err = s.db.Queries.GetSportBet(ctx, req.TransactionID)
	if err != nil {
		err = errors.ErrUnableToGet.Wrap(err, "error fetching sport bet")
		s.log.Error("error fetching sport bet", zap.Error(err))
		return nil, err
	}

	// get user balance by user id
	balance, err := s.db.Queries.GetUserBalanaceByUserIDAndCurrency(ctx, db.GetUserBalanaceByUserIDAndCurrencyParams{
		UserID:       userID,
		Currency: constant.NGN_CURRENCY,
	})
	if err != nil {
		err = errors.ErrUnableToGet.Wrap(err, "error fetching user balance")
		s.log.Error("error fetching user balance", zap.Error(err))
		return nil, err
	}

	// update user balance by the requested amount
	updatedBalance, err := s.db.Queries.UpdateBalance(ctx, db.UpdateBalanceParams{
		UserID:     userID,
		Currency:   constant.NGN_CURRENCY,
		RealMoney:  balance.RealMoney.Decimal.Add(requestedAmount),
		BonusMoney: balance.BonusMoney.Decimal,
		Points:     balance.Points.Int32,
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		err = errors.ErrUnableToUpdate.Wrap(err, "error updating user balance")
		s.log.Error("error updating user balance", zap.Error(err))
		return nil, err
	}

	// Update the sport bet status to WIN and set the actual win amount
	_, err = s.db.Queries.UpdateSportBetStatus(ctx, db.UpdateSportBetStatusParams{
		TransactionID: req.TransactionID,
		Status:        sql.NullString{String: "PLACED", Valid: true},
		ActualWin:     decimal.NullDecimal{Decimal: requestedAmount, Valid: true},
	})
	if err != nil {
		err = errors.ErrUnableToUpdate.Wrap(err, "error updating sport bet status")
		s.log.Error("error updating sport bet status", zap.Error(err))
		return nil, err
	}

	return &dto.SportsServiceAwardWinningsRes{
		Balance:          updatedBalance.RealMoney.Decimal.String(),
		ExtTransactionID: req.TransactionID,
		AlreadyProcessed: "false",
	}, nil
}
