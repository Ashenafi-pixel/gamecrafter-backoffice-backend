package utils

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UserWS interface {
	AddToPlayerBalanceWS(ctx context.Context, userID uuid.UUID, conn *websocket.Conn)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (dto.UserBalanceResp, error)
	TriggerBalanceWS(ctx context.Context, userID uuid.UUID)
}

type User struct {
	log                     *zap.Logger
	balanceStorage          storage.Balance
	UserBalanceSocket       map[uuid.UUID]map[*websocket.Conn]bool
	UserBalanceSocketLocker map[*websocket.Conn]*sync.Mutex
	mutex                   sync.Mutex
}

func InitUserws(
	log *zap.Logger,
	balanceStorage storage.Balance,
) UserWS {
	return &User{
		log:                     log,
		balanceStorage:          balanceStorage,
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
		UserID:   userID,
		Balance:  playerLevel.Balance,
		Currency: playerLevel.Currency,
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

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   userID,
		Currency: constant.DEFAULT_CURRENCY,
	})
	if err != nil {
		b.log.Error("Failed to get user balance", zap.Error(err), zap.String("userID", userID.String()))
		return dto.UserBalanceResp{}, err
	}
	if !exist {
		b.log.Warn("User balance does not exist", zap.String("userID", userID.String()))
		return dto.UserBalanceResp{
			UserID:  userID,
			Balance: decimal.Zero,
		}, nil
	}
	return dto.UserBalanceResp{
		UserID:   userID,
		Balance:  balance.RealMoney,
		Currency: balance.Currency,
	}, nil
}

func (b *User) TriggerBalanceWS(ctx context.Context, userID uuid.UUID) {

	playerLevel, err := b.GetUserBalance(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	pl := dto.UserBalanceResp{
		UserID:   userID,
		Balance:  playerLevel.Balance,
		Currency: playerLevel.Currency,
	}

	msg, err := json.Marshal(pl)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	log.Println("Triggering balance connection for user", b.UserBalanceSocket[userID])

	if conns, exists := b.UserBalanceSocket[userID]; exists {

		for conn := range conns {
			b.getUserBalanceSocketLocker(conn).Lock()
			if conn == nil {
				if b.getUserBalanceSocketLocker(conn) != nil {
					b.getUserBalanceSocketLocker(conn).Unlock()
				}
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send player level response", zap.Error(err), zap.String("userID", userID.String()))
				if b.getUserBalanceSocketLocker(conn) != nil {
					b.getUserBalanceSocketLocker(conn).Unlock()
				}
				continue
			}
			if b.getUserBalanceSocketLocker(conn) != nil {
				b.getUserBalanceSocketLocker(conn).Unlock()
			}

		}
	} else {
		b.log.Info("No user balance socket connections found for user", zap.String("userID", userID.String()))
	}
}
