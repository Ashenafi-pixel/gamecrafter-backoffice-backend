package user

import (
	"context"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
)

func (u *user) GetUsersByDepartmentNotificationTypes(ctx context.Context, notificationTypes []string) ([]dto.GetUsersForNotificationRes, error) {
	var usersRes []dto.GetUsersForNotificationRes
	for _, notificationType := range notificationTypes {
		usrs, err := u.db.Queries.GetUsersByDepartmentNotificationTypes(ctx, notificationType)
		if err != nil && err.Error() != dto.ErrNoRows {
			return []dto.GetUsersForNotificationRes{}, err
		}
		for _, usr := range usrs {
			usersRes = append(usersRes, dto.GetUsersForNotificationRes{
				Email:  usr.Email.String,
				UserID: usr.ID,
				Phone:  usr.PhoneNumber.String,
			})
		}
	}
	return usersRes, nil
}
