package logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type logs struct {
	log *zap.Logger
	module.SystemLogs
}

func Init(log *zap.Logger, sysLogger module.SystemLogs) handler.SystemLogs {
	return &logs{
		log:        log,
		SystemLogs: sysLogger,
	}
}

// GetSystemLogs get system logs
//
//	@Summary		get system logs
//	@Description	get system logs
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string					true	"Bearer <token>"
//	@Param			GetSystemLogsReq	body		dto.GetSystemLogsReq	true	"GetSystemLogsReq"
//	@Success		200					{object}	[]dto.SystemLogs
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/admin/logs [post]
func (l *logs) GetSystemLogs(c *gin.Context) {
	var req dto.GetSystemLogsReq

	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	systemLogs, err := l.SystemLogs.GetSystemLogs(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, systemLogs)
}

// GetAvailableLogModules get available log modules
//
//	@Summary		get available log modules
//	@Description	get available log modules
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	[]string
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/logs/modules [get]
func (l *logs) GetAvailableLogModules(c *gin.Context) {

	systemLogs, err := l.SystemLogs.GetAvailableLogsModule(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, systemLogs)
}
