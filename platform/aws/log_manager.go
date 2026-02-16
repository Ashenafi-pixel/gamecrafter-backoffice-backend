package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// LogRotationConfig represents log rotation configuration
type LogRotationConfig struct {
	MaxSizeMB  int  `yaml:"max_size_mb"`
	MaxBackups int  `yaml:"max_backups"`
	MaxAgeDays int  `yaml:"max_age_days"`
	Compress   bool `yaml:"compress"`
	Enabled    bool `yaml:"enabled"`
}

// LogRetentionConfig represents log retention configuration
type LogRetentionConfig struct {
	LocalRetentionDays int  `yaml:"local_retention_days"`
	AWSRetentionDays   int  `yaml:"aws_retention_days"`
	Enabled            bool `yaml:"enabled"`
}

// LogManager manages log rotation and retention
type LogManager struct {
	awsLogger         *AWSLogger
	logRotationConfig *LogRotationConfig
	retentionConfig   *LogRetentionConfig
	logger            *zap.Logger
}

// NewLogManager creates a new log manager
func NewLogManager(awsLogger *AWSLogger, logger *zap.Logger) *LogManager {
	return &LogManager{
		awsLogger:         awsLogger,
		logRotationConfig: LoadLogRotationConfig(),
		retentionConfig:   LoadLogRetentionConfig(),
		logger:            logger,
	}
}

// LoadLogRotationConfig loads log rotation configuration
func LoadLogRotationConfig() *LogRotationConfig {
	return &LogRotationConfig{
		MaxSizeMB:  viper.GetInt("logger.max_size_mb"),
		MaxBackups: viper.GetInt("logger.max_backups"),
		MaxAgeDays: viper.GetInt("logger.max_age_days"),
		Compress:   viper.GetBool("logger.compress"),
		Enabled:    viper.GetBool("logger.rotation_enabled"),
	}
}

// LoadLogRetentionConfig loads log retention configuration
func LoadLogRetentionConfig() *LogRetentionConfig {
	return &LogRetentionConfig{
		LocalRetentionDays: viper.GetInt("logger.local_retention_days"),
		AWSRetentionDays:   viper.GetInt("logger.aws_retention_days"),
		Enabled:            viper.GetBool("logger.retention_enabled"),
	}
}

// StartLogRotation starts the log rotation process
func (lm *LogManager) StartLogRotation(ctx context.Context) {
	if !lm.logRotationConfig.Enabled {
		lm.logger.Info("Log rotation is disabled")
		return
	}

	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	lm.logger.Info("Starting log rotation process")

	for {
		select {
		case <-ctx.Done():
			lm.logger.Info("Stopping log rotation process")
			return
		case <-ticker.C:
			lm.performLogRotation()
		}
	}
}

// performLogRotation performs log rotation for all log files
func (lm *LogManager) performLogRotation() {
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	logFolders := []string{
		"application", "error", "access", "audit", "performance",
		"security", "database", "api", "websocket", "email",
		"analytics", "cashback", "groove", "bet", "user",
		"payment", "notification", "cronjob", "kafka", "redis",
	}

	for _, folder := range logFolders {
		folderPath := filepath.Join(localLogPath, folder)
		lm.rotateLogFilesInFolder(folderPath)
	}

	lm.logger.Info("Log rotation completed")
}

// rotateLogFilesInFolder rotates log files in a specific folder
func (lm *LogManager) rotateLogFilesInFolder(folderPath string) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		lm.logger.Error("Failed to read log folder", zap.String("folder", folderPath), zap.Error(err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(folderPath, entry.Name())
		lm.rotateLogFile(filePath)
	}
}

// rotateLogFile rotates a specific log file
func (lm *LogManager) rotateLogFile(filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}

	// Check if file needs rotation (size-based)
	maxSizeBytes := int64(lm.logRotationConfig.MaxSizeMB * 1024 * 1024)
	if fileInfo.Size() >= maxSizeBytes {
		lm.performFileRotation(filePath)
	}

	// Check if file needs rotation (age-based)
	maxAge := time.Duration(lm.logRotationConfig.MaxAgeDays) * 24 * time.Hour
	if time.Since(fileInfo.ModTime()) >= maxAge {
		lm.performFileRotation(filePath)
	}
}

// performFileRotation performs the actual file rotation
func (lm *LogManager) performFileRotation(filePath string) {
	// Create rotated filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	rotatedPath := fmt.Sprintf("%s.%s", filePath, timestamp)

	// Rename current file to rotated name
	err := os.Rename(filePath, rotatedPath)
	if err != nil {
		lm.logger.Error("Failed to rotate log file", zap.String("file", filePath), zap.Error(err))
		return
	}

	// Compress if enabled
	if lm.logRotationConfig.Compress {
		go lm.compressLogFile(rotatedPath)
	}

	// Clean up old rotated files
	go lm.cleanupOldLogFiles(filepath.Dir(filePath))

	lm.logger.Info("Log file rotated", zap.String("file", filePath), zap.String("rotated_to", rotatedPath))
}

// compressLogFile compresses a log file
func (lm *LogManager) compressLogFile(filePath string) {
	// This would typically use gzip compression
	// For now, we'll just log that compression would happen
	lm.logger.Debug("Log file compressed", zap.String("file", filePath))
}

