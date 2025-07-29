package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type user struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.User {
	return &user{
		db:  db,
		log: log,
	}
}

func (u *user) CreateUser(ctx context.Context, userRequest dto.User) (dto.User, error) {
	//check if phone is  null or not
	var phone sql.NullString
	var email sql.NullString
	var createdBy uuid.NullUUID
	if userRequest.CreatedBy != uuid.Nil {
		createdBy = uuid.NullUUID{UUID: userRequest.CreatedBy, Valid: true}
	}

	phone = sql.NullString{Valid: false}
	if len(userRequest.PhoneNumber) > 3 {
		phone = sql.NullString{String: userRequest.PhoneNumber, Valid: true}
	}
	email = sql.NullString{Valid: false}
	if len(userRequest.Email) > 3 {
		email = sql.NullString{String: userRequest.Email, Valid: true}
	}
	usr, err := u.db.Queries.CreateUser(ctx, db.CreateUserParams{
		PhoneNumber:     phone,
		Email:           email,
		Password:        userRequest.Password,
		FirstName:       sql.NullString{String: userRequest.FirstName, Valid: true},
		LastName:        sql.NullString{String: userRequest.LastName, Valid: true},
		DefaultCurrency: sql.NullString{String: userRequest.DefaultCurrency, Valid: true},
		Source:          sql.NullString{String: userRequest.Source, Valid: true},
		ReferalCode:     sql.NullString{String: userRequest.ReferralCode, Valid: true},
		DateOfBirth:     sql.NullString{String: userRequest.DateOfBirth, Valid: true},
		IsAdmin:         sql.NullBool{Bool: userRequest.IsAdmin, Valid: true},
		CreatedBy:       createdBy,
		Status:          sql.NullString{String: userRequest.Status, Valid: userRequest.Status != ""},
		UserType:        sql.NullString{String: string(userRequest.Type), Valid: userRequest.Type != ""},
		ReferedByCode:   sql.NullString{String: string(userRequest.ReferedByCode), Valid: userRequest.ReferedByCode != ""},
		ReferalType:     sql.NullString{String: string(userRequest.ReferalType), Valid: userRequest.ReferalType != ""},
	})
	if err != nil {
		u.log.Error("unable to create user ", zap.Error(err), zap.Any("user", userRequest))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create user ")
		return dto.User{}, err
	}

	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Password:        usr.Password,
		DefaultCurrency: usr.DefaultCurrency.String,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		Email:           usr.Email.String,
		ProfilePicture:  usr.Profile.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		ReferalType:     dto.Type(usr.ReferalType.String),
		ReferedByCode:   usr.ReferedByCode.String,
		Type:            dto.Type(usr.UserType.String),
	}, nil
}

func (u *user) GetUserByUserName(ctx context.Context, username string) (dto.User, bool, error) {
	usr, err := u.db.Queries.GetUserByUserName(ctx, sql.NullString{String: username, Valid: true})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error("unable to make get query using username")
		err = errors.ErrUnableToGet.Wrap(err, "unable to get user using username")
		return dto.User{}, false, err
	}
	if err != nil && err.Error() == dto.ErrNoRows {
		return dto.User{}, false, nil
	}
	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Password:        usr.Password,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		Email:           usr.Email.String,
		ProfilePicture:  usr.Profile.String,
		DefaultCurrency: usr.DefaultCurrency.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		StreetAddress:   usr.StreetAddress,
		Country:         usr.Country,
		State:           usr.State,
		City:            usr.City,
		PostalCode:      usr.PostalCode,
		KYCStatus:       usr.KycStatus,
		IsAdmin:         usr.IsAdmin.Bool,
		ReferalType:     dto.Type(usr.ReferalType.String),
		ReferedByCode:   usr.ReferedByCode.String,
		Type:            dto.Type(usr.UserType.String),
	}, true, nil
}

