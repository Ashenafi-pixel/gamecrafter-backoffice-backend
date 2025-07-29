package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	}
}
