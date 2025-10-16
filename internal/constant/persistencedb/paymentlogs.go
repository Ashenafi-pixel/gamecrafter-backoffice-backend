package persistencedb

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/db"
	"go.uber.org/zap"
)

func (p *PersistenceDB) GetBalanceLogs(ctx context.Context, query string, perPage, offset int, conditions ...interface{}) (dto.GetBalanceResData, error) {
	var getBalanceRes []dto.BalanceLogsRes
	var total int
	getBalanceData := dto.GetBalanceResData{}

	rows, err := p.pool.Query(ctx, query, conditions...)
	if err != nil {
		return dto.GetBalanceResData{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var i dto.BalanceLogs
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Component,
			&i.Currency,
			&i.ChangeAmount,
			&i.OperationalGroupID,
			&i.OperationalTypeID,
			&i.OperationalTypeName,
			&i.Description,
			&i.Timestamp,
			&i.Type,
			&i.BalanceAfterUpdate,
			&i.TransactionID,
		); err != nil {
			return dto.GetBalanceResData{}, err
		}
		getBalanceRes = append(getBalanceRes, dto.BalanceLogsRes{
			ID:                 i.ID,
			UserID:             i.UserID,
			Component:          i.Component,
			Currency:           i.Currency,
			Description:        i.Description,
			ChangeAmount:       i.ChangeAmount,
			OperationalGroupID: i.OperationalGroupID,
			OperationalType: dto.OperationalType{
				ID:   i.OperationalTypeID,
				Name: i.OperationalTypeName,
			},
			Type:               i.Type,
			Timestamp:          i.Timestamp,
			BalanceAfterUpdate: i.BalanceAfterUpdate,
			TransactionID:      i.TransactionID,
		})
	}

	remove := fmt.Sprintf("offset %d limit %d", offset, perPage)
	totalQuery := strings.ReplaceAll(fmt.Sprintf("SELECT count(id) FROM (%s) AS subquery", query), remove, "")
	row := p.pool.QueryRow(ctx, totalQuery, conditions...)

	if err := row.Scan(&total); err != nil {
		return dto.GetBalanceResData{}, err
	}
	if len(getBalanceRes) == 0 {
		getBalanceRes = []dto.BalanceLogsRes{}
	}
	getBalanceData.Logs = getBalanceRes
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	getBalanceData.TotalPages = totalPages
	return getBalanceData, nil
}

var (
	Where                   = "where"
	BalanceLogsQuery        = `SELECT bl.id,bl.user_id,bl.component,bl.currency,bl.change_amount,bl.operational_group_id,bl.operational_type_id,ot.name AS operation_type_name,bl.description,bl.timestamp,ops.name as type,balance_after_update,transaction_id FROM balance_logs bl join operational_groups ops on ops.id = bl.operational_group_id join operational_types ot on ot.id = bl.operational_type_id %s %s %s`
	UserIDQuery             = "user_id = '%s'"
	OperationalGroupIDQuery = "operational_group_id = '%s'"
	OperationalTypeIDQuery  = "operational_type_id = '%s'"
	PaginationQuery         = " ORDER BY timestamp DESC offset %d limit %d"
	StartDate               = "timestamp >= '%s'"
	EndDate                 = "timestamp <= '%s'"
	ComponentQuery          = "component = '%s' "
)

var (
	AdminWhere            = "where"
	AdminBalanceLogsQuery = `SELECT 
    bl.id,
    bl.component, 
    bl.transaction_id, 
    bl.currency, 
    bl.change_amount, 
    bl.balance_after_update, 
    bl.timestamp, 
    ot.name AS operation_type, 
	us.first_name, 
    us.last_name, 
    us.email, 
    us.username, 
    us.phone_number,
	bl.description,
	op.name AS transaction_type,
	bl.status
FROM 
    balance_logs bl 
JOIN 
    users us ON bl.user_id = us.id 
JOIN 
    operational_groups op ON op.id = bl.operational_group_id 
JOIN 
    operational_types ot ON ot.id = bl.operational_type_id 
	%s offset %d limit %d`
	AdminUserIDQuery          = "user_id = '%s'"
	AdminTransactionTypeQuery = "og.name = '%s'"
	AdminPaginationQuery      = " ORDER BY timestamp DESC offset %d limit %d"
	AdminStartDateQuery       = "timestamp >= '%s'"
	AdminEndDateQuery         = "timestamp <= '%s'"
	AdminEndAmountQuery       = "change_amount <= '%s'"
	AdminStartAmountQuery     = "change_amount >= '%s'"
	AdminComponentQuery       = "component = '%s' "
	AdminPlayerUsernameQuery  = "us.username = '%s'"

	AdminOrderQuery         = "ORDER BY "
	AdminOrderAmountQuery   = "bl.change_amount"
	AdminOrderDateQuery     = "bl.timestamp"
	AdminOrderUsernameQuery = "us.username"
)

