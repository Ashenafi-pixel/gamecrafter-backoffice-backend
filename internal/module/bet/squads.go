package bet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
)

func (b *bet) SaveToSquads(ctx context.Context, req dto.SquadEarns) error {

	// get current level of squads

	squs, err := b.squadsStorage.GetUserSquads(ctx, req.UserID)
	if err != nil {
		return err
	}
	// get all tournaments
	for _, sq := range squs {
		b.CheckRankAndClaim(ctx, req)
		req.SquadID = sq.ID
		b.squadsStorage.AddSquadEarn(ctx, req)
	}

	return nil
}

func (b *bet) CheckRankAndClaim(ctx context.Context, req dto.SquadEarns) error {
	// get current level of squads
	squs, err := b.squadsStorage.GetUserSquads(ctx, req.UserID)
	if err != nil {
		return err
	}
	// get all tournaments
	tournaments, err := b.squadsStorage.GetTornamentStyles(ctx)
	if err != nil {
	}

	for _, sq := range squs {
		for _, tournament := range tournaments {

			// check if already claimed
			claim, err := b.squadsStorage.GetTournamentClaimBySquadID(ctx, tournament.ID, sq.ID)
			if err != nil {
				return err
			}
			// if the claim is not empty, continue to the next tournament
			if claim.ID != (dto.TournamentClaim{}).ID {
				continue
			}

			// get the squad earns for the tournament
			earn, err := b.squadsStorage.GetUserEarnsForSquad(ctx, sq.ID, req.UserID)
			if err != nil {
				return err
			}

			// check if the earn is greater than or equal to the tournament's required points
			if earn.GreaterThanOrEqual(decimal.NewFromInt(int64(tournament.CumulativePoints))) {

				// update user balance of all squad members
				squadMembers, err := b.squadsStorage.GetSquadMembersBySquadID(ctx, dto.GetSquadMemebersReq{
					Page:    1,
					PerPage: 1000,
					SquadID: sq.ID,
				})

				if err != nil {
					return err
				}
				for _, member := range squadMembers.SquadMemebers {

					for _, reward := range tournament.Rewards {
						if reward.Type == "bucks" {
							// update user balance
							balance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
								UserId:       member.UserID,
								CurrencyCode: constant.POINT_CURRENCY,
							})
							if err != nil {
								continue
							}

							if !exist {
								balance = dto.Balance{
									UserId:       member.UserID,
									CurrencyCode: constant.POINT_CURRENCY,
									AmountUnits:  decimal.Zero,
								}
							}

							operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
							if err != nil {
								b.log.Error(err.Error())
								err = errors.ErrInternalServerError.Wrap(err, err.Error())
								return err
							}

							// save operations logs
							transactionID := utils.GenerateTransactionId()
							_, err = b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
								UserID:             member.UserID,
								Component:          constant.REAL_MONEY,
								Currency:           constant.POINT_CURRENCY,
								Description:        fmt.Sprintf("squad %s claimed tournament %s reward %s", sq.Handle, tournament.Rank, reward.Type),
								ChangeAmount:       reward.Amount,
								OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
								OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
								BalanceAfterUpdate: &balance.AmountUnits,
								TransactionID:      &transactionID,
							})
							if err != nil {
								return err
							}

						}
					}

				}
			}

		}
	}

	return nil
}

func (s *bet) AddFakeBalanceLog(ctx context.Context, userID uuid.UUID, changeAmount decimal.Decimal, currency string) error {
	if userID == uuid.Nil {
		err := fmt.Errorf("invalid user ID")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	if changeAmount.IsZero() {
		err := fmt.Errorf("change amount cannot be zero")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	if currency == "" {
		err := fmt.Errorf("currency cannot be empty")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}

	err := s.betStorage.AddFakeBalanceLog(ctx, userID, changeAmount, currency)
	if err != nil {
		return err
	}

	s.TriggerLevelResponse(ctx, userID)
	return nil
}
