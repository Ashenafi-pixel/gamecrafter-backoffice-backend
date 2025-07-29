package airtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type airtime struct {
	log                         *zap.Logger
	airtimeProviderEndpoint     string
	timeout                     time.Duration
	accessToken                 string
	vaspid                      int
	cron                        *cron.Cron
	password                    string
	airtimeStorage              storage.Airtime
	balanceStorage              storage.Balance
	balanceLogStorage           storage.BalanceLogs
	userStorage                 storage.User
	operationalGroupStorage     storage.OperationalGroup
	operationalGroupTypeStorage storage.OperationalGroupType
}

func Init(log *zap.Logger,
	airtimeProviderEndpoint,
	password string,
	timeout time.Duration,
	vaspid int,
	airtimeStorage storage.Airtime,
	balanceStorage storage.Balance,
	balanceLogs storage.BalanceLogs,
	userStorage storage.User,
	operationalGroupStorage storage.OperationalGroup,
	operationalGroupTypeStorage storage.OperationalGroupType,

) module.AirtimeProvider {

	a := &airtime{
		log:                         log,
		airtimeProviderEndpoint:     airtimeProviderEndpoint,
		timeout:                     timeout,
		vaspid:                      vaspid,
		password:                    password,
		airtimeStorage:              airtimeStorage,
		balanceStorage:              balanceStorage,
		balanceLogStorage:           balanceLogs,
		userStorage:                 userStorage,
		operationalGroupStorage:     operationalGroupStorage,
		operationalGroupTypeStorage: operationalGroupTypeStorage,
	}

	err := a.initCronJob()
	a.Login()
	if err != nil {
		log.Fatal(err.Error())
	}
	return a
}

func (a *airtime) initCronJob() error {
	a.cron = cron.New(cron.WithSeconds())

	_, err := a.cron.AddFunc("@every 86400s", a.Login)
	if err != nil {
		return err
	}
	a.cron.Start()
	return nil
}

func (a *airtime) Login() {
	req := dto.AirtimeLoginReq{
		Vaspid:   a.vaspid,
		Password: a.password,
	}
	var parsedResp dto.AirtimeLoginResp
	endpoint := a.airtimeProviderEndpoint + constant.AIRTIME_LOGIN_ENDPOINT
	resp, err := utils.SendPostHttpRequest(endpoint, req, nil, a.timeout)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return
	}

	byteData, err := json.Marshal(resp)

	if err != nil {
		err := fmt.Errorf("unable to marshal response data")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return
	}

	if err := json.Unmarshal(byteData, &parsedResp); err != nil {
		err := fmt.Errorf("unable to parse the response to dto.AirtimeLoginResp")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return
	}

	if parsedResp.StatusCode != constant.AIRTIME_SUCCESS_STATUS_CODE {
		err := fmt.Errorf("unable to signin  , message %s ", parsedResp.Message)
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return
	}

	a.accessToken = parsedResp.PisiAuthorizationToken
}

func (a *airtime) RefereshUtilies(ctx context.Context) ([]dto.AirtimeUtility, error) {
	var resp interface{}
	existingUtilitiesHandler := make(map[int]bool)
	var airtimeRespHandler dto.AirtimeUtilitiesResp
	var respUtilities []dto.AirtimeUtility
	var err error
	header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
	endpoint := a.airtimeProviderEndpoint + constant.AIRTIME_GET_UTILITIES_ENDPOINT
	resp, err = utils.SendGetHttpRequest(ctx, endpoint, header, a.timeout)

	if err != nil {
		// try to login and refresh again
		a.Login()
		header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
		resp, err = utils.SendGetHttpRequest(ctx, endpoint, header, a.timeout)
		if err != nil {
			a.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return []dto.AirtimeUtility{}, err
		}
	}

	// convert response to valid data types to handle
	byteData, err := json.Marshal(resp)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return []dto.AirtimeUtility{}, err
	}

	if err := json.Unmarshal(byteData, &airtimeRespHandler); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return []dto.AirtimeUtility{}, err
	}

	// get all utilities
	utilities, exist, err := a.airtimeStorage.GetAllAirtimeUtilities(ctx)
	if err != nil {
		return []dto.AirtimeUtility{}, err
	}

	if !exist {
		// save as new data
		for _, utility := range airtimeRespHandler.Data {
			utility.Status = constant.INACTIVE
			utility.Timestamp = time.Now()
			r, err := a.airtimeStorage.CreateUtility(ctx, utility)
			if err != nil {
				return []dto.AirtimeUtility{}, err
			}
			respUtilities = append(respUtilities, r)
		}

	} else {
		// check if the utilities exist or not
		for _, u := range utilities {
			respUtilities = append(respUtilities, u)
			existingUtilitiesHandler[u.ID] = true
		}

		for _, u := range airtimeRespHandler.Data {
			if _, ok := existingUtilitiesHandler[u.ID]; !ok {
				// save
				resp, err := a.airtimeStorage.CreateUtility(ctx, u)
				if err != nil {
					return []dto.AirtimeUtility{}, err
				}
				respUtilities = append(respUtilities, resp)
			}
		}

	}

	return respUtilities, nil
}