func (p *PersistenceDB) GetBalanceLogsForAdmin(ctx context.Context, query string, conditions []any, perPage, offset int) (dto.AdminGetBalanceLogsRes, error) {
	var getBalanceRes []dto.GetAdminBalanceResData
	var total int

	// Execute the query to fetch the balance logs
	rows, err := p.pool.Query(ctx, query, conditions...)
	if err != nil {
		return dto.AdminGetBalanceLogsRes{}, err
	}
	defer rows.Close()

	// Scan the results
	for rows.Next() {
		var i dto.DbAdminBalanceResData
		if err := rows.Scan(
			&i.ID,
			&i.Component,
			&i.TransactionID,
			&i.Currency,
			&i.ChangeAmount,
			&i.BalanceAfterUpdate,
			&i.Timestamp,
			&i.OperationType,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Username,
			&i.PhoneNumber,
			&i.Description,
			&i.TransactionType,
			&i.Status,
		); err != nil {
			return dto.AdminGetBalanceLogsRes{}, err
		}
		getBalanceRes = append(getBalanceRes, dto.GetAdminBalanceResData{
			ID:                 i.ID,
			Component:          i.Component,
			TransactionID:      i.TransactionID,
			Currency:           i.Currency,
			ChangeAmount:       i.ChangeAmount,
			BalanceAfterUpdate: i.BalanceAfterUpdate,
			Timestamp:          i.Timestamp,
			OperationType:      i.OperationType,
			FirstName:          dto.NullToString(i.FirstName),
			LastName:           dto.NullToString(i.LastName),
			Email:              dto.NullToString(i.Email),
			Username:           dto.NullToString(i.Username),
			PhoneNumber:        dto.NullToString(i.PhoneNumber),
			Description:        dto.NullToString(i.Description),
			TransactionType:    i.TransactionType,
			Status:             dto.NullToString(i.Status),
		})
	}

	// Remove pagination for count query and add alias to subquery
	remove := fmt.Sprintf("offset %d limit %d", offset, perPage)
	totalQuery := fmt.Sprintf("SELECT count(id) as total FROM (%s) AS subquery", strings.ReplaceAll(query, remove, ""))
	row := p.pool.QueryRow(ctx, totalQuery, conditions...)
	if err := row.Scan(&total); err != nil {
		return dto.AdminGetBalanceLogsRes{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if len(getBalanceRes) == 0 {
		getBalanceRes = []dto.GetAdminBalanceResData{}
	}

	return dto.AdminGetBalanceLogsRes{
		Message:    constant.SUCCESS,
		Data:       getBalanceRes,
		TotalPages: totalPages,
	}, nil
}

var (
	GetManualBalanceQuery = `SELECT 
	mf.id,
    mf.transaction_id,
    mf.type,
    mf.amount,
    mf.reason,
    mf.currency,
    mf.note,
    mf.created_at,
    
    -- Customer Information
    us.id AS customer_user_id,
    us.username AS customer_username,
    us.phone_number AS customer_phone_number,
    us.first_name AS customer_first_name,
    us.last_name AS customer_last_name,
    us.email AS customer_email,
    us.default_currency AS customer_default_currency,
    us.profile AS customer_profile_picture,
    us.date_of_birth AS customer_date_of_birth,
    
    -- Admin Information
    ad.id AS admin_user_id,
    ad.username AS admin_username,
    ad.phone_number AS admin_phone_number,
    ad.first_name AS admin_first_name,
    ad.last_name AS admin_last_name,
    ad.email AS admin_email,
    ad.profile AS admin_profile_picture,
    ad.date_of_birth AS admin_date_of_birth

FROM 
    manual_funds mf 
JOIN 
    users us ON us.id = mf.user_id
JOIN 
    users ad ON ad.id = mf.admin_id %s offset %d limit %d
`

	ManualUserIDQuery         = "mf.user_id = '%s'"
	ManualAdminIDQuery        = "mf.admin_id = '%s'"
	ManualPaginationQuery     = "offset %d limit %d"
	ManualStartDateQuery      = "timestamp >= '%s'"
	ManualEndDateQuery        = "timestamp <= '%s'"
	ManualEndAmountQuery      = "mf.amount <= '%s'"
	ManualStartAmountQuery    = "mf.amount >= '%s'"
	ManualTypeQuery           = "mf.type = '%s' "
	ManualPlayerUsernameQuery = "us.username = '%s'"

	ManualOrderQuery         = "ORDER BY "
	ManualOrderAmountQuery   = "mf.amount"
	ManualOrderDateQuery     = "mf.timestamp"
	ManualOrderUsernameQuery = "us.username"
)

func (p *PersistenceDB) GetManualFunds(ctx context.Context, query string, conditions []interface{}, limit, offset int) (dto.GetManualFundRes, error) {
	var resp []dto.GetManualDBRes
	var total int
	var data []dto.GetManualFundData
	rows, err := p.pool.Query(ctx, query, conditions...)
	if err != nil {
		return dto.GetManualFundRes{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var i dto.GetManualDBRes
		if err := rows.Scan(
			&i.ID,
			&i.TransactionID,
			&i.Type,
			&i.Amount,
			&i.Reason,
			&i.Currency,
			&i.Note,
			&i.CreatedAt,
			&i.CustomerUserID,
			&i.CustomerUserName,
			&i.CustomerPhoneNumber,
			&i.CustomerFirstName,
			&i.CustomerLastName,
			&i.CustomerEmail,
			&i.CustomerDefaultCurrency,
			&i.CustomerProfilePicture,
			&i.CustomerDateOfBirth,
			&i.AdminUserID,
			&i.AdminUserName,
			&i.AdminPhoneNumber,
			&i.AdminFirstName,
			&i.AdminLastName,
			&i.AdminEmail,
			&i.AdminProfilePicture,
			&i.AdminDateOfBirth,
		); err != nil {
			return dto.GetManualFundRes{}, err
		}
		resp = append(resp, i)

	}
	for _, d := range resp {
		data = append(data, dto.GetManualFundData{
			ManualFund: dto.ManualFundResData{
				ID:            d.ID,
				UserID:        d.CustomerUserID,
				AdminID:       d.AdminUserID,
				TransactionID: d.TransactionID,
				Amount:        d.Amount,
				Reason:        d.Reason,
				Currency:      d.Currency,
				Note:          d.Note,
				CreatedAt:     d.CreatedAt,
			},
			User: dto.User{
				ID:             d.CustomerUserID,
				PhoneNumber:    d.CustomerPhoneNumber,
				FirstName:      d.CustomerFirstName,
				LastName:       d.CustomerLastName,
				Email:          d.CustomerEmail,
				ProfilePicture: d.CustomerProfilePicture,
				DateOfBirth:    d.CustomerDateOfBirth,
			},
			FundBy: dto.User{
				ID:             d.AdminUserID,
				FirstName:      d.AdminFirstName,
				LastName:       d.AdminLastName,
				PhoneNumber:    d.AdminPhoneNumber,
				ProfilePicture: d.AdminProfilePicture,
				DateOfBirth:    d.AdminDateOfBirth,
			},
		})
	}
	remove := fmt.Sprintf("offset %d limit %d", offset, limit)
	totalQuery := fmt.Sprintf("SELECT count(id) as total FROM  (%s) AS subquery", strings.ReplaceAll(query, remove, ""))
	row := p.pool.QueryRow(ctx, totalQuery, conditions...)
	if err := row.Scan(&total); err != nil {
		return dto.GetManualFundRes{}, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return dto.GetManualFundRes{
		Message:   constant.SUCCESS,
		Data:      data,
		TotalPage: totalPages,
	}, nil
}

func (p *PersistenceDB) RemoveFromCasbinRule(ctx context.Context, roleID uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM casbin_rule where v0 = $1", roleID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PersistenceDB) UpdateBalance(ctx context.Context, params db.UpdateBalanceParams) (db.Balance, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		p.log.Error("Failed to begin transaction", zap.Error(err))
		return db.Balance{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			p.log.Warn("Failed to rollback transaction", zap.Error(err))
		}
	}()

	q := db.New(tx)

	// Lock the row
	_, err = q.LockBalance(ctx, db.LockBalanceParams{
		UserID:       params.UserID,
		CurrencyCode: params.CurrencyCode,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			p.log.Warn("No balance found for user and currency", zap.Any("user_id", params.UserID), zap.String("currency", params.CurrencyCode))
			return db.Balance{}, err
		}
		p.log.Error("Failed to lock balance", zap.Error(err))
		return db.Balance{}, err
	}

	// Perform the update
	balance, err := q.UpdateBalance(ctx, db.UpdateBalanceParams{
		CurrencyCode: params.CurrencyCode,
		AmountUnits:    params.AmountUnits,
		ReservedUnits:   params.ReservedUnits,
		ReservedCents:       params.ReservedCents,
		UpdatedAt:    params.UpdatedAt,
		UserID:       params.UserID,
	})
	if err != nil {
		p.log.Error("Failed to update balance", zap.Error(err))
		return db.Balance{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("Failed to commit transaction", zap.Error(err))
		return db.Balance{}, err
	}

	p.log.Info("Successfully updated balance", zap.Any("user_id", params.UserID), zap.String("currency", params.CurrencyCode))
	return balance, nil
}

func (p *PersistenceDB) UpdateMoney(ctx context.Context, params db.UpdateAmountUnitsParams) (db.Balance, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		p.log.Error("Failed to begin transaction", zap.Error(err))
		return db.Balance{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			p.log.Warn("Failed to rollback transaction", zap.Error(err))
		}
	}()

	q := db.New(tx)

	// Lock the row
	_, err = q.LockBalance(ctx, db.LockBalanceParams{
		UserID:       params.UserID,
		CurrencyCode: params.CurrencyCode,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			p.log.Warn("No balance found for user and currency", zap.Any("user_id", params.UserID), zap.String("currency", params.CurrencyCode))
			return db.Balance{}, err
		}
		p.log.Error("Failed to lock balance", zap.Error(err))
		return db.Balance{}, err
	}

	// Perform the update
	balance, err := q.UpdateAmountUnits(ctx, params)
	if err != nil {
		p.log.Error("Failed to update balance", zap.Error(err))
		return db.Balance{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("Failed to commit transaction", zap.Error(err))
		return db.Balance{}, err
	}

	p.log.Info("Successfully updated balance", zap.Any("user_id", params.UserID), zap.String("currency", params.CurrencyCode))
	return balance, nil
}

func (p *PersistenceDB) GetBalanceLog(ctx context.Context, balanceLogID uuid.UUID) (db.BalanceLog, error) {
	query := `
		SELECT
			bl.id,
			bl.user_id,
			bl.component,
			bl.currency,
			bl.description,
			bl.change_amount,
			bl.operational_group_id,
			ops.name AS type,
			bl.operational_type_id,
			ot.name AS operational_type_name,
			bl.timestamp,
			bl.balance_after_update,
			bl.transaction_id,
			bl.status
		FROM
			balance_logs bl
		JOIN
			operational_groups ops ON ops.id = bl.operational_group_id
		JOIN
			operational_types ot ON ot.id = bl.operational_type_id
		WHERE
			bl.id = $1`

	var balanceLog db.BalanceLog
	err := p.pool.QueryRow(ctx, query, balanceLogID).Scan(
		&balanceLog.ID,
		&balanceLog.UserID,
		&balanceLog.Component,
		&balanceLog.Currency,
		&balanceLog.Description,
		&balanceLog.ChangeAmount,
		&balanceLog.OperationalGroupID,
		&balanceLog.Type,
		&balanceLog.OperationalTypeID,
		&balanceLog.OperationalTypeName,
		&balanceLog.Timestamp,
		&balanceLog.BalanceAfterUpdate,
		&balanceLog.TransactionID,
		&balanceLog.Status,
	)
	if err != nil {
		return db.BalanceLog{}, err
	}

	return balanceLog, nil
}

func (p *PersistenceDB) GetBalanceLogByTransactionID(ctx context.Context, transactionID string) (db.BalanceLog, error) {
	query := `
		SELECT
			bl.id,
			bl.user_id,
			bl.component,
			bl.currency,
			bl.description,
			bl.change_amount,
			bl.operational_group_id,
			ops.name AS type,
			bl.operational_type_id,
			ot.name AS operational_type_name,
			bl.timestamp,
			bl.balance_after_update,
			bl.transaction_id,
			bl.status
		FROM
			balance_logs bl
		JOIN
			operational_groups ops ON ops.id = bl.operational_group_id
		JOIN
			operational_types ot ON ot.id = bl.operational_type_id
		WHERE
			bl.transaction_id = $1`

	var balanceLog db.BalanceLog
	err := p.pool.QueryRow(ctx, query, transactionID).Scan(
		&balanceLog.ID,
		&balanceLog.UserID,
		&balanceLog.Component,
		&balanceLog.Currency,
		&balanceLog.Description,
		&balanceLog.ChangeAmount,
		&balanceLog.OperationalGroupID,
		&balanceLog.Type,
		&balanceLog.OperationalTypeID,
		&balanceLog.OperationalTypeName,
		&balanceLog.Timestamp,
		&balanceLog.BalanceAfterUpdate,
		&balanceLog.TransactionID,
		&balanceLog.Status,
	)
	if err != nil {
		return db.BalanceLog{}, err
	}

	return balanceLog, nil
}

func (p *PersistenceDB) SaveBalanceLogs(ctx context.Context, arg db.SaveBalanceLogsParams) (db.BalanceLog, error) {
	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (with correct column names)
		query := `
			INSERT INTO balance_logs (
				user_id,
				component,
				change_cents,
				change_units,
				operational_group_id,
				operational_type_id,
				description,
				timestamp,
				balance_after_cents,
				balance_after_units,
				transaction_id,
				status,
				currency_code
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
			) RETURNING id, user_id, component, change_cents, change_units, operational_group_id, operational_type_id, description, timestamp, balance_after_cents, balance_after_units, transaction_id, status, currency_code`

		// Convert amounts to cents for server database
		changeCents := arg.ChangeAmount.Decimal.Mul(decimal.NewFromInt(100)).IntPart()
		balanceAfterCents := arg.BalanceAfterUpdate.Mul(decimal.NewFromInt(100)).IntPart()

		row := p.pool.QueryRow(ctx, query,
			arg.UserID,
			arg.Component,
			changeCents,
			arg.ChangeAmount.Decimal,
			arg.OperationalGroupID,
			arg.OperationalTypeID,
			arg.Description,
			arg.Timestamp,
			balanceAfterCents,
			arg.BalanceAfterUpdate,
			arg.TransactionID,
			arg.Status,
			arg.Currency.String,
		)

		var balanceLog db.BalanceLog
		var changeCentsResult int64
		var changeUnitsResult decimal.Decimal
		var balanceAfterCentsResult int64
		var balanceAfterUnitsResult decimal.Decimal
		var currencyCodeResult string

		err := row.Scan(
			&balanceLog.ID,
			&balanceLog.UserID,
			&balanceLog.Component,
			&changeCentsResult,
			&changeUnitsResult,
			&balanceLog.OperationalGroupID,
			&balanceLog.OperationalTypeID,
			&balanceLog.Description,
			&balanceLog.Timestamp,
			&balanceAfterCentsResult,
			&balanceAfterUnitsResult,
			&balanceLog.TransactionID,
			&balanceLog.Status,
			&currencyCodeResult,
		)
		if err != nil {
			return balanceLog, err
		}

		// Convert back to original format for compatibility
		balanceLog.Currency = sql.NullString{String: currencyCodeResult, Valid: true}
		balanceLog.ChangeAmount = decimal.NullDecimal{Decimal: changeUnitsResult, Valid: true}
		balanceLog.BalanceAfterUpdate = balanceAfterUnitsResult

		return balanceLog, nil
	}

	// Use original query for local development
	query := `
		INSERT INTO balance_logs (
			user_id,
			component,
			currency,
			change_amount,
			operational_group_id,
			operational_type_id,
			description,
			timestamp,
			balance_after_update,
			transaction_id,
			status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, user_id, component, currency, change_amount, operational_group_id, operational_type_id, description, timestamp, balance_after_update, transaction_id, status`

	row := p.pool.QueryRow(ctx, query,
		arg.UserID,
		arg.Component,
		arg.Currency,
		arg.ChangeAmount,
		arg.OperationalGroupID,
		arg.OperationalTypeID,
		arg.Description,
		arg.Timestamp,
		arg.BalanceAfterUpdate,
		arg.TransactionID,
		arg.Status,
	)

	var balanceLog db.BalanceLog
	err := row.Scan(
		&balanceLog.ID,
		&balanceLog.UserID,
		&balanceLog.Component,
		&balanceLog.Currency,
		&balanceLog.ChangeAmount,
		&balanceLog.OperationalGroupID,
		&balanceLog.OperationalTypeID,
		&balanceLog.Description,
		&balanceLog.Timestamp,
		&balanceLog.BalanceAfterUpdate,
		&balanceLog.TransactionID,
		&balanceLog.Status,
	)

	return balanceLog, err
}
