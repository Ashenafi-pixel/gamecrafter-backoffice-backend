package user

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (u *user) BlockAccount(ctx context.Context, account dto.AccountBlockReq) (dto.AccountBlockReq, error) {
	//validate nullable fields
	blockedFrom := utils.NullTime(account.BlockedFrom)
	blockedTo := utils.NullTime(account.BlockedTo)
	note := utils.NullString(account.Note)

	blockedAcc, err := u.db.BlockAccount(ctx, db.BlockAccountParams{
		UserID:      account.UserID,
		BlockedBy:   account.BlockedBy,
		Type:        account.Type,
		Reason:      sql.NullString{String: account.Reason, Valid: true},
		BlockedFrom: blockedFrom,
		BlockedTo:   blockedTo,
		Note:        note,
		CreatedAt:   sql.NullTime{Time: time.Now().In(time.Now().Location()).UTC(), Valid: true},
		Duration:    account.Duration,
	})
	if err != nil {
		u.log.Error(err.Error(), zap.Any("account", account))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.AccountBlockReq{}, err
	}

	return dto.AccountBlockReq{
		ID:          blockedAcc.ID,
		UserID:      blockedAcc.UserID,
		BlockedBy:   blockedAcc.BlockedBy,
		BlockedFrom: &blockedAcc.BlockedFrom.Time,
		BlockedTo:   &blockedAcc.BlockedTo.Time,
		Type:        blockedAcc.Type,
		Duration:    blockedAcc.Duration,
		CreatedAt:   time.Now(),
	}, nil
}

