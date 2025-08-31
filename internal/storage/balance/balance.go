package balance

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type balance struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Balance {
	return &balance{
		db:  db,
		log: log,
	}
}

func (b *balance) CreateBalance(ctx context.Context, createBalanceReq dto.Balance) (dto.Balance, error) {
	blc, err := b.db.Queries.CreateBalance(ctx, db.CreateBalanceParams{
		UserID:     createBalanceReq.UserId,
		Currency:   createBalanceReq.Currency,
		RealMoney:  decimal.NullDecimal{Decimal: createBalanceReq.RealMoney, Valid: true},
		BonusMoney: decimal.NullDecimal{Decimal: createBalanceReq.BonusMoney, Valid: true},
		UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		b.log.Error("unable to create balance ", zap.Error(err), zap.Any("user", createBalanceReq))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create balance ", zap.Any("user", createBalanceReq))
		return dto.Balance{}, err
	}
	return dto.Balance{
		ID:         blc.ID,
		UserId:     blc.UserID,
		Currency:   blc.Currency,
		RealMoney:  blc.RealMoney.Decimal,
		BonusMoney: blc.BonusMoney.Decimal,
		UpdateAt:   blc.UpdatedAt.Time,
	}, nil
}

func (b *balance) GetUserBalanaceByUserID(ctx context.Context, getBalanceReq dto.Balance) (dto.Balance, bool, error) {
	blc, err := b.db.Queries.GetUserBalanaceByUserIDAndCurrency(ctx, db.GetUserBalanaceByUserIDAndCurrencyParams{
		UserID:   getBalanceReq.UserId,
		Currency: getBalanceReq.Currency,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("unable to make get balance request using user_id")
		err = errors.ErrUnableToGet.Wrap(err, "unable to make get balance request using user_id", zap.Error(err), zap.Any("getBalanceReq", getBalanceReq))
		return dto.Balance{}, false, err

	} else if err != nil && err.Error() == dto.ErrNoRows {
		return dto.Balance{}, false, nil
	}

	return dto.Balance{
		ID:         blc.ID,
		UserId:     blc.UserID,
		Currency:   blc.Currency,
		RealMoney:  blc.RealMoney.Decimal,
		BonusMoney: blc.BonusMoney.Decimal,
		UpdateAt:   blc.UpdatedAt.Time,
	}, true, nil

}

func (b *balance) UpdateBalance(ctx context.Context, updatedBalance dto.Balance) (dto.Balance, error) {

	ubalance, err := b.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Currency:   updatedBalance.Currency,
		RealMoney:  decimal.NullDecimal{Decimal: updatedBalance.RealMoney, Valid: true},
		BonusMoney: decimal.NullDecimal{Decimal: updatedBalance.BonusMoney, Valid: true},
		UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		UserID:     updatedBalance.UserId,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("updateBalance", updatedBalance))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Balance{}, err
	}
	return dto.Balance{
		ID:         ubalance.ID,
		UserId:     ubalance.UserID,
		Currency:   ubalance.Currency,
		RealMoney:  ubalance.RealMoney.Decimal,
		BonusMoney: ubalance.BonusMoney.Decimal,
		UpdateAt:   ubalance.UpdatedAt.Time,
	}, err
}

func (b *balance) GetBalancesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error) {
	balances := []dto.Balance{}
	userBalances, err := b.db.Queries.GetUserBalancesByUserID(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Balance{}, err
	}
	for _, userBalance := range userBalances {
		balances = append(balances, dto.Balance{
			ID:         userBalance.ID,
			UserId:     userBalance.UserID,
			Currency:   userBalance.Currency,
			RealMoney:  userBalance.RealMoney.Decimal,
			BonusMoney: userBalance.BonusMoney.Decimal,
			UpdateAt:   userBalance.UpdatedAt.Time,
		})
	}
	return balances, nil
}

