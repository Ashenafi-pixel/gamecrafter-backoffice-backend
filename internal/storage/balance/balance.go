package balance

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
	var amountUnits decimal.Decimal
	if dbBalance.AmountUnits.Valid {
		amountUnits = dbBalance.AmountUnits.Decimal
	} else {
		amountUnits = decimal.Zero
	}

	var reservedUnits decimal.Decimal
	if dbBalance.ReservedUnits.Valid {
		reservedUnits = dbBalance.ReservedUnits.Decimal
	} else {
		reservedUnits = decimal.Zero
	}

	var updateAt time.Time
	if dbBalance.UpdatedAt.Valid {
		updateAt = dbBalance.UpdatedAt.Time
	}

	return dto.Balance{
		ID:            dbBalance.ID,
		UserId:        dbBalance.UserID,
		CurrencyCode:  dbBalance.CurrencyCode,
		AmountCents:   dbBalance.AmountCents,
		AmountUnits:   amountUnits,
		ReservedCents: dbBalance.ReservedCents,
		ReservedUnits: reservedUnits,
		UpdateAt:      updateAt,
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
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (with correct column names)
		var id uuid.UUID
		var userID uuid.UUID
		var currencyCode string
		var amountCents int64
		var amountUnits decimal.Decimal
		var reservedCents int64
		var reservedUnits decimal.Decimal
		var updatedAt time.Time

		// Convert RealMoney to cents and units for the actual database schema
		amountCents = createBalanceReq.AmountUnits.Mul(decimal.NewFromInt(100)).IntPart()
		amountUnits = createBalanceReq.AmountUnits
		reservedCents = createBalanceReq.ReservedUnits.Mul(decimal.NewFromInt(100)).IntPart()
		reservedUnits = createBalanceReq.ReservedUnits

		err := b.db.GetPool().QueryRow(ctx, `
			INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7) 
			RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
		`, createBalanceReq.UserId, createBalanceReq.CurrencyCode, amountCents, amountUnits, reservedCents, reservedUnits, time.Now()).Scan(
			&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
		)
		if err != nil {
			b.log.Error("unable to create balance ", zap.Error(err), zap.Any("user", createBalanceReq))
			err = errors.ErrUnableTocreate.Wrap(err, "unable to create balance ", zap.Any("user", createBalanceReq))
			return dto.Balance{}, err
		}

		return dto.Balance{
			ID:            id,
			UserId:        userID,
			CurrencyCode:  currencyCode,
			AmountUnits:   amountUnits,   // amount_units maps to real_money
			ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
			ReservedCents: 0,             // Server DB doesn't have points, set to 0
			UpdateAt:      updatedAt,
		}, nil
	}

	// Use manual SQL to avoid SQLC column mapping issues
	var id uuid.UUID
	var userID uuid.UUID
	var currencyCode string
	var amountCents int64
	var amountUnits decimal.Decimal
	var reservedCents int64
	var reservedUnits decimal.Decimal
	var updatedAt time.Time

	err := b.db.GetPool().QueryRow(ctx, `
		INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
	`, createBalanceReq.UserId, createBalanceReq.CurrencyCode, 0, createBalanceReq.AmountUnits, 0, createBalanceReq.ReservedUnits, time.Now()).Scan(
		&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
	)
	if err != nil {
		b.log.Error("unable to create balance", zap.Error(err), zap.Any("user", createBalanceReq))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create balance")
		return dto.Balance{}, err
	}

	return dto.Balance{
		ID:           id,
		UserId:       userID,
		CurrencyCode: currencyCode,
		RealMoney:    amountUnits,   // amount_units maps to real_money
		BonusMoney:   reservedUnits, // reserved_units maps to bonus_money
		Points:       0,             // Points not stored in balances table
		UpdateAt:     updatedAt,
	}, nil
}

func (b *balance) GetUserBalanaceByUserID(ctx context.Context, getBalanceReq dto.Balance) (dto.Balance, bool, error) {
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
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
			ID:            id,
			UserId:        userID,
			CurrencyCode:  currencyCode,
			AmountUnits:   realMoney,
			ReservedUnits: bonusMoney,
			ReservedCents: int64(points),
			UpdateAt:      updateAt,
		}

		return balance, true, nil
	}

	// Use original query for local development
	blc, err := b.db.Queries.GetUserBalanaceByUserIDAndCurrency(ctx, db.GetUserBalanaceByUserIDAndCurrencyParams{
		UserID:       getBalanceReq.UserId,
		CurrencyCode: getBalanceReq.CurrencyCode,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Balance{}, false, nil
		}
		b.log.Error("unable to make get balance request using user_id", zap.Error(err), zap.Any("getBalanceReq", getBalanceReq))
		err = errors.ErrUnableToGet.Wrap(err, "unable to make get balance request using user_id", zap.Error(err), zap.Any("getBalanceReq", getBalanceReq))
		return dto.Balance{}, false, err
	}

	return convertDBBalanceToDTO(blc), true, nil
}