func (u *user) GetBlockedAccountByType(ctx context.Context, userID uuid.UUID, tp, duration string) (dto.AccountBlockReq, bool, error) {
	bacc, err := u.db.Queries.GetPermamentlyBlockedAccountByUserIdAndType(ctx, db.GetPermamentlyBlockedAccountByUserIdAndTypeParams{
		UserID:   userID,
		Type:     tp,
		Duration: duration,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.AccountBlockReq{}, false, err
	}
	if err != nil {
		return dto.AccountBlockReq{}, false, nil
	}

	return dto.AccountBlockReq{
		ID:          bacc.ID,
		UserID:      bacc.UserID,
		BlockedBy:   bacc.BlockedBy,
		BlockedFrom: &bacc.BlockedFrom.Time,
		BlockedTo:   &bacc.BlockedTo.Time,
		Duration:    bacc.Duration,
		Type:        bacc.Type,
		Reason:      bacc.Reason.String,
		Note:        bacc.Note.String,
		CreatedAt:   bacc.CreatedAt.Time,
	}, true, nil
}

func (u *user) GetBlockedAccountByUserID(ctx context.Context, userID uuid.UUID) ([]dto.AccountBlockReq, bool, error) {
	blockedAccs := []dto.AccountBlockReq{}
	baccs, err := u.db.Queries.GetAccountBlockByUserID(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("userID", userID.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.AccountBlockReq{}, false, err
	}

	if err != nil {
		return []dto.AccountBlockReq{}, false, nil
	}

	for _, bacc := range baccs {
		blockedAccs = append(blockedAccs, dto.AccountBlockReq{
			ID:          bacc.ID,
			UserID:      bacc.UserID,
			BlockedBy:   bacc.BlockedBy,
			BlockedFrom: &bacc.BlockedFrom.Time,
			BlockedTo:   &bacc.BlockedTo.Time,
			Duration:    bacc.Duration,
			Type:        bacc.Type,
			Reason:      bacc.Reason.String,
			Note:        bacc.Note.String,
			CreatedAt:   bacc.CreatedAt.Time,
		})
	}
	return blockedAccs, true, nil
}

func (u *user) AaccountUnlock(ctx context.Context, ID uuid.UUID) (dto.AccountBlockReq, error) {
	bacc, err := u.db.Queries.UnlockAccount(ctx, ID)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("ID", ID.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.AccountBlockReq{}, err
	}
	return dto.AccountBlockReq{
		ID:          bacc.ID,
		UserID:      bacc.UserID,
		BlockedBy:   bacc.BlockedBy,
		BlockedFrom: &bacc.BlockedFrom.Time,
		BlockedTo:   &bacc.BlockedTo.Time,
		Duration:    bacc.Duration,
		Type:        bacc.Type,
		Reason:      bacc.Reason.String,
		Note:        bacc.Note.String,
		CreatedAt:   bacc.CreatedAt.Time,
	}, err
}

func (u *user) GetBlockedAllAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAllAccount(ctx, db.GetBlockedAllAccountParams{
		Limit:  int32(getBlockedAcReq.PerPage),
		Offset: int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		ps := float64(bac.Total / int64(getBlockedAcReq.PerPage))
		total := int(math.Ceil(ps))
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
			Total_pages: total,
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByTypeAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAllAccountByType(ctx, db.GetBlockedAllAccountByTypeParams{
		Type:   getBlockedAcReq.Type,
		Limit:  int32(getBlockedAcReq.PerPage),
		Offset: int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByDurationAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAllAccountByDuration(ctx, db.GetBlockedAllAccountByDurationParams{
		Duration: getBlockedAcReq.Duration,
		Limit:    int32(getBlockedAcReq.PerPage),
		Offset:   int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByDurationAndTypeAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAllAccountByTypeAndDuration(ctx, db.GetBlockedAllAccountByTypeAndDurationParams{
		Duration: getBlockedAcReq.Duration,
		Type:     getBlockedAcReq.Type,
		Limit:    int32(getBlockedAcReq.PerPage),
		Offset:   int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByDurationAndTypeAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAllAccountByTypeAndDurationAndUserID(ctx, db.GetBlockedAllAccountByTypeAndDurationAndUserIDParams{
		Duration: getBlockedAcReq.Duration,
		Type:     getBlockedAcReq.Type,
		UserID:   getBlockedAcReq.UserID,
		Limit:    int32(getBlockedAcReq.PerPage),
		Offset:   int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockedAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByDurationAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAccountByUserIDAndDuration(ctx, db.GetBlockedAccountByUserIDAndDurationParams{
		Duration: getBlockedAcReq.Duration,
		UserID:   getBlockedAcReq.UserID,
		Limit:    int32(getBlockedAcReq.PerPage),
		Offset:   int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByTypeAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAccountByUserIDAndType(ctx, db.GetBlockedAccountByUserIDAndTypeParams{
		Type:   getBlockedAcReq.Type,
		UserID: getBlockedAcReq.UserID,
		Limit:  int32(getBlockedAcReq.PerPage),
		Offset: int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) GetBlockedByUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error) {
	var resp []dto.GetBlockedAccountLogRep
	bacs, err := u.db.Queries.GetBlockedAccountByUserID(ctx, db.GetBlockedAccountByUserIDParams{
		UserID: getBlockedAcReq.UserID,
		Limit:  int32(getBlockedAcReq.PerPage),
		Offset: int32(getBlockedAcReq.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("req", getBlockedAcReq))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, false, err
	}
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, false, nil
	}

	for _, bac := range bacs {
		resp = append(resp, dto.GetBlockedAccountLogRep{
			BlockedAccount: dto.AccountBlockReq{
				ID:          bac.ID,
				UserID:      bac.UserID,
				BlockedBy:   bac.BlockedBy,
				BlockedFrom: &bac.BlockedFrom.Time,
				BlockedTo:   &bac.BlockedTo.Time,
				Duration:    bac.Duration,
				Type:        bac.Type,
				Reason:      bac.Reason.String,
				Note:        bac.Note.String,
				CreatedAt:   bac.CreatedAt.Time,
			},
			User: dto.User{
				ID:          bac.BlockdeAccountUserID,
				PhoneNumber: bac.BlockedAccountUserPhone.String,
				FirstName:   bac.BlockedAccountUserFirstName.String,
				LastName:    bac.BlockedAccountUserLastName.String,
				Email:       bac.BlockedAccountUserEmail.String,
			},
			BlockedBy: dto.User{
				ID:          bac.BlockerAccountUserID,
				PhoneNumber: bac.BlockerAccountUserPhone.String,
				FirstName:   bac.BlockerAccountUserFirstName.String,
				LastName:    bac.BlockerAccountUserLastName.String,
				Email:       bac.BlockerAccountUserEmail.String,
			},
		})
	}

	return resp, true, nil

}

func (u *user) AddIpFilter(ctx context.Context, ipFilter dto.IpFilterReq) (dto.IPFilterRes, error) {
	filter, err := u.db.Queries.CreateIPFilter(ctx, db.CreateIPFilterParams{
		CreatedBy:   ipFilter.CreatedBy,
		StartIp:     ipFilter.StartIP,
		EndIp:       ipFilter.EndIP,
		Type:        ipFilter.Type,
		Description: ipFilter.Description,
		CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		u.log.Error(err.Error(), zap.Any("ipFilter", ipFilter))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.IPFilterRes{}, err
	}

	return dto.IPFilterRes{
		Data: dto.IPFilter{
			ID:          filter.ID,
			StartIP:     filter.StartIp,
			EndIP:       filter.EndIp,
			CreatedBy:   filter.CreatedBy,
			Description: filter.Description,
			Type:        filter.Type,
		},
	}, nil
}

func (u *user) GetIPFilterByType(ctx context.Context, tp string) ([]dto.IPFilter, bool, error) {
	var ipRes []dto.IPFilter
	res, err := u.db.Queries.GetIpFilterByType(ctx, tp)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error(), zap.Any("type", tp))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.IPFilter{}, false, err
	}
	if err != nil {
		return []dto.IPFilter{}, false, nil
	}
	for _, r := range res {
		ipRes = append(ipRes, dto.IPFilter{
			ID:          r.ID,
			StartIP:     r.StartIp,
			EndIP:       r.EndIp,
			CreatedBy:   r.CreatedBy,
			Description: r.Description,
			Type:        r.Type,
			Hits:        int(r.Hits),
			LastHit:     r.LastHit.Time,
		})
	}
	return ipRes, true, nil
}

func (u *user) GetIpFilterByTypeWithLimitAndOffset(ctx context.Context, ipFilter dto.GetIPFilterReq) (dto.GetIPFilterRes, bool, error) {
	var totalPage int64
	ipFilerRes := dto.GetIPFilterRes{}
	res, err := u.db.Queries.GetIpFilterByTypeWithLimitAndOffset(ctx, db.GetIpFilterByTypeWithLimitAndOffsetParams{
		Type:   ipFilter.Type,
		Limit:  int32(ipFilter.PerPage),
		Offset: int32(ipFilter.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetIPFilterRes{}, false, err
	}
	if err != nil {
		return dto.GetIPFilterRes{}, false, nil
	}
	for _, ipFiler := range res {
		ipFilerRes.IPFilters = append(ipFilerRes.IPFilters, dto.GetIPFilterResData{
			ID:          ipFiler.ID,
			StartIP:     ipFiler.StartIp,
			EndIP:       ipFiler.EndIp,
			Description: ipFiler.Description,
			Hists:       int(ipFiler.Hits),
			Type:        ipFiler.Type,
			LastHit:     ipFiler.LastHit.Time,
			CreatedBy: dto.User{
				ID:        ipFiler.UserID,
				FirstName: ipFiler.FirstName.String,
				LastName:  ipFiler.LastName.String,
				Email:     ipFiler.Email.String,
			},
		})
		totalPage = ipFiler.Total
	}

	ps := float64(totalPage / int64(ipFilter.PerPage))
	total := int(math.Ceil(ps))
	ipFilerRes.TotalPages = total
	return ipFilerRes, true, nil

}

func (u *user) GetAllIpFilterWithLimitAndOffset(ctx context.Context, ipFilter dto.GetIPFilterReq) (dto.GetIPFilterRes, bool, error) {
	var totalPage int64
	ipFilerRes := dto.GetIPFilterRes{}
	res, err := u.db.Queries.GetAllIpFilterWithLimitAndOffset(ctx, db.GetAllIpFilterWithLimitAndOffsetParams{
		Limit:  int32(ipFilter.PerPage),
		Offset: int32(ipFilter.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetIPFilterRes{}, false, err
	}
	if err != nil {
		return dto.GetIPFilterRes{}, false, nil
	}
	for _, ipFiler := range res {
		ipFilerRes.IPFilters = append(ipFilerRes.IPFilters, dto.GetIPFilterResData{
			ID:          ipFiler.ID,
			StartIP:     ipFiler.StartIp,
			Type:        ipFiler.Type,
			EndIP:       ipFiler.EndIp,
			Description: ipFiler.Description,
			Hists:       int(ipFiler.Hits),
			LastHit:     ipFiler.LastHit.Time,
			CreatedBy: dto.User{
				ID:        ipFiler.UserID,
				FirstName: ipFiler.FirstName.String,
				LastName:  ipFiler.LastName.String,
				Email:     ipFiler.Email.String,
			},
		})
		totalPage = ipFiler.Total
	}

	ps := float64(totalPage / int64(ipFilter.PerPage))
	total := int(math.Ceil(ps))
	if total == 0 {
		total = total + 1
	}
	ipFilerRes.TotalPages = total

	return ipFilerRes, true, nil

}

func (u *user) RemoveIPFilters(ctx context.Context, id uuid.UUID) (dto.RemoveIPBlockRes, error) {
	err := u.db.Queries.RemoveAccountBlock(ctx, id)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.RemoveIPBlockRes{}, err
	}
	return dto.RemoveIPBlockRes{
		Message: constant.SUCCESS,
	}, nil
}

func (u *user) GetIPFilterByID(ctx context.Context, id uuid.UUID) (dto.IPFilter, bool, error) {
	resp, err := u.db.Queries.GetIpFiltersByID(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		u.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.IPFilter{}, false, err
	}

	if err != nil {
		return dto.IPFilter{}, false, nil
	}

	return dto.IPFilter{ID: resp.ID, StartIP: resp.StartIp, EndIP: resp.EndIp, CreatedBy: resp.CreatedBy, Type: resp.Type, Description: resp.Description, Hits: int(resp.Hits), LastHit: resp.LastHit.Time}, true, nil
}

func (u *user) UpdateIpFilter(ctx context.Context, req dto.IPFilter) (dto.IPFilter, error) {
	resp, err := u.db.Queries.UpdateIPfilter(ctx, db.UpdateIPfilterParams{
		Description: req.Description,
		Hits:        int32(req.Hits),
		LastHit:     sql.NullTime{Time: req.LastHit, Valid: true},
		ID:          req.ID,
	})

	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.IPFilter{}, err
	}
	return dto.IPFilter{
		ID:          resp.ID,
		StartIP:     resp.StartIp,
		EndIP:       resp.EndIp,
		Description: resp.Description,
		CreatedBy:   resp.CreatedBy,
		Hits:        int(resp.Hits),
		LastHit:     resp.LastHit.Time,
		Type:        resp.Type,
	}, nil
}
