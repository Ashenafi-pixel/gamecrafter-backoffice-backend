package bet

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
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

func (b *bet) GetPlinkoGameConfig(ctx context.Context) (dto.PlinkoGameConfig, error) {
	// get min
	mulCont := []decimal.Decimal{}
	min, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.PLINKO_MIN_BET)
	if err != nil {
		return dto.PlinkoGameConfig{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to find min bet config")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}

	// get max
	max, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.PLINKO_MAX_BET)
	if err != nil {
		return dto.PlinkoGameConfig{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to find max bet config")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}

	// get max
	rtp, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.PLINKO_RTP)
	if err != nil {
		return dto.PlinkoGameConfig{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to find rtp bet config")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}
	// get max
	multiplier, exist, err := b.ConfigStorage.GetConfigByName(ctx, constant.PLINKO_MULTIPLIERS)
	if err != nil {
		return dto.PlinkoGameConfig{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to find multipliers bet config")
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}
	// convert
	minParsed, err := decimal.NewFromString(min.Value)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}
	maxParsed, err := decimal.NewFromString(max.Value)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}
	rtpParsed, err := decimal.NewFromString(rtp.Value)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlinkoGameConfig{}, err
	}

	//multipliers
	mults := strings.Split(multiplier.Value, ",")
	for i, m := range mults {
		if i == 0 {
			md := strings.Split(m, "{")[1]
			parsed, err := decimal.NewFromString(md)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.PlinkoGameConfig{}, err
			}
			mulCont = append(mulCont, parsed)
			continue
		} else if i == len(mults)-1 {

			md := strings.Split(m, "}")[0]
			parsed, err := decimal.NewFromString(md)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.PlinkoGameConfig{}, err
			}
			mulCont = append(mulCont, parsed)
			continue
		}
		parsed, err := decimal.NewFromString(m)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.PlinkoGameConfig{}, err
		}

		mulCont = append(mulCont, parsed)

	}

	return dto.PlinkoGameConfig{
		Multipliers: mulCont,
		BetLimits: dto.PlinkoBetLimits{
			Min: minParsed,
			Max: maxParsed,
		},
		Rtp: rtpParsed,
	}, nil
}

func (b *bet) PlacePlinkoGame(ctx context.Context, req dto.PlacePlinkoGameReq) (dto.PlacePlinkoGameRes, error) {
	defer b.TriggerLevelResponse(ctx, req.UserID)
	defer b.TriggerPlayerProgressBar(ctx, req.UserID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, req.UserID)
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	// check users balance
	// lock user before making transactions
	userLock := b.getUserLock(req.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	open, err := b.CheckBetLockStatus(ctx, constant.GAME_PLINKO)
	if err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlacePlinkoGameRes{}, err
	}

	if req.Currency != "" && req.Currency != constant.POINT_CURRENCY && !dto.IsValidCurrency(req.Currency) {
		err := fmt.Errorf("invalid currency %s", req.Currency)
		b.log.Warn(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlacePlinkoGameRes{}, err
	}

	if req.Currency == "" {
		req.Currency = constant.POINT_CURRENCY
	}

	//check for the users is blocked or not
	if err := b.CheckGameBlocks(ctx, req.UserID); err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}

	if req.Amount.LessThan(decimal.NewFromFloat(0.10)) || req.Amount.GreaterThan(decimal.NewFromInt(100)) {
		err := fmt.Errorf("minimum bet is $0.10 and maximum bet amount is $100")
		b.log.Warn(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlacePlinkoGameRes{}, err
	}

	//get user balance
	userBalance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: req.Currency,
	})
	if err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}
	if !exist || userBalance.AmountUnits.LessThan(req.Amount) {
		err = fmt.Errorf("insufficient fund %v", req.Currency)
		b.log.Warn(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlacePlinkoGameRes{}, err
	}
	//update user balance
	newBalance := userBalance.AmountUnits.Sub(req.Amount)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  req.Currency,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}
	// save transaction
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})
		return dto.PlacePlinkoGameRes{}, err
	}

	// save transaction logs
	// save operations logs
	balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("place plinko bet amount %v, new balance is %v and  currency %s", req.Amount, userBalance.AmountUnits.Sub(req.Amount), req.Currency),
		ChangeAmount:       req.Amount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}

	// trigger level

	// save transaction logs
	board, err := b.NewBoard(12)
	if err != nil {

		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})
		if err2 != nil {
			return dto.PlacePlinkoGameRes{}, err2
		}
		b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.PlacePlinkoGameRes{}, err
	}
	// to do make sure random offset
	finalSlot, path, err := b.GenerateDrop(0, 3, board)
	if err != nil {
		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})
		if err2 != nil {
			return dto.PlacePlinkoGameRes{}, err2
		}

		return dto.PlacePlinkoGameRes{}, err
	}

	// save plinko bet resp
	winAmount := req.Amount.Mul(board.Multipliers[finalSlot])
	betResult, err := b.betStorage.SavePlinkoBet(ctx, dto.PlacePlinkoGame{
		UserID:        req.UserID,
		Timestamp:     time.Now(),
		BetAmount:     req.Amount,
		DropPath:      path,
		Multiplier:    board.Multipliers[finalSlot],
		FinalPosition: finalSlot,
		WinAmount:     winAmount,
	})
	if err != nil {
		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})
		if err2 != nil {
			return dto.PlacePlinkoGameRes{}, err2
		}
		b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		return dto.PlacePlinkoGameRes{}, err

	}

	//UPDATE user balance
	balanceAfterWin := newBalance.Add(winAmount)
	updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  req.Currency,
		Component: constant.REAL_MONEY,
		Amount:    balanceAfterWin,
	})
	if err != nil {
		return dto.PlacePlinkoGameRes{}, err
	}

	// save cashout balance logs
	operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("cashoutReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		// reverse balance
		_, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.AmountUnits,
		})
		// reverse cashout

		return dto.PlacePlinkoGameRes{}, err
	}
	// save balance log
	transactionId := utils.GenerateTransactionId()

	b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("cash out plinko bet  %v  amount, new balance is %v s currency balance is  %s", winAmount, updatedBalance.AmountUnits, req.Currency),
		ChangeAmount:       updatedBalance.AmountUnits,
		OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
		BalanceAfterUpdate: &balanceAfterWin,
		TransactionID:      &transactionId,
	})

	b.SaveToSquads(ctx, dto.SquadEarns{
		UserID:   req.UserID,
		Currency: req.Currency,
		Earn:     winAmount,
		GameID:   constant.GAME_PLINKO,
	})
	return dto.PlacePlinkoGameRes{
		ID:            betResult.ID,
		Timestamp:     betResult.Timestamp,
		BetAmount:     betResult.BetAmount,
		DropPath:      SortPath(path),
		WinAmount:     winAmount,
		FinalPosition: finalSlot,
		Multiplier:    board.Multipliers[finalSlot],
	}, nil
}

