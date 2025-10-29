package operationsdefinitions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type operationsDefinitions struct {
	log                         *zap.Logger
	operationsDefinitionsModule module.OperationsDefinitions
}

func Init(operationsDefinitionsModule module.OperationsDefinitions, log *zap.Logger) handler.OperationsDefinition {
	return &operationsDefinitions{
		log:                         log,
		operationsDefinitionsModule: operationsDefinitionsModule,
	}
}

// GetOperationsDefinitions Get Operation Definitions  requests.
//
//	@Summary		GetOperationsDefinitions
//	@Description	Get  Operations Definitions
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	[]dto.OperationsDefinition
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/operations/definitions [GET]
func (od *operationsDefinitions) GetOperationsDefinitions(c *gin.Context) {
	ods, err := od.operationsDefinitionsModule.GetOperationalDefinition(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, ods)
}
