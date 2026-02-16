package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/tucanbit/platform/aws"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// EnhancedLogger represents an enhanced logger with AWS CloudWatch integration
type EnhancedLogger interface {
	Logger
	GetAWSLogger() *aws.AWSLogger
	GetLocalLogger() *zap.Logger
	Close() error
}

// enhancedLogger implements EnhancedLogger
type enhancedLogger struct {
	*logger
	awsLogger   *aws.AWSLogger
	localLogger *zap.Logger
	cores       []zapcore.Core
}

// NewEnhancedLogger creates a new enhanced logger with AWS CloudWatch integration
func NewEnhancedLogger() (EnhancedLogger, error) {
	// Create local log folders
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	if err := aws.CreateLocalLogFolders(localLogPath); err != nil {
		return nil, fmt.Errorf("failed to create local log folders: %w", err)
	}

	// Setup AWS logging
	awsLogger, err := aws.SetupAWSLogging()
	if err != nil {
		// If AWS logging fails, continue with local logging only
		fmt.Printf("Warning: AWS logging setup failed: %v\n", err)
		awsLogger = nil
	}

	// Create cores for different log levels
	cores := []zapcore.Core{}

	// Console core for development
	if viper.GetBool("logger.console") {
		consoleConfig := zap.NewDevelopmentConfig()
		consoleConfig.EncoderConfig.TimeKey = "timestamp"
		consoleConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleCore, _ := consoleConfig.Build()
		cores = append(cores, consoleCore.Core())
	}

	// File cores for different log levels
	cores = append(cores, createFileCore(filepath.Join(localLogPath, "application", "app.log"), zapcore.InfoLevel))
	cores = append(cores, createFileCore(filepath.Join(localLogPath, "error", "error.log"), zapcore.ErrorLevel))
	cores = append(cores, createFileCore(filepath.Join(localLogPath, "access", "access.log"), zapcore.InfoLevel))
	cores = append(cores, createFileCore(filepath.Join(localLogPath, "audit", "audit.log"), zapcore.InfoLevel))

	// AWS CloudWatch cores
	if awsLogger != nil {
		cores = append(cores, awsLogger.GetZapCore(zapcore.InfoLevel))
	}

	// Create combined core
	combinedCore := zapcore.NewTee(cores...)

	// Create logger
	zapLogger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Create enhanced logger
	enhancedLogger := &enhancedLogger{
		logger:      &logger{logger: zapLogger},
		awsLogger:   awsLogger,
		localLogger: zapLogger,
		cores:       cores,
	}

	return enhancedLogger, nil
}

// createFileCore creates a file core with rotation
func createFileCore(filename string, level zapcore.Level) zapcore.Core {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	os.MkdirAll(dir, 0755)

	// Create lumberjack logger for rotation
	fileWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    viper.GetInt("logger.max_size_mb"), // MB
		MaxBackups: viper.GetInt("logger.max_backups"),
		MaxAge:     viper.GetInt("logger.max_age_days"), // days
		Compress:   viper.GetBool("logger.compress"),
	}

	// Create encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create core
	return zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level)
}

// GetAWSLogger returns the AWS logger instance
func (el *enhancedLogger) GetAWSLogger() *aws.AWSLogger {
	return el.awsLogger
}

// GetLocalLogger returns the local logger instance
func (el *enhancedLogger) GetLocalLogger() *zap.Logger {
	return el.localLogger
}

// Close closes the enhanced logger
func (el *enhancedLogger) Close() error {
	if el.localLogger != nil {
		el.localLogger.Sync()
	}
	return nil
}

