package balancelogs

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
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

type balance_logs struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.BalanceLogs {
	return &balance_logs{
		db:  db,
		log: log,
	}
}

func (b *balance_logs) SaveBalanceLogs(ctx context.Context, blanceLogReq dto.BalanceLogs) (dto.BalanceLogs, error) {
	balanceStatus := utils.NullString(blanceLogReq.Status)

	// Check if we're using server database (different schema)
	if os.Getenv("SKIP_PERMISSION_INIT") == "true" {
		// Use raw SQL for server database (without currency column)
		var id uuid.UUID
		var userID uuid.UUID
		var component string
		var changeCents int64
		var changeAmount decimal.Decimal
		var operationalGroupID uuid.UUID
		var operationalTypeID uuid.UUID
		var description string
		var timestamp time.Time
		var balanceAfterCents int64
		var balanceAfterUpdate decimal.Decimal
		var transactionID string
		var status string
		var currencyCode string

		// Convert amount to cents and units for server database
		changeCents = blanceLogReq.ChangeAmount.Mul(decimal.NewFromInt(100)).IntPart()
		balanceAfterCents = blanceLogReq.BalanceAfterUpdate.Mul(decimal.NewFromInt(100)).IntPart()

		// Handle transaction ID safely
		var transactionIDValue string
		if blanceLogReq.TransactionID != nil {
			transactionIDValue = *blanceLogReq.TransactionID
		}

		err := b.db.GetPool().QueryRow(ctx, `
			INSERT INTO balance_logs (user_id, component, change_cents, change_units, operational_group_id, operational_type_id, description, timestamp, balance_after_cents, balance_after_units, transaction_id, status, currency_code) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
			RETURNING id, user_id, component, change_cents, change_units, operational_group_id, operational_type_id, description, timestamp, balance_after_cents, balance_after_units, transaction_id, status, currency_code
		`, blanceLogReq.UserID, blanceLogReq.Component, changeCents, blanceLogReq.ChangeAmount, blanceLogReq.OperationalGroupID, blanceLogReq.OperationalTypeID, blanceLogReq.Description, time.Now(), balanceAfterCents, *blanceLogReq.BalanceAfterUpdate, transactionIDValue, balanceStatus.String, constant.DEFAULT_CURRENCY).Scan(
			&id, &userID, &component, &changeCents, &changeAmount, &operationalGroupID, &operationalTypeID, &description, &timestamp, &balanceAfterCents, &balanceAfterUpdate, &transactionID, &status, &currencyCode,
		)
		if err != nil {
			b.log.Error(fmt.Sprintf("unable to save balance logs error %s ", err.Error()), zap.Any("blanceLogReq", blanceLogReq))
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.BalanceLogs{}, err
		}
		return dto.BalanceLogs{
			ID:                 id,
			UserID:             userID,
			Component:          component,
			Currency:           constant.DEFAULT_CURRENCY, // Use default currency for server database
			Description:        description,
			ChangeAmount:       changeAmount,
			OperationalGroupID: operationalGroupID,
			OperationalTypeID:  operationalTypeID,
			Status:             status,
			BalanceAfterUpdate: blanceLogReq.BalanceAfterUpdate,
			TransactionID:      &transactionID,
		}, nil
	}

	// Use original query without currency column
	blanceRes, err := b.db.SaveBalanceLogs(ctx, db.SaveBalanceLogsParams{
		UserID:             uuid.NullUUID{UUID: blanceLogReq.UserID, Valid: true},
		Component:          db.Components(blanceLogReq.Component),
		ChangeAmount:       decimal.NullDecimal{Decimal: blanceLogReq.ChangeAmount, Valid: true},
		OperationalGroupID: uuid.NullUUID{UUID: blanceLogReq.OperationalGroupID, Valid: true},
		OperationalTypeID:  uuid.NullUUID{UUID: blanceLogReq.OperationalTypeID, Valid: true},
		Description:        sql.NullString{String: blanceLogReq.Description, Valid: true},
		Timestamp:          sql.NullTime{Time: time.Now(), Valid: true},
		BalanceAfterUpdate: *blanceLogReq.BalanceAfterUpdate,
		TransactionID:      *blanceLogReq.TransactionID,
		Status:             balanceStatus,
	})
	if err != nil {
		b.log.Error(fmt.Sprintf("unable to save balance logs error %s ", err.Error()), zap.Any("blanceLogReq", blanceLogReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.BalanceLogs{}, err
	}
	return dto.BalanceLogs{
		ID:                 blanceRes.ID,
		UserID:             blanceRes.UserID.UUID,
		Component:          string(blanceRes.Component),
		Currency:           blanceRes.Currency.String,
		Description:        blanceRes.Description.String,
		ChangeAmount:       blanceRes.ChangeAmount.Decimal,
		OperationalGroupID: blanceRes.OperationalGroupID.UUID,
		OperationalTypeID:  blanceRes.OperationalTypeID.UUID,
		Status:             blanceRes.Status.String,
		BalanceAfterUpdate: blanceLogReq.BalanceAfterUpdate,
		TransactionID:      blanceLogReq.TransactionID,
	}, nil
}

func (b *balance_logs) GetBalanceLog(ctx context.Context, req dto.GetBalanceLogReq) (dto.GetBalanceLogRes, error) {
	var balanceQuery string
	var conditions []any
	placeholderIndex := 1
	first := true

	if req.UserID != uuid.Nil {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("user_id = $%d", placeholderIndex)
		conditions = append(conditions, req.UserID)
		placeholderIndex++
		first = false
	}

	if req.Component != "" {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("component = $%d", placeholderIndex)
		conditions = append(conditions, req.Component)
		placeholderIndex++
		first = false
	}

	if req.OperationalGroupID != uuid.Nil {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("operational_group_id = $%d", placeholderIndex)
		conditions = append(conditions, req.OperationalGroupID)
		placeholderIndex++
		first = false
	}

	if req.OperationTypeID != uuid.Nil {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("operational_type_id = $%d", placeholderIndex)
		conditions = append(conditions, req.OperationTypeID)
		placeholderIndex++
		first = false
	}

	if req.StartAmount != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("bl.change_amount >= $%d", placeholderIndex)
		conditions = append(conditions, *req.StartAmount)
		placeholderIndex++
		first = false
	}

	if req.EndAmount != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("bl.change_amount <= $%d", placeholderIndex)
		conditions = append(conditions, *req.EndAmount)
		placeholderIndex++
		first = false
	}

	if req.StartDate != nil {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("timestamp >= $%d", placeholderIndex)
		conditions = append(conditions, req.StartDate.Format("2006-01-02"))
		placeholderIndex++
		first = false
	}

	if req.EndDate != nil {
		if !first {
			balanceQuery += " AND "
		}
		balanceQuery += fmt.Sprintf("timestamp <= $%d", placeholderIndex)
		conditions = append(conditions, req.EndDate.Format("2006-01-02"))
		placeholderIndex++
		first = false
	}

	balanceQuery = fmt.Sprintf(persistencedb.BalanceLogsQuery, persistencedb.Where, balanceQuery, fmt.Sprintf(persistencedb.PaginationQuery, req.Offset, req.PerPage))

	balanceLogs, err := b.db.GetBalanceLogs(ctx, balanceQuery, int(req.PerPage), int(req.Offset), conditions...)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetBalanceLogRes{}, err
	}

	return dto.GetBalanceLogRes{
		Status: constant.SUCCESS,
		Data:   balanceLogs,
	}, nil
}

