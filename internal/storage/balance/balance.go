package balance

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

// convertDBBalanceToDTO safely converts a database Balance to DTO Balance, handling null values
func convertDBBalanceToDTO(dbBalance db.Balance) dto.Balance {
	// Handle null values properly
	var realMoney decimal.Decimal
	if dbBalance.RealMoney.Valid {
		realMoney = dbBalance.RealMoney.Decimal
	} else {
		realMoney = decimal.Zero
	}

	var bonusMoney decimal.Decimal
	if dbBalance.BonusMoney.Valid {
		bonusMoney = dbBalance.BonusMoney.Decimal
	} else {
		bonusMoney = decimal.Zero
	}

	var points int32
	if dbBalance.Points.Valid {
		points = dbBalance.Points.Int32
	}

	var updateAt time.Time
	if dbBalance.UpdatedAt.Valid {
		updateAt = dbBalance.UpdatedAt.Time
	}

	return dto.Balance{
		ID:           dbBalance.ID,
		UserId:       dbBalance.UserID,
		CurrencyCode: dbBalance.Currency,
		RealMoney:    realMoney,
		BonusMoney:   bonusMoney,
		Points:       points,
		UpdateAt:     updateAt,
	}
}

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
		Currency:   createBalanceReq.CurrencyCode,
		RealMoney:  createBalanceReq.RealMoney,
		BonusMoney: createBalanceReq.BonusMoney,
		Points:     createBalanceReq.Points,
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		b.log.Error("unable to create balance ", zap.Error(err), zap.Any("user", createBalanceReq))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create balance ", zap.Any("user", createBalanceReq))
		return dto.Balance{}, err
	}
	return convertDBBalanceToDTO(blc), nil
}

func (b *balance) GetUserBalanaceByUserID(ctx context.Context, getBalanceReq dto.Balance) (dto.Balance, bool, error) {
	query := `SELECT id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at FROM balances WHERE user_id = $1 AND currency_code = $2`
	row := b.db.GetPool().QueryRow(ctx, query, getBalanceReq.UserId, getBalanceReq.CurrencyCode)

	var id uuid.UUID
	var userID uuid.UUID
	var currencyCode string
	var amountCents int64
	var amountUnits decimal.NullDecimal
	var reservedCents int64
	var reservedUnits decimal.NullDecimal
	var updatedAt sql.NullTime

	err := row.Scan(&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Balance{}, false, nil
		}
		b.log.Error("unable to make get balance request using user_id", zap.Error(err), zap.Any("getBalanceReq", getBalanceReq))
		err = errors.ErrUnableToGet.Wrap(err, "unable to make get balance request using user_id", zap.Error(err), zap.Any("getBalanceReq", getBalanceReq))
		return dto.Balance{}, false, err
	}

	// Convert database fields to DTO
	var realMoney decimal.Decimal
	if amountUnits.Valid {
		realMoney = amountUnits.Decimal
	} else {
		realMoney = decimal.Zero
	}

	var bonusMoney decimal.Decimal
	if reservedUnits.Valid {
		bonusMoney = reservedUnits.Decimal
	} else {
		bonusMoney = decimal.Zero
	}

	var points int32
	points = int32(reservedCents)

	var updateAt time.Time
	if updatedAt.Valid {
		updateAt = updatedAt.Time
	}

	balance := dto.Balance{
		ID:           id,
		UserId:       userID,
		CurrencyCode: currencyCode,
		RealMoney:    realMoney,
		BonusMoney:   bonusMoney,
		Points:       points,
		UpdateAt:     updateAt,
	}

	return balance, true, nil
}

func (b *balance) UpdateBalance(ctx context.Context, updatedBalance dto.Balance) (dto.Balance, error) {

	ubalance, err := b.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		Currency:   updatedBalance.CurrencyCode,
		RealMoney:  updatedBalance.RealMoney,
		BonusMoney: updatedBalance.BonusMoney,
		Points:     updatedBalance.Points,
		UpdatedAt:  time.Now(),
		UserID:     updatedBalance.UserId,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("updateBalance", updatedBalance))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Balance{}, err
	}
	return convertDBBalanceToDTO(ubalance), err
}

func (b *balance) GetBalancesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error) {
	balances := []dto.Balance{}

	query := `SELECT id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at FROM balances WHERE user_id = $1`
	rows, err := b.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Balance{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var userID uuid.UUID
		var currencyCode string
		var amountCents int64
		var amountUnits decimal.NullDecimal
		var reservedCents int64
		var reservedUnits decimal.NullDecimal
		var updatedAt sql.NullTime

		err := rows.Scan(&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt)
		if err != nil {
			b.log.Error("Error scanning balance row", zap.Error(err))
			continue
		}

		// Convert database fields to DTO
		var realMoney decimal.Decimal
		if amountUnits.Valid {
			realMoney = amountUnits.Decimal
		} else {
			realMoney = decimal.Zero
		}

		var bonusMoney decimal.Decimal
		if reservedUnits.Valid {
			bonusMoney = reservedUnits.Decimal
		} else {
			bonusMoney = decimal.Zero
		}

		var points int32
		points = int32(reservedCents)

		var updateAt time.Time
		if updatedAt.Valid {
			updateAt = updatedAt.Time
		}

		balance := dto.Balance{
			ID:           id,
			UserId:       userID,
			CurrencyCode: currencyCode,
			RealMoney:    realMoney,
			BonusMoney:   bonusMoney,
			Points:       points,
			UpdateAt:     updateAt,
		}

		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		b.log.Error("Error iterating balance rows", zap.Error(err))
		return []dto.Balance{}, err
	}

	return balances, nil
}

func (b *balance) UpdateMoney(ctx context.Context, updateReq dto.UpdateBalanceReq) (dto.Balance, error) {
	var blnc db.Balance
	var err error
	// check if the user balance exist and if not create balance
	exist, err := b.db.Queries.BalanceExist(ctx, db.BalanceExistParams{
		UserID:   updateReq.UserID,
		Currency: updateReq.Currency,
	})
	if err != nil {
		b.log.Error("unable to get balance ", zap.Error(err), zap.Any("updateReq", updateReq))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to get balance ", zap.Any("updateReq", updateReq))
		return dto.Balance{}, err
	}
	if !exist {
		_, err = b.db.Queries.CreateBalance(ctx, db.CreateBalanceParams{
			UserID:     updateReq.UserID,
			Currency:   updateReq.Currency,
			RealMoney:  decimal.Zero,
			BonusMoney: decimal.Zero,
			Points:     0,
			UpdatedAt:  time.Now(),
		})
		if err != nil {
			b.log.Error("unable to create balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}
	}

	// Convert amount to cents and units
	// amountCents := updateReq.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	switch updateReq.Component {
	case constant.REAL_MONEY:
		blnc, err = b.db.Queries.UpdateAmountUnits(ctx, db.UpdateAmountUnitsParams{
			RealMoney:  updateReq.Amount,
			BonusMoney: decimal.Zero,
			UpdatedAt:  time.Now(),
			UserID:     updateReq.UserID,
			Currency:   updateReq.Currency,
		})
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ", zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}

	case constant.BONUS_MONEY:
		blnc, err = b.db.Queries.UpdateReservedUnits(ctx, db.UpdateReservedUnitsParams{
			Points:    int32(updateReq.Amount.IntPart()),
			UpdatedAt: time.Now(),
			UserID:    updateReq.UserID,
			Currency:  updateReq.Currency,
		})
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ", zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}
	}

	return convertDBBalanceToDTO(blnc), nil
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
