package operationdefinition

import (
	"context"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type operationdefinition struct {
	operationalGroupStorage     storage.OperationalGroup
	operationalGroupTypeStorage storage.OperationalGroupType
	log                         *zap.Logger
}

func Init(operationalGroupStorage storage.OperationalGroup, operationGroupTypeStorage storage.OperationalGroupType, log *zap.Logger) module.OperationsDefinitions {
	return &operationdefinition{
		operationalGroupStorage:     operationalGroupStorage,
		operationalGroupTypeStorage: operationGroupTypeStorage,
		log:                         log,
	}
}
func (od *operationdefinition) GetOperationalDefinition(ctx context.Context) (dto.OperationsDefinition, error) {
	operationalGroupDefinitions := dto.OperationsDefinition{}
	// get operational groups
	oprationalGroups, exist, err := od.operationalGroupStorage.GetOperationalGroups(ctx)
	if err != nil {
		return dto.OperationsDefinition{}, err
	}
	if !exist {
		return dto.OperationsDefinition{}, nil
	}
	operationalGroupDefinitions.Data.Groups = oprationalGroups

	// get operational Group Types
	operationalGroupTypes, _, err := od.operationalGroupTypeStorage.GetOperationalGroupTypes(ctx)
	if err != nil {
		return dto.OperationsDefinition{}, err
	}
	operationalGroupDefinitions.Data.Types = operationalGroupTypes
	operationalGroupDefinitions.Status = "success"
	return operationalGroupDefinitions, nil

}