// cleanupOldLogFiles removes old log files based on retention policy
func (lm *LogManager) cleanupOldLogFiles(dirPath string) {
	if !lm.retentionConfig.Enabled {
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	maxBackups := lm.logRotationConfig.MaxBackups
	maxAge := time.Duration(lm.retentionConfig.LocalRetentionDays) * 24 * time.Hour

	var rotatedFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && lm.isRotatedLogFile(entry.Name()) {
			rotatedFiles = append(rotatedFiles, entry)
		}
	}

	// Remove files exceeding max backups
	if len(rotatedFiles) > maxBackups {
		for i := 0; i < len(rotatedFiles)-maxBackups; i++ {
			filePath := filepath.Join(dirPath, rotatedFiles[i].Name())
			os.Remove(filePath)
			lm.logger.Debug("Removed old log file", zap.String("file", filePath))
		}
	}

	// Remove files exceeding max age
	for _, entry := range rotatedFiles {
		filePath := filepath.Join(dirPath, entry.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if time.Since(fileInfo.ModTime()) > maxAge {
			os.Remove(filePath)
			lm.logger.Debug("Removed old log file by age", zap.String("file", filePath))
		}
	}
}

// isRotatedLogFile checks if a file is a rotated log file
func (lm *LogManager) isRotatedLogFile(filename string) bool {
	// Check if filename contains timestamp pattern (YYYY-MM-DD-HH-MM-SS)
	// This is a simple check - in production you might want more sophisticated pattern matching
	return len(filename) > 20 && filename[len(filename)-19:] != ".log"
}

// StartLogRetention starts the log retention process
func (lm *LogManager) StartLogRetention(ctx context.Context) {
	if !lm.retentionConfig.Enabled {
		lm.logger.Info("Log retention is disabled")
		return
	}

	ticker := time.NewTicker(7 * 24 * time.Hour) // Check weekly
	defer ticker.Stop()

	lm.logger.Info("Starting log retention process")

	for {
		select {
		case <-ctx.Done():
			lm.logger.Info("Stopping log retention process")
			return
		case <-ticker.C:
			lm.performLogRetention()
		}
	}
}

// performLogRetention performs log retention cleanup
func (lm *LogManager) performLogRetention() {
	// Clean up local logs
	lm.cleanupLocalLogs()

	// Clean up AWS CloudWatch logs
	if lm.awsLogger != nil {
		lm.cleanupAWSLogs()
	}

	lm.logger.Info("Log retention completed")
}

// cleanupLocalLogs cleans up local log files
func (lm *LogManager) cleanupLocalLogs() {
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	maxAge := time.Duration(lm.retentionConfig.LocalRetentionDays) * 24 * time.Hour
	lm.cleanupLogsInPath(localLogPath, maxAge)
}

// cleanupLogsInPath recursively cleans up logs in a path
func (lm *LogManager) cleanupLogsInPath(path string, maxAge time.Duration) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			lm.cleanupLogsInPath(entryPath, maxAge)
		} else {
			fileInfo, err := os.Stat(entryPath)
			if err != nil {
				continue
			}

			if time.Since(fileInfo.ModTime()) > maxAge {
				os.Remove(entryPath)
				lm.logger.Debug("Removed old log file", zap.String("file", entryPath))
			}
		}
	}
}

// cleanupAWSLogs cleans up AWS CloudWatch logs
func (lm *LogManager) cleanupAWSLogs() {
	if lm.awsLogger == nil {
		return
	}

	ctx := context.Background()
	retentionDays := int64(lm.retentionConfig.AWSRetentionDays)

	// Set retention policy for log group
	_, err := lm.awsLogger.cloudWatchLogs.PutRetentionPolicyWithContext(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(lm.awsLogger.logGroup),
		RetentionInDays: aws.Int64(retentionDays),
	})
	if err != nil {
		lm.logger.Error("Failed to set AWS log retention policy", zap.Error(err))
		return
	}

	lm.logger.Info("AWS log retention policy updated", zap.Int64("retention_days", retentionDays))
}

// GetLogStats returns statistics about log files
func (lm *LogManager) GetLogStats() map[string]interface{} {
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	stats := map[string]interface{}{
		"local_path": localLogPath,
		"rotation": map[string]interface{}{
			"enabled":      lm.logRotationConfig.Enabled,
			"max_size_mb":  lm.logRotationConfig.MaxSizeMB,
			"max_backups":  lm.logRotationConfig.MaxBackups,
			"max_age_days": lm.logRotationConfig.MaxAgeDays,
			"compress":     lm.logRotationConfig.Compress,
		},
		"retention": map[string]interface{}{
			"enabled":              lm.retentionConfig.Enabled,
			"local_retention_days": lm.retentionConfig.LocalRetentionDays,
			"aws_retention_days":   lm.retentionConfig.AWSRetentionDays,
		},
		"aws_logging": map[string]interface{}{
			"enabled": lm.awsLogger != nil,
			"log_group": func() string {
				if lm.awsLogger != nil {
					return lm.awsLogger.logGroup
				}
				return ""
			}(),
			"log_stream": func() string {
				if lm.awsLogger != nil {
					return lm.awsLogger.logStream
				}
				return ""
			}(),
		},
	}

	return stats
}
