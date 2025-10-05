package bet

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func (b *bet) CreateLeague(ctx context.Context, req dto.League) (dto.League, error) {
	if req.LeagueName == "" {
		err := fmt.Errorf("league name is required")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.League{}, err
	}
	return b.betStorage.CreateLeague(ctx, req)
}

func (b *bet) GetLeagues(ctx context.Context, req dto.GetRequest) (dto.GetLeagueRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return b.betStorage.GetLeagues(ctx, req)
}

func (b *bet) CreateClub(ctx context.Context, req dto.Club) (dto.Club, error) {
	if req.Name == "" {
		err := fmt.Errorf("league name is required")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Club{}, err
	}

	return b.betStorage.CreateClub(ctx, req)
}

func (b *bet) GetClubs(ctx context.Context, req dto.GetRequest) (dto.GetClubRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return b.betStorage.GetClubs(ctx, req)
}

func (b *bet) CreateFootballCardMultiplier(ctx context.Context, req dto.FootballCardMultiplier) (dto.Config, error) {

	if req.CardMultiplier.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("card multiplier must be greater than 0")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return b.betStorage.CreateFootballCardMultiplier(ctx, req.CardMultiplier)

}

func (b *bet) UpdateFootballCardMultiplierValue(ctx context.Context, req dto.FootballCardMultiplier) (dto.Config, error) {
	if req.CardMultiplier.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf("card multiplier must be greater than 0")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return b.betStorage.UpdateFootballCardMultiplierValue(ctx, req.CardMultiplier)
}

