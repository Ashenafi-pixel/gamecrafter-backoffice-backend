package bet

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateLevel(ctx context.Context, level dto.Level) (dto.Level, error) {
	resp, err := b.db.Queries.CreateLevel(ctx, db.CreateLevelParams{
		CreatedBy: level.CreatedBy,
		Level:     decimal.NewFromInt(int64(level.Level)),
		Type:      level.Type,
	})
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableTocreate.Wrap(err, err.Error())
		return dto.Level{}, err
	}
	return dto.Level{
		ID:        resp.ID,
		Level:     int(resp.Level.IntPart()),
		CreatedBy: resp.CreatedBy,
	}, nil
}

func (b *bet) GetLevels(ctx context.Context, req dto.GetRequest) (dto.GetLevelResp, error) {
	var levels []dto.LevelResp

	resp, err := b.db.Queries.GetLevels(ctx, db.GetLevelsParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
		Type:   req.Type,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.GetLevelResp{}, err
	}
	totalPages := 1
	for i, level := range resp {
		requirements := []dto.LevelRequirement{}
		requirementsResp, err := b.db.Queries.GetAallRequirementsByLevelID(ctx, level.ID)
		if err != nil && err.Error() != dto.ErrNoRows {
			b.log.Error(err.Error())
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
			return dto.GetLevelResp{}, err
		}

		for _, requirement := range requirementsResp {
			requirements = append(requirements, dto.LevelRequirement{
				ID:        requirement.ID,
				LevelID:   requirement.LevelID,
				Type:      requirement.Type,
				Value:     requirement.Value,
				CreatedBy: requirement.CreatedBy,
			})
		}
		levels = append(levels, dto.LevelResp{
			ID:           level.ID,
			Level:        int(level.Level.IntPart()),
			CreatedBy:    level.CreatedBy,
			Requirements: requirements,
		})

		if i == 0 {
			totalPages = int(level.TotalRows)
			if int(level.TotalRows)%req.PerPage != 0 {
				totalPages++
			}
		}
	}

	return dto.GetLevelResp{
		Message:   constant.SUCCESS,
		Level:     levels,
		TotalPage: totalPages,
	}, nil
}

func (b *bet) GetLevelByID(ctx context.Context, id uuid.UUID) (dto.Level, error) {
	resp, err := b.db.Queries.GetLevelById(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.Level{}, err
	}

	return dto.Level{
		ID:        resp.ID,
		CreatedBy: resp.CreatedBy,
	}, nil
}

func (b *bet) CreateLevelRequirements(ctx context.Context, req dto.LevelRequirements) (dto.LevelRequirements, error) {
	var respBody []dto.LevelRequirement
	for _, requirement := range req.Requirements {
		if requirement.Type == "" || requirement.Value == "" {
			err := errors.ErrInvalidUserInput.Wrap(nil, "requirement type and value cannot be empty")
			b.log.Error(err.Error())
			continue
		}

		resp, err := b.db.Queries.CreateLevelRequirement(ctx, db.CreateLevelRequirementParams{
			LevelID:   req.LevelID,
			Type:      requirement.Type,
			Value:     requirement.Value,
			CreatedBy: req.CreatedBy,
		})

		if err != nil {
			b.log.Error(err.Error())
			err = errors.ErrUnableTocreate.Wrap(err, err.Error())
			return dto.LevelRequirements{}, err
		}
		respBody = append(respBody, dto.LevelRequirement{
			ID:        resp.ID,
			LevelID:   resp.LevelID,
			Type:      resp.Type,
			Value:     resp.Value,
			CreatedBy: resp.CreatedBy,
		})

	}
	return dto.LevelRequirements{
		LevelID:      req.LevelID,
		Requirements: respBody,
	}, nil
}

func (b *bet) GetLevelRequirementsByLevelID(ctx context.Context, levelID uuid.UUID, req dto.GetRequest) (dto.LevelRequirements, error) {
	resp, err := b.db.Queries.GetLevelRequirementsByLevelID(ctx, db.GetLevelRequirementsByLevelIDParams{
		LevelID: levelID,
		Limit:   int32(req.PerPage),
		Offset:  int32(req.Page),
	})

	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.LevelRequirements{}, err
	}

	var requirementsList []dto.LevelRequirement
	for _, requirement := range resp {
		requirementsList = append(requirementsList, dto.LevelRequirement{
			ID:        requirement.ID,
			LevelID:   requirement.LevelID,
			Type:      requirement.Type,
			Value:     requirement.Value,
			CreatedBy: requirement.CreatedBy,
		})
	}

	return dto.LevelRequirements{
		LevelID:      levelID,
		Requirements: requirementsList,
	}, nil
}