// CreateModuleLogger creates a logger for a specific module
func (el *enhancedLogger) CreateModuleLogger(moduleName string) EnhancedLogger {
	// Create module-specific cores
	moduleCores := []zapcore.Core{}

	// Add existing cores
	moduleCores = append(moduleCores, el.cores...)

	// Add module-specific file core
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	moduleCore := createFileCore(
		filepath.Join(localLogPath, moduleName, fmt.Sprintf("%s.log", moduleName)),
		zapcore.InfoLevel,
	)
	moduleCores = append(moduleCores, moduleCore)

	// Create combined core
	combinedCore := zapcore.NewTee(moduleCores...)

	// Create module logger
	moduleZapLogger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &enhancedLogger{
		logger:      &logger{logger: moduleZapLogger},
		awsLogger:   el.awsLogger,
		localLogger: moduleZapLogger,
		cores:       moduleCores,
	}
}

// LogToFile logs a message to a specific file
func (el *enhancedLogger) LogToFile(filename string, level zapcore.Level, message string, fields ...zap.Field) {
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	filePath := filepath.Join(localLogPath, filename)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)

	// Create file core for this specific log
	fileCore := createFileCore(filePath, level)

	// Create temporary logger
	tempLogger := zap.New(fileCore)
	defer tempLogger.Sync()

	// Log the message
	switch level {
	case zapcore.DebugLevel:
		tempLogger.Debug(message, fields...)
	case zapcore.InfoLevel:
		tempLogger.Info(message, fields...)
	case zapcore.WarnLevel:
		tempLogger.Warn(message, fields...)
	case zapcore.ErrorLevel:
		tempLogger.Error(message, fields...)
	case zapcore.FatalLevel:
		tempLogger.Fatal(message, fields...)
	case zapcore.PanicLevel:
		tempLogger.Panic(message, fields...)
	}
}

// LogToAWS logs a message specifically to AWS CloudWatch
func (el *enhancedLogger) LogToAWS(level zapcore.Level, message string, fields ...zap.Field) {
	if el.awsLogger == nil {
		return
	}

	// Create AWS-specific logger
	awsCore := el.awsLogger.GetZapCore(level)
	awsLogger := zap.New(awsCore)
	defer awsLogger.Sync()

	// Log the message
	switch level {
	case zapcore.DebugLevel:
		awsLogger.Debug(message, fields...)
	case zapcore.InfoLevel:
		awsLogger.Info(message, fields...)
	case zapcore.WarnLevel:
		awsLogger.Warn(message, fields...)
	case zapcore.ErrorLevel:
		awsLogger.Error(message, fields...)
	case zapcore.FatalLevel:
		awsLogger.Fatal(message, fields...)
	case zapcore.PanicLevel:
		awsLogger.Panic(message, fields...)
	}
}

// LogPerformance logs performance metrics
func (el *enhancedLogger) LogPerformance(ctx context.Context, operation string, duration time.Duration, fields ...zap.Field) {
	performanceFields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	}
	performanceFields = append(performanceFields, fields...)

	el.Info(ctx, fmt.Sprintf("Performance: %s completed in %v", operation, duration), performanceFields...)

	// Log to performance-specific file
	el.LogToFile("performance/performance.log", zapcore.InfoLevel,
		fmt.Sprintf("Performance: %s completed in %v", operation, duration), performanceFields...)
}

// LogSecurity logs security events
func (el *enhancedLogger) LogSecurity(ctx context.Context, event string, fields ...zap.Field) {
	securityFields := []zap.Field{
		zap.String("event", event),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "security"),
	}
	securityFields = append(securityFields, fields...)

	el.Warn(ctx, fmt.Sprintf("Security Event: %s", event), securityFields...)

	// Log to security-specific file
	el.LogToFile("security/security.log", zapcore.WarnLevel,
		fmt.Sprintf("Security Event: %s", event), securityFields...)

	// Also log to AWS for security events
	el.LogToAWS(zapcore.WarnLevel, fmt.Sprintf("Security Event: %s", event), securityFields...)
}

// LogAudit logs audit events
func (el *enhancedLogger) LogAudit(ctx context.Context, action string, resource string, fields ...zap.Field) {
	auditFields := []zap.Field{
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "audit"),
	}
	auditFields = append(auditFields, fields...)

	el.Info(ctx, fmt.Sprintf("Audit: %s on %s", action, resource), auditFields...)

	// Log to audit-specific file
	el.LogToFile("audit/audit.log", zapcore.InfoLevel,
		fmt.Sprintf("Audit: %s on %s", action, resource), auditFields...)
}

