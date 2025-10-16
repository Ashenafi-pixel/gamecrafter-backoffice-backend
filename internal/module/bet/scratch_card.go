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
)

func (b *bet) GetScratchGamePrice(ctx context.Context) (dto.GetScratchCardRes, error) {
	price, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.SCRATCH_CARD_PRICE)
	if err != nil {
		return dto.GetScratchCardRes{}, err
	}
	if !exist {
		_, err := b.ConfigStorage.CreateConfig(ctx, dto.Config{
			Name:  constant.SCRATCH_CARD_PRICE,
			Value: fmt.Sprintf("%d", constant.SCRATCH_CARD_PRICE_DEFAULT),
		})
		if err != nil {
			return dto.GetScratchCardRes{}, err
		}
		return dto.GetScratchCardRes{
			Message:  constant.SUCCESS,
			Price:    decimal.NewFromInt(constant.SCRATCH_CARD_PRICE_DEFAULT),
			MaxPrize: b.scratchMaxPrice,
		}, nil
	}
	parsedPrice, err := decimal.NewFromString(price.Value)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.GetScratchCardRes{}, err
	}
	return dto.GetScratchCardRes{
		Message:  constant.SUCCESS,
		Price:    parsedPrice,
		MaxPrize: b.scratchMaxPrice,
	}, nil
}

func (b *bet) PlaceScratchCardBet(ctx context.Context, userID uuid.UUID) (dto.ScratchCard, error) {
	defer b.TriggerLevelResponse(ctx, userID)
	defer b.TriggerPlayerProgressBar(ctx, userID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, userID)
	defer b.userWS.TriggerBalanceWS(ctx, userID)

	open, err := b.CheckBetLockStatus(ctx, constant.GAME_SCRATCH_CARD)
	if err != nil {
		return dto.ScratchCard{}, err
	}
	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ScratchCard{}, err
	}

	wonStatus := constant.LOSE
	wonAmount := decimal.Zero
	var parsedPrice decimal.Decimal
	price, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.SCRATCH_CARD_PRICE)
	if err != nil {
		return dto.ScratchCard{}, err
	}
	if !exist || price.Value == "" {
		parsedPrice = decimal.NewFromInt(constant.SCRATCH_CARD_PRICE_DEFAULT)
	} else {
		parsedPrice = decimal.RequireFromString(price.Value)
	}

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       userID,
		CurrencyCode: constant.POINT_CURRENCY,
	})
	if err != nil {
		return dto.ScratchCard{}, err
	}
	if balance.AmountUnits.LessThan(parsedPrice) || !exist {
		err := fmt.Errorf("insufficient balance")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ScratchCard{}, err
	}

	b.scratchBets = b.scratchBets.Add(parsedPrice)
	posibleWinCards := b.GetPossibleCards(ctx)
	cards := b.GenerateScratchCard(posibleWinCards)

	if cards.Prize.GreaterThan(decimal.Zero) {
		wonStatus = constant.WON
		wonAmount = cards.Prize
		b.scratchBets = b.scratchBets.Sub(cards.Prize)
	}

	_, err = b.betStorage.CreateScratchCardsBet(ctx, dto.ScratchCardsBetData{
		UserID:    userID,
		Status:    constant.CLOSED,
		BetAmount: parsedPrice,
		WonStatus: wonStatus,
		WonAmount: wonAmount,
	})

	if err != nil {
		return dto.ScratchCard{}, err
	}

	newBalance := balance.AmountUnits.Sub(parsedPrice)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    userID,
		Currency:  constant.POINT_CURRENCY,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.ScratchCard{}, err
	}

	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balance.AmountUnits,
		})
		return dto.ScratchCard{}, err
	}

	_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             userID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINT_CURRENCY,
		Description:        fmt.Sprintf("place scratch cards bet amount %v, new balance is %v and currency %s", parsedPrice, balance.AmountUnits.Sub(parsedPrice), constant.POINT_CURRENCY),
		ChangeAmount:       parsedPrice,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.ScratchCard{}, err
	}

	if cards.Prize.GreaterThan(decimal.Zero) {
		balance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
			UserId:       userID,
			CurrencyCode: constant.POINT_CURRENCY,
		})
		if err != nil {
			return dto.ScratchCard{}, err
		}
		balanceAfterWin := balance.AmountUnits.Add(wonAmount)
		updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    userID,
			Currency:  constant.POINT_CURRENCY,
			Component: constant.REAL_MONEY,
			Amount:    balanceAfterWin,
		})
		if err != nil {
			return dto.ScratchCard{}, err
		}
		b.SaveToSquads(ctx, dto.SquadEarns{
			UserID:   userID,
			Currency: constant.POINT_CURRENCY,
			Earn:     wonAmount,
			GameID:   constant.GAME_SCRATCH_CARD,
		})

		operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    userID,
				Currency:  constant.POINT_CURRENCY,
				Component: constant.REAL_MONEY,
				Amount:    balance.AmountUnits,
			})
			return dto.ScratchCard{}, err
		}

		transactionId := utils.GenerateTransactionId()
		b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
			UserID:             userID,
			Component:          constant.REAL_MONEY,
			Currency:           constant.POINT_CURRENCY,
			Description:        fmt.Sprintf("cash out scratch card bet %v amount, new balance is %v currency balance is %s", wonAmount, updatedBalance.AmountUnits, constant.POINT_CURRENCY),
			ChangeAmount:       updatedBalance.AmountUnits,
			OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
			OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
			BalanceAfterUpdate: &balanceAfterWin,
			TransactionID:      &transactionId,
		})
	}

	cards.BetAmount = parsedPrice
	return cards, nil
}

