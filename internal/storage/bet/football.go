package bet

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateLeague(ctx context.Context, league dto.League) (dto.League, error) {
	resp, err := b.db.CreateLeague(ctx, db.CreateLeagueParams{
		LeagueName: league.LeagueName,
		Status:     sql.NullString{String: constant.ACTIVE, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", league))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.League{}, err
	}

	return dto.League{
		ID:         resp.ID,
		LeagueName: resp.LeagueName,
		Status:     resp.Status.String,
	}, nil
}

func (b *bet) GetLeagues(ctx context.Context, req dto.GetRequest) (dto.GetLeagueRes, error) {
	leagues := []dto.League{}
	resp, err := b.db.Queries.GetLeagues(ctx, db.GetLeaguesParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	totalPage := 1

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetLeagueRes{}, err
	}

	if err != nil {
		return dto.GetLeagueRes{}, nil
	}

	for i, league := range resp {
		leagues = append(leagues, dto.League{
			ID:         league.ID,
			LeagueName: league.LeagueName,
			Status:     league.Status.String,
		})
		if i == 0 {
			totalPage := int(int(league.TotalRows) / req.PerPage)
			if int(league.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetLeagueRes{
		TotalPages: totalPage,
		Leagues:    leagues,
	}, nil

}

func (b *bet) GetLeagueByID(ctx context.Context, ID uuid.UUID) (dto.League, bool, error) {
	league, err := b.db.GetLeagueByID(ctx, ID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", ID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.League{}, false, err
	}

	if err != nil {
		return dto.League{}, false, nil
	}

	return dto.League{
		ID:         league.ID,
		LeagueName: league.LeagueName,
		Status:     league.Status.String,
	}, true, nil
}

func (b *bet) CreateClub(ctx context.Context, club dto.Club) (dto.Club, error) {
	resp, err := b.db.Queries.CreateClub(ctx, db.CreateClubParams{
		ClubName:  club.Name,
		Status:    sql.NullString{String: constant.ACTIVE, Valid: true},
		Timestamp: time.Now(),
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", club))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Club{}, err
	}

	return dto.Club{
		ID:        resp.ID,
		Name:      resp.ClubName,
		Status:    resp.Status.String,
		Timestamp: resp.Timestamp,
	}, nil
}

func (b *bet) GetClubs(ctx context.Context, req dto.GetRequest) (dto.GetClubRes, error) {
	clubs := []dto.Club{}
	resp, err := b.db.Queries.GetClubs(ctx, db.GetClubsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	totalPage := 1

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetClubRes{}, err
	}

	if err != nil {
		return dto.GetClubRes{}, nil
	}

	for i, club := range resp {
		clubs = append(clubs, dto.Club{
			ID:        club.ID,
			Name:      club.ClubName,
			Status:    club.Status.String,
			Timestamp: club.Timestamp,
		})
		if i == 0 {
			totalPage := int(int(club.TotalRows) / req.PerPage)
			if int(club.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetClubRes{
		TotalPages: totalPage,
		Clubs:      clubs,
	}, nil
}

func (b *bet) GetClubByID(ctx context.Context, ID uuid.UUID) (dto.Club, bool, error) {
	club, err := b.db.GetClubByID(ctx, ID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", ID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Club{}, false, err
	}

	if err != nil {
		return dto.Club{}, false, nil
	}

	return dto.Club{
		ID:        club.ID,
		Name:      club.ClubName,
		Status:    club.Status.String,
		Timestamp: club.Timestamp,
	}, true, nil
}

func (b *bet) CreateFootballCardMultiplier(ctx context.Context, m decimal.Decimal) (dto.Config, error) {
	resp, err := b.db.Queries.CreateConfig(ctx, db.CreateConfigParams{
		Name:  constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER,
		Value: m.String(),
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", m))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (b *bet) UpdateFootballCardMultiplierValue(ctx context.Context, m decimal.Decimal) (dto.Config, error) {
	resp, err := b.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Name:  constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER,
		Value: m.String(),
	})
	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", m))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Config{}, err
	}
	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (b *bet) GetFootballCardMultiplier(ctx context.Context) (dto.Config, bool, error) {
	resp, err := b.db.GetConfigByName(ctx, constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Config{}, false, err
	}

	if err != nil {
		return dto.Config{}, false, nil
	}

	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, true, nil
}

func (b *bet) CreateFootBallMatchRound(ctx context.Context, req dto.FootballMatchRound) (dto.FootballMatchRound, error) {
	resp, err := b.db.Queries.CreateFootballMatchRound(ctx, db.CreateFootballMatchRoundParams{
		Status:    sql.NullString{String: req.Status, Valid: true},
		Timestamp: time.Now(),
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.FootballMatchRound{}, err
	}

	return dto.FootballMatchRound{
		ID:        resp.ID,
		Status:    resp.Status.String,
		Timestamp: resp.Timestamp,
	}, nil
}

func (b *bet) UpdateFootballMatchRoundStatus(ctx context.Context, req dto.FootballMatchRoundUpdateReq) (dto.FootballMatchRound, error) {
	resp, err := b.db.Queries.UpdateFootballMatchsByID(ctx, db.UpdateFootballMatchsByIDParams{
		ID:     req.ID,
		Status: sql.NullString{String: req.Status, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.FootballMatchRound{}, err
	}

	return dto.FootballMatchRound{
		ID:        resp.ID,
		Status:    resp.Status.String,
		Timestamp: resp.Timestamp,
	}, nil
}

func (b *bet) GetFootballMatchRound(ctx context.Context, req dto.GetRequest) (dto.GetFootballMatchRoundRes, bool, error) {
	rounds := []dto.FootballMatchRound{}
	resp, err := b.db.Queries.GetFootballMatchRound(ctx, db.GetFootballMatchRoundParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetFootballMatchRoundRes{}, false, err
	}

	if err != nil {
		return dto.GetFootballMatchRoundRes{}, false, nil
	}

	totalPage := 1
	for i, round := range resp {
		rounds = append(rounds, dto.FootballMatchRound{
			ID:        round.ID,
			Status:    round.Status.String,
			Timestamp: round.Timestamp,
		})
		if i == 0 {
			totalPage := int(int(round.TotalRows) / req.PerPage)
			if int(round.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}
	return dto.GetFootballMatchRoundRes{
		TotalPages: totalPage,
		Rounds:     rounds,
	}, true, nil

}

func (b *bet) GetFootballRoundByID(ctx context.Context, id uuid.UUID) (dto.FootballMatchRound, bool, error) {
	resp, err := b.db.GetFootballMatchRoundByID(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.FootballMatchRound{}, false, err
	}

	if err != nil {
		return dto.FootballMatchRound{}, false, nil
	}
	return dto.FootballMatchRound{
		ID:        resp.ID,
		Status:    resp.Status.String,
		Timestamp: resp.Timestamp,
	}, true, nil
}

func (b *bet) GetFootballMatchRoundByStatus(ctx context.Context, req dto.GetFootballMatchRoundsByStatusReq) ([]dto.FootballMatchRound, bool, error) {
	var rounds []dto.FootballMatchRound
	resp, err := b.db.GetFootballMatchRoundByStatus(ctx, db.GetFootballMatchRoundByStatusParams{
		Status: sql.NullString{String: req.Status, Valid: true},
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.FootballMatchRound{}, false, err
	}

	if err != nil {
		return []dto.FootballMatchRound{}, false, nil
	}
	for i, round := range resp {
		rounds = append(rounds, dto.FootballMatchRound{
			ID:        round.ID,
			Status:    round.Status.String,
			Timestamp: round.Timestamp,
		})
		if i == 0 {
			totalPage := int(int(round.TotalRows) / req.PerPage)
			if int(round.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return rounds, true, nil
}

func (b *bet) CreateFootballMatch(ctx context.Context, req dto.FootballMatch) (dto.FootballMatch, error) {
	resp, err := b.db.Queries.CreateFootballMatch(ctx, db.CreateFootballMatchParams{
		RoundID:  req.RoundID,
		League:   req.LeagueID,
		Date:     req.MatchDate,
		HomeTeam: req.HomeTeam,
		AwayTeam: sql.NullString{String: req.AwayTeam, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.FootballMatch{}, err
	}

	return dto.FootballMatch{
		ID: resp.ID,

		Status: resp.Status.String,
	}, nil
}

func (b *bet) GetFootballRoundMatchs(ctx context.Context, req dto.GetFootballRoundMatchesReq) (dto.GetFootballRoundMatchesRes, bool, error) {
	var matches []dto.FootballMatch
	resp, err := b.db.Queries.GetFootballRoundMatchs(ctx, db.GetFootballRoundMatchsParams{
		RoundID: req.RoundID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetFootballRoundMatchesRes{}, false, err
	}

	if err != nil {
		return dto.GetFootballRoundMatchesRes{}, false, nil
	}
	totalPage := 1
	for i, match := range resp {
		leagueIDParsed, err := uuid.Parse(match.League)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.GetFootballRoundMatchesRes{}, false, err
		}
		league, err := b.db.Queries.GetLeagueByID(ctx, leagueIDParsed)
		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
		}
		matches = append(matches, dto.FootballMatch{
			ID:         match.ID,
			RoundID:    match.RoundID,
			LeagueID:   match.League,
			HomeTeam:   match.HomeTeam,
			AwayTeam:   match.AwayTeam.String,
			MatchDate:  match.Date,
			Status:     match.Status.String,
			HomeScore:  int(match.HomeScore.Int32),
			AwayScore:  int(match.AwayScore.Int32),
			LeagueName: league.LeagueName,
			WinnerID:   match.Won.String,
		})
		if i == 0 {
			totalPage := int(int(match.TotalRows) / req.PerPage)
			if int(match.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}

	return dto.GetFootballRoundMatchesRes{
		TotalPages: totalPage,
		Matches:    matches,
	}, true, nil
}

func (b *bet) CloseFootballMatch(ctx context.Context, req dto.CloseFootballMatchReq) (dto.FootballMatch, error) {

	resp, err := b.db.Queries.CloseFootballMatchRound(ctx, db.CloseFootballMatchRoundParams{
		ID:        req.ID,
		Status:    sql.NullString{String: constant.CLOSED, Valid: true},
		HomeScore: sql.NullInt32{Int32: int32(req.HomeScore), Valid: true},
		AwayScore: sql.NullInt32{Int32: int32(req.AwayScore), Valid: true},
		Won:       sql.NullString{String: req.Winner, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.FootballMatch{}, err
	}

	return dto.FootballMatch{
		ID:        resp.ID,
		HomeScore: int(resp.HomeScore.Int32),
		AwayScore: int(resp.AwayScore.Int32),
		WinnerID:  resp.Won.String,
		Status:    resp.Status.String,
	}, nil
}

func (b *bet) GetFootballMatchByID(ctx context.Context, ID uuid.UUID) (dto.FootballMatch, bool, error) {
	match, err := b.db.GetFootballMatchByID(ctx, ID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error(), zap.Any("req", ID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.FootballMatch{}, false, err
	}

	if err != nil {
		return dto.FootballMatch{}, false, nil
	}
	leagueIdParsed, err := uuid.Parse(match.League)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.FootballMatch{}, false, err
	}

	league, err := b.db.GetLeagueByID(ctx, leagueIdParsed)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.FootballMatch{}, false, err
	}

	return dto.FootballMatch{
		ID:         match.ID,
		RoundID:    match.RoundID,
		LeagueName: league.LeagueName,
		HomeTeam:   match.HomeTeam,
		AwayTeam:   match.AwayTeam.String,
		MatchDate:  match.Date,
		Status:     match.Status.String,
		HomeScore:  int(match.HomeScore.Int32),
		AwayScore:  int(match.AwayScore.Int32),
		WinnerID:   match.Won.String,
	}, true, nil
}

func (b *bet) SetFootballMatchPrice(ctx context.Context, config dto.Config) (dto.Config, error) {
	resp, err := b.db.Queries.UpdateConfigByName(ctx, db.UpdateConfigByNameParams{
		Value: config.Value,
		Name:  constant.CONFIG_FOOTBALL_MATCH_CARD_PRICE,
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", config))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return dto.Config{
		ID:    resp.ID,
		Name:  resp.Name,
		Value: resp.Value,
	}, nil
}

func (b *bet) GetFootballMatchPrice(ctx context.Context) (dto.Config, error) {
	resp, err := b.db.Queries.GetConfigByName(ctx, constant.CONFIG_FOOTBALL_MATCH_CARD_PRICE)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Config{}, err
	}

	return dto.Config{
		Value: resp.Value,
	}, err
}

func (b *bet) CreateFootballBet(ctx context.Context, req dto.UserFootballMatcheRound) (dto.UserFootballMatcheRound, error) {
	resp, err := b.db.Queries.CreateFootballBet(ctx, db.CreateFootballBetParams{
		UserID:          req.UserID,
		FootballRoundID: uuid.NullUUID{UUID: req.FootballRoundID, Valid: true},
		BetAmount:       decimal.NullDecimal{Decimal: req.BetAmount, Valid: true},
		WonAmount:       req.WonAmount,
		Timestamp:       sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.UserFootballMatcheRound{}, err
	}
	return dto.UserFootballMatcheRound{
		ID:              resp.ID,
		Status:          resp.Status,
		WonStatus:       resp.WonStatus.String,
		UserID:          resp.UserID,
		FootballRoundID: resp.FootballRoundID.UUID,
		BetAmount:       resp.BetAmount.Decimal,
		WonAmount:       resp.WonAmount,
		Currency:        resp.Currency,
	}, nil
}

func (b *bet) CreateFootballBetUserSelection(ctx context.Context, req dto.UserFootballMatchSelection) (dto.UserFootballMatchSelection, error) {
	resp, err := b.db.Queries.CreateFootballBetUserSelection(ctx, db.CreateFootballBetUserSelectionParams{
		MatchID:                    req.MatchID,
		Selection:                  req.Selection,
		Status:                     constant.ACTIVE,
		UsersFootballMatcheRoundID: uuid.NullUUID{UUID: req.UsersFootballMatchRoundID, Valid: true},
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.UserFootballMatchSelection{}, err
	}

	return dto.UserFootballMatchSelection{
		ID:        resp.ID,
		Status:    resp.Status,
		MatchID:   resp.MatchID,
		Selection: resp.Selection,
	}, nil
}

func (b *bet) GetFootballMatchesByRoundID(ctx context.Context, roundID uuid.UUID) ([]dto.FootballMatch, bool, error) {
	var matches []dto.FootballMatch
	resp, err := b.db.Queries.GetFootballMatchesByRoundID(ctx, roundID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.FootballMatch{}, false, err
	}

	if err != nil {
		return []dto.FootballMatch{}, false, nil
	}

	for _, m := range resp {
		matches = append(matches, dto.FootballMatch{
			ID:        m.ID,
			RoundID:   m.RoundID,
			HomeTeam:  m.HomeTeam,
			AwayTeam:  m.AwayTeam.String,
			MatchDate: m.Date,
			LeagueID:  m.League,
			Status:    m.Status.String,
			WinnerID:  m.Won.String,
			HomeScore: int(m.HomeScore.Int32),
			AwayScore: int(m.AwayScore.Int32),
		})
	}
	return matches, true, nil
}

func (b *bet) UpdateUserFootballMatchStatusByMatchID(ctx context.Context, status string, ID uuid.UUID) (dto.UserFootballMatchSelection, error) {
	resp, err := b.db.Queries.UpdateUserFootballMatchStatusByMatchID(ctx, db.UpdateUserFootballMatchStatusByMatchIDParams{
		Status: status,
		ID:     ID,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserFootballMatchSelection{}, err
	}
	return dto.UserFootballMatchSelection{
		ID:        resp.ID,
		Status:    resp.Status,
		Selection: resp.Selection,
		MatchID:   resp.MatchID,
	}, nil
}

func (b *bet) GetUserFootballMatchSelectionsByMatchID(ctx context.Context, matchID uuid.UUID) ([]dto.UserFootballMatchSelection, bool, error) {
	var userMatches []dto.UserFootballMatchSelection
	resp, err := b.db.Queries.GetUserFootballMatchSelectionsByMatchID(ctx, matchID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return []dto.UserFootballMatchSelection{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.UserFootballMatchSelection{}, false, nil
	}

	for _, r := range resp {
		userMatches = append(userMatches, dto.UserFootballMatchSelection{
			ID:        r.ID,
			Status:    r.Status,
			MatchID:   r.MatchID,
			Selection: r.Selection,
		})
	}
	return userMatches, true, nil
}

func (b *bet) GetFootballMatchesByStatus(ctx context.Context, status string, roundID uuid.UUID) ([]dto.FootballMatch, bool, error) {
	var maches []dto.FootballMatch
	resp, err := b.db.Queries.GetFootballMatchesByStatus(ctx, db.GetFootballMatchesByStatusParams{
		RoundID: roundID,
		Status:  sql.NullString{String: status, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.FootballMatch{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.FootballMatch{}, false, nil
	}

	for _, m := range resp {
		maches = append(maches, dto.FootballMatch{
			ID:        m.ID,
			RoundID:   m.RoundID,
			HomeTeam:  m.HomeTeam,
			AwayTeam:  m.AwayTeam.String,
			MatchDate: m.Date,
			LeagueID:  m.League,
			Status:    m.Status.String,
			WinnerID:  m.Won.String,
			HomeScore: int(m.HomeScore.Int32),
			AwayScore: int(m.AwayScore.Int32),
		})
	}

	return maches, true, nil

}

func (b *bet) GetAllFootBallMatchByRoundByRoundID(ctx context.Context, roundID uuid.UUID) ([]dto.UserFootballMatcheRound, bool, error) {

	var userFootballMatches []dto.UserFootballMatcheRound
	resp, err := b.db.GetAllFootBallMatchByRoundByRoundID(ctx, uuid.NullUUID{UUID: roundID, Valid: true})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.UserFootballMatcheRound{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.UserFootballMatcheRound{}, false, nil
	}
	for _, m := range resp {
		userFootballMatches = append(userFootballMatches, dto.UserFootballMatcheRound{
			ID:              m.ID,
			Status:          m.Status,
			WonStatus:       m.WonStatus.String,
			UserID:          m.UserID,
			FootballRoundID: m.FootballRoundID.UUID,
			BetAmount:       m.BetAmount.Decimal,
			WonAmount:       m.WonAmount,
			Currency:        m.Currency,
		})
	}
	return userFootballMatches, true, nil
}

func (b *bet) GetAllUserFootballBetByStatusAndRoundID(ctx context.Context, roundID uuid.UUID, status string) ([]dto.UserFootballMatchSelection, bool, error) {
	var userMatches []dto.UserFootballMatchSelection
	resp, err := b.db.Queries.GetAllUserFootballBetByStatusAndRoundID(ctx, db.GetAllUserFootballBetByStatusAndRoundIDParams{
		Status:                     status,
		UsersFootballMatcheRoundID: uuid.NullUUID{UUID: roundID, Valid: true},
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.UserFootballMatchSelection{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.UserFootballMatchSelection{}, false, nil
	}

	for _, m := range resp {
		userMatches = append(userMatches, dto.UserFootballMatchSelection{
			ID:        m.ID,
			Status:    m.Status,
			MatchID:   m.MatchID,
			Selection: m.Selection,
		})
	}
	return userMatches, true, nil
}

func (b *bet) UpdateUserFootballMatcheRoundsByID(ctx context.Context, roundID uuid.UUID, status, wonStatus string) error {
	err := b.db.Queries.UpdateUserFootballMatcheRoundsByID(ctx, db.UpdateUserFootballMatcheRoundsByIDParams{
		Status:    status,
		ID:        roundID,
		WonStatus: sql.NullString{String: wonStatus, Valid: true},
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) UpdateFootballmatchByRoundID(ctx context.Context, roundID uuid.UUID, status string) error {
	_, err := b.db.Queries.UpdateFootballmatchByRoundID(ctx, db.UpdateFootballmatchByRoundIDParams{
		Status: sql.NullString{String: status, Valid: true},
		ID:     roundID,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) GetUserFootballBets(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetUserFootballBetRes, error) {
	var bets dto.GetUserFootballBetRes

	resp, err := b.db.GetUserFootballBets(ctx, db.GetUserFootballBetsParams{
		UserID: userID,
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetUserFootballBetRes{}, err
	}
	totalPages := 1
	for i, bt := range resp {
		//get selections for bet
		maches := []dto.FootballMatchRes{}

		selections, err := b.db.Queries.GetUserFootballBetMatchesForUserBet(ctx, uuid.NullUUID{UUID: bt.ID, Valid: true})
		if err != nil {
			b.log.Error(err.Error())
			continue
		}
		round := dto.FootballRoundsRes{
			ID:            bt.ID,
			BetAmount:     bt.BetAmount.Decimal,
			WinningAmount: bt.WinningAmount,
			RoundStatus:   bt.RoundStatus,
			Currency:      bt.Currency,
		}

		for _, selection := range selections {
			won := constant.PENDING
			// get match details
			resp, err := b.db.Queries.GetFootballMatchByID(ctx, selection.MatchID)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrUnableToGet.Wrap(err, err.Error())
				return dto.GetUserFootballBetRes{}, err
			}

			// home league name
			parsedLeagueID, err := uuid.Parse(resp.League)
			if err != nil {
				b.log.Error(err.Error())
				err = errors.ErrInternalServerError.Wrap(err, err.Error())
				return dto.GetUserFootballBetRes{}, err
			}
			leagueName, err := b.db.GetLeagueByID(ctx, parsedLeagueID)
			if err != nil {
				b.log.Error(err.Error(), zap.Any("req", resp))
				err = errors.ErrUnableToGet.Wrap(err, err.Error())
				return dto.GetUserFootballBetRes{}, err
			}

			// check if mach end
			if resp.Won.String == constant.FOOTBALL_HOME_WON {
				won = constant.FOOTBALL_HOME_WON
			} else if resp.Won.String == constant.FOOTBALL_AWAY_WON {
				won = constant.FOOTBALL_AWAY_WON
			}

			match := dto.FootballMatchRes{
				ID:         selection.ID,
				Status:     selection.Status,
				Selection:  selection.Selection,
				HomeTeam:   resp.HomeTeam,
				AwayTeam:   resp.AwayTeam.String,
				MatchDate:  resp.Date,
				LeagueName: leagueName.LeagueName,
				Winner:     won,
			}
			maches = append(maches, match)
		}
		round.Matches = maches
		bets.Rounds = append(bets.Rounds, round)
		if i == 0 {
			totalPage := int(int(bt.TotalRows) / req.PerPage)
			if int(bt.TotalRows)%req.PerPage != 0 {
				totalPage++
			}
		}
	}
	bets.Message = constant.SUCCESS
	bets.TotalPages = totalPages
	return bets, nil

}