func (b *balance) UpdateMoney(ctx context.Context, updateReq dto.UpdateBalanceReq) (dto.Balance, error) {
	var blnc db.Balance
	var err error
	// check if the user balance exist and if not create balance
	exist, err := b.db.BalanceExist(ctx, db.BalanceExistParams{
		UserID:   updateReq.UserID,
		Currency: updateReq.Currency,
	})
	if err != nil {
		b.log.Error("unable to get balance ", zap.Error(err), zap.Any("updateReq", updateReq))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to get balance ", zap.Any("updateReq", updateReq))
		return dto.Balance{}, err

	}
	if !exist {
		b.db.CreateBalance(ctx, db.CreateBalanceParams{
			UserID:   updateReq.UserID,
			Currency: updateReq.Currency,
		})
	}
	switch updateReq.Component {
	case constant.REAL_MONEY:
		blnc, err = b.db.UpdateRealMoney(ctx, db.UpdateRealMoneyParams{
			RealMoney: decimal.NullDecimal{Decimal: updateReq.Amount, Valid: true},
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			UserID:    updateReq.UserID,
			Currency:  updateReq.Currency,
		})
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ", zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}

	case constant.BONUS_MONEY:
		blnc, err = b.db.Queries.UpdateBonusMoney(ctx, db.UpdateBonusMoneyParams{
			BonusMoney: decimal.NullDecimal{Decimal: updateReq.Amount, Valid: true},
			UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
			UserID:     updateReq.UserID,
			Currency:   updateReq.Currency,
		})
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ", zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}
	}

	return dto.Balance{
		ID:         blnc.ID,
		UserId:     blnc.UserID,
		Currency:   blnc.Currency,
		RealMoney:  blnc.RealMoney.Decimal,
		BonusMoney: blnc.BonusMoney.Decimal,
		UpdateAt:   blnc.UpdatedAt.Time,
	}, nil
}

func (b *balance) SaveManualFunds(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error) {
	res, err := b.db.Queries.SaveManualFund(ctx, db.SaveManualFundParams{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: fund.TransactionID,
		Type:          fund.Type,
		Amount:        fund.Amount,
		Reason:        fund.Reason,
		Currency:      fund.Currency,
		Note:          fund.Note,
		CreatedAt:     time.Now().In(time.Now().Location()),
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("fund-req", fund))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.ManualFundRes{}, err
	}
	return dto.ManualFundRes{
		Message: constant.SUCCESS,
		Data: dto.ManualFundResData{
			ID:            res.ID,
			UserID:        res.UserID,
			AdminID:       res.AdminID,
			TransactionID: res.TransactionID,
			Amount:        res.Amount,
			Reason:        res.Reason,
			Currency:      res.Currency,
			Note:          res.Note,
			CreatedAt:     res.CreatedAt,
		},
	}, nil
}

func (b *balance) GetManualFundLogs(ctx context.Context, filter dto.GetManualFundReq) (dto.GetManualFundRes, error) {
	var query string
	var conditions []interface{}
	placeholderIndex := 1
	first := true
	orderFirst := true

	if filter.Filter.CustomerUsername != nil {
		query = query + " " + persistencedb.Where + " "
		query += fmt.Sprintf("us.username = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.CustomerUsername)
		placeholderIndex++
		first = false
	}
	if filter.Filter.CustomerEmail != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("us.email = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.CustomerEmail)
		placeholderIndex++
		first = false
	}

	if filter.Filter.CustomerPhone != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("us.phone = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.CustomerPhone)
		placeholderIndex++
		first = false
	}

	if filter.Filter.AdminEmail != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("ad.email = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.AdminEmail)
		placeholderIndex++
		first = false
	}

	if filter.Filter.AdminPhone != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("ad.phone = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.AdminPhone)
		placeholderIndex++
		first = false
	}

	if filter.Filter.AdminUsername != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("ad.username = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.AdminUsername)
		placeholderIndex++
		first = false
	}

	if filter.Filter.Type != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("mf.type = $%d", placeholderIndex)
		conditions = append(conditions, *filter.Filter.Type)
		placeholderIndex++
		first = false
	}

	if filter.Filter.StartDate != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("mf.created_at >= $%d", placeholderIndex)
		conditions = append(conditions, filter.Filter.StartDate.Format("2006-01-02"))
		placeholderIndex++
		first = false
	}
	if filter.Filter.EndDate != nil {
		if !first {
			query += " AND "
		} else {
			query = query + " " + persistencedb.Where + " "
		}

		query += fmt.Sprintf("mf.created_at <= $%d", placeholderIndex)
		conditions = append(conditions, filter.Filter.EndDate.Format("2006-01-02"))
		placeholderIndex++
		first = false
	}

	if filter.Sort.Date != "" {
		if orderFirst {
			query += " ORDER BY "
		} else {
			query += ", "
		}
		query = query + "mf.created_at" + filter.Sort.Date
		orderFirst = false
	}
	if filter.Sort.Amount != "" {
		if orderFirst {
			query += " ORDER BY "
		} else {
			query += ", "
		}
		query = query + "mf.amount" + filter.Sort.Amount
		orderFirst = false
	}

	if filter.Sort.AdminEmail != "" {
		if orderFirst {
			query += " ORDER BY "
		} else {
			query += ", "
		}
		query = query + "ad.email" + filter.Sort.AdminEmail
		orderFirst = false
	}
	query = fmt.Sprintf(persistencedb.GetManualBalanceQuery, query, filter.Page, filter.PerPage)

	return b.db.GetManualFunds(ctx, query, conditions, filter.PerPage, filter.Page)
}
