package system_configs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SystemConfigsHandler struct {
	log *zap.Logger
}

func NewSystemConfigsHandler(log *zap.Logger) *SystemConfigsHandler {
	return &SystemConfigsHandler{
		log: log,
	}
}

// SystemConfig represents a system configuration
type SystemConfig struct {
	ID          string                 `json:"id"`
	ConfigKey   string                 `json:"config_key"`
	ConfigValue map[string]interface{} `json:"config_value"`
	Description string                 `json:"description"`
	CreatedAt   string                 `json:"created_at"`
}

// GetSystemConfigs returns all system configurations
func (h *SystemConfigsHandler) GetSystemConfigs(c *gin.Context) {
	// Mock data for now - replace with actual database queries
	systemConfigs := []SystemConfig{
		{
			ID:        "1",
			ConfigKey: "site_settings",
			ConfigValue: map[string]interface{}{
				"site_name":        "TucanBIT Casino",
				"site_description": "Premier cryptocurrency casino platform",
				"maintenance_mode": false,
			},
			Description: "Main site configuration settings",
			CreatedAt:   "2024-01-01T00:00:00Z",
		},
		{
			ID:        "2",
			ConfigKey: "payment_settings",
			ConfigValue: map[string]interface{}{
				"min_deposit":    10.0,
				"max_deposit":    10000.0,
				"min_withdrawal": 20.0,
				"max_withdrawal": 5000.0,
				"processing_fee": 0.5,
			},
			Description: "Payment processing configuration",
			CreatedAt:   "2024-01-01T00:00:00Z",
		},
		{
			ID:        "3",
			ConfigKey: "security_settings",
			ConfigValue: map[string]interface{}{
				"two_factor_required":  true,
				"session_timeout":      3600,
				"max_login_attempts":   5,
				"ip_whitelist_enabled": false,
			},
			Description: "Security configuration settings",
			CreatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"system_configs": systemConfigs,
			"total":          len(systemConfigs),
		},
		"message": "System configs fetched successfully",
	})
}
