package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/contracts"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type UserWS interface {
	AddToPlayerBalanceWS(ctx context.Context, userID uuid.UUID, conn *websocket.Conn)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (dto.UserBalanceResp, error)
	TriggerBalanceWS(ctx context.Context, userID uuid.UUID)
	TriggerCashbackWS(ctx context.Context, userID uuid.UUID, cashbackData dto.EnhancedUserCashbackSummary)
	TriggerWinnerNotificationWS(ctx context.Context, userID uuid.UUID, winnerData dto.WinnerNotificationData)
	StoreWinnerNotification(ctx context.Context, winnerData dto.WinnerNotificationData) error
	GetWinnerNotifications(ctx context.Context, userID uuid.UUID) ([]dto.WinnerNotificationData, error)
}

type User struct {
	log                     *zap.Logger
	balanceStorage          storage.Balance
	redisClient             contracts.Redis
	UserBalanceSocket       map[uuid.UUID]map[*websocket.Conn]bool
	UserBalanceSocketLocker map[*websocket.Conn]*sync.Mutex
	mutex                   sync.Mutex
}

// CashbackWebSocketMessage represents the WebSocket message for cashback updates
type CashbackWebSocketMessage struct {
	Type    string                          `json:"type"`
	UserID  uuid.UUID                       `json:"user_id"`
	Data    dto.EnhancedUserCashbackSummary `json:"data"`
	Message string                          `json:"message,omitempty"`
}

// WinnerNotificationWebSocketMessage represents the WebSocket message for winner notifications
type WinnerNotificationWebSocketMessage struct {
	Type    string                     `json:"type"`
	UserID  uuid.UUID                  `json:"user_id"`
	Data    dto.WinnerNotificationData `json:"data"`
	Message string                     `json:"message,omitempty"`
}

func InitUserws(
	log *zap.Logger,
	balanceStorage storage.Balance,
	redisClient contracts.Redis,
) UserWS {
	return &User{
		log:                     log,
		balanceStorage:          balanceStorage,
		redisClient:             redisClient,
		UserBalanceSocket:       make(map[uuid.UUID]map[*websocket.Conn]bool),
		UserBalanceSocketLocker: make(map[*websocket.Conn]*sync.Mutex),
	}
}

func (b *User) AddToPlayerBalanceWS(ctx context.Context, userID uuid.UUID, conn *websocket.Conn) {
	if _, exists := b.UserBalanceSocket[userID]; !exists {
		b.UserBalanceSocket[userID] = make(map[*websocket.Conn]bool)
	}
	b.UserBalanceSocket[userID][conn] = true
	if _, exists := b.UserBalanceSocketLocker[conn]; !exists {
		b.UserBalanceSocketLocker[conn] = &sync.Mutex{}
	}

	b.getUserBalanceSocketLocker(conn).Lock()
	defer b.getUserBalanceSocketLocker(conn).Unlock()

	b.log.Info("new user added to balance socket", zap.String("userID", userID.String()))

	// Send ping
	if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
		b.log.Error("Failed to send ping", zap.Error(err))
		return
	}

	// Send connection message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to user balance socket")); err != nil {
		b.log.Error("Failed to send connection message", zap.Error(err))
		return
	}

	// Get and send player level
	playerLevel, err := b.GetUserBalance(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	p := dto.UserBalanceResp{
		UserID:           userID,
		Balance:          playerLevel.Balance,
		BalanceFormatted: FormatCurrency(playerLevel.Balance, playerLevel.Currency),
		Currency:         playerLevel.Currency,
	}

	msg, err := json.Marshal(p)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		b.log.Error("Failed to send player level message", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Player level socket connection established", zap.String("userID", userID.String()))
}

func (b *User) getUserBalanceSocketLocker(conn *websocket.Conn) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if locker, exists := b.UserBalanceSocketLocker[conn]; exists {
		return locker
	}

	locker := &sync.Mutex{}
	b.UserBalanceSocketLocker[conn] = locker
	return locker
}

func (b *User) GetUserBalance(ctx context.Context, userID uuid.UUID) (dto.UserBalanceResp, error) {
	if b.balanceStorage == nil {
		b.log.Error("Balance storage is nil - cannot get user balance", zap.String("userID", userID.String()))
		return dto.UserBalanceResp{}, fmt.Errorf("balance storage not initialized")
	}

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       userID,
		CurrencyCode: constant.DEFAULT_CURRENCY,
	})
	if err != nil {
		b.log.Error("Failed to get user balance", zap.Error(err), zap.String("userID", userID.String()))
		return dto.UserBalanceResp{}, err
	}
	if !exist {
		b.log.Warn("User balance does not exist", zap.String("userID", userID.String()))
		return dto.UserBalanceResp{
			UserID:           userID,
			Balance:          decimal.Zero,
			BalanceFormatted: FormatCurrency(decimal.Zero, constant.DEFAULT_CURRENCY),
			Currency:         constant.DEFAULT_CURRENCY,
		}, nil
	}
	return dto.UserBalanceResp{
		UserID:           userID,
		Balance:          balance.RealMoney,
		BalanceFormatted: FormatCurrency(balance.RealMoney, balance.CurrencyCode),
		Currency:         balance.CurrencyCode,
	}, nil
}

