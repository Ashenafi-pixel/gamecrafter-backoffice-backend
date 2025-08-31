package operationalgrouptype

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type operationalgrouptype struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.OperationalGroupType {
	return &operationalgrouptype{
		db:  db,
		log: log,
	}
}
func (opt *operationalgrouptype) CreateOperationalType(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, error) {
	optRes, err := opt.db.Queries.CreateOperationalGroupType(ctx, db.CreateOperationalGroupTypeParams{
		GroupID:     optReq.GroupID,
		Name:        sql.NullString{String: optReq.Name, Valid: true},
		Description: sql.NullString{String: optReq.Description, Valid: true},
		CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		opt.log.Error(err.Error(), zap.Any("optReq", optReq))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.OperationalGroupType{}, err
	}
	return dto.OperationalGroupType{
		ID:          optRes.ID,
		GroupID:     optRes.GroupID,
		Name:        optRes.Name.String,
		Description: optRes.Description.String,
		CreatedAt:   optReq.CreatedAt,
	}, nil

}

func (opt *operationalgrouptype) GetOperationalGroupByGroupIDandName(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, bool, error) {
	optRes, err := opt.db.Queries.GetOperationalTtypeByGoupIDandOperationalTypeName(ctx, db.GetOperationalTtypeByGoupIDandOperationalTypeNameParams{
		GroupID: optReq.GroupID,
		Name:    sql.NullString{String: optReq.Name, Valid: true},
	})
	if err != nil && err.Error() != dto.ErrNoRows {
		opt.log.Error(err.Error(), zap.Any("optReq", optReq))
		err = errors.ErrUnableTocreate.Wrap(err, err.Error(), zap.Any("optReq", optReq))
		return dto.OperationalGroupType{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return dto.OperationalGroupType{}, false, nil
	}
	return dto.OperationalGroupType{
		ID:          optRes.ID,
		GroupID:     optReq.GroupID,
		Name:        optRes.Name.String,
		Description: optRes.Description.String,
		CreatedAt:   optRes.CreatedAt.Time,
	}, true, nil
}

func (opt *operationalgrouptype) GetOperationalGroupTypeByGroupID(ctx context.Context, groupID uuid.UUID) ([]dto.OperationalGroupType, bool, error) {
	optTypes := []dto.OperationalGroupType{}
	optRes, err := opt.db.Queries.GetOperationalTypesByGroupID(ctx, groupID)
	if err != nil && err.Error() != dto.ErrNoRows {
		opt.log.Error(err.Error(), zap.Any("group_id", groupID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.OperationalGroupType{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return []dto.OperationalGroupType{}, false, nil
	}
	for _, optr := range optRes {
		optTypes = append(optTypes, dto.OperationalGroupType{
			ID:          optr.ID,
			GroupID:     optr.GroupID,
			Name:        optr.Name.String,
			Description: optr.Description.String,
			CreatedAt:   optr.CreatedAt.Time,
		})
	}

	return optTypes, true, nil
}

func (opt *operationalgrouptype) GetOperationalGroupTypes(ctx context.Context) ([]dto.OperationalGroupType, bool, error) {
	operationGroupType := []dto.OperationalGroupType{}
	optRes, err := opt.db.Queries.GetOperationalGroupTypes(ctx)
	if err != nil && err.Error() != dto.ErrNoRows {
		opt.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.OperationalGroupType{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return []dto.OperationalGroupType{}, false, nil
	}
	for _, optg := range optRes {
		operationGroupType = append(operationGroupType, dto.OperationalGroupType{
			ID:          optg.ID,
			GroupID:     optg.GroupID,
			Name:        optg.Name.String,
			Description: optg.Description.String,
			CreatedAt:   optg.CreatedAt.Time,
		})
	}
	return operationGroupType, true, nil
}

func (opt *operationalgrouptype) GetOperationalGroupTypeByID(ctx context.Context, groupTypeID uuid.UUID) (dto.OperationalGroupType, bool, error) {
	opgt, err := opt.db.Queries.GetOperationalGroupTypeByID(ctx, groupTypeID)
	if err != nil && err.Error() != dto.ErrNoRows {
		opt.log.Error(err.Error(), zap.Any("groupTypeID", groupTypeID))
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return dto.OperationalGroupType{}, false, err
	} else if err != nil && err.Error() == dto.ErrNoRows {
		return dto.OperationalGroupType{}, false, nil
	}
	return dto.OperationalGroupType{
		ID:          opgt.ID,
		GroupID:     opgt.GroupID,
		Name:        opgt.Name.String,
		Description: opgt.Description.String,
		CreatedAt:   opgt.CreatedAt.Time,
	}, true, nil
}
