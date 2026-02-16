package squads

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type squads struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Squads {
	return &squads{
		db:  db,
		log: log,
	}
}

func (s *squads) CreateSquads(ctx context.Context, req dto.CreateSquadsReq) (dto.CreateSquadsRes, error) {
	resp, err := s.db.Queries.CreateSquad(ctx, db.CreateSquadParams{
		Handle:    req.Handle,
		Owner:     req.Owner,
		Type:      req.Type,
		CreatedAt: req.CreatedAt,
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateSquadsRes{}, err
	}

	return dto.CreateSquadsRes{
		Message: constant.SUCCESS,
		Squad: dto.Squad{
			ID:     resp.ID,
			Handle: resp.Handle,
			Type:   resp.Type,
			Owner:  resp.Owner,
		},
	}, nil
}

func (s *squads) GetSquadByHandle(ctx context.Context, handle string) (dto.Squad, bool, error) {

	resp, err := s.db.Queries.GetSquadByhandle(ctx, handle)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("handle", handle))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Squad{}, false, err
	}

	if err != nil {
		return dto.Squad{}, false, nil
	}

	return dto.Squad{
		ID:     resp.ID,
		Handle: resp.Handle,
		Type:   resp.Type,
		Owner:  resp.Owner,
	}, true, nil
}

func (s *squads) GetSquadByID(ctx context.Context, id uuid.UUID) (dto.Squad, bool, error) {
	resp, err := s.db.Queries.GetSquadById(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Squad{}, false, err
	}

	if err != nil {
		return dto.Squad{}, false, nil
	}

	return dto.Squad{
		ID:     resp.ID,
		Handle: resp.Handle,
		Type:   resp.Type,
		Owner:  resp.Owner,
	}, true, nil
}