func (u *user) GetUserByPhoneNumber(ctx context.Context, phone string) (dto.User, bool, error) {

	usr, err := u.db.Queries.GetUserByPhone(ctx, sql.NullString{String: phone, Valid: true})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error("unable to get using phone number", zap.Any("phone", phone))
		err = errors.ErrInternalServerError.Wrap(err, "unable to get user using phone number")
		return dto.User{}, false, err
	}
	if err != nil && err.Error() == dto.ErrNoRows {
		return dto.User{}, false, nil
	}
	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Password:        usr.Password,
		Email:           usr.Email.String,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		DefaultCurrency: usr.DefaultCurrency.String,
		ProfilePicture:  usr.Profile.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		StreetAddress:   usr.StreetAddress,
		Country:         usr.Country,
		State:           usr.State,
		City:            usr.City,
		PostalCode:      usr.PostalCode,
		KYCStatus:       usr.KycStatus,
		IsAdmin:         usr.IsAdmin.Bool,
		ReferedByCode:   usr.ReferedByCode.String,
		ReferalType:     dto.Type(usr.ReferalType.String),
		Type:            dto.Type(usr.UserType.String),
		Status:          usr.Status.String,
	}, true, nil
}
func (u *user) GetUserByID(ctx context.Context, userID uuid.UUID) (dto.User, bool, error) {
	usr, err := u.db.Queries.GetUserByID(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.User{}, false, err
	}
	if err != nil {
		return dto.User{}, false, nil
	}
	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Email:           usr.Email.String,
		DefaultCurrency: usr.DefaultCurrency.String,
		ProfilePicture:  usr.Profile.String,
		Password:        usr.Password,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		StreetAddress:   usr.StreetAddress,
		Country:         usr.Country,
		State:           usr.State,
		City:            usr.City,
		PostalCode:      usr.PostalCode,
		KYCStatus:       usr.KycStatus,
		IsAdmin:         usr.IsAdmin.Bool,
		ReferedByCode:   usr.ReferedByCode.String,
		ReferalType:     dto.Type(usr.ReferalType.String),
		Type:            dto.Type(usr.UserType.String),
	}, true, nil
}

func (u *user) UpdateProfilePicuter(ctx context.Context, userID uuid.UUID, filename string) (string, error) {
	updatedUser, err := u.db.UpdateProfilePicuter(ctx, db.UpdateProfilePicuterParams{
		ID:      userID,
		Profile: sql.NullString{String: filename, Valid: true},
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("userID", userID), zap.Any("filename", filename))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return "", err
	}
	return updatedUser.Profile.String, nil
}

func (u *user) UpdatePassword(ctx context.Context, UserID uuid.UUID, newPassword string) (dto.User, error) {
	usr, err := u.db.Queries.UpdatePassword(ctx, db.UpdatePasswordParams{
		Password: newPassword,
		ID:       UserID,
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("userID", UserID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.User{}, err
	}
	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Email:           usr.Email.String,
		Password:        usr.Password,
		ProfilePicture:  usr.Profile.String,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		DefaultCurrency: usr.DefaultCurrency.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		StreetAddress:   usr.StreetAddress,
		Country:         usr.Country,
		State:           usr.State,
		City:            usr.City,
		PostalCode:      usr.PostalCode,
		KYCStatus:       usr.KycStatus,
		IsAdmin:         usr.IsAdmin.Bool,
		ReferedByCode:   usr.ReferedByCode.String,
		ReferalType:     dto.Type(usr.ReferalType.String),
		Type:            dto.Type(usr.UserType.String),
	}, nil
}

func (u *user) GetUserByEmail(ctx context.Context, email string) (dto.User, bool, error) {
	usr, err := u.db.Queries.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("email", email))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.User{}, false, err
	}
	if err != nil {
		return dto.User{}, false, nil
	}

	return dto.User{
		ID:              usr.ID,
		PhoneNumber:     usr.PhoneNumber.String,
		Email:           email,
		Password:        usr.Password,
		ProfilePicture:  usr.Profile.String,
		DefaultCurrency: usr.DefaultCurrency.String,
		FirstName:       usr.FirstName.String,
		LastName:        usr.LastName.String,
		DateOfBirth:     usr.DateOfBirth.String,
		Source:          usr.Source.String,
		ReferralCode:    usr.ReferalCode.String,
		StreetAddress:   usr.StreetAddress,
		Country:         usr.Country,
		State:           usr.State,
		City:            usr.City,
		PostalCode:      usr.PostalCode,
		KYCStatus:       usr.KycStatus,
		IsAdmin:         usr.IsAdmin.Bool,
		ReferedByCode:   usr.ReferedByCode.String,
		ReferalType:     dto.Type(usr.ReferalType.String),
		Type:            dto.Type(usr.UserType.String),
	}, true, nil
}

