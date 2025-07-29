package squads

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type squads struct {
	log           *zap.Logger
	squadsStorage storage.Squads
	usersStorage  storage.User
}

func Init(log *zap.Logger, squadsStorage storage.Squads, userStorage storage.User) module.Squads {
	return &squads{
		log:           log,
		squadsStorage: squadsStorage,
		usersStorage:  userStorage,
	}
}

func (s *squads) CreateSquads(ctx context.Context, req dto.CreateSquadsReq) (dto.CreateSquadsRes, error) {
	rm, _ := s.squadsStorage.GetSquadMembersByUserID(ctx, req.Owner)

	if len(rm) > 0 {
		err := fmt.Errorf("user %s is already a member of squad", req.Owner.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateSquadsRes{}, err
	}
	if req.Handle == "" {
		err := fmt.Errorf("invalid squad handle: %s", req.Handle)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateSquadsRes{}, err
	}
	// Validate squad type
	if err := dto.CheckInvalidSquadType(req.Type); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		s.log.Error("invalid squad type", zap.String("squadType", req.Type), zap.Error(err))
		return dto.CreateSquadsRes{}, err
	}

	_, exist, err := s.squadsStorage.GetSquadByHandle(ctx, req.Handle)
	if err != nil {
		return dto.CreateSquadsRes{}, err
	}

	if exist {
		err := fmt.Errorf("squad with handle %s already exist", req.Handle)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateSquadsRes{}, err
	}

	resp, err := s.squadsStorage.CreateSquads(ctx, req)
	if err != nil {
		return dto.CreateSquadsRes{}, err
	}

	return resp, nil
}

func (s *squads) GetMySquads(ctx context.Context, userID uuid.UUID) ([]dto.GetSquadsResp, error) {
	var resp []dto.GetSquadMembersResp
	squad, err := s.squadsStorage.GetUserSquads(ctx, userID)
	if err != nil {
		squad, err = s.squadsStorage.GetSquadsByOwner(ctx, userID)
		if err != nil {
			resp := []dto.GetSquadsResp{}
			return resp, err
		}
		if len(squad) == 0 {
			resp := []dto.GetSquadsResp{}
			return resp, nil
		}
		members, _ := s.squadsStorage.GetSquadMembersBySquadID(ctx, dto.GetSquadMemebersReq{
			Page:    0,
			PerPage: 1000,
			SquadID: squad[0].ID,
		})

		if len(members.SquadMemebers) > 0 {
			resp = append(resp, members.SquadMemebers...)

		}
		member := dto.GetSquadMembersResp{
			ID:        userID,
			SquadID:   squad[0].ID,
			UserID:    userID,
			FirstName: squad[0].Owener.FirstName,
			LastName:  squad[0].Owener.LastName,
			Phone:     squad[0].Owener.Phone,
			CreatedAt: squad[0].UpdatedAt,
			UpdatedAt: squad[0].UpdatedAt,
		}

		resp = append(resp, member)

		squad[0].SquadMemebers = resp
		return squad, nil
	}

	if len(squad) == 0 || (len(squad) > 0 && squad[0].Handle == "") {
		squad, err = s.squadsStorage.GetSquadsByOwner(ctx, userID)
		if err != nil {
			resp := []dto.GetSquadsResp{}
			return resp, err
		}
		if len(squad) == 0 {
			resp := []dto.GetSquadsResp{}
			return resp, nil
		}
		members, _ := s.squadsStorage.GetSquadMembersBySquadID(ctx, dto.GetSquadMemebersReq{
			Page:    0,
			PerPage: 1000,
			SquadID: squad[0].ID,
		})

		if len(members.SquadMemebers) > 0 {
			resp = append(resp, members.SquadMemebers...)

		}
		member := dto.GetSquadMembersResp{
			ID:        userID,
			SquadID:   squad[0].ID,
			UserID:    userID,
			FirstName: squad[0].Owener.FirstName,
			LastName:  squad[0].Owener.LastName,
			Phone:     squad[0].Owener.Phone,
			CreatedAt: squad[0].UpdatedAt,
			UpdatedAt: squad[0].UpdatedAt,
		}

		resp = append(resp, member)

		squad[0].SquadMemebers = resp
		return squad, nil
	}

	members, err := s.squadsStorage.GetSquadMembersBySquadID(ctx, dto.GetSquadMemebersReq{
		Page:    0,
		PerPage: 1000,
		SquadID: squad[0].ID,
	})

	if err != nil {
		return squad, nil
	}

	if len(members.SquadMemebers) == 0 {
		squad[0].SquadMemebers = []dto.GetSquadMembersResp{}
		return squad, nil
	}

	squad[0].SquadMemebers = members.SquadMemebers
	return squad, nil
}

