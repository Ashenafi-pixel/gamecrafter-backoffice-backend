package operationalgrouptype

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type operationalgrouptype struct {
	operationalGroupTypeStorage storage.OperationalGroupType
	log                         *zap.Logger
}

func Init(operationalgrouptypeStorage storage.OperationalGroupType, log *zap.Logger) module.OperationalGroupType {
	return &operationalgrouptype{
		operationalGroupTypeStorage: operationalgrouptypeStorage,
		log:                         log,
	}
}
func (opt *operationalgrouptype) CreateOperationalGroupType(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, error) {
	// validate operational group type
	if err := dto.ValidateOperationalGroupType(optReq); err != nil {
		opt.log.Warn(err.Error(), zap.Any("optReq", optReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.OperationalGroupType{}, err
	}

	// validate group id
	if optReq.GroupID == uuid.Nil {
		err := fmt.Errorf("invalid operational  group id ")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.OperationalGroupType{}, err
	}

	// check if operational group type for  operational group exist
	_, exist, err := opt.operationalGroupTypeStorage.GetOperationalGroupByGroupIDandName(ctx, optReq)
	if err != nil {
		return dto.OperationalGroupType{}, err
	}
	if exist {
		err = fmt.Errorf("operational group type with this name already exist to this operational group")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.OperationalGroupType{}, err
	}

	//create Operational Group type
	return opt.operationalGroupTypeStorage.CreateOperationalType(ctx, optReq)

}

func (opt *operationalgrouptype) GetOperationalGroupTypeByGroupID(ctx context.Context, groupID uuid.UUID) ([]dto.OperationalGroupType, error) {
	if groupID == uuid.Nil {
		err := fmt.Errorf("empty group id is given")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return []dto.OperationalGroupType{}, err
	}
	groupTypes, _, err := opt.operationalGroupTypeStorage.GetOperationalGroupTypeByGroupID(ctx, groupID)
	if err != nil {
		return []dto.OperationalGroupType{}, err
	}
	return groupTypes, nil
}

func (opt *operationalgrouptype) GetOperationalGroupTypes(ctx context.Context) ([]dto.OperationalTypesRes, error) {
	operationalTypes := []dto.OperationalTypesRes{}
	groupTypes, _, err := opt.operationalGroupTypeStorage.GetOperationalGroupTypes(ctx)

	if err != nil {
		opt.log.Error(err.Error())
		err = errors.ErrUnableToGet.Wrap(err, err.Error())
		return []dto.OperationalTypesRes{}, err
	}

	for _, opts := range groupTypes {
		operationalTypes = append(operationalTypes, dto.OperationalTypesRes{
			ID:   opts.ID,
			Name: opts.Name,
		})
	}

	return operationalTypes, nil
}
