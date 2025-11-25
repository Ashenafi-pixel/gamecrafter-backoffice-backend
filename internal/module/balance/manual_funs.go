package balance

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *balance) AddManualFunds(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error) {

	var err error
	userLock := b.getUserLock(fund.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// Set default currency if not provided or using server database
	if fund.Currency == "" || os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		fund.Currency = constant.DEFAULT_CURRENCY
	}

	// validate inputs
	if err := b.ValidateFundReq(ctx, fund); err != nil {
		return dto.ManualFundRes{}, err
	}

	// Note: brand_id will be fetched and set by CreateBalance and UpdateMoney methods
	// which already query it from the users table, so we don't need to fetch it here

	//fund user with the specified amount
	//create or get operational group and type
	operationalGroupAndType, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.ADD_FUND)
	if err != nil {
		return dto.ManualFundRes{}, err
	}
	usrAmount, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       fund.UserID,
		CurrencyCode: constant.DEFAULT_CURRENCY, // Use default currency for server database
	})
	if err != nil {
		b.log.Error("Failed to get user balance", zap.Error(err), zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY))
		return dto.ManualFundRes{}, err
	}

	b.log.Info("Balance check result", zap.Bool("exists", exist), zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY), zap.Any("currentBalance", usrAmount))

	if !exist {
		b.log.Info("Creating new balance", zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY), zap.String("amount", fund.Amount.String()))
		_, err = b.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:       fund.UserID,
			CurrencyCode: constant.DEFAULT_CURRENCY, // Use default currency for server database
			AmountCents:  fund.Amount.Mul(decimal.NewFromInt(100)).IntPart(),
			AmountUnits:  fund.Amount,
			// BrandID will be fetched by CreateBalance from users table
		})
		if err != nil {
			b.log.Error("Failed to create balance", zap.Error(err))
			return dto.ManualFundRes{}, err
		}
		b.log.Info("Balance created successfully")
	} else {
		// update existing balance by adding the new amount
		b.log.Info("Updating existing balance", zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY), zap.String("amountToAdd", fund.Amount.String()))
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  constant.DEFAULT_CURRENCY, // Use default currency for server database
			Component: constant.REAL_MONEY,
			Amount:    fund.Amount, // Pass only the amount to add, not the total
		})
		if err != nil {
			b.log.Error("Failed to update balance", zap.Error(err))
			return dto.ManualFundRes{}, err
		}
		b.log.Info("Balance updated successfully", zap.String("newAmountUnits", updatedBalance.AmountUnits.String()), zap.Int64("newAmountCents", updatedBalance.AmountCents))
	}
	//save transaction_log
	// Note: brand_id will be fetched by SaveBalanceLogs from the users table
	blog, err := b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
		OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
		OperationalGroupType: operationalGroupAndType.OperationalTypeID,
		UpdateReq: dto.UpdateBalanceReq{
			Component:   constant.REAL_MONEY,
			Description: fund.Reason,
			Amount:      fund.Amount,
		},
		UpdateRes: dto.UpdateBalanceRes{
			Data: dto.BalanceData{
				UserID:     fund.UserID,
				Currency:   constant.DEFAULT_CURRENCY, // Use default currency for server database
				NewBalance: fund.Amount.Add(usrAmount.AmountUnits),
			},
		},
	})
	if err != nil {
		// Log the error but don't rollback the balance update for SaveBalanceLogs
		b.log.Error("SaveBalanceLogs failed, but continuing with manual fund save", zap.Error(err))
	}

	// save it to manual fund
	var transactionID string
	if blog.TransactionID != nil {
		transactionID = *blog.TransactionID
	}
	manualFund, err := b.balanceStorage.SaveManualFunds(ctx, dto.ManualFundReq{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: transactionID,
		Type:          constant.ADD_FUND,
		Amount:        fund.Amount,
		Reason:        fund.Reason,
		Currency:      constant.DEFAULT_CURRENCY, // Use default currency for server database
		Note:          fund.Reason,
	})
	if err != nil {
		// reverse transaction by subtracting the amount we added
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  constant.DEFAULT_CURRENCY, // Use default currency for server database
			Component: constant.REAL_MONEY,
			Amount:    fund.Amount.Neg(), // Subtract the amount we added
		})

		//remove from balance log
		b.balanceLogStorage.DeleteBalanceLog(ctx, blog.ID)
		return dto.ManualFundRes{}, err
	}

	// Check for deposit alerts after successful deposit
	// Use the alert service which handles both trigger creation and email sending
	if fund.Type == constant.ADD_FUND && b.alertService != nil {
		go func() {
			// Small delay to ensure database transaction is committed
			time.Sleep(500 * time.Millisecond)

			// Run in background to not block the response
			// Use the alert service's CheckDepositAlerts which handles both trigger creation and email sending
			b.log.Info("Checking deposit alerts after manual fund addition",
				zap.String("user_id", fund.UserID.String()),
				zap.String("amount", fund.Amount.String()))

			// Skip duplicate check when manually adding funds - we want to create a trigger and send email
			if err := b.alertService.CheckDepositAlerts(context.Background(), true); err != nil {
				b.log.Error("Failed to check deposit alerts after manual fund", zap.Error(err))
			} else {
				b.log.Info("Deposit alerts checked and emails sent after manual fund addition",
					zap.String("user_id", fund.UserID.String()),
					zap.String("amount", fund.Amount.String()))
			}
		}()
	}

	return manualFund, nil
}

