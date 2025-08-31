package adds

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type adds struct {
	storage           storage.Adds
	balanceStorage    storage.Balance
	balanceLogStorage storage.BalanceLogs
	logger            *zap.Logger
}

func Init(storage storage.Adds, balanceStorage storage.Balance, balanceLogStorage storage.BalanceLogs, logger *zap.Logger) module.Adds {
	return &adds{
		storage:           storage,
		balanceStorage:    balanceStorage,
		balanceLogStorage: balanceLogStorage,
		logger:            logger,
	}
}

func (a *adds) SaveAddsService(ctx context.Context, req dto.CreateAddsServiceReq) (*dto.CreateAddsServiceRes, error) {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		a.logger.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return nil, err
	}

	hashPassword, err := utils.Encrypt(req.ServiceSecret)
	if err != nil {
		a.logger.Error("error encrypting password", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, "error encrypting password")
		return nil, err
	}

	req.ServiceSecret = hashPassword
	data, err := a.storage.SaveAddsService(ctx, req)
	if err != nil {
		a.logger.Error("error saving adds service", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "error saving adds service")
	}
	return &data, nil

}

func (a *adds) SignIn(ctx context.Context, req dto.AddSignInReq) (*dto.AddSignInRes, error) {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		a.logger.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return nil, err
	}

	service, found, err := a.storage.GetAddsServiceByServiceID(ctx, req.ServiceID)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.ErrInvalidUserInput.Wrap(err, "service not found")
	}

	hashPassword, err := utils.Decrypt(service.ServiceSecret)
	if err != nil {
		a.logger.Error("error decrypting password", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, "error decrypting password")
		return nil, err
	}

	if hashPassword != req.ServiceSecret {
		return nil, errors.ErrInvalidUserInput.Wrap(err, "invalid service secret")
	}

	token, err := utils.GenerateAddsServiceToken(service.ID, service.Name)
	if err != nil {
		return nil, err
	}

	return &dto.AddSignInRes{Token: token}, nil
}

func (a *adds) GetAddsServices(ctx context.Context, req dto.GetAddServicesRequest) (*dto.GetAddsServicesRes, error) {
	data, err := a.storage.GetAddsServices(ctx, req)
	if err != nil {
		a.logger.Error("error getting adds services", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "error getting adds services")
	}
	return &data, nil
}

func (a *adds) UpdateBalance(ctx context.Context, req dto.AddUpdateBalanceReq) (*dto.AddUpdateBalanceRes, error) {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		a.logger.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return nil, err
	}

	// Parse amount to decimal
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		a.logger.Error("error parsing amount", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid amount format")
		return nil, err
	}

	// Validate amount is positive
	if amount.LessThanOrEqual(decimal.Zero) {
		a.logger.Error("invalid amount", zap.String("amount", req.Amount))
		err = errors.ErrInvalidUserInput.Wrap(err, "amount must be greater than zero")
		return nil, err
	}

	// Validate component
	if req.Component != constant.REAL_MONEY && req.Component != constant.BONUS_MONEY {
		a.logger.Error("invalid component", zap.String("component", req.Component))
		err = errors.ErrInvalidUserInput.Wrap(err, "component must be either 'real_money' or 'bonus_money'")
		return nil, err
	}

	// Get current user balance
	currentBalance, exists, err := a.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   req.UserID,
		Currency: req.Currency,
	})
	if err != nil {
		a.logger.Error("error getting user balance", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "error getting user balance")
	}

	var newBalance decimal.Decimal

	if !exists {
		// Create new balance if it doesn't exist
		realMoney := decimal.Zero
		bonusMoney := decimal.Zero

		if req.Component == constant.REAL_MONEY {
			realMoney = amount
			newBalance = amount
		} else {
			bonusMoney = amount
			newBalance = amount
		}

		_, err = a.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:     req.UserID,
			Currency:   req.Currency,
			RealMoney:  realMoney,
			BonusMoney: bonusMoney,
		})
		if err != nil {
			a.logger.Error("error creating balance", zap.Error(err))
			return nil, errors.ErrInternalServerError.Wrap(err, "error creating balance")
		}
	} else {
		// Update existing balance
		updateReq := dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: req.Component,
			Amount:    amount,
		}

		// Add the new amount to existing balance
		if req.Component == constant.REAL_MONEY {
			updateReq.Amount = currentBalance.RealMoney.Add(amount)
			newBalance = updateReq.Amount
		} else {
			updateReq.Amount = currentBalance.BonusMoney.Add(amount)
			newBalance = updateReq.Amount
		}

		_, err = a.balanceStorage.UpdateMoney(ctx, updateReq)
		if err != nil {
			a.logger.Error("error updating balance", zap.Error(err))
			return nil, errors.ErrInternalServerError.Wrap(err, "error updating balance")
		}
	}

	// Generate transaction ID
	transactionID := utils.GenerateTransactionId()

	// Save balance log
	_, err = a.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          req.Component,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("Adds service balance update: %s. %s", req.Reason, req.Reference),
		ChangeAmount:       amount,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
		Status:             constant.COMPLTE,
	})
	if err != nil {
		a.logger.Error("error saving balance log", zap.Error(err))
		// Note: We don't rollback the balance update here as the balance update was successful
		// The log failure is logged but doesn't affect the balance update
	}

	a.logger.Info("balance update completed successfully",
		zap.String("user_id", req.UserID.String()),
		zap.String("currency", req.Currency),
		zap.String("component", req.Component),
		zap.String("amount", req.Amount),
		zap.String("new_balance", newBalance.String()),
		zap.String("transaction_id", transactionID),
		zap.String("reason", req.Reason),
		zap.String("reference", req.Reference))

	return &dto.AddUpdateBalanceRes{
		Message: "Balance updated successfully",
		Status:  "success",
	}, nil
}