func (b *bet) NewBoard(rows int) (dto.Board, error) {
	mulCont := []decimal.Decimal{}
	slots := rows + 1
	pegs := make([][]int, rows)
	for i := 0; i < rows; i++ {
		pegs[i] = make([]int, i+1)
		for j := 0; j <= i; j++ {
			pegs[i][j] = 1
		}
	}
	mult, exist, err := b.ConfigStorage.GetConfigByName(context.Background(), constant.PLINKO_MULTIPLIERS)
	if err != nil {
		return dto.Board{}, err
	}

	mults := strings.Split(mult.Value, ",")

	if !exist {
		err := fmt.Errorf("unable to find plinko bet multiplier")
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Board{}, err
	}

	for i, m := range mults {
		if i == 0 {
			md := strings.Split(m, "{")[1]
			parsed, err := decimal.NewFromString(md)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.Board{}, err
			}
			mulCont = append(mulCont, parsed)
			continue
		} else if i == len(mults)-1 {

			md := strings.Split(m, "}")[0]
			parsed, err := decimal.NewFromString(md)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.Board{}, err
			}
			mulCont = append(mulCont, parsed)
			continue
		}
		parsed, err := decimal.NewFromString(m)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.Board{}, err
		}

		mulCont = append(mulCont, parsed)

	}
	return dto.Board{
		Rows:        rows,
		Slots:       slots,
		Pegs:        pegs,
		Multipliers: mulCont,
	}, nil
}

func (b *bet) GenerateDrop(startRow, offsetFromCenter int, board dto.Board) (int, map[int]int, error) {
	rand.Seed(time.Now().UnixNano())
	if startRow < 0 || startRow >= board.Rows {
		err := fmt.Errorf("startRow must be between 0 and %d", board.Rows-1)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return 0, nil, err
	}

	// Determine starting position
	numPegs := len(board.Pegs[startRow])
	centerPos := numPegs / 2
	currentPos := centerPos

	if offsetFromCenter > 0 {
		direction := rand.Intn(2)
		if direction == 0 {
			currentPos -= offsetFromCenter
		} else {
			currentPos += offsetFromCenter
		}
		if currentPos < 0 {
			currentPos = 0
		} else if currentPos >= numPegs {
			currentPos = numPegs - 1
		}
	}

	// Initialize path as a map of row to column
	path := make(map[int]int)
	path[startRow] = currentPos

	// Simulate the drop
	currentRow := startRow
	for currentRow < board.Rows-1 {
		bounceStrength := 1
		if rand.Float64() < 0.3 {
			bounceStrength = 2
		}
		nextRow := currentRow + bounceStrength
		if nextRow >= board.Rows {
			nextRow = board.Rows - 1
		}
		maxPos := len(board.Pegs[currentRow]) - 1
		direction := rand.Intn(2)
		if currentPos == 0 && direction == 0 {
			// Stay at edge
		} else if currentPos == maxPos && direction == 1 {
			// Stay at edge
		} else if direction == 0 {
			currentPos--
			if currentPos < 0 {
				currentPos = 0
			}
		} else {
			currentPos++
			if currentPos > len(board.Pegs[nextRow])-1 {
				currentPos = len(board.Pegs[nextRow]) - 1
			}
		}
		currentRow = nextRow
		path[currentRow] = currentPos
		if currentRow >= board.Rows-1 {
			break
		}
	}

	finalSlot := currentPos
	return finalSlot, path, nil
}

func SortPath(path map[int]int) []dto.PathEntry {
	entries := make([]dto.PathEntry, 0, len(path))
	for row, col := range path {
		entries = append(entries, dto.PathEntry{Row: row, Col: col})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Row < entries[j].Row
	})

	return entries
}

func (b *bet) GetMyPlinkoBetHistory(ctx context.Context, req dto.PlinkoBetHistoryReq) (dto.PlinkoBetHistoryRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	resp, exist, err := b.betStorage.GetUserPlinkoBetHistoryByID(ctx, req)
	if err != nil {
		return dto.PlinkoBetHistoryRes{}, err
	}

	if !exist {
		return dto.PlinkoBetHistoryRes{}, nil
	}

	return resp, nil
}

func (b *bet) GetPlinkoGameStats(ctx context.Context, userID uuid.UUID) (dto.PlinkoGameStatRes, error) {
	fmt.Println(userID.String())
	resp, exist, err := b.betStorage.GetPlinkoGameStats(ctx, userID)

	if err != nil {
		return dto.PlinkoGameStatRes{}, err
	}

	if !exist {
		return dto.PlinkoGameStatRes{}, nil
	}

	return resp, nil
}
