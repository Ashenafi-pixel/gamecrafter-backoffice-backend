package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (u *User) UpdateUserPointByReferingUser(ctx context.Context, referalCode, userID string) {
	//get user by referal code

	point, exist, err := u.userStorage.GetUserPointsByReferalPoint(ctx, referalCode)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("referal_code", referalCode))
		return
	}

	if !exist {
		err := fmt.Errorf("unable to get user by referal code %s", referalCode)
		u.log.Error(err.Error())
		return
	}

	userLock := u.getUserLock(point.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	//get multiplier by for referal code
	multiplier, exist, err := u.userStorage.GetReferalMultiplier(ctx)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("referal_code", referalCode))
		return
	}

	if !exist {
		err := fmt.Errorf("unable to get referal code multiplier")
		u.log.Error(err.Error())
		return
	}

	// multiplay the point by the multiplier
	newPoint := multiplier.PointMultiplier + point.Point
	parsedPoint := decimal.NewFromInt(int64(newPoint))
	if err := u.userStorage.UpdateUserPointByUserID(ctx, point.UserID, parsedPoint); err != nil {
		return
	}

	//save balance logs
	operationalGroupAndType, err := u.CreateOrGetOperationalGroupAndType(ctx, constant.DEPOSIT, constant.REFERAL_POINT)
	if err != nil {
		//reverse point
		u.log.Error(err.Error())

		if err := u.userStorage.UpdateUserPointByUserID(ctx, point.UserID, decimal.NewFromInt(int64(point.Point))); err != nil {
			return
		}
		return
	}

	decimalMuL := decimal.NewFromInt(int64(multiplier.PointMultiplier))
	decimalNewBalanc := decimal.NewFromInt(int64(newPoint))
	u.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
		OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
		OperationalGroupType: operationalGroupAndType.OperationalTypeID,
		UpdateReq: dto.UpdateBalanceReq{
			Component:   constant.REAL_MONEY,
			Description: fmt.Sprintf("%s_%s", constant.REFERAL_POINT, userID),
			Amount:      decimalMuL,
		},
		UpdateRes: dto.UpdateBalanceRes{
			Data: dto.BalanceData{
				UserID:     point.UserID,
				Currency:   constant.POINT_CURRENCY,
				NewBalance: decimalNewBalanc,
			},
		},
	})
}

func (b *User) getUserLock(userID uuid.UUID) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.locker[userID]; !exists {
		b.locker[userID] = &sync.Mutex{}
	}
	return b.locker[userID]
}

func (u *User) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
	// get transfer operational  group and  type if not exist create group transfer and type transfer-internal
	var operationalGroup dto.OperationalGroup
	var exist bool
	var err error
	var operationalGroupTypeID dto.OperationalGroupType
	operationalGroup, exist, err = u.operationalGroupStorage.GetOperationalGroupByName(ctx, constant.DEPOSIT)
	if err != nil {
		return dto.OperationalGroupAndType{}, err
	}
	if !exist {
		// create transfer internal group and type
		operationalGroup, err = u.operationalGroupStorage.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        constant.DEPOSIT,
			Description: "this operational group allow user to deposite currency to the system",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = u.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
			GroupID:     operationalGroup.ID,
			Name:        operationalType,
			Description: "internal transactions",
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
	}
	// create or get operational group type if operational group exist
	if exist {
		// get operational group type
		operationalGroupTypeID, exist, err = u.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    operationalType,
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
		if !exist {
			operationalGroupTypeID, err = u.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
				GroupID: operationalGroup.ID,
				Name:    operationalType,
			})
			if err != nil {
				return dto.OperationalGroupAndType{}, err
			}

		}
	}
	return dto.OperationalGroupAndType{
		OperationalGroupID: operationalGroup.ID,
		OperationalTypeID:  operationalGroupTypeID.ID,
	}, nil
}

