package operator

import (
	"context"

	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type operatorStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func NewOperatorStorage(db *persistencedb.PersistenceDB, log *zap.Logger) *operatorStorage {
	return &operatorStorage{
		db:  db,
		log: log,
	}
}

// AssignGamesToOperator creates operator_games rows for the given game IDs.
func (s *operatorStorage) AssignGamesToOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if len(gameIDs) == 0 {
		return nil
	}

	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		return errors.ErrUnableTocreate.Wrap(err, "unable to start transaction for assigning games to operator")
	}
	defer tx.Rollback(ctx)

	for _, gid := range gameIDs {
		if gid == "" {
			continue
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO operator_games (operator_id, game_id)
			VALUES ($1, $2::uuid)
			ON CONFLICT (operator_id, game_id) DO NOTHING
		`, operatorID, gid)
		if err != nil {
			s.log.Error("unable to assign game to operator", zap.Error(err),
				zap.Int32("operator_id", operatorID), zap.String("game_id", gid))
			return errors.ErrUnableTocreate.Wrap(err, "unable to assign game to operator")
		}
	}

	return tx.Commit(ctx)
}

// RevokeGamesFromOperator removes operator_games rows for the given game IDs.
func (s *operatorStorage) RevokeGamesFromOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if len(gameIDs) == 0 {
		return nil
	}

	_, err := s.db.GetPool().Exec(ctx, `
		DELETE FROM operator_games
		WHERE operator_id = $1 AND game_id = ANY($2::uuid[])
	`, operatorID, gameIDs)
	if err != nil {
		s.log.Error("unable to revoke games from operator", zap.Error(err),
			zap.Int32("operator_id", operatorID))
		return errors.ErrDBDelError.Wrap(err, "unable to revoke games from operator")
	}

	return nil
}

// Ensure interface compliance at compile time.
var _ storage.Operator = (*operatorStorage)(nil)

