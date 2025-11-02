package system_config

import (
	"context"

	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

type systemConfigModule struct {
	log            *zap.Logger
	systemConfigStorage *system_config.SystemConfig
}

func Init(systemConfigStorage *system_config.SystemConfig, log *zap.Logger) SystemConfig {
	return &systemConfigModule{
		log:                 log,
		systemConfigStorage: systemConfigStorage,
	}
}

// SystemConfig interface for system configuration operations
type SystemConfig interface {
	GetSecuritySettings(ctx context.Context) (system_config.SecuritySettings, error)
}

func (s *systemConfigModule) GetSecuritySettings(ctx context.Context) (system_config.SecuritySettings, error) {
	return s.systemConfigStorage.GetSecuritySettings(ctx)
}

