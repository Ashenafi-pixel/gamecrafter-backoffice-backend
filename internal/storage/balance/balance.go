package balance

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
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

		// Fetch brand_id from users table
		var brandID *uuid.UUID
		err := b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, createBalanceReq.UserId).Scan(&brandID)
		if err != nil && err != sql.ErrNoRows {
			b.log.Warn("Failed to get brand_id from user, continuing without it", zap.Error(err), zap.String("userID", createBalanceReq.UserId.String()))
		}

		err = b.db.GetPool().QueryRow(ctx, `
			INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
		`, createBalanceReq.UserId, createBalanceReq.CurrencyCode, amountCents, amountUnits, reservedCents, reservedUnits, brandID, time.Now()).Scan(
			&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
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
			AmountCents:   amountCents,
			AmountUnits:   amountUnits,   // amount_units maps to real_money
			ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
			ReservedCents: reservedCents, // Use actual reserved_cents from database
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

	// Fetch brand_id from users table
	var brandID *uuid.UUID
	err := b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, createBalanceReq.UserId).Scan(&brandID)
	if err != nil && err != sql.ErrNoRows {
		b.log.Warn("Failed to get brand_id from user, continuing without it", zap.Error(err), zap.String("userID", createBalanceReq.UserId.String()))
	}

	err = b.db.GetPool().QueryRow(ctx, `
		INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
	`, createBalanceReq.UserId, createBalanceReq.CurrencyCode, 0, createBalanceReq.AmountUnits, 0, createBalanceReq.ReservedUnits, brandID, time.Now()).Scan(
		&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
	)
	if err != nil {
		b.log.Error("unable to create balance", zap.Error(err), zap.Any("user", createBalanceReq))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create balance")
		return dto.Balance{}, err
	}

	return dto.Balance{
		ID:            id,
		UserId:        userID,
		CurrencyCode:  currencyCode,
		AmountCents:   amountCents,
		AmountUnits:   amountUnits,
		ReservedCents: reservedCents,
		ReservedUnits: reservedUnits,
		UpdateAt:      updatedAt,
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
			AmountCents:   amountCents,
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
				AmountCents:   amountCents,
				AmountUnits:   amountUnits,   // amount_units maps to real_money
				ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
				ReservedCents: reservedCents, // Use actual reserved_cents from database
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
			AmountCents:   amountCents,
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

		// Create balance if it doesn't exist - fetch brand_id from user
		if !exists {
			var brandID *uuid.UUID
			err = b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, updateReq.UserID).Scan(&brandID)
			if err != nil && err != sql.ErrNoRows {
				b.log.Error("unable to get brand_id from user", zap.Error(err), zap.String("userID", updateReq.UserID.String()))
			}

			err = b.db.GetPool().QueryRow(ctx, `
				INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
			`, updateReq.UserID, updateReq.Currency, 0, decimal.Zero, 0, decimal.Zero, brandID, time.Now()).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
			)
			if err != nil {
				b.log.Error("unable to create balance", zap.Error(err), zap.Any("updateReq", updateReq))
				return dto.Balance{}, err
			}
		}

		// Update balance based on component
		switch updateReq.Component {
		case constant.REAL_MONEY:
			b.log.Info("UpdateMoney - Starting update", zap.String("amount", updateReq.Amount.String()), zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency))

			// First update amount_units, then recalculate amount_cents from the final amount_units value
			// This ensures accuracy and prevents drift between cents and units
			// Use FLOOR to match Go's IntPart() behavior (truncate to integer)
			err = b.db.GetPool().QueryRow(ctx, `
				UPDATE balances 
				SET amount_units = amount_units + $1, 
				    amount_cents = FLOOR((amount_units + $1) * 100)::BIGINT,
				    updated_at = $2 
				WHERE user_id = $3 AND currency_code = $4
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at
			`, updateReq.Amount, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &updatedAt,
			)

			if err != nil {
				b.log.Error("UpdateMoney - UPDATE query failed", zap.Error(err), zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency))
				return dto.Balance{}, err
			}

			b.log.Info("UpdateMoney - Successfully updated balance", zap.String("userID", updateReq.UserID.String()), zap.String("currency", updateReq.Currency), zap.String("newAmountUnits", amountUnits.String()), zap.Int64("newAmountCents", amountCents))
		case constant.BONUS_MONEY:
			// First update reserved_units, then recalculate reserved_cents from the final reserved_units value
			// Also update brand_id from user if it's different
			// This ensures accuracy and prevents drift between cents and units
			// Use FLOOR to match Go's IntPart() behavior (truncate to integer)
			var brandID *uuid.UUID
			err = b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, updateReq.UserID).Scan(&brandID)
			if err != nil && err != sql.ErrNoRows {
				b.log.Error("unable to get brand_id from user", zap.Error(err), zap.String("userID", updateReq.UserID.String()))
			}

			err = b.db.GetPool().QueryRow(ctx, `
				UPDATE balances 
				SET reserved_units = reserved_units + $1, 
				    reserved_cents = FLOOR((reserved_units + $1) * 100)::BIGINT,
				    brand_id = $2,
				    updated_at = $3 
				WHERE user_id = $4 AND currency_code = $5
				RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
			`, updateReq.Amount, brandID, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
				&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
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
			AmountCents:   amountCents,
			AmountUnits:   amountUnits,   // amount_units maps to real_money
			ReservedUnits: reservedUnits, // reserved_units maps to bonus_money
			ReservedCents: reservedCents, // Use actual reserved_cents from database
			UpdateAt:      updatedAt,
		}, nil
	}

	// Original code for local development...
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
		// Fetch brand_id from users table
		var brandID *uuid.UUID
		err = b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, updateReq.UserID).Scan(&brandID)
		if err != nil && err != sql.ErrNoRows {
			b.log.Warn("Failed to get brand_id from user, continuing without it", zap.Error(err), zap.String("userID", updateReq.UserID.String()))
		}

		// Use raw SQL to include brand_id since sqlc might not have it
		var insertedID uuid.UUID
		err = b.db.GetPool().QueryRow(ctx, `
			INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING id
		`, updateReq.UserID, updateReq.Currency, 0, decimal.Zero, 0, decimal.Zero, brandID, time.Now()).Scan(&insertedID)
		_ = insertedID // insertedID is used to avoid unused variable error
		if err != nil {
			b.log.Error("unable to create balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			return dto.Balance{}, err
		}
	}

	// Use manual SQL to ensure we INCREMENT both amount_cents and amount_units atomically
	// and avoid sqlc helpers that set absolute values.
	var id uuid.UUID
	var userID uuid.UUID
	var currencyCode string
	var amountCents int64
	var amountUnits decimal.NullDecimal
	var reservedCents int64
	var reservedUnits decimal.NullDecimal
	var updatedAt sql.NullTime

	switch updateReq.Component {
	case constant.REAL_MONEY:
		// First update amount_units, then recalculate amount_cents from the final amount_units value
		// This ensures accuracy and prevents drift between cents and units
		// Use FLOOR to match Go's IntPart() behavior (truncate to integer)
		// Fetch brand_id from user
		var brandID *uuid.UUID
		err = b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, updateReq.UserID).Scan(&brandID)
		if err != nil && err != sql.ErrNoRows {
			b.log.Error("unable to get brand_id from user", zap.Error(err), zap.String("userID", updateReq.UserID.String()))
		}

		err = b.db.GetPool().QueryRow(ctx, `
            UPDATE balances
            SET amount_units = amount_units + $1,
                amount_cents = FLOOR((amount_units + $1) * 100)::BIGINT,
                brand_id = $2,
                updated_at   = $3
            WHERE user_id = $4 AND currency_code = $5
            RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
        `, updateReq.Amount, brandID, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
			&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
		)
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ")
			return dto.Balance{}, err
		}
	case constant.BONUS_MONEY:
		// First update reserved_units, then recalculate reserved_cents from the final reserved_units value
		// Also update brand_id from user if it's different
		// This ensures accuracy and prevents drift between cents and units
		// Use FLOOR to match Go's IntPart() behavior (truncate to integer)
		var brandID *uuid.UUID
		err = b.db.GetPool().QueryRow(ctx, `SELECT brand_id FROM users WHERE id = $1`, updateReq.UserID).Scan(&brandID)
		if err != nil && err != sql.ErrNoRows {
			b.log.Error("unable to get brand_id from user", zap.Error(err), zap.String("userID", updateReq.UserID.String()))
		}

		err = b.db.GetPool().QueryRow(ctx, `
            UPDATE balances
            SET reserved_units = reserved_units + $1,
                reserved_cents = FLOOR((reserved_units + $1) * 100)::BIGINT,
                brand_id = $2,
                updated_at     = $3
            WHERE user_id = $4 AND currency_code = $5
            RETURNING id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, brand_id, updated_at
        `, updateReq.Amount, brandID, time.Now(), updateReq.UserID, updateReq.Currency).Scan(
			&id, &userID, &currencyCode, &amountCents, &amountUnits, &reservedCents, &reservedUnits, &brandID, &updatedAt,
		)
		if err != nil {
			b.log.Error("unable to update balance ", zap.Error(err), zap.Any("updateReq", updateReq))
			err = errors.ErrUnableToUpdate.Wrap(err, "unable to update balance ")
			return dto.Balance{}, err
		}
	}

	// Convert to DTO while handling potential NULLs
	var amountUnitsVal decimal.Decimal
	if amountUnits.Valid {
		amountUnitsVal = amountUnits.Decimal
	}
	var reservedUnitsVal decimal.Decimal
	if reservedUnits.Valid {
		reservedUnitsVal = reservedUnits.Decimal
	}
	var updatedAtVal time.Time
	if updatedAt.Valid {
		updatedAtVal = updatedAt.Time
	}

	return dto.Balance{
		ID:            id,
		UserId:        userID,
		CurrencyCode:  currencyCode,
		AmountCents:   amountCents,
		AmountUnits:   amountUnitsVal,
		ReservedCents: int64(reservedCents),
		ReservedUnits: reservedUnitsVal,
		UpdateAt:      updatedAtVal,
	}, nil
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
	// For now, let's use a simpler query to test
	simpleQuery := `
		SELECT 
			mf.id,
			mf.transaction_id,
			mf.type,
			mf.amount_cents,
			mf.reason,
			mf.currency_code,
			mf.note,
			mf.created_at,
			mf.user_id,
			mf.admin_id,
			COALESCE(NULLIF(TRIM(admin_user.first_name || ' ' || admin_user.last_name), ''), admin_user.username, 'Unknown Admin') as admin_name
		FROM manual_funds mf
		LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
		ORDER BY mf.created_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (filter.Page - 1) * filter.PerPage
	b.log.Info("Executing manual funds query", zap.String("query", simpleQuery), zap.Int("perPage", filter.PerPage), zap.Int("offset", offset))

	rows, err := b.db.GetPool().Query(ctx, simpleQuery, filter.PerPage, offset)
	if err != nil {
		b.log.Error("Failed to execute manual funds query", zap.Error(err))
		return dto.GetManualFundRes{}, err
	}
	defer rows.Close()

	var funds []dto.GetManualFundData
	rowCount := 0
	for rows.Next() {
		rowCount++
		var fund dto.ManualFundResData
		var amountCents int64
		var adminName string

		err := rows.Scan(
			&fund.ID,
			&fund.TransactionID,
			&fund.Type,
			&amountCents,
			&fund.Reason,
			&fund.Currency,
			&fund.Note,
			&fund.CreatedAt,
			&fund.UserID,
			&fund.AdminID,
			&adminName,
		)
		if err != nil {
			b.log.Error("Failed to scan manual fund row", zap.Error(err))
			continue
		}

		// Convert cents to decimal
		fund.Amount = decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
		fund.AdminName = adminName

		// Create dummy user objects for now (we can enhance this later)
		user := dto.User{
			ID: fund.UserID,
		}
		fundBy := dto.User{
			ID: fund.AdminID,
		}

		funds = append(funds, dto.GetManualFundData{
			ManualFund: fund,
			User:       user,
			FundBy:     fundBy,
		})
	}

	b.log.Info("Manual funds query completed", zap.Int("rowsProcessed", rowCount), zap.Int("fundsReturned", len(funds)))
	totalPages := (len(funds) + filter.PerPage - 1) / filter.PerPage

	return dto.GetManualFundRes{
		Message:   "Manual funds retrieved successfully",
		Data:      funds,
		TotalPage: totalPages,
	}, nil
}

func (b *balance) GetAllManualFunds(ctx context.Context, filter dto.GetAllManualFundsFilter) (dto.GetAllManualFundsRes, error) {
	// Build the query with filters
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base query with joins for usernames
	baseQuery := `
		SELECT 
			mf.id,
			mf.transaction_id,
			mf.type,
			mf.amount_cents,
			mf.reason,
			mf.currency_code,
			mf.note,
			mf.created_at,
			mf.user_id,
			mf.admin_id,
			COALESCE(
				NULLIF(TRIM(user_table.first_name || ' ' || user_table.last_name), ''),
				NULLIF(TRIM(user_table.username), ''),
				user_table.email,
				'User ' || SUBSTRING(user_table.id::text, 1, 8)
			) as username,
			COALESCE(NULLIF(TRIM(admin_user.first_name || ' ' || admin_user.last_name), ''), admin_user.username, 'Unknown Admin') as admin_name
		FROM manual_funds mf
		LEFT JOIN users user_table ON mf.user_id = user_table.id
		LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
	`

	// Add search filter
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(user_table.username ILIKE $%d OR user_table.email ILIKE $%d OR mf.transaction_id ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	// Add type filter
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("mf.type = $%d", argIndex))
		args = append(args, filter.Type)
		argIndex++
	}

	// Add currency filter
	if filter.Currency != "" {
		conditions = append(conditions, fmt.Sprintf("mf.currency_code = $%d", argIndex))
		args = append(args, filter.Currency)
		argIndex++
	}

	// Add date filters
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("mf.created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom+" 00:00:00")
		argIndex++
	}

	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("mf.created_at <= $%d", argIndex))
		args = append(args, filter.DateTo+" 23:59:59")
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query for total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM manual_funds mf LEFT JOIN users user_table ON mf.user_id = user_table.id LEFT JOIN users admin_user ON mf.admin_id = admin_user.id %s", whereClause)

	b.log.Info("Executing count query", zap.String("query", countQuery), zap.Any("args", args))

	var totalCount int64
	err := b.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		b.log.Error("Failed to execute count query", zap.Error(err))
		return dto.GetAllManualFundsRes{}, err
	}

	// Calculate pagination
	offset := (filter.Page - 1) * filter.PerPage
	totalPages := int((totalCount + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	// Main query with pagination
	mainQuery := fmt.Sprintf("%s %s ORDER BY mf.created_at DESC LIMIT $%d OFFSET $%d", baseQuery, whereClause, argIndex, argIndex+1)
	args = append(args, filter.PerPage, offset)

	b.log.Info("Executing main query", zap.String("query", mainQuery), zap.Any("args", args))

	rows, err := b.db.GetPool().Query(ctx, mainQuery, args...)
	if err != nil {
		b.log.Error("Failed to execute main query", zap.Error(err))
		return dto.GetAllManualFundsRes{}, err
	}
	defer rows.Close()

	var funds []dto.ManualFundResData
	var totalFundsUSD decimal.Decimal

	for rows.Next() {
		var fund dto.ManualFundResData
		var amountCents int64
		var username, adminName string

		err := rows.Scan(
			&fund.ID,
			&fund.TransactionID,
			&fund.Type,
			&amountCents,
			&fund.Reason,
			&fund.Currency,
			&fund.Note,
			&fund.CreatedAt,
			&fund.UserID,
			&fund.AdminID,
			&username,
			&adminName,
		)
		if err != nil {
			b.log.Error("Failed to scan manual fund row", zap.Error(err))
			continue
		}

		// Convert cents to decimal
		fund.Amount = decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))
		fund.Username = username
		fund.AdminName = adminName

		// Calculate total USD (only for USD transactions)
		if fund.Currency == "USD" {
			totalFundsUSD = totalFundsUSD.Add(fund.Amount)
		}

		funds = append(funds, fund)
	}

	b.log.Info("Manual funds query completed",
		zap.Int("fundsReturned", len(funds)),
		zap.Int64("totalCount", totalCount),
		zap.Int("totalPages", totalPages),
		zap.String("totalFundsUSD", totalFundsUSD.String()))

	return dto.GetAllManualFundsRes{
		Message:       "Manual funds retrieved successfully",
		Data:          funds,
		Total:         totalCount,
		TotalPages:    totalPages,
		CurrentPage:   filter.Page,
		PerPage:       filter.PerPage,
		TotalFundsUSD: totalFundsUSD.String(),
	}, nil
}

func (b *balance) GetManualFundsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.ManualFundResData, error) {
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database with admin name join
		rows, err := b.db.GetPool().Query(ctx, `
			SELECT mf.id, mf.user_id, mf.admin_id, mf.transaction_id, mf.type, mf.amount_cents, mf.currency_code, mf.reason, mf.note, mf.created_at,
			       COALESCE(admin_user.first_name || ' ' || admin_user.last_name, admin_user.username, 'Unknown Admin') as admin_name
			FROM manual_funds mf
			LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
			WHERE mf.user_id = $1 
			ORDER BY mf.created_at DESC
		`, userID)
		if err != nil {
			b.log.Error("unable to get manual funds by user_id", zap.Error(err), zap.String("userID", userID.String()))
			return []dto.ManualFundResData{}, err
		}
		defer rows.Close()

		var funds []dto.ManualFundResData
		for rows.Next() {
			var id uuid.UUID
			var userID uuid.UUID
			var adminID uuid.UUID
			var transactionID string
			var fundType string
			var amountCents int64
			var currencyCode string
			var reason string
			var note string
			var createdAt time.Time
			var adminName string

			err := rows.Scan(&id, &userID, &adminID, &transactionID, &fundType, &amountCents, &currencyCode, &reason, &note, &createdAt, &adminName)
			if err != nil {
				b.log.Error("unable to scan manual fund row", zap.Error(err))
				continue
			}

			// Convert cents back to units for response
			amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))

			funds = append(funds, dto.ManualFundResData{
				ID:            id,
				UserID:        userID,
				AdminID:       adminID,
				AdminName:     adminName,
				TransactionID: transactionID,
				Type:          fundType,
				Amount:        amount,
				Reason:        reason,
				Currency:      currencyCode,
				Note:          note,
				CreatedAt:     createdAt,
			})
		}

		return funds, nil
	}

	// Use original query for local development with admin name join
	query := `SELECT mf.id, mf.user_id, mf.admin_id, mf.transaction_id, mf.type, mf.amount_cents, mf.currency_code, mf.reason, mf.note, mf.created_at,
	         COALESCE(admin_user.first_name || ' ' || admin_user.last_name, admin_user.username, 'Unknown Admin') as admin_name
	         FROM manual_funds mf
	         LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
	         WHERE mf.user_id = $1 ORDER BY mf.created_at DESC`
	rows, err := b.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		return []dto.ManualFundResData{}, err
	}
	defer rows.Close()

	var funds []dto.ManualFundResData
	for rows.Next() {
		var id uuid.UUID
		var userID uuid.UUID
		var adminID uuid.UUID
		var transactionID string
		var fundType string
		var amountCents int64
		var currencyCode string
		var reason string
		var note string
		var createdAt time.Time
		var adminName string

		err := rows.Scan(&id, &userID, &adminID, &transactionID, &fundType, &amountCents, &currencyCode, &reason, &note, &createdAt, &adminName)
		if err != nil {
			b.log.Error("unable to scan manual fund row", zap.Error(err))
			continue
		}

		// Convert cents back to units for response
		amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))

		funds = append(funds, dto.ManualFundResData{
			ID:            id,
			UserID:        userID,
			AdminID:       adminID,
			AdminName:     adminName,
			TransactionID: transactionID,
			Type:          fundType,
			Amount:        amount,
			Reason:        reason,
			Currency:      currencyCode,
			Note:          note,
			CreatedAt:     createdAt,
		})
	}

	return funds, nil
}

func (b *balance) GetManualFundsByUserIDPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]dto.ManualFundResData, int64, error) {
	offset := (page - 1) * perPage

	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// First get total count
		var totalCount int64
		err := b.db.GetPool().QueryRow(ctx, `
			SELECT COUNT(*) FROM manual_funds WHERE user_id = $1
		`, userID).Scan(&totalCount)
		if err != nil {
			b.log.Error("unable to get manual funds count", zap.Error(err), zap.String("userID", userID.String()))
			return []dto.ManualFundResData{}, 0, err
		}

		// Use raw SQL for server database with admin name join and pagination
		rows, err := b.db.GetPool().Query(ctx, `
			SELECT mf.id, mf.user_id, mf.admin_id, mf.transaction_id, mf.type, mf.amount_cents, mf.currency_code, mf.reason, mf.note, mf.created_at,
			       COALESCE(admin_user.first_name || ' ' || admin_user.last_name, admin_user.username, 'Unknown Admin') as admin_name
			FROM manual_funds mf
			LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
			WHERE mf.user_id = $1 
			ORDER BY mf.created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, perPage, offset)
		if err != nil {
			b.log.Error("unable to get manual funds by user_id", zap.Error(err), zap.String("userID", userID.String()))
			return []dto.ManualFundResData{}, 0, err
		}
		defer rows.Close()

		var funds []dto.ManualFundResData
		for rows.Next() {
			var id uuid.UUID
			var userID uuid.UUID
			var adminID uuid.UUID
			var transactionID string
			var fundType string
			var amountCents int64
			var currencyCode string
			var reason string
			var note string
			var createdAt time.Time
			var adminName string

			err := rows.Scan(&id, &userID, &adminID, &transactionID, &fundType, &amountCents, &currencyCode, &reason, &note, &createdAt, &adminName)
			if err != nil {
				b.log.Error("unable to scan manual fund row", zap.Error(err))
				continue
			}

			// Convert cents back to units for response
			amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))

			funds = append(funds, dto.ManualFundResData{
				ID:            id,
				UserID:        userID,
				AdminID:       adminID,
				AdminName:     adminName,
				TransactionID: transactionID,
				Type:          fundType,
				Amount:        amount,
				Reason:        reason,
				Currency:      currencyCode,
				Note:          note,
				CreatedAt:     createdAt,
			})
		}

		return funds, totalCount, nil
	}

	// Use original query for local development with admin name join and pagination
	// First get total count
	var totalCount int64
	err := b.db.GetPool().QueryRow(ctx, `
		SELECT COUNT(*) FROM manual_funds WHERE user_id = $1
	`, userID).Scan(&totalCount)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		return []dto.ManualFundResData{}, 0, err
	}

	query := `SELECT mf.id, mf.user_id, mf.admin_id, mf.transaction_id, mf.type, mf.amount_cents, mf.currency_code, mf.reason, mf.note, mf.created_at,
	         COALESCE(admin_user.first_name || ' ' || admin_user.last_name, admin_user.username, 'Unknown Admin') as admin_name
	         FROM manual_funds mf
	         LEFT JOIN users admin_user ON mf.admin_id = admin_user.id
	         WHERE mf.user_id = $1 ORDER BY mf.created_at DESC
	         LIMIT $2 OFFSET $3`
	rows, err := b.db.GetPool().Query(ctx, query, userID, perPage, offset)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		return []dto.ManualFundResData{}, 0, err
	}
	defer rows.Close()

	var funds []dto.ManualFundResData
	for rows.Next() {
		var id uuid.UUID
		var userID uuid.UUID
		var adminID uuid.UUID
		var transactionID string
		var fundType string
		var amountCents int64
		var currencyCode string
		var reason string
		var note string
		var createdAt time.Time
		var adminName string

		err := rows.Scan(&id, &userID, &adminID, &transactionID, &fundType, &amountCents, &currencyCode, &reason, &note, &createdAt, &adminName)
		if err != nil {
			b.log.Error("unable to scan manual fund row", zap.Error(err))
			continue
		}

		// Convert cents back to units for response
		amount := decimal.NewFromInt(amountCents).Div(decimal.NewFromInt(100))

		funds = append(funds, dto.ManualFundResData{
			ID:            id,
			UserID:        userID,
			AdminID:       adminID,
			AdminName:     adminName,
			TransactionID: transactionID,
			Type:          fundType,
			Amount:        amount,
			Reason:        reason,
			Currency:      currencyCode,
			Note:          note,
			CreatedAt:     createdAt,
		})
	}

	return funds, totalCount, nil
}

// GetAdminFundingLimit retrieves the maximum funding limit for an admin from their roles
func (b *balance) GetAdminFundingLimit(ctx context.Context, adminID uuid.UUID) (*decimal.Decimal, error) {
	result, err := b.db.Queries.GetAdminFundingLimit(ctx, adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No funding limit set (unlimited)
			return nil, nil
		}
		b.log.Error("Failed to get admin funding limit", zap.Error(err), zap.String("adminID", adminID.String()))
		return nil, err
	}

	if !result.MaxFundingLimit.Valid {
		// No funding limit set (unlimited)
		return nil, nil
	}

	limit := result.MaxFundingLimit.Decimal
	return &limit, nil
}
