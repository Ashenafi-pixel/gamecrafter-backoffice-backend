package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func SystemLogs(module string, log *zap.Logger, sysLogger module.SystemLogs) gin.HandlerFunc {
	return func(c *gin.Context) {

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error("Failed to read request body", zap.Error(err))
			// Don't send error response here, just log and continue
			// This prevents duplicate JSON responses
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var systemLogCopy map[string]interface{}

		// Check if this is a multipart form request
		contentType := c.GetHeader("Content-Type")
		if len(bodyBytes) > 0 && !isMultipartRequest(contentType) {
			if err := json.Unmarshal(bodyBytes, &systemLogCopy); err != nil {
				log.Error("Failed to decode request body", zap.String("body", string(bodyBytes)), zap.Error(err))
				// Don't send error response here, just log and continue
				// This prevents duplicate JSON responses
				c.Next()
				return
			}
		} else {
			// For multipart requests or empty body, create empty map
			systemLogCopy = make(map[string]interface{})
		}

		userID := c.GetString("user-id")
		remoteIP := c.GetString("ip")

		// Handle missing user ID gracefully for unauthenticated endpoints
		if userID == "" {
			// For unauthenticated requests (like login), skip system logging
			// but still log the request for audit purposes
			log.Info("System log skipped - no user ID",
				zap.String("module", module),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", remoteIP))
			c.Next()
			return
		}

		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			log.Error("Failed to parse userID", zap.String("userID", userID), zap.Error(err))
			// Don't send error response here, just log and continue
			// This prevents duplicate JSON responses
			c.Next()
			return
		}

		start := time.Now()
		systemLog := dto.SystemLogs{
			UserID:    userIDParsed,
			Module:    module,
			IPAddress: remoteIP,
			Timestamp: start,
			Detail:    systemLogCopy,
		}

		_, err = sysLogger.CreateSystemLogs(c, systemLog)
		if err != nil {
			log.Error("Failed to create system log", zap.Any("systemLog", systemLog), zap.Error(err))
			// Don't send error response here, just log and continue
			// This prevents duplicate JSON responses
		}
		c.Next()

	}
}

// isMultipartRequest checks if the request is a multipart form request
func isMultipartRequest(contentType string) bool {
	return len(contentType) > 0 && len(contentType) > 19 && contentType[:19] == "multipart/form-data"
}