func (s *squads) GetSquadsByOwnerID(ctx context.Context, userID uuid.UUID) ([]dto.Squad, error) {
	res, exist, err := s.squadsStorage.GetSquadByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	return res, nil
}

func (s *squads) GetSquadsByType(ctx context.Context, squadType string) ([]dto.Squad, error) {
	if err := dto.CheckInvalidSquadType(squadType); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		s.log.Error("invalid squad type", zap.String("squadType", squadType), zap.Error(err))
		return nil, err
	}

	return s.squadsStorage.GetSquadsByType(ctx, squadType)
}

func (s *squads) UpdateSquadHandle(ctx context.Context, sq dto.Squad, userID uuid.UUID) (dto.Squad, error) {
	if sq.Handle == "" {
		err := fmt.Errorf("invalid squad handle: %s", sq.Handle)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Squad{}, err
	}

	sq, exist, err := s.squadsStorage.GetSquadByHandle(ctx, sq.Handle)
	if err != nil {
		return dto.Squad{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", sq.ID)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Squad{}, err
	}

	if sq.Owner != userID {
		err := fmt.Errorf("user %s is not the owner of squad %s", userID.String(), sq.ID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.Squad{}, err
	}

	resp, err := s.squadsStorage.UpdateSquadHundle(ctx, sq)
	if err != nil {
		return dto.Squad{}, err
	}

	return resp, nil
}

func (s *squads) DeleteSquad(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {

	sq, exist, err := s.squadsStorage.GetSquadByID(ctx, id)
	if err != nil {
		return err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", id.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	if sq.Owner != userID {
		err := fmt.Errorf("user %s is not the owner of squad %s", userID.String(), sq.ID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	return s.squadsStorage.DeleteSquad(ctx, id)
}

func (s *squads) CreateSquadMember(ctx context.Context, req dto.CreateSquadMemeberReq) (dto.SquadMember, error) {
	usr, exist, err := s.usersStorage.GetUserByPhoneNumber(ctx, req.Phone)
	if err != nil {
		return dto.SquadMember{}, err
	}

	if !exist {
		err := fmt.Errorf("user with %s phone dose not exist", req.Phone)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	rm, _ := s.squadsStorage.GetSquadMembersByUserID(ctx, usr.ID)

	if len(rm) > 0 {
		err := fmt.Errorf("user %s is already a member of squad", usr.ID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	resp, exist, err := s.usersStorage.GetUserByID(ctx, usr.ID)
	if err != nil {
		return dto.SquadMember{}, err
	}

	if !exist {
		err := fmt.Errorf("player not found with this user ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	sq, exist, err := s.squadsStorage.GetSquadByID(ctx, req.SquadID)
	if err != nil {
		return dto.SquadMember{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", req.SquadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	if sq.Owner != req.OwnerID {
		err := fmt.Errorf("user %s is not the owner of squad %s", req.OwnerID.String(), sq.ID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMember{}, err
	}

	return s.squadsStorage.CreateSquadMember(ctx, dto.CreateSquadReq{
		SquadID: req.SquadID,
		UserID:  resp.ID,
	})
}

func (s *squads) GetSquadMembersBySquadID(ctx context.Context, req dto.GetSquadMemebersReq) (dto.GetSquadMemebersRes, error) {
	_, exist, err := s.squadsStorage.GetSquadByID(ctx, req.SquadID)
	if err != nil {
		return dto.GetSquadMemebersRes{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", req.SquadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSquadMemebersRes{}, err
	}

	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return s.squadsStorage.GetSquadMembersBySquadID(ctx, req)
}

func (s *squads) DeleteSquadMember(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	sq, err := s.squadsStorage.GetSquadMemberByID(ctx, id)
	if err != nil {
		return err
	}
	if sq == nil || sq.Owner != userID {
		err := fmt.Errorf("user %s is not the owner of squad member %s", userID.String(), id.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	return s.squadsStorage.DeleteSquadMemberByID(ctx, id)
}

func (s *squads) DeleteSquadMembersBySquadID(ctx context.Context, id, userID uuid.UUID) error {
	sq, exist, err := s.squadsStorage.GetSquadByID(ctx, id)
	if err != nil {
		return err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", id.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	if sq.Owner != userID {
		err := fmt.Errorf("user %s is not the owner of squad %s", userID.String(), sq.ID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	return s.squadsStorage.DeleteSquadMembersBySquadID(ctx, id)
}

func (s *squads) GetSquadEarns(ctx context.Context, req dto.GetSquadEarnsReq) (dto.GetSquadEarnsResp, error) {
	_, exist, err := s.squadsStorage.GetSquadByID(ctx, req.SquadID)
	if err != nil {
		return dto.GetSquadEarnsResp{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", req.SquadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSquadEarnsResp{}, err
	}

	return s.squadsStorage.GetSquadEarns(ctx, req)
}

func (s *squads) GetMySquadEarns(ctx context.Context, req dto.GetSquadEarnsReq, UserID uuid.UUID) (dto.GetSquadEarnsResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset
	return s.squadsStorage.GetSquadEarnsByUserID(ctx, req, UserID)
}

func (s *squads) GetSquadTotalEarns(ctx context.Context, squadID uuid.UUID) (decimal.Decimal, error) {
	_, exist, err := s.squadsStorage.GetSquadByID(ctx, squadID)
	if err != nil {
		return decimal.Zero, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", squadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return decimal.Zero, err
	}

	return s.squadsStorage.GetSquadTotalEarns(ctx, squadID)
}

func (s *squads) GetSquadByName(ctx context.Context, name string) (*dto.Squad, error) {
	if name == "" {
		err := fmt.Errorf("invalid squad name: %s", name)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return nil, err
	}

	sq, exist, err := s.squadsStorage.GetSquadByHandle(ctx, name)
	if err != nil {
		return nil, err
	}

	if !exist {
		err := fmt.Errorf("squad with name %s does not exist", name)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return nil, err
	}

	return &sq, nil
}

func (s *squads) GetTornamentStyleRanking(ctx context.Context, req dto.GetTornamentStyleRankingReq) (dto.GetTornamentStyleRankingResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	resp, err := s.squadsStorage.GetTornamentStyleRanking(ctx, req)
	if err != nil {
		return dto.GetTornamentStyleRankingResp{}, err
	}
	if len(resp.Ranking) == 0 {
		response := dto.GetTornamentStyleRankingResp{
			Ranking:    []dto.GetTornamentStyleRank{},
			TotalPages: 0,
		}
		return response, nil
	}
	return resp, nil
}

func (s *squads) CreateTournaments(ctx context.Context, req dto.CreateTournamentReq) (dto.CreateTournamentResp, error) {

	if len(req.Rewards) > 0 {
		valid := false
		for _, reward := range req.Rewards {
			valid = false
			for _, tp := range dto.RewardTypes {
				if reward.Type == tp {
					valid = true
					break
				}
			}

			if !valid {
				msg := ""
				msg = fmt.Sprintf("%s not valid type, valid types are", reward.Type)
				for _, m := range dto.RewardTypes {
					msg = fmt.Sprintf(" %s , %s", m)
				}

				err := fmt.Errorf("%s", msg)
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())

			}
		}
	}

	resp, err := s.squadsStorage.CreateTournaments(ctx, dto.CreateTournamentReq{
		Rank:             req.Rank,
		Level:            req.Level,
		CumulativePoints: req.CumulativePoints,
		Rewards:          req.Rewards,
		CreatedAt:        req.CreatedAt,
		UpdatedAt:        req.CreatedAt,
	})

	if err != nil {
		return dto.CreateTournamentResp{}, err
	}

	return resp, nil
}

func (s *squads) GetTornamentStyles(ctx context.Context) ([]dto.Tournament, error) {
	resp, err := s.squadsStorage.GetTornamentStyles(ctx)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		res := []dto.Tournament{}
		return res, nil
	}

	return resp, nil
}

func (s *squads) GetSquadMembersEarnings(ctx context.Context, req dto.GetSquadMembersEarningsReq, ownerID uuid.UUID) (dto.GetSquadMembersEarningsResp, error) {
	// Validate squad exists and user is the owner
	sq, exist, err := s.squadsStorage.GetSquadByID(ctx, req.SquadID)
	if err != nil {
		return dto.GetSquadMembersEarningsResp{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", req.SquadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSquadMembersEarningsResp{}, err
	}

	if sq.Owner != ownerID {
		err := fmt.Errorf("user %s is not the owner of squad %s", ownerID.String(), req.SquadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.GetSquadMembersEarningsResp{}, err
	}

	// Set default pagination values
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return s.squadsStorage.GetSquadMembersEarnings(ctx, req, ownerID)
}

func (s *squads) LeaveSquad(ctx context.Context, userID, squadID uuid.UUID) error {
	if userID == uuid.Nil || squadID == uuid.Nil {
		err := fmt.Errorf("invalid user ID or squad ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	// check if squad exists
	_, exist, err := s.squadsStorage.GetSquadByID(ctx, squadID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to get squad by ID")
		return err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", squadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	return s.squadsStorage.LeaveSquad(ctx, userID, squadID)
}

func (s *squads) JoinSquad(ctx context.Context, userID, squadID uuid.UUID) (dto.SquadMemberResp, error) {
	if userID == uuid.Nil || squadID == uuid.Nil {
		err := fmt.Errorf("invalid user ID or squad ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMemberResp{}, err
	}

	// check if squad exists
	sqd, exist, err := s.squadsStorage.GetSquadByID(ctx, squadID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to get squad by ID")
		return dto.SquadMemberResp{}, err
	}

	if !exist {
		err := fmt.Errorf("squad with id %s does not exist", squadID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMemberResp{}, err
	}

	if dto.CheckJoinableSquadType(sqd.Type) != nil {
		s.log.Error("squad not ", zap.String("squadType", sqd.Type))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMemberResp{}, err
	}
	// check if user is already a member of the squad
	members, _ := s.squadsStorage.GetSquadMembersByUserID(ctx, userID)

	if len(members) > 0 {
		// return user member of other squad
		err := fmt.Errorf("user already member of other squad")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SquadMemberResp{}, nil
	}

	// if type is SquadTypeRequestToJoin

	if sqd.Type == string(dto.SquadTypeOpen) {
		rs, err := s.squadsStorage.CreateSquadMember(ctx, dto.CreateSquadReq{
			SquadID: squadID,
			UserID:  userID,
		})

		if err != nil {
			return dto.SquadMemberResp{}, err
		}

		return dto.SquadMemberResp{
			Message: constant.SQUAD_CREATED_SUCCESS,
			Data:    rs,
		}, nil
	}

	resp, err := s.squadsStorage.AddToWaitingSquadMembers(ctx, dto.CreateSquadReq{
		SquadID: squadID,
		UserID:  userID,
	})

	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to add user to waiting squad members")
		s.log.Error("failed to add user to waiting squad members", zap.Error(err))
		return dto.SquadMemberResp{}, err
	}

	return dto.SquadMemberResp{
		Message: constant.SQUAD_ADDED_TO_WAITING,
		Data: dto.SquadMember{
			SquadID:   resp.SquadID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (s *squads) SquadTypeRequestToJoinHandler(ctx context.Context, userID, squadID uuid.UUID) (dto.SquadMemberResp, error) {
	resp, err := s.squadsStorage.AddWaitingSquadMember(ctx, userID, squadID)
	if err != nil {
		return dto.SquadMemberResp{}, err
	}

	return dto.SquadMemberResp{
		Message: "successfully added to waiting squad members",
		Data:    resp,
	}, nil
}

func (s *squads) GetWaitingSquadMembers(ctx context.Context, userID uuid.UUID) ([]dto.WaitingSquadMember, error) {
	// Get User squad with type dto.SquadTypeRequestToJoin

	resp, err := s.squadsStorage.GetSquadsByOwner(ctx, userID)
	if err != nil {
		resp := []dto.WaitingSquadMember{}
		s.log.Error("failed to get squads by owner", zap.Error(err), zap.String("userID", userID.String()))
		return resp, err
	}

	if len(resp) == 0 {
		s.log.Info("no squads found for user", zap.String("userID", userID.String()))
		err := fmt.Errorf("no squads found for user %s", userID.String())
		return []dto.WaitingSquadMember{}, err
	}

	members, err := s.squadsStorage.GetWaitingSquadMembers(ctx, resp[0].ID)
	if err != nil {
		resp := []dto.WaitingSquadMember{}
		s.log.Error("failed to get waiting squad members", zap.Error(err), zap.String("squadID", resp[0].ID.String()))
		return resp, err
	}

	if len(members) == 0 {
		s.log.Info("no waiting squad members found", zap.String("squadID", resp[0].ID.String()))
		resp := []dto.WaitingSquadMember{}
		return resp, nil
	}

	return members, nil
}

func (s *squads) RemoveWaitingSquadWaitingUser(ctx context.Context, req dto.SquadMemberReq) error {
	if req.MemberID == uuid.Nil {
		err := fmt.Errorf("invalid member ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	err := s.squadsStorage.DeleteWaitingSquadMember(ctx, req.MemberID)
	if err != nil {
		s.log.Error("failed to delete waiting squad member", zap.Error(err), zap.String("memberID", req.MemberID.String()))
		return err
	}

	return nil
}

func (s *squads) ApproveWaitingSquadMember(ctx context.Context, req dto.SquadMemberReq) error {

	if req.MemberID == uuid.Nil {
		err := fmt.Errorf("invalid member ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	err := s.squadsStorage.ApproveWaitingSquadMember(ctx, req.MemberID)
	if err != nil {
		s.log.Error("failed to approve waiting squad member", zap.Error(err), zap.String("memberID", req.MemberID.String()))
		return err
	}

	return nil
}
