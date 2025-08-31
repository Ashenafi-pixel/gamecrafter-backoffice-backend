package bet

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (b *bet) CreateLevel(ctx context.Context, level dto.Level) (dto.Level, error) {

	if level.Type == "" || (level.Type != constant.LEVEL_TYPE_PLAYER && level.Type != constant.LEVEL_TYPE_SQUAD) {
		err := fmt.Errorf("invalid level type %s", level.Type)
		b.log.Error("invalid level type", zap.String("type", level.Type))
		return dto.Level{}, err
	}

	return b.betStorage.CreateLevel(ctx, level)
}

func (b *bet) GetLevels(ctx context.Context, req dto.GetRequest) (dto.GetLevelResp, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	return b.betStorage.GetLevels(ctx, req)
}

func (b *bet) CreateLevelReqirements(ctx context.Context, req dto.LevelRequirements) (dto.LevelRequirements, error) {

	lv, err := b.betStorage.GetLevelByID(ctx, req.LevelID)
	if err != nil {
		return dto.LevelRequirements{}, err
	}

	if lv.ID == uuid.Nil {
		err = fmt.Errorf("level with ID %s does not exist", req.LevelID)
		b.log.Error("level does not exist", zap.String("level_id", req.LevelID.String()))
		return dto.LevelRequirements{}, err
	}

	for _, requirement := range req.Requirements {
		if requirement.Type != dto.LevelRequirementTypeBetAmount {
			err = fmt.Errorf("invalid requirement type %s", requirement.Type)
			b.log.Error("invalid requirement type", zap.String("type", requirement.Type))
			return dto.LevelRequirements{}, err
		}
	}

	return b.betStorage.CreateLevelRequirements(ctx, req)
}

func (b *bet) UpdateLevelRequirement(ctx context.Context, req dto.UpdateLevelRequirementReq) (dto.LevelRequirement, error) {
	if req.Type != dto.LevelRequirementTypeBetAmount {
		err := fmt.Errorf("invalid requirement type %s", req.Type)
		b.log.Error("invalid requirement type", zap.String("type", req.Type))
		return dto.LevelRequirement{}, err
	}

	return b.betStorage.UpdateLevelRequirement(ctx, dto.LevelRequirement{
		ID:    req.ID,
		Type:  req.Type,
		Value: req.Value,
	})
}

func SortLevels(levels []dto.Level) []dto.Level {
	sortedLevels := make([]dto.Level, len(levels))
	copy(sortedLevels, levels)

	sort.Slice(sortedLevels, func(i, j int) bool {
		return sortedLevels[i].Level < sortedLevels[j].Level
	})

	return sortedLevels
}

