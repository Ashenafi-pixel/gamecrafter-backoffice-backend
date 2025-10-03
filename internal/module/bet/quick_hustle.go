package bet

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) PlaceQuickHustleBet(ctx context.Context, req dto.CreateQuickHustleBetReq) (dto.CreateQuickHustelBetRes, error) {
	defer b.TriggerLevelResponse(ctx, req.UserID)
	defer b.TriggerPlayerProgressBar(ctx, req.UserID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, req.UserID)
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	// check if bet amount is less that 1
	open, err := b.CheckBetLockStatus(ctx, constant.GAME_QUICK_HUSTLE)
	if err != nil {
		return dto.CreateQuickHustelBetRes{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateQuickHustelBetRes{}, err
	}

	if req.BetAmount.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("bet amount can not be less than or equal to zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateQuickHustelBetRes{}, err
	}

	// check balance
	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: constant.POINT_CURRENCY,
	})

	if err != nil {
		return dto.CreateQuickHustelBetRes{}, err
	}

	if !exist || balance.RealMoney.LessThan(req.BetAmount) {
		err := fmt.Errorf("insufficient balance ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateQuickHustelBetRes{}, err
	}

	// get random card
	firstCard := b.GetRandomCard()
	req.FirstCard = firstCard
	// save to database
	resp, err := b.betStorage.CreateQuickHustleBet(ctx, req)
	if err != nil {
		return dto.CreateQuickHustelBetRes{}, err
	}

	// do transaction
	//update user balance
	newBalance := balance.RealMoney.Sub(req.BetAmount)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.CreateQuickHustelBetRes{}, err
	}

	// save transaction logs
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return dto.CreateQuickHustelBetRes{}, err
	}

	// save operations logs
	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place quick hustle bet amount %v, new balance is %v and  currency %s", req.BetAmount, balance.RealMoney.Sub(req.BetAmount), constant.POINT_CURRENCY),
		ChangeAmount:       req.BetAmount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.CreateQuickHustelBetRes{}, err
	}
	// trigger level

	return dto.CreateQuickHustelBetRes{
		Message: constant.SUCCESS,
		Data:    resp,
	}, nil
}

func (b *bet) GetRandomCard() string {
	numberGuess := int(utils.RandomDecimal(2, 25))
	i := 0
	for card := range b.quickHustleCardsMap {
		if i <= numberGuess {
			i++
			continue
		}

		return card
	}
	return "s"
}

func (b *bet) UserSelectCard(ctx context.Context, req dto.SelectQuickHustlePossibilityReq) (dto.CloseQuickHustleResp, error) {
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	won := constant.LOSE
	wonAmount := decimal.Zero
	// check active bet exist with that id
	if req.UserGuess != constant.QUICK_HUSTLE_HIGHER && req.UserGuess != constant.QUICK_HUSTLE_LOWER {
		err := fmt.Errorf("user guys only HIGHER or LOWER  invalid input %s", req.UserGuess)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CloseQuickHustleResp{}, err
	}

	resp, exist, err := b.betStorage.GetQuickHustleByID(ctx, req.ID)
	if err != nil {
		return dto.CloseQuickHustleResp{}, err
	}

	if !exist {
		err := fmt.Errorf("quick hustle bet dose not exist with this id")
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CloseQuickHustleResp{}, err
	}

	if resp.Status != constant.ACTIVE {
		err := fmt.Errorf("quick hustle bet dose not have active bet with this id")
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CloseQuickHustleResp{}, err
	}

	// generate quick hustle card
	generatedCard, err := b.getCard(20, resp.FirstCard, req.UserGuess)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CloseQuickHustleResp{}, err
	}

	//check if generatedCardEqualToGusedcard
	if (b.quickHustleCardsMap[resp.FirstCard] < b.quickHustleCardsMap[generatedCard] &&
		req.UserGuess == constant.QUICK_HUSTLE_HIGHER) ||
		(b.quickHustleCardsMap[resp.FirstCard] > b.quickHustleCardsMap[generatedCard] &&
			req.UserGuess == constant.QUICK_HUSTLE_LOWER) {
		// won
		won = constant.WON
		if constant.QUICK_HUSTLE_HIGHER == req.UserGuess {
			wonAmount = b.quickHustelHigherMultiplier[generatedCard].Mul(resp.BetAmount)
		} else {
			wonAmount = b.quickHustelLowerMultiplier[generatedCard].Mul(resp.BetAmount)
		}

	}

	//save to database
	respClosed, err := b.betStorage.CloseQuickHustelBet(ctx, dto.CloseQuickHustleBetData{
		ID:         resp.ID,
		UserGuess:  req.UserGuess,
		WonStatus:  won,
		SecondCard: generatedCard,
		WonAmount:  wonAmount,
	})
	if err != nil {
		return dto.CloseQuickHustleResp{}, err
	}

	if won == constant.WON {
		balance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:       req.UserID,
			CurrencyCode: constant.POINT_CURRENCY,
		})

		if err != nil {
			return dto.CloseQuickHustleResp{}, err
		}
		balanceAfterWin := balance.RealMoney.Add(wonAmount)
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balanceAfterWin,
		})
		if err != nil {
			return dto.CloseQuickHustleResp{}, err
		}
		// save cashout balance logs
		operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("cashoutReq", req))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			// reverse balance
			_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    req.UserID,
				Currency:  constant.POINT_CURRENCY,
				Component: constant.REAL_MONEY,
				Amount:    balance.RealMoney,
			})
			// reverse cashout

			return dto.CloseQuickHustleResp{}, err
		}
		// save balance log
		transactionId := utils.GenerateTransactionId()

		b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             req.UserID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("cash out quick hustle kings bet  %v  amount, new balance is %v s currency balance is  %s", wonAmount, updatedBalance.RealMoney, constant.POINT_CURRENCY),
			ChangeAmount:       updatedBalance.RealMoney,
			OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
			BalanceAfterUpdate: &balanceAfterWin,
			TransactionID:      &transactionId,
		})
	}

	b.SaveToSquads(ctx, dto.SquadEarns{
		UserID:   req.UserID,
		Currency: constant.POINT_CURRENCY,
		Earn:     wonAmount,
		GameID:   constant.GAME_QUICK_HUSTLE,
	})
	return dto.CloseQuickHustleResp{
		Message: constant.SUCCESS,
		Data:    respClosed,
	}, nil

}

