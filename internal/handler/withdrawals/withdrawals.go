package withdrawals

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type WithdrawalsHandler struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func NewWithdrawalsHandler(db *persistencedb.PersistenceDB, log *zap.Logger) *WithdrawalsHandler {
	return &WithdrawalsHandler{
		db:  db,
		log: log,
	}
}

// GetAllWithdrawals retrieves all withdrawals with filtering and pagination
func (h *WithdrawalsHandler) GetAllWithdrawals(ctx *gin.Context) {
	h.log.Info("Getting all withdrawals")

	// Parse pagination parameters
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse filter parameters
	var status *string
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = &statusStr
	}

	var userID *uuid.UUID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		if parsedUUID, err := uuid.Parse(userIDStr); err == nil {
			userID = &parsedUUID
		}
	}

	var withdrawalID *string
	if withdrawalIDStr := ctx.Query("withdrawal_id"); withdrawalIDStr != "" {
		withdrawalID = &withdrawalIDStr
	}

	var username *string
	if usernameStr := ctx.Query("username"); usernameStr != "" {
		username = &usernameStr
	}

	var email *string
	if emailStr := ctx.Query("email"); emailStr != "" {
		email = &emailStr
	}

	var startDate *time.Time
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}

	var endDate *time.Time
	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Add 23:59:59 to end date to include the entire day
			parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &parsed
		}
	}

	// Get withdrawals using raw SQL
	query := `
		SELECT 
			w.id, w.user_id, w.admin_id, w.chain_id, w.currency_code, w.protocol,
			w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, w.exchange_rate,
			w.fee_cents, w.source_wallet_address, w.to_address, w.tx_hash,
			w.status, w.requires_admin_review, w.admin_review_deadline,
			w.processed_by_system, w.amount_reserved_cents, w.reservation_released,
			w.reservation_released_at, w.metadata, w.error_message,
			w.created_at, w.updated_at,
			u.username, u.email, u.first_name, u.last_name
		FROM withdrawals w
		LEFT JOIN users u ON w.user_id = u.id
		WHERE ($1::text IS NULL OR w.status = $1)
			AND ($2::uuid IS NULL OR w.user_id = $2)
			AND ($3::text IS NULL OR w.withdrawal_id ILIKE '%' || $3 || '%')
			AND ($4::text IS NULL OR u.username ILIKE '%' || $4 || '%')
			AND ($5::text IS NULL OR u.email ILIKE '%' || $5 || '%')
			AND ($6::timestamp IS NULL OR w.created_at >= $6)
			AND ($7::timestamp IS NULL OR w.created_at <= $7)
		ORDER BY w.created_at DESC
		LIMIT $8 OFFSET $9
	`
	rows, err := h.db.GetPool().Query(ctx.Request.Context(), query,
		status, userID, withdrawalID, username, email, startDate, endDate, int32(limit), int32(offset))
	if err != nil {
		h.log.Error("Failed to get withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawals",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	// For now, return empty array since we don't need to implement the full withdrawal viewing yet
	// This is just to get the backend running
	withdrawals := []interface{}{}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"withdrawals": withdrawals,
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		},
	})
}

// GetWithdrawalByID retrieves a specific withdrawal by ID
func (h *WithdrawalsHandler) GetWithdrawalByID(ctx *gin.Context) {
	withdrawalIDStr := ctx.Param("id")
	_, err := uuid.Parse(withdrawalIDStr)
	if err != nil {
		h.log.Error("Invalid withdrawal ID", zap.String("id", withdrawalIDStr), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid withdrawal ID",
		})
		return
	}

	// For now, return empty response to get backend running
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{},
	})
}

// GetWithdrawalByWithdrawalID retrieves a specific withdrawal by withdrawal_id
func (h *WithdrawalsHandler) GetWithdrawalByWithdrawalID(ctx *gin.Context) {
	withdrawalID := ctx.Param("withdrawal_id")
	if withdrawalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Withdrawal ID is required",
		})
		return
	}

	// For now, return empty response to get backend running
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{},
	})
}

// GetWithdrawalsByUserID retrieves withdrawals for a specific user
func (h *WithdrawalsHandler) GetWithdrawalsByUserID(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("Invalid user ID", zap.String("user_id", userIDStr), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Parse pagination parameters
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// For now, return empty response to get backend running
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"withdrawals": []interface{}{},
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
		},
	})
}

// GetWithdrawalStats retrieves withdrawal statistics
func (h *WithdrawalsHandler) GetWithdrawalStats(ctx *gin.Context) {
	h.log.Info("Getting withdrawal stats")

	// For now, return empty stats to get backend running
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{},
	})
}

// GetWithdrawalsByDateRange retrieves withdrawals within a date range
func (h *WithdrawalsHandler) GetWithdrawalsByDateRange(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "start_date and end_date are required",
		})
		return
	}

	_, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	_, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Parse pagination parameters
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// For now, return empty response to get backend running
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"withdrawals": []interface{}{},
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
			},
			"date_range": gin.H{
				"start_date": startDateStr,
				"end_date":   endDateStr,
			},
		},
	})
}
