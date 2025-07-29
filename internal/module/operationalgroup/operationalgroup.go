package operationalgroup

import (
	"context"
	"fmt"

	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"go.uber.org/zap"
)

type operational_group struct {
	operationalGroupStorage storage.OperationalGroup
	log                     *zap.Logger
}

func Init(operationalGroupStorage storage.OperationalGroup, log *zap.Logger) module.OperationalGroup {

	return &operational_group{
		operationalGroupStorage: operationalGroupStorage,
		log:                     log,
	}
}
func (op *operational_group) CreateOperationalGroup(ctx context.Context, opReq dto.OperationalGroup) (dto.OperationalGroup, error) {
	// validate user request
	if err := dto.ValidateOperationalGroup(opReq); err != nil {
		op.log.Warn(err.Error(), zap.Any("createOperationalGroupRequest", opReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.OperationalGroup{}, err
	}

	// check if the operational Group with this name already exist
	_, exist, err := op.operationalGroupStorage.GetOperationalGroupByName(ctx, opReq.Name)
	if err != nil {
		return dto.OperationalGroup{}, err
	}
	if exist {
		err = fmt.Errorf("operational group already exist")
		op.log.Warn(err.Error(), zap.Any("operationalGroupReq", opReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.OperationalGroup{}, err
	}

	// create Operational Group
	return op.operationalGroupStorage.CreateOperationalGroup(ctx, opReq)
}

func (op *operational_group) GetOperationalGroups(ctx context.Context) ([]dto.OperationalGroup, error) {
	opGroup, _, err := op.operationalGroupStorage.GetOperationalGroups(ctx)
	return opGroup, err
}
