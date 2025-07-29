package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joomcode/errorx"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/spf13/viper"
)

func SendSuccessResponse(ctx *gin.Context, statusCode int, data interface{}) {

	ctx.JSON(
		statusCode,
		data,
	)
}

func SendErrorResponse(ctx *gin.Context, err *ErrorResponse) {
	ctx.AbortWithStatusJSON(err.Code, err)
}

func GetErrorFrom(err error) *ErrorResponse {
	debugMode := viper.GetBool("debug")

	for _, e := range errors.Error {
		if errorx.IsOfType(err, e.Type) {
			er := errorx.Cast(err)
			res := ErrorResponse{
				Code:       e.StatusCode,
				Message:    er.Message(),
				FieldError: ErrorFields(er.Cause()),
			}

			if debugMode {
				res.Description = fmt.Sprintf("Error: %v", er)
				res.StackTrace = fmt.Sprintf("%+v", errorx.EnsureStackTrace(err))
			}

			return &res
		}
	}

	return &ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "Unknown server error",
	}
}

func ErrorFields(err error) []FieldError {
	var errs []FieldError

	if data, ok := err.(validation.Errors); ok {
		for i, v := range data {
			errs = append(errs, FieldError{
				Name:        i,
				Description: v.Error(),
			},
			)
		}

		return errs
	}

	return nil
}
