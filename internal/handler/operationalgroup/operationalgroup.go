package operationalgroup

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type operationalGroup struct {
	operationalGroupModule module.OperationalGroup
	log                    *zap.Logger
}

func Init(operationalGroupModule module.OperationalGroup, log *zap.Logger) handler.OpeartionalGroup {
	return &operationalGroup{
		operationalGroupModule: operationalGroupModule,
		log:                    log,
	}
}

// CreateOperationalGroup Creating Operational group requests.
//	@Summary		CreateOperationalGroup
//	@Description	Create Operational Group using name and description, only for admins
//	@Tags			OperationalGroup
//	@Accept			json
//	@Produce		json
//	@Param			opReq	body		dto.OperationalGroup	true	"Create Operational Group Request"
//	@Success		201		{object}	dto.OperationalGroup
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/operationalgroup [post]
func (op *operationalGroup) CreateOperationalGroup(c *gin.Context) {
	var opReq dto.OperationalGroup
	if err := c.ShouldBind(&opReq); err != nil {
		op.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	operationalGroup, err := op.operationalGroupModule.CreateOperationalGroup(c, opReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, operationalGroup)
}

// GetOperationalGroups Creating Operational group requests.
//	@Summary		GetOperationalGroup
//	@Description	Get Operational Groups, only for admins
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]dto.OperationalGroup
//	@Router			/api/admin/operationalgroup [get]
func (op *operationalGroup) GetOperationalGroups(c *gin.Context) {
	opGroups, err := op.operationalGroupModule.GetOperationalGroups(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, opGroups)
}