func (b *bet) GetUserLevel(ctx context.Context, userID uuid.UUID) (dto.GetUserLevelResp2, error) {
	if userID == uuid.Nil {
		err := fmt.Errorf("invalid user ID")
		b.log.Error("invalid user ID", zap.String("user_id", userID.String()))
		return dto.GetUserLevelResp2{}, err
	}

	b.log.Info("Getting user level", zap.String("user_id", userID.String()))

	levels, err := b.betStorage.GetAllLevels(ctx, constant.LEVEL_TYPE_PLAYER)
	if err != nil {
		b.log.Error("Failed to get all levels", zap.Error(err))
		return dto.GetUserLevelResp2{}, err
	}

	b.log.Info("Retrieved levels from database", zap.Int("level_count", len(levels)))

	sortedLevels := SortLevels(levels)
	if len(sortedLevels) == 0 {
		b.log.Info("No levels configured in database, returning default level 0")
		return dto.GetUserLevelResp2{
			ID:    uuid.New(),
			Level: 0,
		}, nil
	}

	var highestLevel dto.GetUserLevelResp2
	highestLevel.Level = -1
	totalBetAmount, err := b.CalculateTotalBetAmount(ctx, userID)
	if err != nil {
		b.log.Error("unable to calculate total bet amount", zap.Error(err))
		return dto.GetUserLevelResp2{}, err
	}
	var nextLevelRequirments decimal.Decimal
	var currentLevelRequirements decimal.Decimal
	nextLevel := 1

	for _, level := range sortedLevels {
		b.log.Info("Processing level",
			zap.String("level_id", level.ID.String()),
			zap.Int("level_number", level.Level))

		requirements, err := b.betStorage.GetAllLevelRequirementsByLevelID(ctx, level.ID)
		if err != nil {
			b.log.Error("unable to get level requirements", zap.Error(err))
			return dto.GetUserLevelResp2{}, err
		}

		b.log.Info("Retrieved level requirements",
			zap.String("level_id", level.ID.String()),
			zap.Int("requirement_count", len(requirements.Requirements)))

		meetsLevel := true
		for _, requirement := range requirements.Requirements {
			if requirement.Type == dto.LevelRequirementTypeBetAmount {

				b.log.Info("Calculated total bet amount",
					zap.String("user_id", userID.String()),
					zap.String("total_bet_amount", totalBetAmount.String()),
					zap.String("requirement_value", requirement.Value))

				amount, err := strconv.Atoi(requirement.Value)
				if err != nil {
					b.log.Error("invalid requirement value", zap.String("value", requirement.Value), zap.Error(err))
					return dto.GetUserLevelResp2{}, fmt.Errorf("invalid requirement value: %s", requirement.Value)
				}
				nextLevelRequirments = decimal.NewFromInt(int64(amount))

				if !totalBetAmount.GreaterThanOrEqual(decimal.NewFromInt(int64(amount))) {
					b.log.Info("User does not meet level requirement",
						zap.String("user_id", userID.String()),
						zap.String("total_bet_amount", totalBetAmount.String()),
						zap.Int("required_amount", amount))
					meetsLevel = false
					nextLevel = level.Level
					break
				} else {
					currentLevelRequirements = decimal.NewFromInt(int64(amount))
				}

				b.log.Info("User meets level requirement",
					zap.String("user_id", userID.String()),
					zap.String("total_bet_amount", totalBetAmount.String()),
					zap.Int("required_amount", amount),
					zap.Int("level", level.Level))
			}
		}

		if meetsLevel && level.Level > highestLevel.Level {
			current := totalBetAmount.Sub(currentLevelRequirements)
			next := nextLevelRequirments.Sub(currentLevelRequirements)
			highestLevel = dto.GetUserLevelResp2{
				ID:                      level.ID,
				Level:                   level.Level,
				NextLevel:               nextLevel,
				Bucks:                   totalBetAmount,
				AmountSpentToReachLevel: current,
				NextLevelRequirement:    next,
			}
		} else {
			break
		}

	}

	current := totalBetAmount.Sub(currentLevelRequirements)
	next := nextLevelRequirments.Sub(currentLevelRequirements)
	if highestLevel.Level == -1 {

		b.log.Info("User does not meet any level requirements, returning default level 0")
		return dto.GetUserLevelResp2{
			ID:                      uuid.New(),
			Level:                   0,
			Bucks:                   totalBetAmount,
			NextLevel:               nextLevel,
			AmountSpentToReachLevel: current,
			NextLevelRequirement:    next,
		}, nil
	}

	b.log.Info("Returning user level",
		zap.String("user_id", userID.String()),
		zap.Int("level", highestLevel.Level))
	if highestLevel.Level == sortedLevels[len(sortedLevels)-1].Level {
		highestLevel.IsFinalLevel = true
	} else {
		highestLevel.IsFinalLevel = false
	}
	highestLevel.AmountSpentToReachLevel = current
	highestLevel.NextLevelRequirement = next
	highestLevel.Bucks = totalBetAmount

	return highestLevel, nil
}

func (b *bet) CalculateTotalBetAmount(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	if userID == uuid.Nil {
		err := fmt.Errorf("invalid user ID")
		b.log.Error("invalid user ID", zap.String("user_id", userID.String()))
		return decimal.Zero, err
	}

	b.log.Info("Calculating total bet amount", zap.String("user_id", userID.String()))

	totalBetAmount, err := b.betStorage.CalculateTotalBetAmount(ctx, userID)
	if err != nil {
		b.log.Error("unable to calculate total bet amount", zap.Error(err))
		return decimal.Zero, err
	}

	b.log.Info("Total bet amount calculated",
		zap.String("user_id", userID.String()),
		zap.String("total_bet_amount", totalBetAmount.String()))

	return totalBetAmount, nil
}