func (b *balance_logs) GetBalanceLogByID(ctx context.Context, balanceLogID uuid.UUID) (dto.BalanceLogsRes, error) {
	if balanceLogID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid balance logs UUID")
		return dto.BalanceLogsRes{}, err
	}

	balanceLog, err := b.db.GetBalanceLog(ctx, balanceLogID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("balanceLogID", balanceLogID.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.BalanceLogsRes{}, err
	}
	return dto.BalanceLogsRes{
		ID:                 balanceLog.ID,
		UserID:             balanceLog.UserID.UUID,
		Component:          string(balanceLog.Component),
		Currency:           balanceLog.Currency.String,
		Description:        balanceLog.Description.String,
		ChangeAmount:       balanceLog.ChangeAmount.Decimal,
		OperationalGroupID: balanceLog.OperationalGroupID.UUID,
		OperationalType: dto.OperationalType{
			ID:   dto.NullToUUID(balanceLog.OperationalTypeID),
			Name: dto.NullToString(balanceLog.OperationalTypeName),
		},
		Timestamp:          dto.NullToTime(balanceLog.Timestamp),
		Type:               dto.NullToString(balanceLog.Type),
		BalanceAfterUpdate: &balanceLog.BalanceAfterUpdate,
		TransactionID:      &balanceLog.TransactionID,
		Status:             balanceLog.Status.String,
	}, nil
}

