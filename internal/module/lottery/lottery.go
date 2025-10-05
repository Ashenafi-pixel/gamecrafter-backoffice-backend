package lottery

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform"
	httpclient "github.com/tucanbit/platform/http_client"
	"github.com/tucanbit/platform/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type lottery struct {
	log                         *zap.Logger
	lotteryStorage              storage.Lottery
	kafkaAgent                  platform.Kafka
	balanceLogStorage           storage.BalanceLogs
	balanceStorage              storage.Balance
	operationalGroupStorage     storage.OperationalGroup
	operationalGroupTypeStorage storage.OperationalGroupType
	lotteryServiceToken         string
	loggeryTokenLocker          *sync.Mutex
}

func Init(lotteryStorage storage.Lottery,
	log *zap.Logger,
	kafkaAgent platform.Kafka,
	balanceLogStorage storage.BalanceLogs,
	balanceStorage storage.Balance,
	operationalGroupStorage storage.OperationalGroup,
	operationalGroupTypeStorage storage.OperationalGroupType,
) module.Lottery {
	lottery := lottery{
		log:                         log,
		lotteryStorage:              lotteryStorage,
		kafkaAgent:                  kafkaAgent,
		balanceLogStorage:           balanceLogStorage,
		balanceStorage:              balanceStorage,
		operationalGroupStorage:     operationalGroupStorage,
		operationalGroupTypeStorage: operationalGroupTypeStorage,
		loggeryTokenLocker:          &sync.Mutex{},
	}
	lottery.kafkaAgent.RegisterKafkaEventHandler(constant.LOTTERY_DRAW_SYNC, lottery.HandleLotteryEvent)
	return &lottery
}

func (l *lottery) CreateLotteryService(ctx context.Context, req dto.CreateLotteryServiceReq) (dto.CreateLotteryServiceRes, error) {
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		l.log.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return dto.CreateLotteryServiceRes{}, err
	}

	avialbleService, _ := l.lotteryStorage.GetAvailableLotteryService(ctx)
	if avialbleService.ID != uuid.Nil {
		l.log.Error("lottery service already exists", zap.Error(errors.ErrInvalidUserInput.New("lottery service already exists")))
		return dto.CreateLotteryServiceRes{}, errors.ErrInvalidUserInput.New("lottery service already exists")
	}

	hashPassword, err := utils.Encrypt(req.ServiceSecret)
	if err != nil {
		l.log.Error("error encrypting password", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, "error encrypting password")
		return dto.CreateLotteryServiceRes{}, err
	}

	req.ServiceSecret = hashPassword
	resp, err := l.lotteryStorage.CreateLotteryService(ctx, req)
	if err != nil {
		return dto.CreateLotteryServiceRes{}, err
	}

	return resp, err
}

func (l *lottery) CreateLotteryRequest(ctx context.Context, req dto.LotteryRequestCreate) (dto.LotteryRequestCreate, error) {
	validate := validator.New()
	var response dto.LotteryRequestCreate
	err := validate.Struct(req)
	if err != nil {
		l.log.Error("error validating request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error validating request")
		return dto.LotteryRequestCreate{}, err
	}

	lotteryService, err := l.lotteryStorage.GetAvailableLotteryService(ctx)
	if err != nil {
		l.log.Error("error getting lottery service", zap.Error(err))
		return dto.LotteryRequestCreate{}, errors.ErrInternalServerError.Wrap(err, "error getting lottery service by ID")
	}

	if l.lotteryServiceToken == "" {
		if err := l.UpdateLotteryToken(ctx); err != nil {
			l.log.Error("error updating lottery token", zap.Error(err))
			return dto.LotteryRequestCreate{}, errors.ErrInternalServerError.Wrap(err, "error updating lottery token")
		}
	}

	header := map[string]string{
		"Authorization": "Bearer " + l.lotteryServiceToken,
	}

	if err := httpclient.SendPostHttpRequest(lotteryService.CallbackURL+"/api/lottery", req, &response, header, time.Minute); err != nil {
		// update token and retry
		if err := l.UpdateLotteryToken(ctx); err != nil {
			l.log.Error("error updating lottery token", zap.Error(err))
			return dto.LotteryRequestCreate{}, errors.ErrInternalServerError.Wrap(err, "error updating lottery token")
		}
		if err := httpclient.SendPostHttpRequest(lotteryService.CallbackURL+"/api/lottery", req, &response, header, time.Minute); err != nil {
			l.log.Error("error sending post HTTP request", zap.Error(err))
			return dto.LotteryRequestCreate{}, errors.ErrInternalServerError.Wrap(err, "error sending post HTTP request")
		}
	}

	return response, nil
}