func (b *bet) GenerateScratchCard(possibleCardsToWon []string) dto.ScratchCard {
	for {
		rand.Seed(time.Now().UnixNano())
		var card dto.ScratchCard

		allSymbols := []string{constant.SCRATCH_CAR, constant.SCRATCH_DOLLAR, constant.SCRATCH_CRAWN, constant.SCRATCH_CENT, constant.SCRATCH_DIAMOND, constant.SCRATCH_CUP}
		rand.Shuffle(len(allSymbols), func(i, j int) {
			allSymbols[i], allSymbols[j] = allSymbols[j], allSymbols[i]
		})
		card.Symbols = allSymbols[:3]
		card.Board = [3][3]string{}

		const matchProbability = 0.01 // Reduced to 0.01 (1% chance of winning)

		symbolCounts := map[string]int{
			card.Symbols[0]: 3,
			card.Symbols[1]: 3,
			card.Symbols[2]: 3,
		}

		var symbolList []string
		for symbol, count := range symbolCounts {
			for i := 0; i < count; i++ {
				symbolList = append(symbolList, symbol)
			}
		}
		rand.Shuffle(len(symbolList), func(i, j int) {
			symbolList[i], symbolList[j] = symbolList[j], symbolList[i]
		})

		// Choose either row or diagonal match, not both
		matchType := rand.Intn(2) // 0 for row, 1 for diagonal
		allowRowMatch := matchType == 0 && rand.Float64() < matchProbability
		allowDiagonalMatch := matchType == 1 && rand.Float64() < matchProbability

		idx := 0
		if allowRowMatch {
			symbol := card.Symbols[rand.Intn(3)]
			card.Board[0][0], card.Board[0][1], card.Board[0][2] = symbol, symbol, symbol
			tempList := []string{}
			count := 0
			for _, s := range symbolList {
				if s != symbol || count >= 3 {
					tempList = append(tempList, s)
				} else {
					count++
				}
			}
			symbolList = tempList
			idx = 3
		} else if allowDiagonalMatch {
			symbol := card.Symbols[rand.Intn(3)]
			card.Board[0][0], card.Board[1][1], card.Board[2][2] = symbol, symbol, symbol
			tempList := []string{}
			count := 0
			for _, s := range symbolList {
				if s != symbol || count >= 3 {
					tempList = append(tempList, s)
				} else {
					count++
				}
			}
			symbolList = tempList
			idx = 3
		}

		remainingCells := [][2]int{}
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if card.Board[i][j] == "" {
					remainingCells = append(remainingCells, [2]int{i, j})
				}
			}
		}
		rand.Shuffle(len(remainingCells), func(i, j int) {
			remainingCells[i], remainingCells[j] = remainingCells[j], remainingCells[i]
		})

		for _, cell := range remainingCells {
			i, j := cell[0], cell[1]
			var validSymbol string
			for len(symbolList) > 0 {
				validSymbol = symbolList[0]
				symbolList = symbolList[1:]

				rowMatch := i < 2 && card.Board[i][0] == validSymbol && card.Board[i][1] == validSymbol && card.Board[i][2] == ""
				if i == 2 {
					rowMatch = card.Board[i][0] == validSymbol && card.Board[i][1] == validSymbol
				}
				diagMatch := i == j && i < 2 && card.Board[0][0] == validSymbol && card.Board[1][1] == validSymbol && card.Board[2][2] == ""
				secDiagMatch := (i+j == 2) && i < 2 && card.Board[0][2] == validSymbol && card.Board[1][1] == validSymbol && card.Board[2][0] == ""
				if rowMatch || (diagMatch && !allowDiagonalMatch) || (secDiagMatch && !allowDiagonalMatch) {
					symbolList = append(symbolList, validSymbol)
					continue
				}
				break
			}
			card.Board[i][j] = validSymbol
		}

		if len(symbolList) > 0 {
			for _, cell := range remainingCells[idx:] {
				i, j := cell[0], cell[1]
				if card.Board[i][j] == "" && len(symbolList) > 0 {
					card.Board[i][j] = symbolList[0]
					symbolList = symbolList[1:]
				}
			}
		}

		matches := make(map[string]int)
		var matchCells [][2]int
		for i := 0; i < 3; i++ {
			if card.Board[i][0] != "" && card.Board[i][0] == card.Board[i][1] && card.Board[i][1] == card.Board[i][2] {
				matches[card.Board[i][0]]++
				if len(matchCells) == 0 {
					matchCells = append(matchCells, [2]int{i, 0}, [2]int{i, 1}, [2]int{i, 2})
				}
			}
		}
		for j := 0; j < 3; j++ {
			if card.Board[0][j] != "" && card.Board[0][j] == card.Board[1][j] && card.Board[1][j] == card.Board[2][j] {
				matches[card.Board[0][j]]++
				if len(matchCells) == 0 {
					matchCells = append(matchCells, [2]int{0, j}, [2]int{1, j}, [2]int{2, j})
				}
			}
		}
		if card.Board[0][0] != "" && card.Board[0][0] == card.Board[1][1] && card.Board[1][1] == card.Board[2][2] {
			matches[card.Board[0][0]]++
			if len(matchCells) == 0 {
				matchCells = append(matchCells, [2]int{0, 0}, [2]int{1, 1}, [2]int{2, 2})
			}
		}
		if card.Board[0][2] != "" && card.Board[0][2] == card.Board[1][1] && card.Board[1][1] == card.Board[2][0] {
			matches[card.Board[0][2]]++
			if len(matchCells) == 0 {
				matchCells = append(matchCells, [2]int{0, 2}, [2]int{1, 1}, [2]int{2, 0})
			}
		}

		winningSymbols := []string{}
		for symbol, count := range matches {
			if count > 0 {
				winningSymbols = append(winningSymbols, symbol)
			}
		}

		if len(winningSymbols) <= 1 {
			card.WinningSymbol, card.Prize, card.MatchCells = b.CheckWins(card)
			return card
		}
	}
}