func (b *balance_logs) DeleteBalanceLog(ctx context.Context, balanceLogID uuid.UUID) error {
	err := b.db.DeleteBalanceLog(ctx, balanceLogID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("error", "unable to delete balance log"), zap.Any("balanceLogID", balanceLogID.String()))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *balance_logs) GetBalanceLogsForAdmin(ctx context.Context, req dto.AdminGetBalanceLogsReq) (dto.AdminGetBalanceLogsRes, error) {
	var balanceQuery string
	var conditions []any
	placeholderIndex := 1
	first := true
	orderClauses := []string{}

	if req.Filter.PlayerUsername != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("us.username = $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.PlayerUsername)
		placeholderIndex++
		first = false
	}

	if req.Filter.TransactionType != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("og.name = $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.TransactionType)
		placeholderIndex++
		first = false
	}

	if req.Filter.StartAmount != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("bl.change_amount >= $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.StartAmount)
		placeholderIndex++
		first = false
	}
	if req.Filter.Status != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("bl.status = $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.Status)
		placeholderIndex++
		first = false
	}
	if req.Filter.EndAmount != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("bl.change_amount <= $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.EndAmount)
		placeholderIndex++
		first = false
	}

	if req.Filter.StartDate != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("timestamp >= $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.StartDate)
		placeholderIndex++
		first = false
	}

	if req.Filter.EndDate != nil {
		if !first {
			balanceQuery += " AND "
		} else {
			balanceQuery = balanceQuery + " " + persistencedb.Where + " "
		}
		balanceQuery += fmt.Sprintf("timestamp <= $%d", placeholderIndex)
		conditions = append(conditions, *req.Filter.EndDate)
		placeholderIndex++
		first = false
	}

	if req.Sort.Amount != "" {
		order := strings.ToUpper(req.Sort.Amount)
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		orderClauses = append(orderClauses, fmt.Sprintf("bl.change_amount %s", order))
	}

	if req.Sort.Date != "" {
		order := strings.ToUpper(req.Sort.Date)
		if order != "ASC" && order != "DESC" {
			order = "DESC"
		}
		orderClauses = append(orderClauses, fmt.Sprintf("bl.timestamp %s", order))
	}

	if len(orderClauses) > 0 {
		balanceQuery += " ORDER BY " + strings.Join(orderClauses, ", ")
	}

	balanceQuery = fmt.Sprintf(persistencedb.AdminBalanceLogsQuery, balanceQuery, req.Page, req.PerPage)

	res, err := b.db.GetBalanceLogsForAdmin(ctx, balanceQuery, conditions, req.PerPage, req.Page)

	if err != nil {
		b.log.Error(err.Error(), zap.Any("balance_logs request", res))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.AdminGetBalanceLogsRes{}, err
	}

	return res, nil
}

func (b *balance_logs) GetBalanceLogByTransactionID(ctx context.Context, transactionID string) (dto.BalanceLogsRes, error) {
	res, err := b.db.GetBalanceLogByTransactionID(ctx, transactionID)
	if err != nil {
		return dto.BalanceLogsRes{}, nil
	}
	return dto.BalanceLogsRes{
		ID:                 res.ID,
		UserID:             res.UserID.UUID,
		Component:          string(res.Component),
		Currency:           res.Currency.String,
		Description:        res.Description.String,
		ChangeAmount:       res.ChangeAmount.Decimal,
		OperationalGroupID: res.OperationalGroupID.UUID,
		OperationalType: dto.OperationalType{
			ID:   dto.NullToUUID(res.OperationalTypeID),
			Name: dto.NullToString(res.OperationalTypeName),
		},
		Timestamp:          dto.NullToTime(res.Timestamp),
		Type:               dto.NullToString(res.Type),
		BalanceAfterUpdate: &res.BalanceAfterUpdate,
		TransactionID:      &res.TransactionID,
		Status:             res.Status.String,
	}, nil
}