func (l *lottery) HandleLotteryEvent(ctx context.Context, message []byte) (bool, error) {

	// Unmarshal the message into a LotteryEvent struct
	var event dto.KafkaLotteryEvent
	if err := json.Unmarshal(message, &event); err != nil {
		l.log.Error("failed to unmarshal lottery event", zap.Error(err))
		return false, errors.ErrInvalidUserInput.Wrap(err, "failed to unmarshal lottery event")
	}

	// check if the event has readed before or not
	logs, err := l.lotteryStorage.GetLotteryLogsByUniqIdentifier(ctx, event.UniqueID)
	if err != nil {
		l.log.Error("failed to get lottery logs by uniq identifier", zap.Error(err))
		return false, errors.ErrInternalServerError.Wrap(err, "failed to get lottery logs by uniq identifier")
	}

	if len(logs) > 0 {
		l.log.Info("lottery event already processed", zap.String("unique_id", event.UniqueID.String()))
		return false, nil
	}

	if len(event.Winners) == 0 {
		l.log.Warn("no winners found in lottery event", zap.String("lottery_id", event.LotteryID.String()))
		// just update the draw balls
		for _, reward := range event.Rewards {
			_, err := l.lotteryStorage.CreateLotteryLog(ctx, dto.LotteryKafkaLog{
				LotteryID:       event.LotteryID,
				LotteryRewardID: reward.ID,
				DrawNumbers:     reward.DrawedNumbers,
				Prize:           reward.Prize,
				UniqIdentifier:  event.UniqueID,
			})

			if err != nil {
				l.log.Error("failed to create lottery log", zap.Error(err))
				return false, errors.ErrInternalServerError.Wrap(err, "failed to create lottery log")
			}
		}

		return false, nil
	} else {
		return l.RewardWinnersAndSaveLogs(ctx, event)
	}

}

func (l *lottery) CheckUserBalanceAndDeductBalance(ctx context.Context, req dto.LotteryVerifyAndDeductBalanceReq) (dto.LotteryVerifyAndDeductBalanceRes, error) {
	balance, exist, err := l.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: req.Currency,
	})

	if err != nil {
		l.log.Error("error getting user balance", zap.Error(err), zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInternalServerError.Wrap(err, "error getting user balance")
	}

	if !exist {
		l.log.Warn("user balance does not exist", zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInvalidUserInput.Wrap(nil, "user balance does not exist")
	}

	if balance.RealMoney.LessThanOrEqual(decimal.Zero) {
		l.log.Warn("user balance is zero", zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInvalidUserInput.Wrap(nil, "user balance is zero")
	}

	if balance.RealMoney.LessThan(req.Amount) {
		l.log.Warn("user balance is less than required amount", zap.Any("user_id", req.UserID), zap.Any("required_amount", req.Amount))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInvalidUserInput.Wrap(nil, "user balance is less than required amount")
	}
	newBalance := balance.RealMoney.Sub(req.Amount)
	transactionID := utils.GenerateTransactionId()
	_, err = l.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  req.Currency,
		Component: req.Component,
		Amount:    newBalance,
	})

	if err != nil {
		l.log.Error("error updating user balance", zap.Error(err), zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInternalServerError.Wrap(err, "error updating user balance")
	}

	// Save balance log
	operationalGroupAndTypeIDs, err := l.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_LOTTERY_BUY)
	if err != nil {
		l.log.Error("failed to create or get operational group and type", zap.Error(err), zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInternalServerError.Wrap(err, "failed to create or get operational group and type")
	}
	// save operations logs
	_, err = l.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:    req.UserID,
		Component: constant.REAL_MONEY,
		Currency:  req.Currency,
		Description: "lottery buy, new balance is " + newBalance.String() +
			" and currency " + req.Currency,
		ChangeAmount:       req.Amount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		l.log.Error("failed to save balance log", zap.Error(err), zap.Any("user_id", req.UserID))
		return dto.LotteryVerifyAndDeductBalanceRes{}, errors.ErrInternalServerError.Wrap(err, "failed to save balance log")
	}
	return dto.LotteryVerifyAndDeductBalanceRes{
		UserID:   req.UserID,
		Currency: req.Currency,
		Amount:   req.Amount,
	}, nil
}

func (l *lottery) UpdateLotteryToken(ctx context.Context) error {
	l.loggeryTokenLocker.Lock()
	defer l.loggeryTokenLocker.Unlock()
	var response dto.LotteryServiceLoginRes
	lotteryService, err := l.lotteryStorage.GetAvailableLotteryService(ctx)
	if err != nil {
		l.log.Error("error getting lottery servicea", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "error getting lottery service by ID")
	}

	decryptedLotteryServiceSecret, err := utils.Decrypt(lotteryService.ServiceSecret)
	if err != nil {
		l.log.Error("error decrypting lottery service secret", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "error decrypting lottery service secret")
	}

	loginReq := dto.LotteryServiceLoginReq{
		ClientID:     lotteryService.ServiceClientID,
		ClientSecret: decryptedLotteryServiceSecret,
	}

	if err := httpclient.SendPostHttpRequest(lotteryService.CallbackURL+"/api/service/auth", loginReq, &response, nil, time.Minute); err != nil {
		l.log.Error("error sending post HTTP request", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "error sending post HTTP request")
	}
	token := response.Token
	if token == "" {
		l.log.Error("received empty token from lottery service", zap.String("service_name", lotteryService.ServiceName))
		return errors.ErrInternalServerError.New("received empty token from lottery service")
	}
	l.lotteryServiceToken = token
	return nil
}