func (b *bet) CheckWins(sc dto.ScratchCard) (string, decimal.Decimal, [][2]int) {
	var winningSymbol string
	prize := decimal.Zero
	matches := make(map[string]int)
	var matchCells [][2]int

	for i := 0; i < 3; i++ {
		if sc.Board[i][0] != "" && sc.Board[i][0] == sc.Board[i][1] && sc.Board[i][1] == sc.Board[i][2] {
			matches[sc.Board[i][0]]++
			if len(matchCells) == 0 {
				matchCells = append(matchCells, [2]int{i, 0}, [2]int{i, 1}, [2]int{i, 2})
			}
		}
	}
	for j := 0; j < 3; j++ {
		if sc.Board[0][j] != "" && sc.Board[0][j] == sc.Board[1][j] && sc.Board[1][j] == sc.Board[2][j] {
			matches[sc.Board[0][j]]++
			if len(matchCells) == 0 {
				matchCells = append(matchCells, [2]int{0, j}, [2]int{1, j}, [2]int{2, j})
			}
		}
	}
	if sc.Board[0][0] != "" && sc.Board[0][0] == sc.Board[1][1] && sc.Board[1][1] == sc.Board[2][2] {
		matches[sc.Board[0][0]]++
		if len(matchCells) == 0 {
			matchCells = append(matchCells, [2]int{0, 0}, [2]int{1, 1}, [2]int{2, 2})
		}
	}
	if sc.Board[0][2] != "" && sc.Board[0][2] == sc.Board[1][1] && sc.Board[1][1] == sc.Board[2][0] {
		matches[sc.Board[0][2]]++
		if len(matchCells) == 0 {
			matchCells = append(matchCells, [2]int{0, 2}, [2]int{1, 1}, [2]int{2, 0})
		}
	}

	winningSymbols := []string{}
	for symbol, count := range matches {
		if count > 0 {
			winningSymbols = append(winningSymbols, symbol)
		}
	}

	if len(winningSymbols) == 1 && matches[winningSymbols[0]] == 1 {
		winningSymbol = winningSymbols[0]
		prizeOfWinning, exist, err := b.ConfigStorage.GetConfigByName(context.Background(), winningSymbol)
		if err != nil {
			prize = decimal.Zero
		}
		if exist && prizeOfWinning.Value != "" {
			prize, err = decimal.NewFromString(strings.TrimSpace(prizeOfWinning.Value))
			if err != nil {
				return winningSymbol, prize, matchCells
			}
		}
	} else {
		matchCells = nil
	}

	return winningSymbol, prize, matchCells
}

