package operationalgrouptype

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type operationalGroupType struct {
	operationalGroupModule module.OperationalGroupType
	log                    *zap.Logger
}

func Init(operationalGroupTypeModule module.OperationalGroupType, log *zap.Logger) handler.OperationalGroupType {
	return &operationalGroupType{
		operationalGroupModule: operationalGroupTypeModule,
		log:                    log,
	}
}

// CreateOperationalGroupType Creating Operational type requests.
//	@Summary		CreateOperationalType
//	@Description	Create Operational Type using name and description, only for admins
//	@Tags			OperationalType
//	@Accept			json
//	@Produce		json
//	@Param			optReq	body		dto.OperationalGroupType	true	"Create Operational Type Request"
//	@Success		201		{object}	dto.OperationalGroup
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/operationalgrouptype [post]
func (opt *operationalGroupType) CreateOperationalGroupType(c *gin.Context) {
	var optReq dto.OperationalGroupType
	if err := c.ShouldBind(&optReq); err != nil {
		opt.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	optRes, err := opt.operationalGroupModule.CreateOperationalGroupType(c, optReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, optRes)
}

// GetOperationalGroupTypesByGroupID Creating Operational type requests.
//	@Summary		GetOperationalType
//	@Description	Get Operational Type, only for admins
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			groupID	path		string	true	"Group ID"
//	@Success		200		{object}	[]dto.OperationalGroup
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/operationalgrouptype:groupID [get]
func (opt *operationalGroupType) GetOperationalGroupTypesByGroupID(c *gin.Context) {
	groupID := c.Param("groupID")
	groupIDParsed, err := uuid.Parse(groupID)
	if err != nil {
		opt.log.Warn(err.Error(), zap.Any("groupID", groupID))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	oprationGroupTypes, err := opt.operationalGroupModule.GetOperationalGroupTypeByGroupID(c, groupIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, oprationGroupTypes)
}

// GetOperationalGroupTypes Get all operational group types.
//	@Summary		Get all operational group types
//	@Description	Retrieve a list of all operational group types.
//	@Tags			OperationalType
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	[]dto.OperationalTypesRes
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/operationaltypes [get]
func (opt *operationalGroupType) GetOperationalGroupTypes(c *gin.Context) {
	operationalGroupTypes, err := opt.operationalGroupModule.GetOperationalGroupTypes(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, operationalGroupTypes)
}
