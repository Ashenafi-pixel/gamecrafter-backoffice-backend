package user

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"go.uber.org/zap"
)

func (u *user) CreateReferalCodeMultiplier(ctx context.Context, req dto.ReferalMultiplierReq) (dto.ReferalData, error) {
	res, err := u.db.Queries.CreateConfig(ctx, db.CreateConfigParams{
		Name:  constant.CONFIG_POINT_MULTIPLIER,
		Value: fmt.Sprintf("%d", req.PointMultiplier),
	})

	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.ReferalData{}, err
	}

	// change from string to decimal
	decimalMultiplier, err := strconv.Atoi(res.Value)
	if err != nil {
		u.log.Error(err.Error())
	}

	return dto.ReferalData{
		ID:              res.ID,
		Name:            res.Name,
		PointMultiplier: decimalMultiplier,
	}, nil
}

func (u *user) GetReferalMultiplier(ctx context.Context) (dto.ReferalData, bool, error) {

	res, err := u.db.Queries.GetConfigByName(ctx, constant.CONFIG_POINT_MULTIPLIER)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.ReferalData{}, false, err
	}
	if err != nil {
		return dto.ReferalData{}, false, nil
	}

	decimalMultiplier, err := strconv.Atoi(res.Value)
	if err != nil {
		u.log.Error(err.Error())
	}

	return dto.ReferalData{
		ID:              res.ID,
		Name:            res.Name,
		PointMultiplier: decimalMultiplier,
	}, true, nil
}

func (u *user) UpdateReferalMultiplier(ctx context.Context, mul decimal.Decimal) (dto.ReferalData, error) {
	res, err := u.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Value: mul.String(),
		Name:  constant.CONFIG_POINT_MULTIPLIER,
	})

	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.ReferalData{}, err
	}
	decimalMultiplier, err := strconv.Atoi(res.Value)
	if err != nil {
		u.log.Error(err.Error())
	}

	return dto.ReferalData{
		ID:              res.ID,
		Name:            res.Name,
		PointMultiplier: decimalMultiplier,
	}, nil
}

func (u *user) GetUserPointsByReferalPoint(ctx context.Context, referal string) (dto.UserPoint, bool, error) {
	res, err := u.db.Queries.GetUserPointsByReferals(ctx, db.GetUserPointsByReferalsParams{
		ReferalCode: sql.NullString{String: referal, Valid: true},
		Currency:    constant.POINT_CURRENCY,
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return dto.UserPoint{}, false, nil
		}
		u.log.Error(err.Error(), zap.Any("referal_code", referal))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.UserPoint{}, false, err
	}
	intp := res.RealMoney.Decimal.IntPart()
	return dto.UserPoint{
		UserID: res.UserID,
		Point:  int(intp),
	}, true, nil
}

func (u *user) UpdateUserPointByUserID(ctx context.Context, userID uuid.UUID, points decimal.Decimal) error {
	_, err := u.db.Queries.UpdatePointByUserID(ctx, db.UpdatePointByUserIDParams{
		RealMoney: decimal.NullDecimal{Decimal: points, Valid: true},
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("user_id", userID.String()), zap.Any("new_point", points))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (u *user) GetUsersDoseNotHaveReferalCode(ctx context.Context) ([]dto.User, error) {
	var users []dto.User
	res, err := u.db.Queries.GetUsersDoseNotHaveReferalCode(ctx)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return []dto.User{}, nil
		}
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.User{}, err
	}

	for _, usr := range res {
		users = append(users, dto.User{
			ID:             usr.ID,
			PhoneNumber:    usr.PhoneNumber.String,
			FirstName:      usr.FirstName.String,
			LastName:       usr.LastName.String,
			Email:          usr.Email.String,
			ProfilePicture: usr.Profile.String,
			DateOfBirth:    usr.DateOfBirth.String,
		})
	}
	return users, nil
}

