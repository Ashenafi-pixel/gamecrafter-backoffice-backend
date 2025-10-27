package system_configs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/model/db"
	"go.uber.org/zap"
)

type SystemConfigsHandler struct {
	log *zap.Logger
	db  *db.Queries
}

func NewSystemConfigsHandler(log *zap.Logger, queries *db.Queries) *SystemConfigsHandler {
	return &SystemConfigsHandler{
		log: log,
		db:  queries,
	}
}

// SystemConfig represents a system configuration from the database
type SystemConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

// GetSystemConfigs returns all system configurations
func (h *SystemConfigsHandler) GetSystemConfigs(c *gin.Context) {
	configs, err := h.db.GetAllConfigs(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to fetch configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch system configs",
		})
		return
	}

	var systemConfigs []SystemConfig
	for _, config := range configs {
		systemConfigs = append(systemConfigs, SystemConfig{
			ID:        config.ID.String(),
			Name:      config.Name,
			Value:     config.Value,
			CreatedAt: config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"configs": systemConfigs,
			"total":   len(systemConfigs),
		},
		"message": "System configs fetched successfully",
	})
}

// UpdateSystemConfig updates a system configuration
func (h *SystemConfigsHandler) UpdateSystemConfig(c *gin.Context) {
	var request struct {
		ID    string `json:"id" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	configID, err := uuid.Parse(request.ID)
	if err != nil {
		h.log.Error("Invalid config ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid config ID",
		})
		return
	}

	config, err := h.db.UpdateConfigs(c.Request.Context(), db.UpdateConfigsParams{
		ID:    configID,
		Value: request.Value,
	})
	if err != nil {
		h.log.Error("Failed to update config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update system config",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": SystemConfig{
			ID:        config.ID.String(),
			Name:      config.Name,
			Value:     config.Value,
			CreatedAt: config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"message": "System config updated successfully",
	})
}
