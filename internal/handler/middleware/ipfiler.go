package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/module"
)

func IpFilter(userModule module.User) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userIP := ctx.ClientIP()
		ok, err := userModule.EnforceIPFilerRule(ctx, userIP)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !ok {
			ctx.JSON(http.StatusBadRequest, "service not available in your area")
			ctx.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		ctx.Next()
	}
}