func (u *user) AddReferalCode(ctx context.Context, userID uuid.UUID, referalCode string) error {
	if err := u.db.Queries.AddReferalCode(ctx, db.AddReferalCodeParams{
		ReferalCode: sql.NullString{String: referalCode, Valid: true},
		ID:          userID,
	}); err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (u *user) GetUserReferalUsersByUserID(ctx context.Context, userID uuid.UUID) (dto.MyRefferedUsers, bool, error) {
	var refferedUsers []dto.RefferedUsers
	var userIDs []uuid.UUID
	var resp dto.MyRefferedUsers
	userIDToPoint := make(map[uuid.UUID]int)
	amount := 0
	referralLogs, err := u.db.Queries.GetUserReferalUsersByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})

	if err != nil {
		if err.Error() == "no rows in result set" {
			return dto.MyRefferedUsers{}, false, nil
		}
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.MyRefferedUsers{}, false, err
	}

	for _, reffered := range referralLogs {
		amount = amount + int(reffered.ChangeAmount.Decimal.IntPart())
		usrIDString := strings.Split(reffered.Description.String, "_")[2]
		if usrIDString == "admin" {
			refferedUsers = append(refferedUsers, dto.RefferedUsers{
				Username:  usrIDString,
				CreatedAt: reffered.Timestamp.Time,
				Amount:    int(reffered.ChangeAmount.Decimal.IntPart()),
			})
		} else {

			usrIDParsed, err := uuid.Parse(usrIDString)

			if err != nil {
				u.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.MyRefferedUsers{}, false, err
			}

			userIDs = append(userIDs, usrIDParsed)
			userIDToPoint[usrIDParsed] = int(reffered.ChangeAmount.Decimal.IntPart())
		}

	}

	if len(userIDs) > 0 {

		reffered, err := u.db.Queries.GetUsersInArrayOfUserIDs(ctx, userIDs)
		if err != nil {
			u.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
			return dto.MyRefferedUsers{}, false, err
		}

		for _, rf := range reffered {
			refferedUsers = append(refferedUsers, dto.RefferedUsers{
				Username:  rf.Username.String,
				CreatedAt: rf.CreatedAt,
				Amount:    userIDToPoint[rf.ID],
			})
		}

	}

	resp.Amount = amount
	resp.Users = refferedUsers

	return resp, true, nil
}

func (u *user) GetCurrentReferralMultiplier(ctx context.Context) (int, error) {
	currentMultiplier, err := u.db.GetConfigByName(ctx, constant.CONFIG_POINT_MULTIPLIER)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return 0, err
	}

	multiplier, err := strconv.Atoi(currentMultiplier.Value)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return 0, err
	}
	return multiplier, nil
}

func (u *user) UpdateReferralMultiplier(ctx context.Context, newValue int) (dto.ReferalData, error) {
	_, err := u.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Name:  constant.CONFIG_POINT_MULTIPLIER,
		Value: strconv.Itoa(newValue),
	})

	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.ReferalData{}, err
	}

	return dto.ReferalData{
		ID:              uuid.Nil, // Since we're not returning the ID from the update
		Name:            constant.CONFIG_POINT_MULTIPLIER,
		PointMultiplier: newValue,
	}, nil
}

func (u *user) GetAdminAssignedPoints(ctx context.Context, limit, offset int) (dto.GetAdminAssignedResp, bool, error) {
	var resp dto.GetAdminAssignedResp
	var respdata []dto.GetAdminAssignedData

	logs, err := u.db.Queries.GetAddminAssignedPoints(ctx, db.GetAddminAssignedPointsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetAdminAssignedResp{}, false, err
	}

	if err != nil {
		return dto.GetAdminAssignedResp{}, false, nil
	}

	for _, log := range logs {
		usrIDString := strings.Split(log.Description.String, "_")[3]
		userIDParsed, err := uuid.Parse(usrIDString)
		if err != nil {
			u.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.GetAdminAssignedResp{}, false, err
		}

		respdata = append(respdata, dto.GetAdminAssignedData{

			UserID:            log.UserID.UUID,
			AddedPoints:       int(log.ChangeAmount.Decimal.IntPart()),
			Timestamp:         log.Timestamp.Time,
			PointsAfterUpdate: int(log.BalanceAfterUpdate.IntPart()),
			AdminID:           userIDParsed,
			TransactionID:     log.TransactionID,
		})

		// total pages
		resp.TotalPage = int(math.Ceil(float64(log.Total) / float64(limit)))

	}
	resp.Data = respdata
	return resp, true, nil
}

func (u *user) CreateUserPoint(ctx context.Context, userID uuid.UUID, points decimal.Decimal) (dto.UserPoint, error) {
	resp, err := u.db.Queries.CreateBalance(ctx, db.CreateBalanceParams{
		UserID:     userID,
		Currency:   constant.POINT_CURRENCY,
		RealMoney:  decimal.NullDecimal{Decimal: points, Valid: true},
		BonusMoney: decimal.NullDecimal{Decimal: decimal.Zero, Valid: true},
		UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.UserPoint{}, err
	}
	return dto.UserPoint{
		UserID: userID,
		Point:  int(resp.RealMoney.Decimal.IntPart()),
	}, nil
}