func (b *balance) RemoveFundManualy(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error) {

	var err error
	userLock := b.getUserLock(fund.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// Set default currency if not provided or using server database
	if fund.Currency == "" || os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		fund.Currency = constant.DEFAULT_CURRENCY
	}

	// validate inputs
	if err := b.ValidateFundReq(ctx, fund); err != nil {
		return dto.ManualFundRes{}, err
	}

	//fund user with the specified amount
	//create or get operational group and type
	operationalGroupAndType, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.REMOVE_FUND)
	if err != nil {
		return dto.ManualFundRes{}, err
	}
	usrAmount, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       fund.UserID,
		CurrencyCode: constant.DEFAULT_CURRENCY, // Use default currency for server database
	})
	if err != nil {
		b.log.Error("Failed to get user balance", zap.Error(err), zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY))
		return dto.ManualFundRes{}, err
	}

	b.log.Info("Balance check result", zap.Bool("exists", exist), zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY), zap.Any("currentBalance", usrAmount))

	if !exist {
		err = fmt.Errorf("user dose not have balance with this currency")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ManualFundRes{}, err
	} else {
		// update existing balance
		//check if the balance is greater than or equal
		if fund.Amount.GreaterThan(usrAmount.AmountUnits) {
			err = fmt.Errorf("user dose not have enough balance with this currency to substract. user current balance %v", usrAmount.AmountUnits)
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.ManualFundRes{}, err
		}
		// update existing balance by subtracting the amount
		negativeAmount := fund.Amount.Neg()
		b.log.Info("Updating existing balance", zap.String("userID", fund.UserID.String()), zap.String("currency", constant.DEFAULT_CURRENCY), zap.String("amountToSubtract", fund.Amount.String()), zap.String("negativeAmount", negativeAmount.String()))
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  constant.DEFAULT_CURRENCY, // Use default currency for server database
			Component: constant.REAL_MONEY,
			Amount:    negativeAmount, // Pass negative amount to subtract from balance
		})
		if err != nil {
			b.log.Error("Failed to update balance", zap.Error(err))
			return dto.ManualFundRes{}, err
		}
		b.log.Info("Balance updated successfully", zap.String("newAmountUnits", updatedBalance.AmountUnits.String()), zap.Int64("newAmountCents", updatedBalance.AmountCents), zap.String("originalBalance", usrAmount.AmountUnits.String()))
	}
	//save transaction_log
	blog, err := b.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
		OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
		OperationalGroupType: operationalGroupAndType.OperationalTypeID,
		UpdateReq: dto.UpdateBalanceReq{
			Component:   constant.REAL_MONEY,
			Description: fund.Reason,
			Amount:      fund.Amount,
		},
		UpdateRes: dto.UpdateBalanceRes{
			Data: dto.BalanceData{
				UserID:     fund.UserID,
				Currency:   constant.DEFAULT_CURRENCY,              // Use default currency for server database
				NewBalance: usrAmount.AmountUnits.Sub(fund.Amount), // Calculate new balance after subtraction
			},
		},
	})
	if err != nil {
		// reverse amount by adding back the amount we subtracted
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  constant.DEFAULT_CURRENCY, // Use default currency for server database
			Component: constant.REAL_MONEY,
			Amount:    fund.Amount, // Add back the amount we subtracted
		})
	}

	// save it to manual fund
	var transactionID string
	if blog.TransactionID != nil {
		transactionID = *blog.TransactionID
	}
	manualFund, err := b.balanceStorage.SaveManualFunds(ctx, dto.ManualFundReq{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: transactionID,
		Type:          constant.REMOVE_FUND,
		Amount:        fund.Amount,
		Reason:        fund.Reason,
		Currency:      constant.DEFAULT_CURRENCY, // Use default currency for server database
		Note:          fund.Note,
	})
	if err != nil {
		// reverse transaction by adding back the amount we subtracted
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    fund.UserID,
			Currency:  constant.DEFAULT_CURRENCY, // Use default currency for server database
			Component: constant.REAL_MONEY,
			Amount:    fund.Amount, // Add back the amount we subtracted
		})

		//remove from balance log
		b.balanceLogStorage.DeleteBalanceLog(ctx, blog.ID)
		return dto.ManualFundRes{}, err
	}
	return manualFund, nil
}
func (b *balance) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
	// get transfer operational  group and  type if not exist create group transfer and type transfer-internal
	var operationalGroup dto.OperationalGroup
	var exist bool
	var err error
	var operationalGroupTypeID dto.OperationalGroupType
	operationalGroup, exist, err = b.operationalGroupStorage.GetOperationalGroupByName(ctx, constant.TRANSFER)
	if err != nil {
		return dto.OperationalGroupAndType{}, err
	}
	if !exist {
		// create transfer internal group and type
		operationalGroup, err = b.operationalGroupStorage.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        constant.TRANSFER,
			Description: "internal transaction",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
			GroupID:     operationalGroup.ID,
			Name:        operationalType,
			Description: "internal transactions",
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
	}
	// create or get operational group type if operational group exist
	if exist {
		// get operational group type
		operationalGroupTypeID, exist, err = b.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    operationalType,
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
		if !exist {
			operationalGroupTypeID, err = b.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
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

func (b *balance) ValidateFundReq(ctx context.Context, fund dto.ManualFundReq) error {
	var err error
	if fund.Amount.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("invalid amount is given. fund can not be less than zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	//validate currency - skip validation for server database compatibility
	// if yes := dto.IsValidCurrency(fund.Currency); !yes {
	//	err = fmt.Errorf("invalid currency is given.")
	//	err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
	//	return err
	// }
	// validate and verify user
	if fund.UserID == uuid.Nil {
		err = fmt.Errorf("invalid user_id is given. ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
	}

	if fund.Reason == "" {
		err = fmt.Errorf("reason can not be empty")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	_, exist, err := b.userStorage.GetUserByID(ctx, fund.UserID)
	if err != nil {
		return err
	}
	if !exist {
		err = fmt.Errorf("user not found ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	//validate admin
	_, exist, err = b.userStorage.GetUserByID(ctx, fund.AdminID)
	if err != nil {
		return err
	}
	if !exist {
		err = fmt.Errorf("unable to find admin with this id")
		b.log.Error(err.Error(), zap.Any("adminID", fund.AdminID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	// Check admin's funding limit from role_permissions
	maxLimit, err := b.balanceStorage.GetAdminFundingLimit(ctx, fund.AdminID)
	if err != nil {
		b.log.Error("Failed to get admin funding limit", zap.Error(err), zap.String("adminID", fund.AdminID.String()))
		// Don't fail validation if we can't get the limit, just log it and allow
	} else if maxLimit != nil {
		// Admin has a funding limit set
		if fund.Amount.GreaterThan(*maxLimit) {
			err = fmt.Errorf("funding amount %s exceeds admin's limit of %s", fund.Amount.String(), maxLimit.String())
			b.log.Warn("Funding limit exceeded",
				zap.String("adminID", fund.AdminID.String()),
				zap.String("requestedAmount", fund.Amount.String()),
				zap.String("limit", maxLimit.String()))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return err
		}
	}
	// If maxLimit is nil, admin has unlimited funding (no limit)

	return nil
}

func (b *balance) GetManualFundLogs(ctx context.Context, filter dto.GetManualFundReq) (dto.GetManualFundRes, error) {
	//offset := (filter.Page - 1) * filter.PerPage
	if filter.Sort.Amount != "" {
		if err := utils.ValidateSortOptions("amount", filter.Sort.Amount); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	if filter.Sort.AdminEmail != "" {
		if err := utils.ValidateSortOptions("admin amount", filter.Sort.AdminEmail); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	if filter.Sort.Date != "" {
		if err := utils.ValidateSortOptions("date", filter.Sort.Date); err != nil {
			return dto.GetManualFundRes{}, err
		}
	}

	return b.balanceStorage.GetManualFundLogs(ctx, filter)
}

func (b *balance) GetAllManualFunds(ctx context.Context, filter dto.GetAllManualFundsFilter) (dto.GetAllManualFundsRes, error) {
	return b.balanceStorage.GetAllManualFunds(ctx, filter)
}