func (a *airtime) GetAvailableAirtimeUtilities(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	resp, exist, err := a.airtimeStorage.GetAvailableAirtime(ctx, req)
	if err != nil {
		return dto.GetAirtimeUtilitiesResp{}, err
	}

	if !exist {
		resp2, err := a.RefereshUtilies(ctx)
		if err != nil {
			return dto.GetAirtimeUtilitiesResp{}, err
		}

		if len(resp2) > 0 {
			return a.GetAvailableAirtimeUtilities(ctx, req)
		}

		return dto.GetAirtimeUtilitiesResp{}, nil
	}

	return resp, nil
}

func (a *airtime) UpdateAirtimeStatus(ctx context.Context, req dto.UpdateAirtimeStatusReq) (dto.UpdateAirtimeStatusResp, error) {
	if req.Status != constant.AIRTIME_STATUS_ACTIVE && req.Status != constant.AIRTIME_STATUS_INACTIVE {
		err := fmt.Errorf("only  %s and %s status are available ", constant.AIRTIME_STATUS_ACTIVE, constant.AIRTIME_STATUS_INACTIVE)
		err = errors.ErrInactiveUserStatus.Wrap(err, err.Error())
		return dto.UpdateAirtimeStatusResp{}, err
	}

	_, exist, err := a.airtimeStorage.GetAirtimeUtilityByLocalID(ctx, req.LocalID)
	if err != nil {
		return dto.UpdateAirtimeStatusResp{}, err
	}

	if !exist {
		// update the value  the list
		err := fmt.Errorf("unable to find utility using localID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateAirtimeStatusResp{}, err
	}

	updatedResp, err := a.airtimeStorage.UpdateAirtimeStatus(ctx, req.LocalID, req.Status)
	if err != nil {
		return dto.UpdateAirtimeStatusResp{}, err
	}

	return dto.UpdateAirtimeStatusResp{
		Message: constant.SUCCESS,
		Data:    updatedResp.Data,
	}, nil
}

func (a *airtime) UpdateUtilityPrice(ctx context.Context, req dto.UpdateAirtimeUtilityPriceReq) (dto.UpdateAirtimeUtilityPriceRes, error) {
	// check utility price not less than zero
	if req.Price.LessThan(decimal.Zero) {
		err := fmt.Errorf("price can not be less than zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateAirtimeUtilityPriceRes{}, err
	}

	// update price
	_, exist, err := a.airtimeStorage.GetAirtimeUtilityByLocalID(ctx, req.LocalID)
	if err != nil {
		return dto.UpdateAirtimeUtilityPriceRes{}, err
	}

	if !exist {
		err := fmt.Errorf("airtime utitility not found with this local id")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateAirtimeUtilityPriceRes{}, err
	}

	resp, err := a.airtimeStorage.UpdateAirtimeUtilityPrice(ctx, req)
	if err != nil {
		return dto.UpdateAirtimeUtilityPriceRes{}, err
	}

	return dto.UpdateAirtimeUtilityPriceRes{
		Message: constant.SUCCESS,
		Data:    resp,
	}, nil
}

func (a *airtime) ClaimPoints(ctx context.Context, req dto.ClaimPointsReq) (dto.ClaimPointsResp, error) {
	// get airtime offser price
	resp, exist, err := a.airtimeStorage.GetAirtimeUtilityByLocalID(ctx, req.AirtimeLocalID)
	if err != nil {
		return dto.ClaimPointsResp{}, nil
	}

	if !exist {
		err := fmt.Errorf("airtime utitility not found with this local id")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	if resp.Status != constant.ACTIVE {
		err := fmt.Errorf("airtime utitility not active")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	// get user balance
	balance, exist, err := a.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:   req.UserID,
		Currency: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.ClaimPointsResp{}, err
	}

	if !exist {
		err := fmt.Errorf("insuficent balance")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	// check user balance
	if balance.RealMoney.LessThan(resp.Price) {
		err := fmt.Errorf("insuficent balance")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	// get customer phone number
	usr, exist, err := a.userStorage.GetUserByID(ctx, req.UserID)
	if err != nil {
		return dto.ClaimPointsResp{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to get user")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	if strings.TrimSpace(usr.PhoneNumber) == "" {
		err := fmt.Errorf("unable to get customer's phone number")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	// updat balance
	//update user balance
	transactionID := uuid.New().String()
	newBalance := balance.RealMoney.Sub(resp.Price)
	_, err = a.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.ClaimPointsResp{}, err
	}

	// save transaction logs
	operationalGroupAndTypeIDs, err := a.CreateOrGetOperationalGroupAndType(ctx, constant.CASHOUT, constant.WITHDRAWAL)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		a.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return dto.ClaimPointsResp{}, err
	}

	// save operations logs
	_, err = a.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("airtime cashout %s, new balance is %s and  currency %s", resp.Price.String(), balance.RealMoney.Sub(resp.Price).String(), constant.POINT_CURRENCY),
		ChangeAmount:       resp.Price,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.ClaimPointsResp{}, err
	}

	// claim point
	amount, err := strconv.Atoi(resp.Amount)
	if err != nil {

		a.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})

		err := fmt.Errorf("unable to get amount")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimPointsResp{}, err
	}

	parsedPhoneNumber := strings.Split(usr.PhoneNumber, "+")
	phone := parsedPhoneNumber[len(parsedPhoneNumber)-1]
	d, err := a.PayAndValidateTransaction(ctx, dto.ClaimAirtimeReq{
		Msisdn:               phone,
		CustomerId:           phone,
		UtilityPackageId:     resp.ID,
		TransactionReference: transactionID,
		Amount:               amount,
	})

	if err != nil {
		a.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return dto.ClaimPointsResp{}, err
	}

	// save transaction
	a.airtimeStorage.SaveAirtimeTransactions(ctx, dto.AirtimeTransactions{
		UserID:           req.UserID,
		TransactionID:    transactionID,
		Cashout:          resp.Price,
		BillerName:       d.Data.BillerName,
		UtilityPackageId: d.Data.UtilityPackageId,
		PackageName:      d.Data.BillerName,
		Amount:           decimal.NewFromInt(int64(d.Data.Amount)),
		Status:           constant.SUCCESS,
	})

	return dto.ClaimPointsResp{
		Message: constant.SUCCESS,
		Data: dto.ClaimPointsData{
			TransactionID: transactionID,
			Data: dto.ClaimAirtimeRespData{
				ResultCode: d.ResultCode,
				Data: dto.ClaimAirtimeData{
					UtilityPackageId:     d.Data.UtilityPackageId,
					Amount:               d.Data.Amount,
					TransactionStatus:    d.Data.TransactionStatus,
					TransactionReference: d.Data.TransactionReference,
					BillerName:           d.Data.BillerName,
					PackageName:          d.Data.PackageName,
				},
			},
		},
	}, nil
}

func (a *airtime) PayAndValidateTransaction(ctx context.Context, req dto.ClaimAirtimeReq) (dto.ClaimAirtimeRespData, error) {
	var resp interface{}
	var err error
	var parsedResp dto.ClaimAirtimeRespFromProvider
	endpoint := a.airtimeProviderEndpoint + constant.AIRTIME_PAY_UTILITY
	header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
	resp, err = utils.SendPostHttpRequest(endpoint, req, header, a.timeout)
	if err != nil {
		// login agian and try
		a.Login()
		header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
		resp, err = utils.SendPostHttpRequest(endpoint, req, header, a.timeout)
		if err != nil {
			err := fmt.Errorf("This airtime package cannot be provided for the number  %s", req.CustomerId)
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.ClaimAirtimeRespData{}, err

		}
	}

	byteData, err := json.Marshal(resp)

	if err != nil {
		a.log.Error(err.Error())
		err := fmt.Errorf("unable to marshal response data")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimAirtimeRespData{}, err
	}

	if err := json.Unmarshal(byteData, &parsedResp); err != nil {
		err := fmt.Errorf("unable to parse the response to dto.AirtimeLoginResp")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimAirtimeRespData{}, err
	}

	if parsedResp.ResultCode != 200 {
		err := fmt.Errorf("unable to signin  , message %s ", parsedResp.Description)
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimAirtimeRespData{}, err
	}

	testCounter := 3

	for {
		if testCounter < 1 {
			break
		}
		testCounter--
		result, err := a.CheckAirtimeTransactions(ctx, req.TransactionReference)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		} else {
			// confirm update
			return result, nil
		}
	}
	err = fmt.Errorf("unable to update response ")
	err = errors.ErrInternalServerError.Wrap(err, err.Error())
	return dto.ClaimAirtimeRespData{}, err
}

func (a *airtime) CheckAirtimeTransactions(ctx context.Context, transactionID string) (dto.ClaimAirtimeRespData, error) {
	var resp interface{}
	var claimResponseHandler dto.ClaimAirtimeRespData

	var err error
	header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
	endpoint := a.airtimeProviderEndpoint + constant.AIRTIME_CHECK_TRANSACTION_STATUS + transactionID
	resp, err = utils.SendGetHttpRequest(ctx, endpoint, header, a.timeout)

	if err != nil {
		// try to login and refresh again
		a.Login()
		header := map[string]string{"pisi-authorization-token": "Bearer " + a.accessToken, "vaspid": strconv.Itoa(a.vaspid)}
		resp, err = utils.SendGetHttpRequest(ctx, endpoint, header, a.timeout)
		if err != nil {
			a.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.ClaimAirtimeRespData{}, err
		}
	}

	// convert response to valid data types to handle
	byteData, err := json.Marshal(resp)
	if err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimAirtimeRespData{}, err
	}

	if err := json.Unmarshal(byteData, &claimResponseHandler); err != nil {
		a.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ClaimAirtimeRespData{}, err
	}

	return claimResponseHandler, nil
}