func (s *squads) GetUserSquads(ctx context.Context, id uuid.UUID) ([]dto.GetSquadsResp, error) {
	var sqds []dto.GetSquadsResp
	resp, err := s.db.Queries.GetSquadByUserID(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("user_id", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetSquadsResp{}, err
	}

	sqds = append(sqds, dto.GetSquadsResp{
		ID:     resp.ID,
		Handle: resp.Handle,
		Type:   resp.Type,
		Owener: dto.Owener{
			ID:        resp.Owner,
			FirstName: resp.FirstName.String,
			LastName:  resp.LastName.String,
			Phone:     resp.PhoneNumber.String,
		},
	})

	return sqds, nil
}

func (s *squads) GetSquadByOwnerID(ctx context.Context, id uuid.UUID) ([]dto.Squad, bool, error) {
	var squadsList []dto.Squad
	resp, err := s.db.GetSquadByOwner(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Squad{}, false, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.Squad{}, false, nil
	}

	for _, squad := range resp {
		squadsList = append(squadsList, dto.Squad{
			ID:     squad.ID,
			Handle: squad.Handle,
			Type:   squad.Type,
			Owner:  id,
		})
	}

	return squadsList, true, nil
}

func (s *squads) UpdateSquadHundle(ctx context.Context, sd dto.Squad) (dto.Squad, error) {
	resp, err := s.db.UpdateSquad(ctx, db.UpdateSquadParams{
		Handle:    sd.Handle,
		Type:      sd.Type,
		UpdatedAt: time.Now(),
		ID:        sd.ID,
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("squad", sd))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.Squad{}, err
	}

	return dto.Squad{
		ID:     resp.ID,
		Handle: resp.Handle,
		Type:   resp.Type,
		Owner:  resp.Owner,
	}, nil
}

func (s *squads) DeleteSquad(ctx context.Context, id uuid.UUID) error {
	if err := s.db.Queries.DeleteSquad(ctx, db.DeleteSquadParams{
		ID:        id,
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	}); err != nil {
		s.log.Error(err.Error(), zap.Any("squad_id", id))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}

	return nil
}

func (s *squads) CreateSquadMember(ctx context.Context, req dto.CreateSquadReq) (dto.SquadMember, error) {
	resp, err := s.db.Queries.CreateSquadMember(ctx, db.CreateSquadMemberParams{
		SquadID:   req.SquadID,
		UserID:    req.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("squad", req))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	return dto.SquadMember{
		ID:        resp.ID,
		SquadID:   resp.SquadID,
		UserID:    resp.UserID,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

func (s *squads) GetSquadMembersBySquadID(ctx context.Context, req dto.GetSquadMemebersReq) (dto.GetSquadMemebersRes, error) {
	var members []dto.GetSquadMembersResp
	resp, err := s.db.Queries.GetSquadMembersBySquadId(ctx, db.GetSquadMembersBySquadIdParams{
		SquadID: req.SquadID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})
	if err != nil {
		s.log.Error(err.Error(), zap.Any("get_squad_members_req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetSquadMemebersRes{}, err
	}

	totalPages := 1
	for i, member := range resp {
		members = append(members, dto.GetSquadMembersResp{
			ID:        member.ID,
			SquadID:   member.SquadID,
			UserID:    member.UserID,
			FirstName: member.FirstName.String,
			LastName:  member.LastName.String,
			Phone:     member.PhoneNumber.String,
			CreatedAt: member.CreatedAt,
			UpdatedAt: member.UpdatedAt,
		})
		if i == 0 {
			totalPages = int(int(member.Total) / req.PerPage)
			if int(member.Total)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSquadMemebersRes{
		TotalPages:    totalPages,
		SquadMemebers: members,
	}, nil
}

func (s *squads) DeleteSquadMemberByID(ctx context.Context, id uuid.UUID) error {
	if err := s.db.Queries.DeleteSquadMemberByID(ctx, id); err != nil {
		s.log.Error(err.Error(), zap.Any("squad_member_id", id))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (s *squads) DeleteSquadMembersBySquadID(ctx context.Context, id uuid.UUID) error {
	if err := s.db.Queries.DeleteSquadMembersBySquadId(ctx, db.DeleteSquadMembersBySquadIdParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		SquadID:   id,
	}); err != nil {
		s.log.Error(err.Error(), zap.Any("squad_id", id))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (s *squads) AddSquadEarn(ctx context.Context, req dto.SquadEarns) (dto.SquadEarns, error) {
	resp, err := s.db.Queries.CreateSquadEarn(ctx, db.CreateSquadEarnParams{
		SquadID:   req.SquadID,
		UserID:    req.UserID,
		Currency:  req.Currency,
		Earned:    req.Earn,
		GameID:    req.GameID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("squad_ear", req))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.SquadEarns{}, err
	}

	return dto.SquadEarns{
		SquadID:   resp.SquadID,
		UserID:    resp.UserID,
		Currency:  resp.Currency,
		Earn:      resp.Earned,
		GameID:    resp.GameID,
		CreatedAt: resp.CreatedAt,
		UpdateAt:  resp.UpdatedAt,
	}, nil
}

func (s *squads) GetSquadEarns(ctx context.Context, req dto.GetSquadEarnsReq) (dto.GetSquadEarnsResp, error) {
	var sqds []dto.SquadEarns
	resp, err := s.db.Queries.GetSquadEarnsBySquadId(ctx, db.GetSquadEarnsBySquadIdParams{
		SquadID: req.SquadID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetSquadEarnsResp{}, err
	}

	if err != nil {
		return dto.GetSquadEarnsResp{}, nil
	}
	totalPages := 1
	for i, earn := range resp {
		sqds = append(sqds, dto.SquadEarns{
			SquadID:   earn.SquadID,
			Currency:  earn.Currency,
			Earn:      earn.Earned,
			GameID:    earn.GameID,
			CreatedAt: earn.CreatedAt,
			UserID:    earn.UserID,
			UpdateAt:  earn.UpdatedAt,
		})

		if i == 0 {
			totalPages = int(int(earn.Total) / req.PerPage)
			if int(earn.Total)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSquadEarnsResp{
		TotalPages:  totalPages,
		SquadsEarns: sqds,
	}, nil
}

func (s *squads) GetSquadEarnsByUserID(ctx context.Context, req dto.GetSquadEarnsReq, UserID uuid.UUID) (dto.GetSquadEarnsResp, error) {
	var sqds []dto.SquadEarns
	resp, err := s.db.Queries.GetSquadEarnsByUserIdAndSquadID(ctx, db.GetSquadEarnsByUserIdAndSquadIDParams{
		SquadID: req.SquadID,
		UserID:  UserID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetSquadEarnsResp{}, err
	}

	if err != nil {
		return dto.GetSquadEarnsResp{}, nil
	}
	totalPages := 1
	for i, earn := range resp {
		sqds = append(sqds, dto.SquadEarns{
			SquadID:   earn.SquadID,
			Currency:  earn.Currency,
			Earn:      earn.Earned,
			GameID:    earn.GameID,
			CreatedAt: earn.CreatedAt,
			UserID:    earn.UserID,
			UpdateAt:  earn.UpdatedAt,
		})

		if i == 0 {
			totalPages = int(int(earn.Total) / req.PerPage)
			if int(earn.Total)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSquadEarnsResp{
		TotalPages:  totalPages,
		SquadsEarns: sqds,
	}, nil
}

func (s *squads) GetSquadTotalEarns(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {

	resp, err := s.db.Queries.GetSquadTotalEarnsBySquadID(ctx, id)

	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return decimal.Decimal{}, err
	}
	if err != nil {
		return decimal.Zero, nil
	}

	return resp, nil
}

func (s *squads) GetUserEarnsForSquad(ctx context.Context, squadID, userID uuid.UUID) (decimal.Decimal, error) {

	resp, err := s.db.Queries.GetUserEarnsForSquad(ctx, db.GetUserEarnsForSquadParams{
		SquadID: squadID,
		UserID:  userID,
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", squadID), zap.Any("user_id", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return decimal.Decimal{}, err
	}
	if err != nil {
		return decimal.Zero, nil
	}

	return resp, nil
}

func (s *squads) GetSquadsByType(ctx context.Context, squadType string) ([]dto.Squad, error) {
	var squadsList []dto.Squad
	resp, err := s.db.Queries.GetSquadsByType(ctx, squadType)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_type", squadType))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Squad{}, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.Squad{}, nil
	}

	for _, squad := range resp {
		squadsList = append(squadsList, dto.Squad{
			ID:     squad.ID,
			Handle: squad.Handle,
			Type:   squad.Type,
			Owner:  squad.Owner,
		})
	}

	return squadsList, nil
}

func (s *squads) GetTornamentStyleRanking(ctx context.Context, req dto.GetTornamentStyleRankingReq) (dto.GetTornamentStyleRankingResp, error) {

	resp, err := s.db.Queries.GetTornamentStyleRanking(ctx, db.GetTornamentStyleRankingParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("get_tornament_style_ranking_req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetTornamentStyleRankingResp{}, err
	}

	if err != nil {
		return dto.GetTornamentStyleRankingResp{}, nil
	}
	var ranking []dto.GetTornamentStyleRank
	totalPages := 1
	for i, rank := range resp {
		ranking = append(ranking, dto.GetTornamentStyleRank{
			Rank:       rank.Rank,
			Handle:     rank.Handle,
			TotalEarns: rank.TotalEarned,
		})
		if i == 0 {
			totalPages = int(int(rank.Total) / req.PerPage)
			if int(rank.Total)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetTornamentStyleRankingResp{
		TotalPages: totalPages,
		Ranking:    ranking,
	}, nil
}

func (s *squads) CreateTournaments(ctx context.Context, req dto.CreateTournamentReq) (dto.CreateTournamentResp, error) {
	byteRewardData, err := json.Marshal(req.Rewards)

	if err != nil {
		s.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateTournamentResp{}, err
	}

	resp, err := s.db.Queries.CreateTournaments(ctx, db.CreateTournamentsParams{
		Rank:             req.Rank,
		Level:            int32(req.Level),
		CumulativePoints: req.CumulativePoints,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Rewards:          pgtype.JSONB{Bytes: byteRewardData, Status: pgtype.Present},
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("create_tournament_req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateTournamentResp{}, err
	}
	return dto.CreateTournamentResp{
		Message: constant.SUCCESS,
		Tournament: dto.Tournament{
			ID:               resp.ID,
			Rank:             resp.Rank,
			Level:            int(resp.Level),
			CumulativePoints: resp.CumulativePoints,
			Rewards:          req.Rewards,
			CreatedAt:        resp.CreatedAt,
			UpdatedAt:        resp.UpdatedAt,
		},
	}, nil
}

func (s *squads) GetTornamentStyles(ctx context.Context) ([]dto.Tournament, error) {
	var tournaments []dto.Tournament
	resp, err := s.db.Queries.GetTornamentStyles(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("get_tornament_styles", ""))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.Tournament{}, err
	}

	if err != nil || len(resp) == 0 {
		res := []dto.Tournament{}
		return res, nil
	}

	for _, tournament := range resp {
		var rewards []dto.Reward
		if tournament.Rewards.Status == pgtype.Present {
			if err := json.Unmarshal(tournament.Rewards.Bytes, &rewards); err != nil {
				s.log.Error(err.Error(), zap.Any("tournament_rewards", tournament.Rewards.Bytes))
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				return []dto.Tournament{}, err
			}
		}

		tournaments = append(tournaments, dto.Tournament{
			ID:               tournament.ID,
			Rank:             tournament.Rank,
			Level:            int(tournament.Level),
			CumulativePoints: tournament.CumulativePoints,
			Rewards:          rewards,
			CreatedAt:        tournament.CreatedAt,
			UpdatedAt:        tournament.UpdatedAt,
		})
	}

	return tournaments, nil
}

func (s *squads) CreateTournamentClaim(ctx context.Context, tournamentID, squadID uuid.UUID) (dto.TournamentClaim, error) {
	resp, err := s.db.Queries.CreateTournamentClaim(ctx, db.CreateTournamentClaimParams{
		TournamentID: tournamentID,
		SquadID:      squadID,
		ClaimedAt:    time.Now(),
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("tournament_id", tournamentID), zap.Any("squad_id", squadID))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		res := dto.TournamentClaim{}
		return res, err
	}

	return dto.TournamentClaim{
		ID:           resp.ID,
		TournamentID: resp.TournamentID,
		SquadID:      resp.SquadID,
		ClaimedAt:    resp.ClaimedAt,
	}, nil

}

func (s *squads) GetTournamentClaimBySquadID(ctx context.Context, tournamentID, squadID uuid.UUID) (dto.TournamentClaim, error) {
	resp, err := s.db.Queries.GetTournamentClaimBySquadID(ctx, db.GetTournamentClaimBySquadIDParams{
		SquadID:      squadID,
		TournamentID: tournamentID,
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", squadID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.TournamentClaim{}, err
	}

	return dto.TournamentClaim{
		ID:           resp.ID,
		TournamentID: resp.TournamentID,
		SquadID:      resp.SquadID,
		ClaimedAt:    resp.ClaimedAt,
	}, nil
}

func (s *squads) GetSquadMemberByID(ctx context.Context, id uuid.UUID) (*dto.GetSquadMemberByIDresp, error) {
	resp, err := s.db.Queries.GetSquadMemberById(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_member_id", id))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	if err != nil {
		return nil, nil
	}

	return &dto.GetSquadMemberByIDresp{
		ID:        resp.ID,
		SquadID:   resp.SquadID,
		Handle:    resp.SquadHandle,
		Owner:     resp.SquadOwner,
		CreatedAt: resp.CreatedAt,
	}, nil
}

func (s *squads) GetSquadMembersEarnings(ctx context.Context, req dto.GetSquadMembersEarningsReq, ownerID uuid.UUID) (dto.GetSquadMembersEarningsResp, error) {
	var members []dto.SquadMemberEarnings
	resp, err := s.db.Queries.GetSquadMembersEarnings(ctx, db.GetSquadMembersEarningsParams{
		SquadID: req.SquadID,
		Owner:   ownerID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_members_earnings_req", req))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetSquadMembersEarningsResp{}, err
	}

	if err != nil {
		return dto.GetSquadMembersEarningsResp{}, nil
	}
	totalPages := 1
	for i, member := range resp {
		// convert internal time.Time to dto.SquadMemberEarnings
		var memberLastActiveAt *time.Time

		t, ok := member.LastEarnedAt.(time.Time)
		if ok {
			memberLastActiveAt = &t
		} else {
			memberLastActiveAt = &time.Time{} // default zero value if conversion fails
		}

		members = append(members, dto.SquadMemberEarnings{
			UserID:       member.UserID,
			FirstName:    member.FirstName.String,
			LastName:     member.LastName.String,
			PhoneNumber:  member.PhoneNumber.String,
			TotalEarned:  member.TotalEarned,
			TotalGames:   member.TotalGames,
			LastEarnedAt: memberLastActiveAt,
		})

		if i == 0 {
			totalPages = int(int(member.Total) / req.PerPage)
			if int(member.Total)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetSquadMembersEarningsResp{
		TotalPages:      totalPages,
		Members_Earning: members,
	}, nil
}

func (s *squads) LeaveSquad(ctx context.Context, userID, squadID uuid.UUID) error {
	if err := s.db.Queries.LeaveSquad(ctx, db.LeaveSquadParams{
		UserID:  userID,
		SquadID: squadID,
	}); err != nil {
		s.log.Error(err.Error(), zap.Any("user_id", userID), zap.Any("squad_id", squadID))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (s *squads) GetSquadMembersByUserID(ctx context.Context, userID uuid.UUID) ([]dto.GetSquadMemberByIDresp, error) {
	var members []dto.GetSquadMemberByIDresp
	resp, err := s.db.Queries.GetSquadByUserIDFromSquadMember(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("user_id", userID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetSquadMemberByIDresp{}, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.GetSquadMemberByIDresp{}, nil
	}

	for _, member := range resp {
		members = append(members, dto.GetSquadMemberByIDresp{
			ID:        member.MemberID,
			SquadID:   member.ID,
			Handle:    member.Handle,
			Owner:     member.Owner,
			CreatedAt: member.CreatedAt,
		})
	}

	return members, nil
}

func (s *squads) AddToWaitingSquadMembers(ctx context.Context, req dto.CreateSquadReq) (dto.CreateSquadReq, error) {
	resp, err := s.db.Queries.AddWaitingSquadMember(ctx, db.AddWaitingSquadMemberParams{
		SquadID:   req.SquadID,
		UserID:    req.UserID,
		CreatedAt: time.Now(),
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("req", req))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.CreateSquadReq{}, err
	}

	return dto.CreateSquadReq{
		SquadID: resp.SquadID,
		UserID:  resp.UserID,
	}, nil
}

func (s *squads) GetWaitingSquadMembers(ctx context.Context, squadID uuid.UUID) ([]dto.WaitingSquadMember, error) {
	var members []dto.WaitingSquadMember
	resp, err := s.db.Queries.GetWaitingSquadMembersBySquadID(ctx, squadID)

	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", squadID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.WaitingSquadMember{}, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.WaitingSquadMember{}, nil
	}

	for _, member := range resp {
		members = append(members, dto.WaitingSquadMember{
			ID:        member.ID,
			SquadID:   member.SquadID,
			UserID:    member.UserID,
			CreatedAt: member.CreatedAt,
		})

	}

	return members, nil
}

func (s *squads) DeleteWaitingSquadMember(ctx context.Context, id uuid.UUID) error {
	if err := s.db.Queries.DeleteWaitingSquadMember(ctx, id); err != nil {
		s.log.Error(err.Error(), zap.Any("waiting_squad_member_id", id))
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (s *squads) GetWaitingSquadsOwner(ctx context.Context, squadID uuid.UUID) (dto.WaitingsquadOwner, error) {
	resp, err := s.db.Queries.GetWaitingSquadMemberOwner(ctx, squadID)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("squad_id", squadID.String()))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.WaitingsquadOwner{}, err
	}

	return dto.WaitingsquadOwner{
		FirstName: resp.FirstName.String,
		LastName:  resp.LastName.String,
		Phone:     resp.PhoneNumber.String,
		OwnerID:   resp.OwnerID,
	}, nil
}

func (s *squads) GetSquadsByOwner(ctx context.Context, ownerID uuid.UUID) ([]dto.GetSquadsResp, error) {
	resp, err := s.db.Queries.GetSquadsByOwner(ctx, ownerID)
	if err != nil && err.Error() != dto.ErrNoRows {
		s.log.Error(err.Error(), zap.Any("owner_id", ownerID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.GetSquadsResp{}, err
	}

	if err != nil || len(resp) == 0 {
		return []dto.GetSquadsResp{}, nil
	}
	var squadsResp []dto.GetSquadsResp
	for _, squad := range resp {
		squadsResp = append(squadsResp, dto.GetSquadsResp{
			ID:        squad.ID,
			Handle:    squad.Handle,
			Type:      squad.Type,
			CreatedAt: squad.CreatedAt,
			UpdatedAt: squad.UpdatedAt,
			Owener: dto.Owener{
				ID:        squad.Owner,
				FirstName: squad.FirstName.String,
				LastName:  squad.LastName.String,
				Phone:     squad.PhoneNumber.String,
			},
		})
	}

	return squadsResp, nil
}

func (s *squads) AddWaitingSquadMember(ctx context.Context, userID, squadID uuid.UUID) (dto.SquadMember, error) {
	resp, err := s.db.AddWaitingSquadMember(ctx, db.AddWaitingSquadMemberParams{
		UserID:    userID,
		SquadID:   squadID,
		CreatedAt: time.Now(),
	})

	if err != nil {
		s.log.Error(err.Error(), zap.Any("user_id", userID), zap.Any("squad_id", squadID))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	return dto.SquadMember{
		ID:        resp.ID,
		SquadID:   resp.SquadID,
		UserID:    resp.UserID,
		CreatedAt: time.Now(),
	}, nil
}

func (s *squads) ApproveWaitingSquadMember(ctx context.Context, ID uuid.UUID) error {
	err := s.db.ApproveWaitingSquadMember(ctx, ID)

	if err != nil {
		s.log.Error(err.Error(), zap.Any("waiting user id", ID))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return err
	}

	return nil
}