func (b *bet) GetPossibleCards(ctx context.Context) []string {
	var resp []string
	for k, r := range b.scratchCardsPriceHolder {
		if b.scratchBets.GreaterThan(r) {
			resp = append(resp, k)
		}
	}
	return resp
}

func (b *bet) GetUserScratchCardBetHistories(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetScratchBetHistoriesResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	resp, _, err := b.betStorage.GetUserScratchCardBetHistories(ctx, req, userID)
	if err != nil {
		return dto.GetScratchBetHistoriesResp{}, err
	}
	return resp, nil
}

func (b *bet) GetScratchCardsConfig(ctx context.Context) (dto.GetScratchCardConfigs, error) {
	return b.ConfigStorage.GetScratchCardConfigs(ctx)
}

func (b *bet) UpdateScratchGameConfig(ctx context.Context, req dto.UpdateScratchGameConfigRequest) (dto.UpdateScratchGameConfigResponse, error) {
	if req.Name != constant.SCRATCH_CAR &&
		req.Name != constant.SCRATCH_DOLLAR &&
		req.Name != constant.SCRATCH_CENT &&
		req.Name != constant.SCRATCH_CUP &&
		req.Name != constant.SCRATCH_CRAWN &&
		req.Name != constant.SCRATCH_DIAMOND {
	}
	_, err := b.ConfigStorage.UpdateConfigByName(ctx, dto.Config{
		Name:  req.Name,
		Value: req.Prize.String(),
		ID:    req.Id,
	})
	if err != nil {
		return dto.UpdateScratchGameConfigResponse{}, err
	}
	return dto.UpdateScratchGameConfigResponse{
		Message: constant.SUCCESS,
	}, nil
}
