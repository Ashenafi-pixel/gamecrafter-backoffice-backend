package operationalgroup

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/db"
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type operational_group struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.OperationalGroup {
	return &operational_group{
		db:  db,
		log: log,
	}
}

func (o *operational_group) CreateOperationalGroup(ctx context.Context, op dto.OperationalGroup) (dto.OperationalGroup, error) {
	opRet, err := o.db.Queries.CreateOperationalGroup(ctx, db.CreateOperationalGroupParams{
		Name:        sql.NullString{String: op.Name, Valid: true},
		Description: sql.NullString{String: op.Description, Valid: true},
		CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		o.log.Error(err.Error(), zap.Any("op", op))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.OperationalGroup{}, err
	}
	return dto.OperationalGroup{
		ID:          opRet.ID,
		Name:        opRet.Name.String,
		Description: opRet.Description.String,
		CreatedAt:   opRet.CreatedAt.Time,
	}, nil
}

func (o *operational_group) GetOperationalGroupByName(ctx context.Context, name string) (dto.OperationalGroup, bool, error) {
	opRes, err := o.db.Queries.GetOperationalGroupByName(ctx, sql.NullString{String: name, Valid: true})
	if err != nil && err.Error() != dto.ErrNoRows {
		o.log.Error(err.Error(), zap.String("name", name))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.OperationalGroup{}, false, err
	}
	if err != nil && err.Error() == dto.ErrNoRows {
		return dto.OperationalGroup{}, false, nil
	}
	return dto.OperationalGroup{
		ID:          opRes.ID,
		Name:        opRes.Name.String,
		Description: opRes.Description.String,
		CreatedAt:   opRes.CreatedAt.Time,
	}, true, nil
}

func (o *operational_group) GetOperationalGroups(ctx context.Context) ([]dto.OperationalGroup, bool, error) {
	opGroupsRes := []dto.OperationalGroup{}
	opGroups, err := o.db.Queries.GetOperationalGroups(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		o.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return []dto.OperationalGroup{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return []dto.OperationalGroup{}, false, nil
	}
	for _, opGroup := range opGroups {
		opGroupsRes = append(opGroupsRes, dto.OperationalGroup{
			ID:          opGroup.ID,
			Name:        opGroup.Name.String,
			Description: opGroup.Description.String,
			CreatedAt:   opGroup.CreatedAt.Time,
		})
	}
	return opGroupsRes, true, nil
}

func (o *operational_group) GetOperationalGroupByID(ctx context.Context, groupID uuid.UUID) (dto.OperationalGroup, bool, error) {
	opg, err := o.db.Queries.GetOperationalGroupByID(ctx, groupID)
	if err != nil && err.Error() != dto.ErrNoRows {
		o.log.Error(err.Error(), zap.Any("groupID", groupID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.OperationalGroup{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return dto.OperationalGroup{}, false, nil
	}
	return dto.OperationalGroup{
		ID:          opg.ID,
		Name:        opg.Name.String,
		Description: opg.Description.String,
		CreatedAt:   opg.CreatedAt.Time,
	}, true, nil
}
