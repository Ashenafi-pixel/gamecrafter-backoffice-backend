package notification

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type Notification struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) *Notification {
	return &Notification{
		db:  db,
		log: log,
	}
}

func (n *Notification) StoreNotification(ctx context.Context, req dto.NotificationPayload, delivered bool) (dto.NotificationResponse, error) {
	var metaJSON pgtype.JSONB
	metaJSON.Status = pgtype.Null // Initialize as null

	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			n.log.Error("unable to marshal notification metadata", zap.Error(err), zap.Any("metadata", req.Metadata))
			err = errors.ErrUnableTocreate.Wrap(err, "unable to marshal notification metadata")
			return dto.NotificationResponse{}, err
		}
		if err := metaJSON.Set(b); err != nil {
			n.log.Error("unable to set JSONB for notification metadata", zap.Error(err), zap.Any("metadata", req.Metadata))
			err = errors.ErrUnableTocreate.Wrap(err, "unable to set JSONB for notification metadata")
			return dto.NotificationResponse{}, err
		}
	}

	// Handle nullable CreatedBy field
	var createdBy uuid.NullUUID
	if req.CreatedBy != nil {
		createdBy = uuid.NullUUID{
			UUID:  *req.CreatedBy,
			Valid: true,
		}
	} else {
		// For system/third-party notifications, set as NULL
		createdBy = uuid.NullUUID{
			Valid: false,
		}
	}

	res, err := n.db.Queries.InsertUserNotification(ctx, db.InsertUserNotificationParams{
		UserID:    req.UserID,
		Title:     req.Title,
		Content:   req.Content,
		Type:      string(req.Type),
		Metadata:  metaJSON,
		Read:      false, // Default to unread
		Delivered: delivered,
		CreatedBy: createdBy,
	})
	if err != nil {
		n.log.Error("unable to store notification", zap.Error(err), zap.Any("user_id", req.UserID))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to store notification")
		return dto.NotificationResponse{}, err
	}
	return dto.NotificationResponse{
		ID:        res.ID,
		Delivered: res.Delivered,
		Timestamp: res.CreatedAt,
	}, nil
}

