package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AWSLogConfig represents AWS logging configuration
type AWSLogConfig struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	LogGroupName    string `yaml:"log_group_name"`
	LogStreamName   string `yaml:"log_stream_name"`
	RetentionDays   int    `yaml:"retention_days"`
	Enabled         bool   `yaml:"enabled"`
}

// AWSLogger represents AWS CloudWatch logger
type AWSLogger struct {
	config         *AWSLogConfig
	cloudWatchLogs *cloudwatchlogs.CloudWatchLogs
	logGroup       string
	logStream      string
	sequenceToken  *string
	logger         *zap.Logger
}

// AWSLogWriter implements io.Writer for AWS CloudWatch
type AWSLogWriter struct {
	awsLogger *AWSLogger
	level     zapcore.Level
}

// NewAWSLogger creates a new AWS CloudWatch logger
func NewAWSLogger(config *AWSLogConfig) (*AWSLogger, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("AWS logging is disabled")
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyID,
			config.SecretAccessKey,
			"", // token
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	cloudWatchLogs := cloudwatchlogs.New(sess)

	awsLogger := &AWSLogger{
		config:         config,
		cloudWatchLogs: cloudWatchLogs,
		logGroup:       config.LogGroupName,
		logStream:      config.LogStreamName,
		logger:         nil, // Will be set later
	}

	// Initialize CloudWatch log group and stream
	if err := awsLogger.initializeCloudWatch(); err != nil {
		return nil, fmt.Errorf("failed to initialize CloudWatch: %w", err)
	}

	return awsLogger, nil
}

// initializeCloudWatch sets up CloudWatch log group and stream
func (al *AWSLogger) initializeCloudWatch() error {
	ctx := context.Background()

	// Create log group if it doesn't exist
	_, err := al.cloudWatchLogs.CreateLogGroupWithContext(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(al.logGroup),
	})
	if err != nil {
		// Check if it's because the log group already exists
		if !strings.Contains(err.Error(), "ResourceAlreadyExistsException") {
			return fmt.Errorf("failed to create log group: %w", err)
		}
	}

	// Set retention policy
	if al.config.RetentionDays > 0 {
		_, err = al.cloudWatchLogs.PutRetentionPolicyWithContext(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(al.logGroup),
			RetentionInDays: aws.Int64(int64(al.config.RetentionDays)),
		})
		if err != nil {
			return fmt.Errorf("failed to set retention policy: %w", err)
		}
	}

	// Create log stream if it doesn't exist
	_, err = al.cloudWatchLogs.CreateLogStreamWithContext(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(al.logGroup),
		LogStreamName: aws.String(al.logStream),
	})
	if err != nil {
		// Check if it's because the log stream already exists
		if !strings.Contains(err.Error(), "ResourceAlreadyExistsException") {
			return fmt.Errorf("failed to create log stream: %w", err)
		}
	}

	return nil
}

// Write implements io.Writer interface for AWS CloudWatch
func (w *AWSLogWriter) Write(p []byte) (n int, err error) {
	if w.awsLogger == nil {
		return len(p), nil
	}

	message := strings.TrimSpace(string(p))
	if message == "" {
		return len(p), nil
	}

	// Send log to CloudWatch
	err = w.awsLogger.sendLogToCloudWatch(message, w.level)
	return len(p), err
}

// Sync implements zapcore.WriteSyncer interface
func (w *AWSLogWriter) Sync() error {
	// AWS CloudWatch doesn't require explicit sync
	return nil
}

