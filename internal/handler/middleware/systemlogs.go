package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

func SystemLogs(module string, log *zap.Logger, sysLogger module.SystemLogs) gin.HandlerFunc {
	return func(c *gin.Context) {

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error("Failed to read request body", zap.Error(err))
			err = errors.ErrInternalServerError.Wrap(err, "failed to read request body")
			_ = c.Error(err)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var systemLogCopy map[string]interface{}

		// Check if this is a multipart form request
		contentType := c.GetHeader("Content-Type")
		if len(bodyBytes) > 0 && !isMultipartRequest(contentType) {
			if err := json.Unmarshal(bodyBytes, &systemLogCopy); err != nil {
				log.Error("Failed to decode request body", zap.String("body", string(bodyBytes)), zap.Error(err))
				err = errors.ErrInternalServerError.Wrap(err, "failed to decode request body")
				_ = c.Error(err)
				return
			}
		} else {
			// For multipart requests or empty body, create empty map
			systemLogCopy = make(map[string]interface{})
		}

		userID := c.GetString("user-id")
		remoteIP := c.GetString("ip")
		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			log.Error("Failed to parse userID", zap.String("userID", userID), zap.Error(err))
			err = errors.ErrInternalServerError.Wrap(err, "failed to parse userID")
			_ = c.Error(err)
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
			err = errors.ErrInternalServerError.Wrap(err, "failed to create system log")
			_ = c.Error(err)
			return
		}
		c.Next()

	}
}

// isMultipartRequest checks if the request is a multipart form request
func isMultipartRequest(contentType string) bool {
	return len(contentType) > 0 && len(contentType) > 19 && contentType[:19] == "multipart/form-data"
}