func (n *Notification) GetUserNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error) {
	notifs, err := n.db.Queries.GetUserNotifications(ctx, db.GetUserNotificationsParams{
		UserID: req.UserID,
		Limit:  int32(req.PerPage),
		Offset: int32((req.Page - 1) * req.PerPage),
	})

	if err != nil {
		n.log.Error("unable to get user notifications", zap.Error(err), zap.Any("user_id", req.UserID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get user notifications")
		return dto.GetNotificationsResponse{}, err
	}

	var total int32
	if len(notifs) > 0 {
		total = int32(notifs[0].Total)
	}

	unreadCount, err := n.GetUnreadNotificationCount(ctx, req.UserID)
	result := make([]dto.UserNotification, len(notifs))
	for i, r := range notifs {
		var meta dto.NotificationMetadata
		if r.Metadata.Status == pgtype.Present && r.Metadata.Bytes != nil {
			_ = json.Unmarshal(r.Metadata.Bytes, &meta)
		}
		// Handle nullable CreatedBy field
		createdByUUID := dto.NullToUUID(r.CreatedBy)
		var createdBy *uuid.UUID
		if createdByUUID != uuid.Nil {
			createdBy = &createdByUUID
		}

		result[i] = dto.UserNotification{
			ID:        r.ID,
			UserID:    r.UserID,
			Title:     r.Title,
			Content:   r.Content,
			Type:      dto.NotificationType(r.Type),
			Metadata:  meta,
			Read:      r.Read,
			Delivered: r.Delivered,
			CreatedBy: createdBy,
			ReadAt:    dto.NullToTime(r.ReadAt),
			CreatedAt: r.CreatedAt,
		}
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return dto.GetNotificationsResponse{
		Message:       "Notifications retrieved successfully",
		Notifications: result,
		Total:         int(total),
		TotalPages:    totalPages,
		Page:          req.Page,
		PerPage:       req.PerPage,
		UnreadCount:   int(unreadCount),
	}, nil
}

func (n *Notification) GetAllNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error) {
	n.log.Info("GetAllNotifications called", zap.Int("page", req.Page), zap.Int("per_page", req.PerPage))
	
	offset := (req.Page - 1) * req.PerPage
	n.log.Info("Calculated offset", zap.Int("offset", offset))
	
	notifs, err := n.db.Queries.GetAllNotifications(ctx, db.GetAllNotificationsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(offset),
	})

	if err != nil {
		n.log.Error("unable to get all notifications", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get all notifications")
		return dto.GetNotificationsResponse{}, err
	}

	var total int32
	if len(notifs) > 0 {
		total = int32(notifs[0].Total)
	}

	// Get additional counts
	unreadCount, err := n.GetAllUnreadNotificationCount(ctx)
	if err != nil {
		n.log.Error("unable to get unread count", zap.Error(err))
		unreadCount = 0
	}

	deliveredCount, err := n.GetAllDeliveredNotificationCount(ctx)
	if err != nil {
		n.log.Error("unable to get delivered count", zap.Error(err))
		deliveredCount = 0
	}

	readCount, err := n.GetAllReadNotificationCount(ctx)
	if err != nil {
		n.log.Error("unable to get read count", zap.Error(err))
		readCount = 0
	}

	result := make([]dto.UserNotification, len(notifs))
	for i, r := range notifs {
		var meta dto.NotificationMetadata
		if r.Metadata.Status == pgtype.Present && r.Metadata.Bytes != nil {
			_ = json.Unmarshal(r.Metadata.Bytes, &meta)
		}
		// Handle nullable CreatedBy field
		createdByUUID := dto.NullToUUID(r.CreatedBy)
		var createdBy *uuid.UUID
		if createdByUUID != uuid.Nil {
			createdBy = &createdByUUID
		}

		result[i] = dto.UserNotification{
			ID:        r.ID,
			UserID:    r.UserID,
			Title:     r.Title,
			Content:   r.Content,
			Type:      dto.NotificationType(r.Type),
			Metadata:  meta,
			Read:      r.Read,
			Delivered: r.Delivered,
			CreatedBy: createdBy,
			ReadAt:    dto.NullToTime(r.ReadAt),
			CreatedAt: r.CreatedAt,
		}
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return dto.GetNotificationsResponse{
		Message:        "All notifications retrieved successfully",
		Notifications:  result,
		Total:          int(total),
		TotalPages:     totalPages,
		Page:           req.Page,
		PerPage:        req.PerPage,
		UnreadCount:    int(unreadCount),
		DeliveredCount: int(deliveredCount),
		ReadCount:      int(readCount),
	}, nil
}

func (n *Notification) GetAllUnreadNotificationCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM user_notifications WHERE read = FALSE`
	var count int64
	err := n.db.GetPool().QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (n *Notification) GetAllDeliveredNotificationCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM user_notifications WHERE delivered = TRUE`
	var count int64
	err := n.db.GetPool().QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (n *Notification) GetAllReadNotificationCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM user_notifications WHERE read = TRUE`
	var count int64
	err := n.db.GetPool().QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (n *Notification) MarkNotificationRead(ctx context.Context, req dto.MarkNotificationReadRequest) (dto.MarkNotificationReadResponse, error) {
	err := n.db.Queries.MarkNotificationRead(ctx, db.MarkNotificationReadParams{
		ID:     req.NotificationID,
		UserID: req.UserID,
	})
	if err != nil {
		n.log.Error("unable to mark notification as read", zap.Error(err), zap.Any("notif_id", req.NotificationID), zap.Any("user_id", req.UserID))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to mark notification as read")
		return dto.MarkNotificationReadResponse{}, err
	}
	return dto.MarkNotificationReadResponse{
		Message: "Notification marked as read successfully",
		Read:    true,
	}, nil
}

func (n *Notification) MarkAllNotificationsRead(ctx context.Context, req dto.MarkAllNotificationsReadRequest) (dto.MarkAllNotificationsReadResponse, error) {
	res, err := n.db.Queries.MarkAllNotificationsRead(ctx, req.UserID)
	if err != nil {
		n.log.Error("unable to mark all notifications as read", zap.Error(err), zap.Any("user_id", req.UserID))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to mark all notifications as read")
		return dto.MarkAllNotificationsReadResponse{}, err
	}
	return dto.MarkAllNotificationsReadResponse{
		Message:      "All notifications marked as read successfully",
		UpdatedCount: int(res),
	}, nil
}

func (n *Notification) DeleteNotification(ctx context.Context, req dto.DeleteNotificationRequest) (dto.DeleteNotificationResponse, error) {
	err := n.db.Queries.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     req.NotificationID,
		UserID: req.UserID,
	})
	if err != nil {
		n.log.Error("unable to delete notification", zap.Error(err), zap.Any("notif_id", req.NotificationID), zap.Any("user_id", req.UserID))
		err = errors.ErrDBDelError.Wrap(err, "unable to delete notification")
		return dto.DeleteNotificationResponse{}, err
	}
	return dto.DeleteNotificationResponse{
		Message: "Notification deleted successfully",
		Deleted: true,
	}, nil
}

func (n *Notification) GetUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int32, error) {
	count, err := n.db.Queries.GetUnreadNotificationCount(ctx, userID)
	if err != nil {
		n.log.Error("unable to get unread notification count", zap.Error(err), zap.Any("user_id", userID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get unread notification count")
		return 0, err
	}
	return int32(count), nil
}