func (a *airtime) CreateOrGetOperationalGroupAndType(ctx context.Context, operationalGroupName, operationalType string) (dto.OperationalGroupAndType, error) {
	// get transfer operational  group and  type if not exist create group transfer and type transfer-internal
	var operationalGroup dto.OperationalGroup
	var exist bool
	var err error
	var operationalGroupTypeID dto.OperationalGroupType
	operationalGroup, exist, err = a.operationalGroupStorage.GetOperationalGroupByName(ctx, constant.TRANSFER)
	if err != nil {
		return dto.OperationalGroupAndType{}, err
	}
	if !exist {
		// create transfer internal group and type
		operationalGroup, err = a.operationalGroupStorage.CreateOperationalGroup(ctx, dto.OperationalGroup{
			Name:        constant.TRANSFER,
			Description: "internal transaction",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}

		// create operation group type
		operationalGroupTypeID, err = a.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
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
		operationalGroupTypeID, exist, err = a.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, dto.OperationalGroupType{
			GroupID: operationalGroup.ID,
			Name:    operationalType,
		})
		if err != nil {
			return dto.OperationalGroupAndType{}, err
		}
		if !exist {
			operationalGroupTypeID, err = a.operationalGroupTypeStorage.CreateOperationalType(ctx, dto.OperationalGroupType{
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

func (a *airtime) GetActiveAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return a.airtimeStorage.GetActiveAvailableAirtime(ctx, req)
}

func (a *airtime) GetUserAirtimeTransactions(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetAirtimeTransactionsResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return a.airtimeStorage.GetUserAirtimeTransactions(ctx, req, userID)
}

func (a *airtime) GetAllAirtimeUtilitiesTransactions(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeTransactionsResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return a.airtimeStorage.GetAllAirtimeUtilitiesTransactions(ctx, req)
}

func (a *airtime) UpdateAirtimeAmount(ctx context.Context, req dto.UpdateAirtimeAmountReq) (dto.AirtimeUtility, error) {
	if req.Amount.LessThan(decimal.Zero) {
		err := fmt.Errorf("amount can not be less than zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AirtimeUtility{}, err
	}

	return a.airtimeStorage.UpdateAirtimeAmount(ctx, req)
}

func (a *airtime) GetAirtimeUtilitiesStats(ctx context.Context) (dto.AirtimeUtilitiesStats, error) {
	return a.airtimeStorage.GetAirtimeUtilitiesStats(ctx)
}