func (b *User) TriggerBalanceWS(ctx context.Context, userID uuid.UUID) {

	playerLevel, err := b.GetUserBalance(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	pl := dto.UserBalanceResp{
		UserID:           userID,
		Balance:          playerLevel.Balance,
		BalanceFormatted: FormatCurrency(playerLevel.Balance, playerLevel.Currency),
		Currency:         playerLevel.Currency,
	}

	msg, err := json.Marshal(pl)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	log.Println("Triggering balance connection for user", b.UserBalanceSocket[userID])

	if conns, exists := b.UserBalanceSocket[userID]; exists {

		for conn := range conns {
			locker := b.getUserBalanceSocketLocker(conn)
			locker.Lock()
			defer locker.Unlock()
			
			if conn == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send player level response", zap.Error(err), zap.String("userID", userID.String()))
				// Remove the broken connection from the map
				delete(b.UserBalanceSocket[userID], conn)
				delete(b.UserBalanceSocketLocker, conn)
				continue
			}
		}
	} else {
		b.log.Info("No user balance socket connections found for user", zap.String("userID", userID.String()))
	}
}

// TriggerCashbackWS sends real-time cashback updates to connected WebSocket clients
func (b *User) TriggerCashbackWS(ctx context.Context, userID uuid.UUID, cashbackData dto.EnhancedUserCashbackSummary) {
	cashbackMessage := CashbackWebSocketMessage{
		Type:    "cashback_update",
		UserID:  userID,
		Data:    cashbackData,
		Message: "Cashback availability updated",
	}

	msg, err := json.Marshal(cashbackMessage)
	if err != nil {
		b.log.Error("Failed to marshal cashback message", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Triggering cashback WebSocket update",
		zap.String("userID", userID.String()),
		zap.String("available_cashback", cashbackData.AvailableCashback.String()),
		zap.String("current_tier", cashbackData.CurrentTier.TierName))

	if conns, exists := b.UserBalanceSocket[userID]; exists {
		for conn := range conns {
			locker := b.getUserBalanceSocketLocker(conn)
			locker.Lock()
			defer locker.Unlock()
			
			if conn == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send cashback update", zap.Error(err), zap.String("userID", userID.String()))
				continue
			}
		}
	} else {
		b.log.Info("No user balance socket connections found for cashback update", zap.String("userID", userID.String()))
	}
}

// TriggerWinnerNotificationWS sends real-time winner notifications to connected WebSocket clients
func (b *User) TriggerWinnerNotificationWS(ctx context.Context, userID uuid.UUID, winnerData dto.WinnerNotificationData) {
	// Store winner notification in Redis with 1-hour expiration
	if err := b.StoreWinnerNotification(ctx, winnerData); err != nil {
		b.log.Error("Failed to store winner notification in Redis", zap.Error(err))
	}

	winnerMessage := WinnerNotificationWebSocketMessage{
		Type:    "winner_notification",
		UserID:  userID,
		Data:    winnerData,
		Message: "Congratulations! You won!",
	}

	msg, err := json.Marshal(winnerMessage)
	if err != nil {
		b.log.Error("Failed to marshal winner notification message", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Triggering winner notification WebSocket update",
		zap.String("userID", userID.String()),
		zap.String("username", winnerData.Username),
		zap.String("game_name", winnerData.GameName),
		zap.String("game_id", winnerData.GameID),
		zap.String("bet_amount", winnerData.BetAmount.String()),
		zap.String("win_amount", winnerData.WinAmount.String()))

	if conns, exists := b.UserBalanceSocket[userID]; exists {
		for conn := range conns {
			locker := b.getUserBalanceSocketLocker(conn)
			locker.Lock()
			defer locker.Unlock()
			
			if conn == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send winner notification", zap.Error(err), zap.String("userID", userID.String()))
				continue
			}
		}
	} else {
		b.log.Info("No user balance socket connections found for winner notification", zap.String("userID", userID.String()))
	}
}

// StoreWinnerNotification stores winner notification data in Redis with 1-hour expiration
func (b *User) StoreWinnerNotification(ctx context.Context, winnerData dto.WinnerNotificationData) error {
	if b.redisClient == nil {
		b.log.Warn("Redis client not available, skipping winner notification storage")
		return nil
	}

	// Create a unique key for this winner notification
	key := fmt.Sprintf("winner_notification:%s:%s", winnerData.Username, winnerData.TransactionID)

	// Convert to JSON for Redis storage
	jsonData, err := json.Marshal(winnerData)
	if err != nil {
		b.log.Error("Failed to marshal winner notification data",
			zap.String("username", winnerData.Username),
			zap.String("transaction_id", winnerData.TransactionID),
			zap.Error(err))
		return err
	}

	// Store with 1-hour expiration
	err = b.redisClient.Set(ctx, key, string(jsonData), time.Hour)
	if err != nil {
		b.log.Error("Failed to store winner notification in Redis",
			zap.String("key", key),
			zap.String("username", winnerData.Username),
			zap.String("transaction_id", winnerData.TransactionID),
			zap.Error(err))
		return err
	}

	b.log.Info("Winner notification stored in Redis",
		zap.String("key", key),
		zap.String("username", winnerData.Username),
		zap.String("game_name", winnerData.GameName),
		zap.String("win_amount", winnerData.WinAmount.String()))

	return nil
}

// GetWinnerNotifications retrieves all winner notifications for a user from Redis
func (b *User) GetWinnerNotifications(ctx context.Context, userID uuid.UUID) ([]dto.WinnerNotificationData, error) {
	if b.redisClient == nil {
		b.log.Warn("Redis client not available, returning empty winner notifications")
		return []dto.WinnerNotificationData{}, nil
	}

	// For now, we'll return empty since we need to implement pattern matching
	// In a real implementation, you'd use Redis SCAN with pattern matching
	// to find all keys matching "winner_notification:*" and filter by user
	b.log.Info("Retrieving winner notifications from Redis", zap.String("userID", userID.String()))

	// TODO: Implement Redis pattern matching to get all winner notifications
	// This would require extending the Redis interface to support SCAN operations
	return []dto.WinnerNotificationData{}, nil
}