func (u *User) SaveBalanceLogs(ctx context.Context, saveLogsReq dto.SaveBalanceLogsReq) (dto.BalanceLogs, error) {
	transactionID := utils.GenerateTransactionId()
	return u.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             saveLogsReq.UpdateRes.Data.UserID,
		Component:          saveLogsReq.UpdateReq.Component,
		Currency:           saveLogsReq.UpdateRes.Data.Currency,
		Description:        saveLogsReq.UpdateReq.Description,
		ChangeAmount:       saveLogsReq.UpdateReq.Amount,
		OperationalGroupID: saveLogsReq.OperationalGroupID,
		OperationalTypeID:  saveLogsReq.OperationalGroupType,
		BalanceAfterUpdate: &saveLogsReq.UpdateRes.Data.NewBalance,
		TransactionID:      &transactionID,
	})
}

func (u *User) GetMyReferralCode(ctx context.Context, userID uuid.UUID) (string, error) {
	usr, exist, err := u.userStorage.GetUserByID(ctx, userID)

	if err != nil {
		return "", err
	}

	if !exist {
		err := fmt.Errorf("user not found with this id %s", userID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return "", err
	}
	return usr.ReferralCode, nil
}

func (u *User) GetUserReferalUsersByUserID(ctx context.Context, userID uuid.UUID) (dto.MyRefferedUsers, error) {
	referredUsers, exist, err := u.userStorage.GetUserReferalUsersByUserID(ctx, userID)
	if err != nil {
		return dto.MyRefferedUsers{}, err
	}

	if !exist {
		usr := dto.MyRefferedUsers{}
		return usr, nil
	}

	return referredUsers, nil
}

func (u *User) GetReferalMultiplier(ctx context.Context) (dto.ReferalUpdateResp, error) {

	resp, err := u.userStorage.GetCurrentReferralMultiplier(ctx)
	if err != nil {
		return dto.ReferalUpdateResp{}, err
	}

	return dto.ReferalUpdateResp{
		Message:         constant.SUCCESS,
		PointMultiplier: resp,
	}, nil
}

func (u *User) UpdateReferalMultiplier(ctx context.Context, mul dto.UpdateReferralPointReq) (dto.ReferalUpdateResp, error) {
	if mul.Multiplier <= 0 {
		err := fmt.Errorf("referral multiplier can not be less than zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ReferalUpdateResp{}, err
	}
	result, err := u.userStorage.UpdateReferralMultiplier(ctx, mul.Multiplier)
	if err != nil {
		return dto.ReferalUpdateResp{}, err
	}
	return dto.ReferalUpdateResp{
		Message:         constant.SUCCESS,
		PointMultiplier: result.PointMultiplier,
	}, nil

}

func (u *User) UpdateUsersPointsForReferrances(ctx context.Context, adminID uuid.UUID, req []dto.MassReferralReq) (dto.MassReferralRes, error) {
	userToUpdate := make(map[uuid.UUID]int)
	newPointMap := make(map[uuid.UUID]int)
	var balance dto.UserPoint
	var err error
	var exist bool
	var usr dto.User
	var respData []dto.MassReferralResData
	for _, r := range req {
		// get users by user id
		usr, exist, err = u.userStorage.GetUserByID(ctx, r.UserID)
		if err != nil {
			return dto.MassReferralRes{}, err

		}

		if !exist {
			err := fmt.Errorf("unable to find user with user id %s", r.UserID.String())
			u.log.Error(err.Error())
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.MassReferralRes{}, err
		}

		// get users current points
		balance, exist, err = u.userStorage.GetUserPointsByReferalPoint(ctx, usr.ReferralCode)
		if err != nil {
			return dto.MassReferralRes{}, err
		}

		if !exist {
			// create balance with point
			balance, err = u.userStorage.CreateUserPoint(ctx, r.UserID, decimal.NewFromInt(int64(r.Point)))
			if err != nil {
				return dto.MassReferralRes{}, err
			}
		}
		newPointMap[usr.ID] = r.Point

		//add new point to the existing values
		newValue := balance.Point + r.Point
		userToUpdate[usr.ID] = newValue

	}

	for userID, newPoint := range userToUpdate {
		//update user balance
		if err := u.userStorage.UpdateUserPointByUserID(ctx, userID, decimal.NewFromInt(int64(newPoint))); err != nil {
			return dto.MassReferralRes{}, err
		}

		//save the log about the updated points
		//save balance logs
		operationalGroupAndType, err := u.CreateOrGetOperationalGroupAndType(ctx, constant.DEPOSIT, constant.REFERAL_POINT)
		if err != nil {
			return dto.MassReferralRes{}, err
		}

		decimalMuL := decimal.NewFromInt(int64(newPointMap[userID]))
		decimalNewBalanc := decimal.NewFromInt(int64(newPoint))
		u.SaveBalanceLogs(ctx, dto.SaveBalanceLogsReq{
			OperationalGroupID:   operationalGroupAndType.OperationalGroupID,
			OperationalGroupType: operationalGroupAndType.OperationalTypeID,
			UpdateReq: dto.UpdateBalanceReq{
				Component:   constant.POINTS,
				Description: fmt.Sprintf("%s_admin_%s", constant.REFERAL_POINT, adminID.String()),
				Amount:      decimalMuL,
			},
			UpdateRes: dto.UpdateBalanceRes{
				Data: dto.BalanceData{
					UserID:     userID,
					Currency:   constant.POINT_CURRENCY,
					NewBalance: decimalNewBalanc,
				},
			},
		})
		respData = append(respData, dto.MassReferralResData{
			UserID: userID,
			Point:  newPoint,
		})
	}

	return dto.MassReferralRes{
		Message:      constant.SUCCESS,
		UpdatedUsers: respData,
	}, nil
}

func (u *User) GetAdminAssignedPoints(ctx context.Context, req dto.GetAdminAssignedPointsReq) (dto.GetAdminAssignedPointsRes, error) {
	var resp []dto.GetAdminAssignedPointsData
	if req.PerPage <= 0 || req.Page <= 0 {
		err := fmt.Errorf("please provide page and per_page")
		u.log.Warn(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetAdminAssignedPointsRes{}, err
	}
	// page and per_page logic here
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	usrs, exist, err := u.userStorage.GetAdminAssignedPoints(ctx, req.PerPage, req.Page)

	if err != nil {
		return dto.GetAdminAssignedPointsRes{}, err
	}

	if !exist {
		return dto.GetAdminAssignedPointsRes{}, nil
	}

	for _, usr := range usrs.Data {
		// get user
		usrRes, exist, err := u.userStorage.GetUserByID(ctx, usr.UserID)
		if err != nil {
			return dto.GetAdminAssignedPointsRes{}, err
		}

		if !exist {
			err := fmt.Errorf("user dose not exist with id of %s ", usr.UserID.String())
			u.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
			return dto.GetAdminAssignedPointsRes{}, err
		}

		// get admin

		adminRes, exist, err := u.userStorage.GetUserByID(ctx, usr.UserID)
		if err != nil {
			return dto.GetAdminAssignedPointsRes{}, err
		}

		if !exist {
			err := fmt.Errorf("admin dose not exist with id of %s ", usr.AdminID.String())
			u.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
			return dto.GetAdminAssignedPointsRes{}, err
		}

		// apped response
		resp = append(resp, dto.GetAdminAssignedPointsData{
			Transaction: usr.TransactionID,
			Admin: dto.AdminInfo{
				UserID:    adminRes.ID,
				FirstName: adminRes.FirstName,
				LastName:  adminRes.LastName,
				Email:     adminRes.Email,
				Phone:     adminRes.PhoneNumber,
				Profile:   adminRes.ProfilePicture,
			},
			Amount: usr.AddedPoints,
			User: dto.User{
				ID:              usr.UserID,
				PhoneNumber:     usrRes.PhoneNumber,
				FirstName:       usrRes.FirstName,
				LastName:        usrRes.LastName,
				Email:           usrRes.Email,
				DefaultCurrency: usrRes.DefaultCurrency,
				ProfilePicture:  usrRes.ProfilePicture,
				DateOfBirth:     usrRes.DateOfBirth,
			},
		})

	}
	return dto.GetAdminAssignedPointsRes{
		Message:             constant.SUCCESS,
		AdminAssignedPoints: resp,
		TotalPages:          int32(usrs.TotalPage),
	}, nil

}