func (b *balance) UpdateBalance(ctx context.Context, updatedBalance dto.Balance) (dto.Balance, error) {

	ubalance, err := b.db.UpdateBalance(ctx, db.UpdateBalanceParams{
		CurrencyCode:  updatedBalance.CurrencyCode,
		AmountUnits:   updatedBalance.AmountUnits,
		ReservedUnits: updatedBalance.ReservedUnits,
		ReservedCents: int32(updatedBalance.ReservedCents),
		UpdatedAt:     time.Now(),
		UserID:        updatedBalance.UserId,
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("updateBalance", updatedBalance))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Balance{}, err
	}
	return convertDBBalanceToDTO(ubalance), err
}

func (b *balance) GetBalancesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error) {
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (with correct column names)
		rows, err := b.db.GetPool().Query(ctx, `
			SELECT id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at 
			FROM balances 
			WHERE user_id = $1
		`, userID)
		if err != nil {
			b.log.Error("unable to get balances by user_id", zap.Error(err), zap.String("userID", userID.String()))
			return []dto.Balance{}, err
		}
		defer rows.Close()

		var balances []dto.Balance
		for rows.Next() {
			var id uuid.UUID
			var userID uuid.UUID
			var currencyCode string
			var amountCents int64
			var amountUnits decimal.Decimal
			var reservedCents int64
			var reservedUnits decimal.Decimal
			var updatedAt time.Time

			err := rows.Scan(&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt)
			if err != nil {
				b.log.Error("unable to scan balance row", zap.Error(err))
				continue
			}

			// Convert server database format to DTO format
			balances = append(balances, dto.Balance{
				ID:            id,
				UserId:        userID,
				CurrencyCode:  currencyCode,
				AmountUnits:   amountUnits,   // amount_units maps to real_money
				ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
				ReservedCents: 0,             // Server DB doesn't have points, set to 0
				UpdateAt:      updatedAt,
			})
		}

		return balances, nil
	}

	// Use original query for local development
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
			ID:            id,
			UserId:        userID,
			CurrencyCode:  currencyCode,
			AmountUnits:   realMoney,
			ReservedUnits: bonusMoney,
			ReservedCents: int64(points),
			UpdateAt:      updateAt,
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
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (with correct column names)
		var id uuid.UUID
		var userID uuid.UUID
		var currencyCode string
		var amountCents int64
		var amountUnits decimal.Decimal
		var reservedCents int64
		var reservedUnits decimal.Decimal
		var updatedAt time.Time

		// Check if balance exists for this user and currency
		var exists bool
		err := b.db.GetPool().QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM balances WHERE user_id = $1 AND currency_code = $2)
		`, updateReq.UserID, updateReq.Currency).Scan(&exists)
		if err != nil {
			b.log.Error("unable to check balance existence", zap.Error(err), zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}

		b.log.Info("UpdateMoney - Balance existence check", zap.Bool("exists", exists), zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency), zap.String("amount", updateReq.Amount.String()))

		// Create balance if it doesn't exist
		if !exists {
			err = b.db.GetPool().QueryRow(ctx, `
				INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7) 
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
			`, updateReq.UserID, updateReq.Currency, 0, decimal.Zero, 0, decimal.Zero, time.Now()).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
			)
			if err != nil {
				b.log.Error("unable to create balance", zap.Error(err), zap.Any("updateReq", updateReq))
				return dto.Balance{}, err
			}
		}

		// Update balance based on component
		switch updateReq.Component {
		case constant.REAL_MONEY:
			amountCentsToAdd := updateReq.Amount.Mul(decimal.NewFromInt(100)).IntPart()
			b.log.Info("UpdateMoney - Executing UPDATE query", zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency), zap.String("amount", updateReq.Amount.String()), zap.Int64("amountCentsToAdd", amountCentsToAdd))
			err = b.db.GetPool().QueryRow(ctx, `
				UPDATE balances 
				SET amount_cents = amount_cents + $1, amount_units = amount_units + $2::decimal, updated_at = $3 
				WHERE user_id = $4 AND currency_code = $5
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
			`, amountCentsToAdd, updateReq.Amount, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
			)
			if err != nil {
				b.log.Error("UpdateMoney - UPDATE query failed", zap.Error(err), zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency))
			} else {
				b.log.Info("UpdateMoney - UPDATE query successful", zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency), zap.String("newAmountUnits", amountUnits.String()), zap.Int64("newAmountCents", amountCents))
			}
		case constant.BONUS_MONEY:
			reservedCents := updateReq.Amount.Mul(decimal.NewFromInt(100)).IntPart()
			err = b.db.GetPool().QueryRow(ctx, `
				UPDATE balances 
				SET reserved_cents = reserved_cents + $1, reserved_units = reserved_units + $2, updated_at = $3 
				WHERE user_id = $4 AND currency_code = $5
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
			`, reservedCents, updateReq.Amount, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
			)
		}

		if err != nil {
			b.log.Error("unable to update balance", zap.Error(err), zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}

		return dto.Balance{
			ID:            id,
			UserId:        userID,
			CurrencyCode:  currencyCode,
			AmountUnits:   amountUnits,   // amount_units maps to real_money
			ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
			ReservedCents: 0,             // Server DB doesn't have points, set to 0
			UpdateAt:      updatedAt,
		}, nil
	}

	// Original code for local development...
	var blnc db.Balance
	var err error
	// check if the user balance exist and if not create balance
	exist, err := b.db.Queries.BalanceExist(ctx, db.BalanceExistParams{
		UserID:       updateReq.UserID,
		CurrencyCode: updateReq.Currency,
	})
	if err != nil {
		b.log.Error("unable to get balance ", zap.Error(err), zap.Any("updateReq", updateReq))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to get balance ", zap.Any("updateReq", updateReq))
		return dto.Balance{}, err
	}
	if !exist {
		_, err = b.db.Queries.CreateBalance(ctx, db.CreateBalanceParams{
			UserID:        updateReq.UserID,
			CurrencyCode:  updateReq.Currency,
			AmountCents:   0,
			AmountUnits:   decimal.Zero,
			ReservedCents: 0,
			ReservedUnits: decimal.Zero,
			UpdatedAt:     time.Now(),
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
			AmountUnits:   updateReq.Amount,
			ReservedUnits: decimal.Zero,
			UpdatedAt:     time.Now(),
			UserID:        updateReq.UserID,
			CurrencyCode:  updateReq.Currency,
		})
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ", zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}

	case constant.BONUS_MONEY:
		blnc, err = b.db.Queries.UpdateReservedUnits(ctx, db.UpdateReservedUnitsParams{
			ReservedCents: int32(updateReq.Amount.IntPart()),
			UpdatedAt:     time.Now(),
			UserID:        updateReq.UserID,
			CurrencyCode:  updateReq.Currency,
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
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (with correct column names)
		var id uuid.UUID
		var userID uuid.UUID
		var adminID uuid.UUID
		var transactionID string
		var fundType string
		var currencyCode string
		var reason string
		var note string
		var createdAt time.Time

		// Convert amount to cents (assuming amount is in units, convert to cents)
		amountCents := fund.Amount.Mul(decimal.NewFromInt(100)).IntPart()

		err := b.db.GetPool().QueryRow(ctx, `
			INSERT INTO manual_funds (user_id, admin_id, transaction_id, type, amount_cents, currency_code, reason, note, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
			RETURNING id, user_id, admin_id, transaction_id, type, amount_cents, currency_code, reason, note, created_at
		`, fund.UserID, fund.AdminID, fund.TransactionID, fund.Type, amountCents, fund.Currency, fund.Reason, fund.Note, time.Now()).Scan(
			&id, &userID, &adminID, &transactionID, &fundType, &amountCents, &currencyCode, &reason, &note, &createdAt,
		)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("fund-req", fund))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.ManualFundRes{}, err
		}

		// Convert cents back to units for response
		amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))

		return dto.ManualFundRes{
			Message: constant.SUCCESS,
			Data: dto.ManualFundResData{
				ID:            id,
				UserID:        userID,
				AdminID:       adminID,
				TransactionID: transactionID,
				Amount:        amount,
				Reason:        reason,
				Currency:      currencyCode,
				Note:          note,
				CreatedAt:     createdAt,
			},
		}, nil
	}

	// Use original query with currency for local development
	// Convert amount to cents for database storage
	amountCents := fund.Amount.Mul(decimal.NewFromInt(100)).IntPart()
	res, err := b.db.Queries.SaveManualFund(ctx, db.SaveManualFundParams{
		UserID:        fund.UserID,
		AdminID:       fund.AdminID,
		TransactionID: fund.TransactionID,
		Type:          fund.Type,
		AmountCents:   amountCents,
		Reason:        fund.Reason,
		CurrencyCode:  fund.Currency,
		Note:          fund.Note,
		CreatedAt:     time.Now().In(time.Now().Location()),
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("fund-req", fund))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.ManualFundRes{}, err
	}
	// Convert cents back to units for response
	amount := decimal.NewFromInt(res.AmountCents).Div(decimal.NewFromInt(100))

	return dto.ManualFundRes{
		Message: constant.SUCCESS,
		Data: dto.ManualFundResData{
			ID:            res.ID,
			UserID:        res.UserID,
			AdminID:       res.AdminID,
			TransactionID: res.TransactionID,
			Amount:        amount,
			Reason:        res.Reason,
			Currency:      res.CurrencyCode,
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