// LogAPI logs API requests
func (el *enhancedLogger) LogAPI(ctx context.Context, method string, path string, statusCode int, duration time.Duration, fields ...zap.Field) {
	apiFields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "api"),
	}
	apiFields = append(apiFields, fields...)

	level := zapcore.InfoLevel
	if statusCode >= 400 {
		level = zapcore.ErrorLevel
	}

	el.Info(ctx, fmt.Sprintf("API: %s %s - %d (%v)", method, path, statusCode, duration), apiFields...)

	// Log to API-specific file
	el.LogToFile("api/api.log", level,
		fmt.Sprintf("API: %s %s - %d (%v)", method, path, statusCode, duration), apiFields...)
}

// LogDatabase logs database operations
func (el *enhancedLogger) LogDatabase(ctx context.Context, operation string, table string, duration time.Duration, fields ...zap.Field) {
	dbFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "database"),
	}
	dbFields = append(dbFields, fields...)

	el.Debug(ctx, fmt.Sprintf("Database: %s on %s (%v)", operation, table, duration), dbFields...)

	// Log to database-specific file
	el.LogToFile("database/database.log", zapcore.DebugLevel,
		fmt.Sprintf("Database: %s on %s (%v)", operation, table, duration), dbFields...)
}

// LogWebSocket logs WebSocket events
func (el *enhancedLogger) LogWebSocket(ctx context.Context, event string, userID string, fields ...zap.Field) {
	wsFields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "websocket"),
	}
	wsFields = append(wsFields, fields...)

	el.Info(ctx, fmt.Sprintf("WebSocket: %s for user %s", event, userID), wsFields...)

	// Log to WebSocket-specific file
	el.LogToFile("websocket/websocket.log", zapcore.InfoLevel,
		fmt.Sprintf("WebSocket: %s for user %s", event, userID), wsFields...)
}

// LogEmail logs email operations
func (el *enhancedLogger) LogEmail(ctx context.Context, action string, recipient string, subject string, fields ...zap.Field) {
	emailFields := []zap.Field{
		zap.String("action", action),
		zap.String("recipient", recipient),
		zap.String("subject", subject),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "email"),
	}
	emailFields = append(emailFields, fields...)

	el.Info(ctx, fmt.Sprintf("Email: %s to %s - %s", action, recipient, subject), emailFields...)

	// Log to email-specific file
	el.LogToFile("email/email.log", zapcore.InfoLevel,
		fmt.Sprintf("Email: %s to %s - %s", action, recipient, subject), emailFields...)
}

// LogKafka logs Kafka events
func (el *enhancedLogger) LogKafka(ctx context.Context, action string, topic string, fields ...zap.Field) {
	kafkaFields := []zap.Field{
		zap.String("action", action),
		zap.String("topic", topic),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "kafka"),
	}
	kafkaFields = append(kafkaFields, fields...)

	el.Info(ctx, fmt.Sprintf("Kafka: %s on topic %s", action, topic), kafkaFields...)

	// Log to Kafka-specific file
	el.LogToFile("kafka/kafka.log", zapcore.InfoLevel,
		fmt.Sprintf("Kafka: %s on topic %s", action, topic), kafkaFields...)
}

// LogRedis logs Redis operations
func (el *enhancedLogger) LogRedis(ctx context.Context, operation string, key string, duration time.Duration, fields ...zap.Field) {
	redisFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Duration("duration", duration),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
		zap.String("type", "redis"),
	}
	redisFields = append(redisFields, fields...)

	el.Debug(ctx, fmt.Sprintf("Redis: %s key %s (%v)", operation, key, duration), redisFields...)

	// Log to Redis-specific file
	el.LogToFile("redis/redis.log", zapcore.DebugLevel,
		fmt.Sprintf("Redis: %s key %s (%v)", operation, key, duration), redisFields...)
}