func (u *user) SaveUserOTP(ctx context.Context, otpReq dto.ForgetPasswordOTPReq) error {
	err := u.db.Queries.SaveOTP(ctx, db.SaveOTPParams{
		UserID:    otpReq.UserID,
		Otp:       otpReq.OTP,
		CreatedAt: time.Now().In(time.Now().Location()).UTC(),
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("otpReq", otpReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (u *user) GetUserOTP(ctx context.Context, userID uuid.UUID) (dto.OTPHolder, bool, error) {
	otp, err := u.db.Queries.GetOTP(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.OTPHolder{}, false, err
	}
	if err != nil {
		return dto.OTPHolder{}, false, nil
	}

	return dto.OTPHolder{
		TmpOTP:    otp.Otp,
		CreatedAT: otp.CreatedAt,
		Attempts:  0,
	}, true, nil
}

func (u *user) DeleteOTP(ctx context.Context, userID uuid.UUID) error {
	err := u.db.Queries.DeleteOTP(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("user_id", userID))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (u *user) UpdateUser(ctx context.Context, updateProfile dto.UpdateProfileReq) (dto.User, error) {
	updatedUser, err := u.db.Queries.UpdateProfile(ctx, db.UpdateProfileParams{
		FirstName:     sql.NullString{String: updateProfile.FirstName, Valid: true},
		LastName:      sql.NullString{String: updateProfile.LastName, Valid: true},
		Email:         sql.NullString{String: updateProfile.Email, Valid: true},
		DateOfBirth:   sql.NullString{String: updateProfile.DateOfBirth, Valid: true},
		PhoneNumber:   sql.NullString{String: updateProfile.Phone, Valid: true},
		ID:            updateProfile.UserID,
		Username:      sql.NullString{String: updateProfile.Username, Valid: true},
		StreetAddress: updateProfile.StreetAddress,
		City:          updateProfile.City,
		PostalCode:    updateProfile.PostalCode,
		State:         updateProfile.State,
		Country:       updateProfile.Country,
		KycStatus:     updateProfile.KYCStatus,
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("updateRequest", updateProfile))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.User{}, err
	}
	return dto.User{
		ID:              updateProfile.UserID,
		PhoneNumber:     updatedUser.PhoneNumber.String,
		FirstName:       updateProfile.FirstName,
		LastName:        updateProfile.LastName,
		Email:           updateProfile.Email,
		DefaultCurrency: updatedUser.DefaultCurrency.String,
		ProfilePicture:  updatedUser.Profile.String,
		DateOfBirth:     updateProfile.DateOfBirth,
		ReferralCode:    updatedUser.ReferalCode.String,
		ReferedByCode:   updatedUser.ReferedByCode.String,
		ReferalType:     dto.Type(updatedUser.ReferalType.String),
		Type:            dto.Type(updatedUser.UserType.String),
	}, nil
}

func (u *user) GetAllUsers(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error) {
	var users []dto.User
	usrs, err := u.db.Queries.GetAllUsers(ctx, db.GetAllUsersParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetPlayersRes{}, err
	}
	totalPages := 1
	totalCount := int64(0)
	for i, usr := range usrs {
		users = append(users, dto.User{
			ID:              usr.ID,
			PhoneNumber:     usr.PhoneNumber.String,
			FirstName:       usr.FirstName.String,
			LastName:        usr.LastName.String,
			Email:           usr.Email.String,
			DefaultCurrency: usr.DefaultCurrency.String,
			ProfilePicture:  usr.Profile.String,
			DateOfBirth:     usr.DateOfBirth.String,
			Source:          usr.Source.String,
			ReferralCode:    usr.ReferalCode.String,
			StreetAddress:   usr.StreetAddress,
			Country:         usr.Country,
			State:           usr.State,
			City:            usr.City,
			PostalCode:      usr.PostalCode,
			KYCStatus:       usr.KycStatus,
			IsAdmin:         usr.IsAdmin.Bool,
			ReferedByCode:   usr.ReferedByCode.String,
			ReferalType:     dto.Type(usr.ReferalType.String),
			Type:            dto.Type(usr.UserType.String),
		})

		if i == 0 {
			totalPages = int(int(usr.TotalRows) / req.PerPage)
			if int(usr.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
			totalCount = usr.TotalRows
		}
	}
	return dto.GetPlayersRes{
		Message:    constant.SUCCESS,
		Users:      users,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}, nil
}

func (u *user) GetUserPoints(ctx context.Context, userID uuid.UUID) (decimal.Decimal, bool, error) {
	blc, err := u.db.Queries.GetUserBalanaceByUserIDAndCurrency(ctx, db.GetUserBalanaceByUserIDAndCurrencyParams{
		UserID:   userID,
		Currency: constant.POINT_CURRENCY,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error("unable to make get balance request using user_id")
		err = errors.ErrUnableToGet.Wrap(err, "unable to make get balance request using user_id")
		return decimal.Zero, false, err

	} else if err != nil && err.Error() == dto.ErrNoRows {
		return decimal.Zero, false, nil
	}

	return blc.RealMoney.Decimal, true, nil
}

func (u *user) UpdateUserPoints(ctx context.Context, userID uuid.UUID, points decimal.Decimal) (decimal.Decimal, error) {
	resp, err := u.db.Queries.UpdateRealMoney(ctx, db.UpdateRealMoneyParams{
		RealMoney: decimal.NullDecimal{Decimal: points, Valid: true},
		UserID:    userID,
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Currency:  constant.POINT_CURRENCY,
	})
	if err != nil {
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
	}
	return resp.RealMoney.Decimal, nil
}

func (u *user) GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	var admins []dto.Admin

	adminResp, err := u.db.Queries.GetAdmins(ctx, db.GetAdminsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	for _, admin := range adminResp {

		var adminRoles []dto.AdminRoleRes
		if admin.Roles.Status == pgtype.Present {
			err := json.Unmarshal(admin.Roles.Bytes, &adminRoles)
			if err != nil {
				u.log.Error("Failed to unmarshal roles", zap.Error(err), zap.Any("raw", string(admin.Roles.Bytes)))
				continue
			}
		}

		admins = append(admins, dto.Admin{
			ID:          admin.UserID,
			PhoneNumber: admin.PhoneNumber.String,
			FirstName:   admin.FirstName.String,
			LastName:    admin.LastName.String,
			Email:       admin.Email.String,
			Status:      admin.Status.String,
			Roles:       adminRoles,
		})
	}
	return admins, nil
}

func (u *user) GetAdminsByRole(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	var admins []dto.Admin
	adminResp, err := u.db.Queries.GetAdminsByRole(ctx, db.GetAdminsByRoleParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		RoleID: req.RoleID,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	for _, admin := range adminResp {
		admins = append(admins, dto.Admin{
			ID:          admin.ID,
			PhoneNumber: admin.PhoneNumber.String,
			FirstName:   admin.FirstName.String,
			LastName:    admin.LastName.String,
			Email:       admin.Email.String,
			Status:      admin.Status.String,
		})
	}
	return admins, nil
}

func (u *user) GetAdminsByStatus(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	var admins []dto.Admin

	adminResp, err := u.db.Queries.GetAdminsByStatus(ctx, db.GetAdminsByStatusParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		Status: sql.NullString{String: req.Status, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	for _, admin := range adminResp {

		var adminRoles []dto.AdminRoleRes
		if admin.Roles.Status == pgtype.Present {
			err := json.Unmarshal(admin.Roles.Bytes, &adminRoles)
			if err != nil {
				u.log.Error("Failed to unmarshal roles", zap.Error(err), zap.Any("raw", string(admin.Roles.Bytes)))
				continue
			}
		}

		admins = append(admins, dto.Admin{
			ID:          admin.UserID,
			PhoneNumber: admin.PhoneNumber.String,
			FirstName:   admin.FirstName.String,
			LastName:    admin.LastName.String,
			Email:       admin.Email.String,
			Status:      admin.Status.String,
			Roles:       adminRoles,
		})
	}
	return admins, nil
}

func (u *user) GetAdminsByRoleAndStatus(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	var admins []dto.Admin
	adminResp, err := u.db.Queries.GetAdminsByRoleAndStatus(ctx, db.GetAdminsByRoleAndStatusParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		RoleID: req.RoleID,
		Status: sql.NullString{String: req.Status, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	for _, admin := range adminResp {
		admins = append(admins, dto.Admin{
			ID:          admin.ID,
			PhoneNumber: admin.PhoneNumber.String,
			FirstName:   admin.FirstName.String,
			LastName:    admin.LastName.String,
			Email:       admin.Email.String,
			Status:      admin.Status.String,
		})
	}
	return admins, nil
}

func (u *user) GetUserByReferalCode(ctx context.Context, code string) (*dto.UserProfile, error) {
	user, err := u.db.GetUserByReferalCode(ctx, sql.NullString{
		String: code,
		Valid:  code != "",
	})

	return &dto.UserProfile{
		UserID:       user.ID,
		Email:        user.Email.String,
		PhoneNumber:  user.Email.String,
		ReferralCode: user.ReferalCode.String,
	}, err
}

func (u *user) GetUsersByEmailAndPhone(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error) {
	var users []dto.User
	userResp, err := u.db.Queries.GetUserEmailOrPhoneNumber(ctx, db.GetUserEmailOrPhoneNumberParams{
		Column1: sql.NullString{String: req.Filter.Phone, Valid: req.Filter.Phone != ""},
		Column2: sql.NullString{String: req.Filter.Email, Valid: req.Filter.Email != ""},
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("email", req.Filter.Email), zap.Any("phone", req.Filter.Phone))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetPlayersRes{}, err
	}
	totalPages := 1
	totalCount := int64(0)
	for i, user := range userResp {
		// Get user balance
		balance, err := u.db.Queries.GetUserBalancesByUserID(ctx, user.ID)
		if err != nil && err.Error() != dto.ErrNoRows {
			u.log.Error("unable to get user balance", zap.Error(err), zap.Any("userID", user.ID))
			err = errors.ErrUnableToGet.Wrap(err, "unable to get user balance")
			return dto.GetPlayersRes{}, err
		}

		var accounts []dto.Balance
		for _, bal := range balance {
			accounts = append(accounts, dto.Balance{
				ID:         bal.ID,
				Currency:   bal.Currency,
				RealMoney:  bal.RealMoney.Decimal,
				BonusMoney: bal.BonusMoney.Decimal,
			})
		}

		users = append(users, dto.User{
			ID:              user.ID,
			PhoneNumber:     user.PhoneNumber.String,
			Email:           user.Email.String,
			DefaultCurrency: user.DefaultCurrency.String,
			ProfilePicture:  user.Profile.String,
			FirstName:       user.FirstName.String,
			LastName:        user.LastName.String,
			DateOfBirth:     user.DateOfBirth.String,
			Source:          user.Source.String,
			ReferralCode:    user.ReferalCode.String,
			Type:            dto.Type(user.UserType.String),
			IsAdmin:         user.IsAdmin.Bool,
			ReferedByCode:   user.ReferedByCode.String,
			ReferalType:     dto.Type(user.ReferalType.String),
			Accounts:        accounts,
		})
		if i == 0 {
			totalPages = int(int(user.TotalRows) / req.PerPage)
			if int(user.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
			totalCount = user.TotalRows
		}
	}
	return dto.GetPlayersRes{
		TotalCount: totalCount,
		Message:    constant.SUCCESS,
		TotalPages: totalPages,
		Users:      users,
	}, nil
}

func (u *user) SaveToTemp(ctx context.Context, req dto.UserReferals) error {
	_, err := u.db.Queries.CreateTemp(ctx, db.CreateTempParams{
		UserID: req.UserID,
		Data:   pgtype.JSONB{Bytes: req.Data, Status: pgtype.Present},
	})

	if err != nil {
		u.log.Error("unable to save to temp", zap.Error(err), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to save to temp")
		return err
	}

	return nil
}

func (u *user) GetTempData(ctx context.Context, userID uuid.UUID) (dto.GetUserReferals, error) {
	var referals dto.ReferralData

	tempData, err := u.db.Queries.GetTempByUserID(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error("unable to get temp data", zap.Error(err), zap.Any("userID", userID))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get temp data")
		return dto.GetUserReferals{}, err
	}

	if err != nil {
		return dto.GetUserReferals{}, nil
	}

	if tempData.Data.Status != pgtype.Present {
		u.log.Error("temp data is not present", zap.Any("userID", userID))
		return dto.GetUserReferals{}, nil
	}

	err = json.Unmarshal(tempData.Data.Bytes, &referals)
	if err != nil {
		u.log.Error("unable to unmarshal temp data", zap.Error(err), zap.Any("data", string(tempData.Data.Bytes)))
		err = errors.ErrInternalServerError.Wrap(err, "unable to unmarshal temp data")
		return dto.GetUserReferals{}, err
	}

	return dto.GetUserReferals{
		ID:           tempData.ID,
		UserID:       tempData.UserID,
		ReferralData: referals,
		CreatedAt:    tempData.CreatedAt.Time,
	}, nil
}

func (u *user) DeleteTempData(ctx context.Context, ID uuid.UUID) error {
	err := u.db.Queries.DeleteTempByID(ctx, ID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error("unable to delete temp data", zap.Error(err), zap.Any("ID", ID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to delete temp data")
		return err
	}
	return nil
}