func (b *bet) GetSquadLevel(ctx context.Context, squadID uuid.UUID) (dto.GetUserLevelResp2, error) {
	if squadID == uuid.Nil {
		err := fmt.Errorf("invalid user ID")
		b.log.Error("invalid user ID", zap.String("user_id", squadID.String()))
		return dto.GetUserLevelResp2{}, err
	}

	b.log.Info("Getting user level", zap.String("user_id", squadID.String()))

	levels, err := b.betStorage.GetAllLevels(ctx, constant.LEVEL_TYPE_SQUAD)
	if err != nil {
		b.log.Error("Failed to get all levels", zap.Error(err))
		return dto.GetUserLevelResp2{}, err
	}

	b.log.Info("Retrieved levels from database", zap.Int("level_count", len(levels)))

	sortedLevels := SortLevels(levels)
	if len(sortedLevels) == 0 {
		b.log.Info("No levels configured in database, returning default level 0")
		return dto.GetUserLevelResp2{
			ID:    uuid.New(),
			Level: 0,
		}, nil
	}

	var highestLevel dto.GetUserLevelResp2
	highestLevel.Level = -1
	totalBetAmount, err := b.CalculateSquadBets(ctx, squadID)
	if err != nil {
		b.log.Error("unable to calculate total bet amount", zap.Error(err))
		return dto.GetUserLevelResp2{}, err
	}
	var nextLevelRequirments decimal.Decimal
	var currentLevelRequirements decimal.Decimal
	nextLevel := 1

	for _, level := range sortedLevels {
		b.log.Info("Processing level",
			zap.String("level_id", level.ID.String()),
			zap.Int("level_number", level.Level))

		requirements, err := b.betStorage.GetAllLevelRequirementsByLevelID(ctx, level.ID)
		if err != nil {
			b.log.Error("unable to get level requirements", zap.Error(err))
			return dto.GetUserLevelResp2{}, err
		}

		b.log.Info("Retrieved level requirements",
			zap.String("level_id", level.ID.String()),
			zap.Int("requirement_count", len(requirements.Requirements)))

		meetsLevel := true
		for _, requirement := range requirements.Requirements {
			if requirement.Type == dto.LevelRequirementTypeBetAmount {

				b.log.Info("Calculated total bet amount",
					zap.String("user_id", squadID.String()),
					zap.String("total_bet_amount", totalBetAmount.String()),
					zap.String("requirement_value", requirement.Value))

				amount, err := strconv.Atoi(requirement.Value)
				if err != nil {
					b.log.Error("invalid requirement value", zap.String("value", requirement.Value), zap.Error(err))
					return dto.GetUserLevelResp2{}, fmt.Errorf("invalid requirement value: %s", requirement.Value)
				}
				nextLevelRequirments = decimal.NewFromInt(int64(amount))

				if !totalBetAmount.GreaterThanOrEqual(decimal.NewFromInt(int64(amount))) {
					b.log.Info("User does not meet level requirement",
						zap.String("user_id", squadID.String()),
						zap.String("total_bet_amount", totalBetAmount.String()),
						zap.Int("required_amount", amount))
					meetsLevel = false
					nextLevel = level.Level
					break
				} else {
					currentLevelRequirements = decimal.NewFromInt(int64(amount))
				}

				b.log.Info("User meets level requirement",
					zap.String("user_id", squadID.String()),
					zap.String("total_bet_amount", totalBetAmount.String()),
					zap.Int("required_amount", amount),
					zap.Int("level", level.Level))
			}
		}

		if meetsLevel && level.Level > highestLevel.Level {
			current := totalBetAmount.Sub(currentLevelRequirements)
			next := nextLevelRequirments.Sub(currentLevelRequirements)
			if !highestLevel.IsFinalLevel {
				nextLevel = level.Level + 1
			}
			highestLevel = dto.GetUserLevelResp2{
				ID:                      level.ID,
				Level:                   level.Level,
				NextLevel:               nextLevel,
				Bucks:                   totalBetAmount,
				AmountSpentToReachLevel: current,
				NextLevelRequirement:    next,
				SquadID:                 squadID,
			}
		} else {
			break
		}

	}

	current := totalBetAmount.Sub(currentLevelRequirements)
	next := nextLevelRequirments.Sub(currentLevelRequirements)
	if highestLevel.Level == -1 {

		b.log.Info("User does not meet any level requirements, returning default level 0")
		return dto.GetUserLevelResp2{
			ID:                      uuid.New(),
			Level:                   0,
			Bucks:                   totalBetAmount,
			NextLevel:               nextLevel,
			AmountSpentToReachLevel: current,
			NextLevelRequirement:    next,
			SquadID:                 squadID,
		}, nil
	}

	b.log.Info("Returning user level",
		zap.String("user_id", squadID.String()),
		zap.Int("level", highestLevel.Level))
	if highestLevel.Level == sortedLevels[len(sortedLevels)-1].Level {
		highestLevel.IsFinalLevel = true
	} else {
		highestLevel.IsFinalLevel = false
	}
	highestLevel.AmountSpentToReachLevel = current
	highestLevel.NextLevelRequirement = next
	highestLevel.Bucks = totalBetAmount
	highestLevel.SquadID = squadID

	return highestLevel, nil
}

func (b *bet) CalculateSquadBets(ctx context.Context, squadID uuid.UUID) (decimal.Decimal, error) {
	if squadID == uuid.Nil {
		err := fmt.Errorf("invalid user ID")
		b.log.Error("invalid user ID", zap.String("squad_id", squadID.String()))
		return decimal.Zero, err
	}

	b.log.Info("Calculating total bet amount", zap.String("squad_id", squadID.String()))

	totalBetAmount, err := b.betStorage.CalculateSquadBets(ctx, squadID)
	if err != nil {
		b.log.Error("unable to calculate total bet amount", zap.Error(err))
		return decimal.Zero, err
	}

	b.log.Info("Total bet amount calculated",
		zap.String("squad_id", squadID.String()),
		zap.String("total_bet_amount", totalBetAmount.String()))

	return totalBetAmount, nil
}
