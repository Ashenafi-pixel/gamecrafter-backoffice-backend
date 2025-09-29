package initiator

import (
	"log"

	"github.com/spf13/viper"
	"github.com/tucanbit/platform/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() logger.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(viper.GetInt("logger.level")))

	lg, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	return logger.New(lg)
}

// InitEnhancedLogger creates an enhanced logger with AWS CloudWatch integration
func InitEnhancedLogger() logger.EnhancedLogger {
	enhancedLogger, err := logger.NewEnhancedLogger()
	if err != nil {
		log.Fatalf("failed to initialize enhanced logger: %v", err)
	}
	return enhancedLogger
}