func (b *bet) DeleteLevel(ctx context.Context, levelID uuid.UUID) error {
	_, err := b.db.Queries.DeleteLevel(ctx, levelID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (b *bet) UpdateLevelRequirement(ctx context.Context, req dto.LevelRequirement) (dto.LevelRequirement, error) {
	resp, err := b.db.Queries.UpdateLevelRequirement(ctx, db.UpdateLevelRequirementParams{
		ID:    req.ID,
		Type:  req.Type,
		Value: req.Value,
	})

	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnableToUpdate.Wrap(err, err.Error())
		return dto.LevelRequirement{}, err
	}

	return dto.LevelRequirement{
		ID:        resp.ID,
		LevelID:   resp.LevelID,
		Type:      resp.Type,
		Value:     resp.Value,
		CreatedBy: resp.CreatedBy,
	}, nil
}

func (b *bet) GetAllLevels(ctx context.Context, tp string) ([]dto.Level, error) {
	b.log.Info("Getting all levels from database")

	resp, err := b.db.Queries.GetAllLevels(ctx, tp)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to get all levels from database", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	b.log.Info("Retrieved levels from database", zap.Int("level_count", len(resp)))

	var levels []dto.Level
	for _, level := range resp {
		levels = append(levels, dto.Level{
			ID:        level.ID,
			Level:     int(level.Level.IntPart()),
			CreatedBy: level.CreatedBy,
		})
	}

	return levels, nil
}

func (b *bet) CalculateTotalBetAmount(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	b.log.Info("Calculating total bet amount from database", zap.String("user_id", userID.String()))

	totalBetAmount, err := b.db.Queries.CalculateUserBets(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to calculate total bet amount from database", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return decimal.Zero, err
	}

	b.log.Info("Total bet amount calculated from database",
		zap.String("user_id", userID.String()),
		zap.String("total_bet_amount", totalBetAmount.String()))

	return totalBetAmount, nil
}

func (b *bet) GetAllLevelRequirementsByLevelID(ctx context.Context, levelID uuid.UUID) (dto.LevelRequirements, error) {
	b.log.Info("Getting all level requirements by level ID", zap.String("level_id", levelID.String()))

	resp, err := b.db.Queries.GetAallRequirementsByLevelID(ctx, levelID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to get level requirements by level ID", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.LevelRequirements{}, err
	}

	b.log.Info("Retrieved level requirements from database",
		zap.String("level_id", levelID.String()),
		zap.Int("requirement_count", len(resp)))

	var requirements []dto.LevelRequirement
	for _, requirement := range resp {
		requirements = append(requirements, dto.LevelRequirement{
			ID:        requirement.ID,
			LevelID:   requirement.LevelID,
			Type:      requirement.Type,
			Value:     requirement.Value,
			CreatedBy: requirement.CreatedBy,
		})
	}

	return dto.LevelRequirements{
		LevelID:      levelID,
		Requirements: requirements,
	}, nil
}

func (b *bet) CalculateSquadBets(ctx context.Context, squadID uuid.UUID) (decimal.Decimal, error) {
	b.log.Info("Calculating total bet amount for squad", zap.String("squad_id", squadID.String()))

	totalBetAmount, err := b.db.Queries.CalculateSquadBets(ctx, squadID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to calculate total bet amount for squad", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return decimal.Zero, err
	}

	b.log.Info("Total bet amount calculated for squad",
		zap.String("squad_id", squadID.String()),
		zap.String("total_bet_amount", totalBetAmount.String()))

	return totalBetAmount, nil
}

func (b *bet) GetUserSquads(ctx context.Context, userID uuid.UUID) ([]dto.Squad, error) {
	b.log.Info("Getting squads for user", zap.String("user_id", userID.String()))

	squads, err := b.db.Queries.GetUserSquads(ctx, userID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to get squads for user", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	var resp []dto.Squad
	for _, squad := range squads {
		resp = append(resp, dto.Squad{
			ID: squad,
		})
	}
	if len(resp) == 0 {
		sq, err := b.db.GetSquadIDByOwnerID(ctx, userID)
		if err != nil && err.Error() != dto.ErrNoRows {
			b.log.Error("Failed to get squads for user", zap.Error(err))
			err = errors.ErrUnableToGet.Wrap(err, err.Error())
			return nil, err
		}
		if len(sq) > 0 {
			resp = append(resp, dto.Squad{
				ID: sq,
			},
			)
			return resp, nil
		}
	}

	return resp, nil
}

func (b *bet) GetAllSquadMembersBySquadId(ctx context.Context, squadID uuid.UUID) ([]dto.SquadMember, error) {
	b.log.Info("Getting all squad members by squad ID", zap.String("squad_id", squadID.String()))

	members, err := b.db.Queries.GetAllSquadMembersBySquadId(ctx, squadID)
	if err != nil && err.Error() != dto.ErrNoRows {
		b.log.Error("Failed to get squad members by squad ID", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return nil, err
	}

	b.log.Info("Retrieved squad members from database",
		zap.String("squad_id", squadID.String()),
		zap.Int("member_count", len(members)))

	var resp []dto.SquadMember
	for _, member := range members {
		resp = append(resp, dto.SquadMember{
			ID:        member.ID,
			UserID:    member.UserID,
			SquadID:   member.SquadID,
			CreatedAt: member.CreatedAt,
		})
	}

	return resp, nil
}
