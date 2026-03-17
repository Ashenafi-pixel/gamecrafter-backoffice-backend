package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GinLogger(log zap.Logger) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.Query()
		id := uuid.New()
		ctx.Set("x-request-id", id)
		ctx.Set("x-start-time", start)
		ctx.Set("user-agent", ctx.Request.UserAgent())
		ctx.Set("ip", ctx.ClientIP())
		ctx.Next()
		end := time.Now()
		latency := end.Sub(start)
		fields := []zapcore.Field{
			zap.Int("status", ctx.Writer.Status()),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("ip", ctx.ClientIP()),
			zap.Any("query", query),
			zap.String("user-agent", ctx.Request.UserAgent()),
			zap.Float64("latency", latency.Minutes()),
		}
		log.Info("Gin Request", fields...)

		// Also write a simple line to the HTTP log file following
		// docs/LOGGING_STRUCTURE.md (logs/app/http/...).
		status := ctx.Writer.Status()
		success := status < 400

		// Collect any error attached to the Gin context for inclusion in the log line.
		var errMsg string
		if len(ctx.Errors) > 0 {
			errs := make([]string, 0, len(ctx.Errors))
			for _, e := range ctx.Errors {
				if e != nil && e.Err != nil {
					errs = append(errs, e.Err.Error())
				}
			}
			if len(errs) > 0 {
				errMsg = strings.Join(errs, " | ")
			}
		}

		line := fmt.Sprintf(
			"%s status=%d method=%s path=%s ip=%s latency_ms=%d error=%s",
			end.Format(time.RFC3339),
			status,
			ctx.Request.Method,
			path,
			ctx.ClientIP(),
			latency.Milliseconds(),
			errMsg,
		)
		utils.WriteHTTPLogLine(success, line, end)
	}
}
