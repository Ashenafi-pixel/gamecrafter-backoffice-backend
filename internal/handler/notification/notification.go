package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type Notification struct {
	log                *zap.Logger
	notificationModule module.Notification
}

func Init(log *zap.Logger, notificationModule module.Notification) *Notification {
	return &Notification{
		log:                log,
		notificationModule: notificationModule,
	}
}

// CreateNotification Create a new notification.
//
//	@Summary		CreateNotification
//	@Description	Create a new notification for a user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Bearer <token>"
//	@Param			notification	body		dto.NotificationPayload	true	"Notification payload"
//	@Success		200				{object}	dto.NotificationResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/notifications [post]
func (n *Notification) CreateNotification(ctx *gin.Context) {
	var req dto.NotificationPayload
	userID := ctx.GetString("user-id")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		n.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}
	req.CreatedBy = &userIDParsed

	// Try to deliver in real time
	delivered := n.notificationModule.SendNotificationToUser(req.UserID, req)

	resp, err := n.notificationModule.StoreNotification(ctx, req, delivered)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// GetNotifications Get user notifications.
//
//	@Summary		GetNotifications
//	@Description	Get paginated notifications for the current user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//
//	@Param			Authorization	header		string	true	"Bearer <token>"
//
//	@Param			page			query		int		false	"Page number"
//	@Param			per_page		query		int		false	"Items per page"
//	@Success		200				{object}	dto.GetNotificationsResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/notifications [get]
func (n *Notification) GetUserNotifications(ctx *gin.Context) {
	var req dto.GetNotificationsRequest
	userID := ctx.GetString("user-id")

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid query parameters")
		_ = ctx.Error(err)
		return
	}

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		n.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.UserID = userIDParsed

	resp, err := n.notificationModule.GetUserNotifications(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// MarkNotificationRead Mark a notification as read.
//
//	@Summary		MarkNotificationRead
//	@Description	Mark a notification as read
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			id				path		string	true	"Notification ID"
//	@Success		200				{object}	dto.MarkNotificationReadResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/notifications/{id}/mark-read [patch]
func (n *Notification) MarkNotificationRead(ctx *gin.Context) {
	idStr := ctx.Param("id")
	userID := ctx.GetString("user-id")

	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid notification ID format")
		_ = ctx.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		n.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req := dto.MarkNotificationReadRequest{
		UserID:         userIDParsed,
		NotificationID: notificationID,
	}
	resp, err := n.notificationModule.MarkNotificationRead(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// MarkAllNotificationsRead Mark all notifications as read.
//
//	@Summary		MarkAllNotificationsRead
//	@Description	Mark all notifications as read for the current user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	dto.MarkAllNotificationsReadResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/notifications/mark-all-read [patch]
func (n *Notification) MarkAllNotificationsRead(ctx *gin.Context) {
	userID := ctx.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		n.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req := dto.MarkAllNotificationsReadRequest{UserID: userIDParsed}
	resp, err := n.notificationModule.MarkAllNotificationsRead(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// DeleteNotification Delete a notification.
//
//	@Summary		DeleteNotification
//	@Description	Delete a notification by ID
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//
//	@Param			Authorization	header		string	true	"Bearer <token>"
//
//	@Param			id				path		string	true	"Notification ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/notifications/{id} [delete]
func (n *Notification) DeleteNotification(ctx *gin.Context) {
	idStr := ctx.Param("id")
	userID := ctx.GetString("user-id")

	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid notification ID format")
		_ = ctx.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		n.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req := dto.DeleteNotificationRequest{
		UserID:         userIDParsed,
		NotificationID: notificationID,
	}
	resp, err := n.notificationModule.DeleteNotification(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusNoContent, resp)
}