func (b *bet) getCard(winChance float64, selectedCard string, userGuess string) (string, error) {
	// Validate inputs
	if winChance < 0 || winChance > 100 {
		return "", fmt.Errorf("win chance must be between 0 and 100, got %.2f", winChance)
	}
	userGuess = strings.ToLower(userGuess)
	if userGuess != "higher" && userGuess != "lower" {
		return "", fmt.Errorf("user guess must be 'higher' or 'lower', got %s", userGuess)
	}
	selectedCard = strings.ToUpper(selectedCard)

	// Get card value map
	cardValues := b.quickHustleCardsMap

	// Validate selected card
	selectedValue, exists := cardValues[selectedCard]
	if !exists {
		return "", fmt.Errorf("invalid card: %s", selectedCard)
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Collect winning cards (cards that satisfy the guess)
	var winningCards []dto.QuickHustleCard
	for name, value := range cardValues {
		if (userGuess == "higher" && value > selectedValue) || (userGuess == "lower" && value < selectedValue) {
			winningCards = append(winningCards, dto.QuickHustleCard{Name: name, Value: value})
		}
	}

	// Collect losing cards (cards that don't satisfy the guess)
	var losingCards []dto.QuickHustleCard
	for name, value := range cardValues {
		if (userGuess == "higher" && value <= selectedValue) || (userGuess == "lower" && value >= selectedValue) {
			losingCards = append(losingCards, dto.QuickHustleCard{Name: name, Value: value})
		}
	}

	// Check if there are no winning cards
	if len(winningCards) == 0 {
		return "", fmt.Errorf("no cards are %s than %s (value: %d)", userGuess, selectedCard, selectedValue)
	}

	// Determine if the user wins
	isWin := rand.Float64()*100 < winChance

	if isWin {
		// Return a random winning card
		return winningCards[rand.Intn(len(winningCards))].Name, nil
	}

	// If no losing cards, return a winning card with a warning
	if len(losingCards) == 0 {
		fmt.Printf("Warning: No losing cards available for guess %s on card %s, returning winning card\n", userGuess, selectedCard)
		return winningCards[rand.Intn(len(winningCards))].Name, nil
	}

	// Return a random losing card
	return losingCards[rand.Intn(len(losingCards))].Name, nil
}

func (b *bet) GetQuickHustleBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetQuickHustleResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	res, exist, err := b.betStorage.GetQuickHustleBetHistoryByUserID(ctx, req, userID)
	if err != nil {
		return dto.GetQuickHustleResp{}, err
	}

	if !exist {
		return dto.GetQuickHustleResp{}, nil
	}

	return res, nil
}