// sendLogToCloudWatch sends a log message to CloudWatch
func (al *AWSLogger) sendLogToCloudWatch(message string, level zapcore.Level) error {
	ctx := context.Background()

	// Prepare log event
	logEvent := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(message),
		Timestamp: aws.Int64(time.Now().UnixMilli()),
	}

	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(al.logGroup),
		LogStreamName: aws.String(al.logStream),
		LogEvents:     []*cloudwatchlogs.InputLogEvent{logEvent},
	}

	// Add sequence token if available
	if al.sequenceToken != nil {
		input.SequenceToken = al.sequenceToken
	}

	// Send log event
	result, err := al.cloudWatchLogs.PutLogEventsWithContext(ctx, input)
	if err != nil {
		// If it's a sequence token error, try to get the correct token
		if strings.Contains(err.Error(), "InvalidSequenceTokenException") {
			// Get the next sequence token
			describeResp, describeErr := al.cloudWatchLogs.DescribeLogStreamsWithContext(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName:        aws.String(al.logGroup),
				LogStreamNamePrefix: aws.String(al.logStream),
			})
			if describeErr == nil && len(describeResp.LogStreams) > 0 {
				al.sequenceToken = describeResp.LogStreams[0].UploadSequenceToken
				input.SequenceToken = al.sequenceToken
				result, err = al.cloudWatchLogs.PutLogEventsWithContext(ctx, input)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to send log to CloudWatch: %w", err)
	}

	// Update sequence token for next log
	al.sequenceToken = result.NextSequenceToken
	return nil
}

// CreateAWSLogWriter creates a new AWS log writer
func (al *AWSLogger) CreateAWSLogWriter(level zapcore.Level) *AWSLogWriter {
	return &AWSLogWriter{
		awsLogger: al,
		level:     level,
	}
}

// GetZapCore creates a zapcore.Core for AWS logging
func (al *AWSLogger) GetZapCore(level zapcore.Level) zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	writer := al.CreateAWSLogWriter(level)

	return zapcore.NewCore(encoder, writer, level)
}

// CreateLocalLogFolders creates local log folders for backup
func CreateLocalLogFolders(basePath string) error {
	folders := []string{
		"application",
		"error",
		"access",
		"audit",
		"performance",
		"security",
		"database",
		"api",
		"websocket",
		"email",
		"analytics",
		"cashback",
		"groove",
		"bet",
		"user",
		"payment",
		"notification",
		"cronjob",
		"kafka",
		"redis",
	}

	for _, folder := range folders {
		folderPath := filepath.Join(basePath, folder)
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			return fmt.Errorf("failed to create log folder %s: %w", folderPath, err)
		}
	}

	return nil
}

// LoadAWSLogConfig loads AWS logging configuration from viper
func LoadAWSLogConfig() *AWSLogConfig {
	return &AWSLogConfig{
		Region:          viper.GetString("aws.logging.region"),
		AccessKeyID:     viper.GetString("aws.logging.access_key_id"),
		SecretAccessKey: viper.GetString("aws.logging.secret_access_key"),
		LogGroupName:    viper.GetString("aws.logging.log_group_name"),
		LogStreamName:   viper.GetString("aws.logging.log_stream_name"),
		RetentionDays:   viper.GetInt("aws.logging.retention_days"),
		Enabled:         viper.GetBool("aws.logging.enabled"),
	}
}

// SetupAWSLogging sets up AWS CloudWatch logging with local backup
func SetupAWSLogging() (*AWSLogger, error) {
	// Create local log folders first
	localLogPath := viper.GetString("aws.logging.local_path")
	if localLogPath == "" {
		localLogPath = "./logs"
	}

	if err := CreateLocalLogFolders(localLogPath); err != nil {
		return nil, fmt.Errorf("failed to create local log folders: %w", err)
	}

	// Load AWS configuration
	config := LoadAWSLogConfig()
	if !config.Enabled {
		return nil, fmt.Errorf("AWS logging is disabled in configuration")
	}

	// Validate required fields
	if config.Region == "" || config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS logging configuration is incomplete")
	}

	// Set defaults
	if config.LogGroupName == "" {
		config.LogGroupName = "/tucanbit/application"
	}
	if config.LogStreamName == "" {
		config.LogStreamName = fmt.Sprintf("tucanbit-%s", time.Now().Format("2006-01-02"))
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 30
	}

	// Create AWS logger
	awsLogger, err := NewAWSLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS logger: %w", err)
	}

	return awsLogger, nil
}
