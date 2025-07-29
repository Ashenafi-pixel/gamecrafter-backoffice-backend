package notification

import (
	"context"
	"encoding/json"

	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type Notification struct {
	log                 *zap.Logger
	notificationStorage storage.Notification

	notificationSockets       map[uuid.UUID]map[*websocket.Conn]bool
	notificationSocketLockers map[*websocket.Conn]*sync.Mutex
	mutex                     sync.Mutex
}

func Init(notificationStorage storage.Notification, log *zap.Logger) *Notification {
	return &Notification{
		log:                       log,
		notificationStorage:       notificationStorage,
		notificationSockets:       make(map[uuid.UUID]map[*websocket.Conn]bool),
		notificationSocketLockers: make(map[*websocket.Conn]*sync.Mutex),
	}
}

func (n *Notification) StoreNotification(ctx context.Context, req dto.NotificationPayload, delivered bool) (dto.NotificationResponse, error) {
	if err := dto.ValidateNotificationPayload(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.NotificationResponse{}, err
	}

	return n.notificationStorage.StoreNotification(ctx, req, delivered)
}

func (n *Notification) GetUserNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	offset := (req.Page - 1) * req.PerPage

	req.Page = offset

	notifRes, err := n.notificationStorage.GetUserNotifications(ctx, req)
	if err != nil {
		return dto.GetNotificationsResponse{}, err
	}

	return notifRes, nil
}

func (n *Notification) MarkNotificationRead(ctx context.Context, req dto.MarkNotificationReadRequest) (dto.MarkNotificationReadResponse, error) {
	return n.notificationStorage.MarkNotificationRead(ctx, req)
}

func (n *Notification) MarkAllNotificationsRead(ctx context.Context, req dto.MarkAllNotificationsReadRequest) (dto.MarkAllNotificationsReadResponse, error) {
	return n.notificationStorage.MarkAllNotificationsRead(ctx, req)
}

func (n *Notification) DeleteNotification(ctx context.Context, req dto.DeleteNotificationRequest) (dto.DeleteNotificationResponse, error) {
	return n.notificationStorage.DeleteNotification(ctx, req)
}

func (n *Notification) GetUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int32, error) {
	if userID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid user ID")
		return 0, err
	}

	return n.notificationStorage.GetUnreadNotificationCount(ctx, userID)
}

// AddNotificationSocketConnection registers a websocket connection for a user
func (n *Notification) AddNotificationSocketConnection(userID uuid.UUID, conn *websocket.Conn) {
	locker := n.getNotificationSocketLocker(conn)
	locker.Lock()
	if _, exists := n.notificationSockets[userID]; !exists {
		n.notificationSockets[userID] = make(map[*websocket.Conn]bool)
	}
	n.notificationSockets[userID][conn] = true
	if _, exists := n.notificationSocketLockers[conn]; !exists {
		n.notificationSocketLockers[conn] = &sync.Mutex{}
	}
	locker.Unlock()

	// Optionally send a connection message
	locker = n.getNotificationSocketLocker(conn)
	locker.Lock()
	defer locker.Unlock()
	_ = conn.WriteMessage(websocket.TextMessage, []byte("Connected to notification socket"))
}

// RemoveNotificationSocketConnection removes a websocket connection for a user
func (n *Notification) RemoveNotificationSocketConnection(userID uuid.UUID, conn *websocket.Conn) {
	locker := n.getNotificationSocketLocker(conn)
	locker.Lock()
	if conns, exists := n.notificationSockets[userID]; exists {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(n.notificationSockets, userID)
		}
	}
	delete(n.notificationSocketLockers, conn)
	locker.Unlock()
}

// getNotificationSocketLocker returns the mutex for a connection
func (n *Notification) getNotificationSocketLocker(conn *websocket.Conn) *sync.Mutex {
	locker := n.notificationSocketLockers[conn]
	if locker == nil {
		locker = &sync.Mutex{}
		n.notificationSocketLockers[conn] = locker
	}
	return locker
}

// SendNotificationToUser attempts to send a notification to all active sockets for a user
func (n *Notification) SendNotificationToUser(userID uuid.UUID, notif dto.NotificationPayload) bool {
	locker := n.getNotificationSocketLocker(nil) // Use a temporary locker for the main map
	locker.Lock()
	conns, exists := n.notificationSockets[userID]
	locker.Unlock()

	if !exists || len(conns) == 0 {
		n.log.Info("No active notification sockets for user", zap.String("userID", userID.String()))
		return false
	}

	msg, err := json.Marshal(notif)
	if err != nil {
		n.log.Error("Failed to marshal notification payload", zap.Error(err))
		return false
	}

	delivered := false
	for conn := range conns {
		connLocker := n.getNotificationSocketLocker(conn)
		connLocker.Lock()
		if conn == nil {
			connLocker.Unlock()
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			n.log.Warn("Failed to send notification to user", zap.Error(err), zap.String("userID", userID.String()))
			connLocker.Unlock()
			n.RemoveNotificationSocketConnection(userID, conn)
			continue
		}
		delivered = true
		connLocker.Unlock()
	}
	return delivered
}