func (b *bet) GetFootballCardMultiplier(ctx context.Context) (dto.Config, error) {
	resp, exist, err := b.betStorage.GetFootballCardMultiplier(ctx)
	if err != nil {
		return dto.Config{}, err
	}
	if !exist {
		err := fmt.Errorf("card multiplier not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return resp, nil
}

func (b *bet) CreateFootballMatchRound(ctx context.Context) (dto.FootballMatchRound, error) {
	return b.betStorage.CreateFootBallMatchRound(ctx, dto.FootballMatchRound{
		Status:    constant.PENDING,
		Timestamp: time.Now(),
	})
}

func (b *bet) GetFootballMatchRounds(ctx context.Context, req dto.GetRequest) (dto.GetFootballMatchRoundRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	m, _, err := b.betStorage.GetFootballMatchRound(ctx, req)
	return m, err
}

func (b *bet) CreateFootballMatch(ctx context.Context, req []dto.FootballMatchReq) ([]dto.FootballMatch, error) {
	var matchs []dto.FootballMatch
	for _, m := range req {
		if m.LeagueID == uuid.Nil {
			err := fmt.Errorf("league id is required")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		// check if league exist
		_, exist, err := b.betStorage.GetLeagueByID(ctx, m.LeagueID)
		if err != nil {
			return nil, err
		}
		if !exist {
			err := fmt.Errorf("league not found")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		// check if home club exist
		if m.HomeTeam == uuid.Nil {
			err := fmt.Errorf("home club id is required")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		home, exist, err := b.betStorage.GetClubByID(ctx, m.HomeTeam)
		if err != nil {
			return nil, err
		}

		if !exist {
			err := fmt.Errorf("home club not found")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		// check if away club exist
		if m.AwayTeam == uuid.Nil {
			err := fmt.Errorf("away club id is required")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		away, exist, err := b.betStorage.GetClubByID(ctx, m.AwayTeam)
		if err != nil {
			return nil, err
		}

		if !exist {
			err := fmt.Errorf("away club not found")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		// check if round exist
		if m.RoundID == uuid.Nil {
			err := fmt.Errorf("round id is required")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		round, exist, err := b.betStorage.GetFootballRoundByID(ctx, m.RoundID)
		if err != nil {
			return nil, err
		}

		if !exist || round.ID == uuid.Nil {
			err := fmt.Errorf("round not found")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return nil, err
		}
		// create match

		match, err := b.betStorage.CreateFootballMatch(ctx, dto.FootballMatch{
			LeagueID:  m.LeagueID.String(),
			RoundID:   m.RoundID,
			HomeTeam:  home.Name,
			AwayTeam:  away.Name,
			Status:    constant.ACTIVE,
			MatchDate: m.MatchDate,
		})

		if err != nil {
			return nil, err
		}
		matchs = append(matchs, dto.FootballMatch{
			ID:        match.ID,
			LeagueID:  m.LeagueID.String(),
			RoundID:   m.RoundID,
			HomeTeam:  home.Name,
			AwayTeam:  away.Name,
			Status:    constant.ACTIVE,
			HomeScore: 0,
			AwayScore: 0,
		})
	}
	return matchs, nil
}

func (b *bet) GetFootballRoundMatchs(ctx context.Context, req dto.GetFootballRoundMatchesReq) (dto.GetFootballRoundMatchesRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	// check if round exist
	rw, exist, err := b.betStorage.GetFootballRoundByID(ctx, req.RoundID)
	if err != nil {
		return dto.GetFootballRoundMatchesRes{}, err
	}

	if !exist || rw.ID == uuid.Nil {
		err := fmt.Errorf("round not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetFootballRoundMatchesRes{}, err
	}

	resp, exist, err := b.betStorage.GetFootballRoundMatchs(ctx, req)
	if err != nil {
		return dto.GetFootballRoundMatchesRes{}, err
	}

	if !exist {
		err := fmt.Errorf("round not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetFootballRoundMatchesRes{}, err
	}
	resp.Round = rw
	return resp, nil
}

func (b *bet) GetCurrentFootballRound(ctx context.Context) (dto.GetFootballRoundMatchesRes, error) {
	resp, exist, err := b.betStorage.GetFootballMatchRoundByStatus(ctx, dto.GetFootballMatchRoundsByStatusReq{
		Status:  constant.ACTIVE,
		Page:    0,
		PerPage: 1,
	})
	if err != nil {
		return dto.GetFootballRoundMatchesRes{}, err
	}
	if !exist || len(resp) == 0 {
		err := fmt.Errorf("round not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetFootballRoundMatchesRes{}, err
	}

	resp2, exist, err := b.betStorage.GetFootballRoundMatchs(ctx, dto.GetFootballRoundMatchesReq{
		RoundID: resp[0].ID,
		Page:    0,
		PerPage: 10})

	if err != nil {
		return dto.GetFootballRoundMatchesRes{}, err
	}

	if !exist {
		err := fmt.Errorf("round not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetFootballRoundMatchesRes{}, err
	}
	resp2.Round = resp[0]

	return resp2, nil
}

func (b *bet) CloseFootballMatch(ctx context.Context, req dto.CloseFootballMatchReq) (dto.FootballMatch, error) {
	// check if match exist
	match, exist, err := b.betStorage.GetFootballMatchByID(ctx, req.ID)
	won := constant.FOOTBALL_DRAW
	if err != nil {
		return dto.FootballMatch{}, err
	}

	if !exist {
		err := fmt.Errorf("match not found")
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.FootballMatch{}, err
	}

	if req.HomeScore > req.AwayScore {
		won = constant.FOOTBALL_HOME_WON
	} else if req.AwayScore > req.HomeScore {
		won = constant.FOOTBALL_AWAY_WON
	}

	resp, err := b.betStorage.CloseFootballMatch(ctx, dto.CloseFootballMatchReq{
		ID:        req.ID,
		HomeScore: req.HomeScore,
		AwayScore: req.AwayScore,
		Winner:    won,
	})

	if err != nil {
		return dto.FootballMatch{}, err
	}

	resp.HomeTeam = match.HomeTeam
	resp.AwayTeam = match.AwayTeam
	resp.LeagueID = match.LeagueID
	resp.RoundID = match.RoundID
	//check bet logic
	b.UpdateUserFootballBets(ctx, req, resp.RoundID)
	return resp, nil
}

func (b *bet) UpdateFootballRoundPrice(ctx context.Context, req dto.UpdateFootballBetPriceReq) (dto.UpdateFootballBetPriceRes, error) {

	if req.Price.LessThan(decimal.Zero) {
		err := fmt.Errorf("price can not be less that zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateFootballBetPriceRes{}, err
	}

	resp, err := b.betStorage.SetFootballMatchPrice(ctx, dto.Config{
		Value: req.Price.String(),
	})

	if err != nil {
		return dto.UpdateFootballBetPriceRes{}, err
	}

	return dto.UpdateFootballBetPriceRes{
		Message: constant.SUCCESS,
		Price:   resp.Value,
	}, nil
}

func (b *bet) GetFootballMatchPrice(ctx context.Context) (dto.UpdateFootballBetPriceRes, error) {
	resp, err := b.betStorage.GetFootballMatchPrice(ctx)
	if err != nil {
		return dto.UpdateFootballBetPriceRes{}, err
	}

	return dto.UpdateFootballBetPriceRes{
		Message: constant.SUCCESS,
		Price:   resp.Value,
	}, nil
}

//place bet for football round

func (b *bet) PleceBetOnFootballRound(ctx context.Context, req dto.UserFootballMatchBetReq) (dto.UserFootballMatchBetRes, error) {
	// lock user before making transactions
	defer b.TriggerLevelResponse(ctx, req.UserID)
	defer b.TriggerPlayerProgressBar(ctx, req.UserID)
	defer b.InitiateTriggerSquadsProgressBar(ctx, req.UserID)
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)

	userLock := b.getUserLock(req.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	open, err := b.CheckBetLockStatus(ctx, constant.GAME_FOOTBALL_FIXTURES)
	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	if !open {
		err := fmt.Errorf("game currently unavailable")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	machesHolder := make(map[uuid.UUID]bool)

	// check selections

	//check if user has enough amount to by ticket
	if req.Currency == "" {
		req.Currency = constant.DEFAULT_CURRENCY
	}
	tiketPrice, err := b.betStorage.GetFootballMatchPrice(ctx)
	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	// get user balance

	balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		CurrencyCode: req.Currency,
		UserId:       req.UserID,
	})

	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	if !exist || balance.ID == uuid.Nil {
		err := fmt.Errorf("insufficient balance ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	price, err := decimal.NewFromString(tiketPrice.Value)

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	if balance.RealMoney.LessThan(price) {
		err := fmt.Errorf("insufficient balance  %v", tiketPrice.Value)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	//check for the users is blocked or not
	if err := b.CheckGameBlocks(ctx, req.UserID); err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}
	// check  if round is exist or not
	round, exist, err := b.betStorage.GetFootballRoundByID(ctx, req.RoundID)
	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	if !exist || round.ID == uuid.Nil || round.Status != constant.ACTIVE {
		err := fmt.Errorf("not active round found with this round id")
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	// get all matchs for selected round
	matches, exist, err := b.betStorage.GetFootballMatchesByRoundID(ctx, req.RoundID)
	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	if !exist || len(matches) == 0 {
		err := fmt.Errorf("unable to find matches for this round")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	// check if all selections are given or not

	for _, m := range req.Selections {
		machesHolder[m.MatchID] = true
	}

	if len(machesHolder) > len(matches) || len(matches) > len(machesHolder) {
		err := fmt.Errorf("requested match selections dose not match with available selections  selected matches %d available Matches %d", len(req.Selections), len(matches))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	//check selections
	for _, v := range req.Selections {
		if v.Selection == constant.FOOTBALL_DRAW || v.Selection == constant.FOOTBALL_HOME_WON || v.Selection == constant.FOOTBALL_AWAY_WON {
			continue
		}
		err := fmt.Errorf("possible  football match selections are %s %s or %s  %s not acceptable ", constant.FOOTBALL_DRAW, constant.FOOTBALL_AWAY_WON, constant.FOOTBALL_HOME_WON, v.Selection)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}

	for _, mResp := range matches {

		if ok := machesHolder[mResp.ID]; !ok {
			err := fmt.Errorf("for match %s vs %s selections is not given", mResp.HomeTeam, mResp.AwayTeam)
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.UserFootballMatchBetRes{}, err
		}
	}

	// create mache selection
	//get bet mutiplier
	betMultiplier, exist, err := b.betStorage.GetFootballCardMultiplier(ctx)

	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to find football bet multiplier")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}
	parsedBetMultiplier, err := decimal.NewFromString(betMultiplier.Value)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserFootballMatchBetRes{}, err
	}
	//save transaction
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		return dto.UserFootballMatchBetRes{}, err
	}

	// save transaction logs
	//update user balance
	newBalance := balance.RealMoney.Sub(price)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  req.Currency,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}
	// trigger level

	// save operations logs
	balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("place football bet amount %v, new balance is %v and  currency %s", price, balance.RealMoney.Sub(price), req.Currency),
		ChangeAmount:       price,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		//reverse balance
		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		if err2 != nil {
			return dto.UserFootballMatchBetRes{}, err2
		}
		b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		return dto.UserFootballMatchBetRes{}, err
	}
	possibleWoneValue := price.Mul(parsedBetMultiplier)
	//create rounds for user bet
	footballBet, err := b.betStorage.CreateFootballBet(ctx, dto.UserFootballMatcheRound{
		Status:          constant.ACTIVE,
		WonStatus:       constant.PENDING,
		UserID:          req.UserID,
		FootballRoundID: round.ID,
		BetAmount:       price,
		WonAmount:       possibleWoneValue,
	})

	if err != nil {
		return dto.UserFootballMatchBetRes{}, err
	}
	// create bet maches

	for _, selection := range req.Selections {
		_, err := b.betStorage.CreateFootballBetUserSelection(ctx, dto.UserFootballMatchSelection{
			MatchID:                   selection.MatchID,
			Selection:                 selection.Selection,
			UsersFootballMatchRoundID: footballBet.ID,
		})

		if err != nil {
			_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
				UserID:    req.UserID,
				Currency:  req.Currency,
				Component: constant.REAL_MONEY,
				Amount:    balance.RealMoney,
			})
			if err2 != nil {
				return dto.UserFootballMatchBetRes{}, err2
			}
			b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
			return dto.UserFootballMatchBetRes{}, err
		}
	}

	return dto.UserFootballMatchBetRes{
		Message: constant.SUCCESS,
		Data: dto.UserFootballMatchBetReq{
			UserID:     req.UserID,
			Currency:   req.Currency,
			RoundID:    req.RoundID,
			Selections: req.Selections,
		},
	}, nil
}

func (b *bet) UpdateUserFootballBets(ctx context.Context, req dto.CloseFootballMatchReq, roundID uuid.UUID) error {
	// update status of all user matches
	// get all user bet for this match
	won := constant.FOOTBALL_HOME_WON
	resp, exist, err := b.betStorage.GetUserFootballMatchSelectionsByMatchID(ctx, req.ID)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	if req.HomeScore < req.AwayScore {
		won = constant.FOOTBALL_AWAY_WON
	}

	if req.HomeScore == req.AwayScore {
		won = constant.FOOTBALL_DRAW
	}

	for _, m := range resp {

		if m.Selection == won {
			_, err := b.betStorage.UpdateUserFootballMatchStatusByMatchID(ctx, constant.FOOTBALL_WON, m.ID)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return err
			}

		} else {
			_, err := b.betStorage.UpdateUserFootballMatchStatusByMatchID(ctx, constant.FOOTBALL_LOSE, m.ID)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return err
			}
		}
	}
	//  check if it is the last match if it is the last match then update round status for players
	_, exist, err = b.betStorage.GetFootballMatchesByStatus(ctx, constant.ACTIVE, roundID)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	// update user balance if they won

	return b.CloseFootballMatchRound(ctx, roundID)
}

func (b *bet) CloseFootballMatchRound(ctx context.Context, roundID uuid.UUID) error {
	// get winners by winner matches
	resp, exist, err := b.betStorage.GetAllFootBallMatchByRoundByRoundID(ctx, roundID)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	//check each users all bet for this round
	for _, m := range resp {
		// check for lose exist or not
		_, exist, err := b.betStorage.GetAllUserFootballBetByStatusAndRoundID(ctx, m.ID, constant.FOOTBALL_LOSE)
		if err != nil {
			b.log.Error(err.Error())
			continue
		}
		if exist {
			err := b.betStorage.UpdateUserFootballMatcheRoundsByID(ctx, m.ID, constant.CLOSED, constant.FOOTBALL_LOSE)
			if err != nil {
				b.log.Error(err.Error())
			}
			continue
		}
		b.betStorage.UpdateUserFootballMatcheRoundsByID(ctx, m.ID, constant.CLOSED, constant.FOOTBALL_WON)
		b.UpdateFootballBetWinnerBalance(ctx, m)
	}
	// close round
	b.betStorage.UpdateFootballmatchByRoundID(ctx, roundID, constant.CLOSED)
	return nil

}

// update transactions

func (b *bet) UpdateFootballBetWinnerBalance(ctx context.Context, req dto.UserFootballMatcheRound) {
	// lock user before making transactions
	defer b.userWS.TriggerBalanceWS(ctx, req.UserID)
	userLock := b.getUserLock(req.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	// get user balance
	balance, _, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       req.UserID,
		CurrencyCode: req.Currency,
	})

	if err != nil {
		b.log.Error(err.Error())
		return
	}

	newBalance := balance.RealMoney.Add(req.WonAmount)
	updatedBalance, err := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    req.UserID,
		Currency:  req.Currency,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})

	if err != nil {
		b.log.Error(err.Error())
	}
	b.SaveToSquads(ctx, dto.SquadEarns{
		UserID:   req.UserID,
		Currency: req.Currency,
		Earn:     req.WonAmount,
		GameID:   constant.GAME_FOOTBALL_FIXTURES,
	})
	// save cashout balance logs
	operationalGroupAndTypeIDsResp, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CASHOUT)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("cashoutReq", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		// reverse balance
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    req.UserID,
			Currency:  req.Currency,
			Component: constant.REAL_MONEY,
			Amount:    balance.RealMoney,
		})
		// reverse cashout

		return
	}
	// save balance log
	transactionId := utils.GenerateTransactionId()

	b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             req.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           req.Currency,
		Description:        fmt.Sprintf("cash out football bet  %v  amount, new balance is %v s currency balance is  %s", req.WonAmount, updatedBalance.RealMoney, req.Currency),
		ChangeAmount:       updatedBalance.RealMoney,
		OperationalGroupID: operationalGroupAndTypeIDsResp.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDsResp.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionId,
	})
}

func (b *bet) GetUserFootballBets(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetUserFootballBetRes, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return b.betStorage.GetUserFootballBets(ctx, req, userID)
}
