package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"go.uber.org/zap"
)

func (u *User) BlockUser(ctx context.Context, blockAcc dto.AccountBlockReq) (dto.AccountBlockRes, error) {
	// validate user input
	if blockAcc.Duration != constant.BLOCK_DURATION_PERMANENT && blockAcc.Duration != constant.BLOCK_DURATION_TEMPORARY {
		err := fmt.Errorf("invalid block duration ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}
	if blockAcc.Type != constant.BLOCK_TYPE_FINANCIAL &&
		blockAcc.Type != constant.BLOCK_TYPE_GAMING &&
		blockAcc.Type != constant.BLOCK_TYPE_LOGIN &&
		blockAcc.Type != constant.BLOCK_TYPE_COMPLETE {
		err := fmt.Errorf("invalid block type ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}
	if blockAcc.UserID == uuid.Nil {
		err := fmt.Errorf("invalid user id ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}
	//get user
	blockedUser, exist, err := u.userStorage.GetUserByID(ctx, blockAcc.UserID)
	if err != nil {
		return dto.AccountBlockRes{}, err
	}
	if !exist {
		err = fmt.Errorf("unable to find user")
		u.log.Error(err.Error(), zap.Any("userID", blockAcc.UserID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}
	blockederUser, exist, err := u.userStorage.GetUserByID(ctx, blockAcc.BlockedBy)
	if err != nil {
		return dto.AccountBlockRes{}, err
	}
	if !exist {
		err = fmt.Errorf("unable to find user")
		u.log.Error(err.Error(), zap.Any("userID", blockAcc.BlockedBy.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}

	if blockAcc.Type == constant.BLOCK_DURATION_TEMPORARY && (blockAcc.BlockedFrom == nil || blockAcc.BlockedTo == nil) {
		//check the user is already blocked or not
		err = fmt.Errorf("please provide blocked_from and blocked_to for temporary duration")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}
	//check for if the user is already block with this type of not

	_, exist, err = u.userStorage.GetBlockedAccountByType(ctx, blockAcc.UserID, blockAcc.Type, constant.BLOCK_DURATION_PERMANENT)
	if err != nil {
		return dto.AccountBlockRes{}, err
	}

	if exist {
		err = fmt.Errorf("user already blocked")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AccountBlockRes{}, err
	}

	if blockAcc.Duration != constant.BLOCK_DURATION_PERMANENT {
		// check for from to account going to be blocked
		if blockAcc.BlockedFrom == nil || blockAcc.BlockedTo == nil {
			err = fmt.Errorf("please provide blocked_from and blocked_to  to block user temporarily ")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.AccountBlockRes{}, err
		}
		tempBlock, exist, err := u.userStorage.GetBlockedAccountByType(ctx, blockAcc.UserID, blockAcc.Type, blockAcc.Duration)
		if err != nil {
			return dto.AccountBlockRes{}, err
		}

		if exist {
			//check if it is expired or not
			if tempBlock.BlockedTo != nil {
				tempbHolder := *tempBlock.BlockedTo
				if tempbHolder.Before(time.Now().In(time.Now().Location()).UTC()) {
					// then unblock the previous then lock it with new lock
					_, err = u.userStorage.AaccountUnlock(ctx, tempBlock.ID)
					if err != nil {
						return dto.AccountBlockRes{}, err
					}
				} else {
					err = fmt.Errorf("user already blocked")
					err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
					return dto.AccountBlockRes{}, err
				}
			}
		}
	}

	blockedAcc, err := u.userStorage.BlockAccount(ctx, blockAcc)
	if err != nil {
		return dto.AccountBlockRes{}, err
	}

	//notify departments about the blocked user
	usrsToNoTify, _ := u.userStorage.GetUsersByDepartmentNotificationTypes(ctx, []string{blockAcc.Type})
	u.NotifyUsers(ctx, dto.NotifyDepartmentsReq{
		BlockerUser:   blockedUser,
		BlockedUser:   blockederUser,
		BlockReq:      blockAcc,
		UsersToNotify: usrsToNoTify,
	})
	return dto.AccountBlockRes{
		UserID:   blockedAcc.ID,
		Message:  constant.BLOCK_USER_SUCCESS,
		Type:     blockedAcc.Type,
		Duration: blockAcc.Duration,
	}, nil
}

func (u *User) GetBlockedAccount(ctx context.Context, blockAccountReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, error) {

	if blockAccountReq.PerPage <= 0 || blockAccountReq.Page <= 0 {
		err := fmt.Errorf("please provide page and per_page")
		u.log.Warn(err.Error(), zap.Any("get_blocked_req", blockAccountReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return []dto.GetBlockedAccountLogRep{}, err
	}
	// page and per_page logic here
	offset := (blockAccountReq.Page - 1) * blockAccountReq.PerPage
	blockAccountReq.Page = offset

	if blockAccountReq.Duration != "" && blockAccountReq.Type != "" && blockAccountReq.UserID != uuid.Nil {
		res, _, err := u.userStorage.GetBlockedByDurationAndTypeAndUserIDAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.Type != "" && blockAccountReq.UserID != uuid.Nil {
		res, _, err := u.userStorage.GetBlockedByTypeAndUserIDAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.Duration != "" && blockAccountReq.UserID != uuid.Nil {
		res, _, err := u.userStorage.GetBlockedByDurationAndUserIDAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.Duration != "" && blockAccountReq.Type != "" {
		res, _, err := u.userStorage.GetBlockedByDurationAndTypeAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.Duration != "" {
		res, _, err := u.userStorage.GetBlockedByDurationAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.Type != "" {
		res, _, err := u.userStorage.GetBlockedByTypeAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	if blockAccountReq.UserID != uuid.Nil {
		res, _, err := u.userStorage.GetBlockedByUserIDAccount(ctx, blockAccountReq)
		if err != nil {
			return []dto.GetBlockedAccountLogRep{}, err
		}
		return res, nil
	}

	res, _, err := u.userStorage.GetBlockedAllAccount(ctx, blockAccountReq)
	if err != nil {
		return []dto.GetBlockedAccountLogRep{}, err
	}
	return res, nil
}

func (u *User) NotifyUsers(ctx context.Context, notification dto.NotifyDepartmentsReq) {
	blockedNotificationTemplate := constant.GetBlockedUserString(notification)
	for _, usr := range notification.UsersToNotify {
		if err := utils.SendEmail(ctx, dto.EmailReq{
			Subject: constant.FORGOT_PASSWORD_SUBJECT,
			To:      []string{usr.Email},
			Body:    []byte(blockedNotificationTemplate),
		}); err != nil {
			u.log.Error(err.Error())
		}
	}
}

func (u *User) AddIpFilter(ctx context.Context, ipFilter dto.IpFilterReq) (dto.IPFilterRes, error) {
	//validate filter
	if ok := utils.ValidateIP(ipFilter.StartIP); !ok {
		err := fmt.Errorf("invalid start_ip is given")
		u.log.Warn(err.Error(), zap.Any("ipFilterReq", ipFilter))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.IPFilterRes{}, err
	}

	if ipFilter.EndIP != "" && !utils.ValidateIP(ipFilter.EndIP) {
		err := fmt.Errorf("invalid end_ip is given")
		u.log.Warn(err.Error(), zap.Any("ipFilterReq", ipFilter))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.IPFilterRes{}, err
	}

	if ipFilter.Type != constant.IP_FILTER_TYPE_ALLOW && ipFilter.Type != constant.IP_FILTER_TYPE_DENY {
		err := fmt.Errorf("invalid type is given only allow or deny is allowed")
		u.log.Warn(err.Error(), zap.Any("ipFilterReq", ipFilter))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.IPFilterRes{}, err
	}

	//validate blockeder identity
	_, exist, err := u.userStorage.GetUserByID(ctx, ipFilter.CreatedBy)
	if err != nil {
		return dto.IPFilterRes{}, err
	}

	if !exist {
		err := fmt.Errorf("admin user not found")
		u.log.Error(err.Error(), zap.Any("ipFilterReq", ipFilter))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.IPFilterRes{}, err
	}

	ipFilterRes, err := u.userStorage.AddIpFilter(ctx, ipFilter)
	if err != nil {
		return dto.IPFilterRes{}, err
	}
	u.UpdateIpFilterMap(ctx)
	ipFilterRes.Message = constant.SUCCESS
	return ipFilterRes, nil
}

func (u *User) UpdateIpFilterMap(ctx context.Context) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	//get ip list for  allow
	allow, _, err := u.userStorage.GetIPFilterByType(ctx, constant.IP_FILTER_TYPE_ALLOW)
	if err != nil {
		return err
	}
	deny, _, err := u.userStorage.GetIPFilterByType(ctx, constant.IP_FILTER_TYPE_DENY)
	if err != nil {
		return err
	}
	u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW] = allow
	u.IpFilterMap[constant.IP_FILTER_TYPE_DENY] = deny

	return nil
}
func (u *User) EnforceIPFilerRule(ctx context.Context, ip string) (bool, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// If no IP filters exist, allow by default
	if len(u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW]) == 0 && len(u.IpFilterMap[constant.IP_FILTER_TYPE_DENY]) == 0 {
		return true, nil
	}

	// Deny list only
	if len(u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW]) == 0 && len(u.IpFilterMap[constant.IP_FILTER_TYPE_DENY]) > 0 {
		for i, denyList := range u.IpFilterMap[constant.IP_FILTER_TYPE_DENY] {
			if len(strings.TrimSpace(denyList.EndIP)) < 5 {
				// Single IP check
				if denyList.StartIP == ip {
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].LastHit = time.Now()
					return false, nil
				}
			} else {
				// IP range check
				if utils.CheckIPInRanage(ip, denyList.StartIP, denyList.EndIP) {
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].LastHit = time.Now()
					return false, nil
				}
			}
		}
		return true, nil
	}

	// Allow list only
	if len(u.IpFilterMap[constant.IP_FILTER_TYPE_DENY]) == 0 && len(u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW]) > 0 {
		for i, allowList := range u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW] {
			if len(strings.TrimSpace(allowList.EndIP)) < 5 {
				// Single IP check
				if allowList.StartIP == ip {
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].Hits++
					return true, nil
				}
			} else {
				// IP range check
				if utils.CheckIPInRanage(ip, allowList.StartIP, allowList.EndIP) {
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].Hits++
					return true, nil
				}
			}
		}
		return false, nil
	}

	// Mixed allow and deny lists
	if len(u.IpFilterMap[constant.IP_FILTER_TYPE_DENY]) > 0 && len(u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW]) > 0 {
		existInAllowedList := false
		// Check allow list first
		for i, allowList := range u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW] {
			if len(strings.TrimSpace(allowList.EndIP)) < 5 {
				if allowList.StartIP == ip {
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].LastHit = time.Now()
					existInAllowedList = true
					break
				}
			} else {
				if utils.CheckIPInRanage(ip, allowList.StartIP, allowList.EndIP) {
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_ALLOW][i].LastHit = time.Now()
					existInAllowedList = true
					break
				}
			}
		}

		// If not in allow list, deny
		if !existInAllowedList {
			return false, nil
		}

		// Check deny list
		for i, denyList := range u.IpFilterMap[constant.IP_FILTER_TYPE_DENY] {
			if len(strings.TrimSpace(denyList.EndIP)) < 5 {
				if denyList.StartIP == ip {
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].LastHit = time.Now()
					return false, nil
				}
			} else {
				if utils.CheckIPInRanage(ip, denyList.StartIP, denyList.EndIP) {
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].Hits++
					u.IpFilterMap[constant.IP_FILTER_TYPE_DENY][i].LastHit = time.Now()
					return false, nil
				}
			}
		}
		return true, nil
	}

	return true, nil
}

func (u *User) GetIPFilters(ctx context.Context, getIPFilterReq dto.GetIPFilterReq) (dto.GetIPFilterRes, error) {
	if getIPFilterReq.PerPage <= 0 {
		getIPFilterReq.PerPage = 10
	}
	if getIPFilterReq.Page <= 0 {
		getIPFilterReq.Page = 1
	}
	offset := (getIPFilterReq.Page - 1) * getIPFilterReq.PerPage
	getIPFilterReq.Page = offset

	if getIPFilterReq.Type != "" {
		if getIPFilterReq.Type != constant.IP_FILTER_TYPE_ALLOW && getIPFilterReq.Type != constant.IP_FILTER_TYPE_DENY {
			err := fmt.Errorf("invalid type is given")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.GetIPFilterRes{}, err
		}

		resp, _, err := u.userStorage.GetIpFilterByTypeWithLimitAndOffset(ctx, getIPFilterReq)
		if err != nil {
			return dto.GetIPFilterRes{}, err
		}

		return resp, nil
	}
	resp, _, err := u.userStorage.GetAllIpFilterWithLimitAndOffset(ctx, getIPFilterReq)
	if err != nil {
		return dto.GetIPFilterRes{}, err
	}

	return resp, nil
}

func (u *User) RemoveIPFilter(ctx context.Context, req dto.RemoveIPBlockReq) (dto.RemoveIPBlockRes, error) {
	if req.ID == uuid.Nil {
		err := fmt.Errorf("invalid uuid is given")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.RemoveIPBlockRes{}, err
	}

	// get ip filter by id
	_, exist, err := u.userStorage.GetIPFilterByID(ctx, req.ID)
	if err != nil {
		return dto.RemoveIPBlockRes{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to find ip filter with this id  %s", req.ID.String())
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.RemoveIPBlockRes{}, err
	}

	// remove from ip filters
	resp, err := u.userStorage.RemoveIPFilters(ctx, req.ID)
	if err != nil {
		return dto.RemoveIPBlockRes{}, err
	}
	u.UpdateIpFilterMap(ctx)
	return resp, nil
}

func (u *User) updateipfilterDatabase() {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	for _, bl := range u.IpFilterMap {
		for _, b := range bl {
			u.userStorage.UpdateIpFilter(context.Background(), b)
		}
	}
}
