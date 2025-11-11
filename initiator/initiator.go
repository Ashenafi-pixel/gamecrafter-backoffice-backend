package initiator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/handler/middleware"
	alertModule "github.com/tucanbit/internal/module/alert"
	analyticsModule "github.com/tucanbit/internal/module/analytics"
	emailModule "github.com/tucanbit/internal/module/email"
	analyticsStorage "github.com/tucanbit/internal/storage/analytics"
	"github.com/tucanbit/internal/storage/groove"
	"github.com/tucanbit/platform"
	"github.com/tucanbit/platform/clickhouse"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

func Initiate() {
	fmt.Println("Initializing application components")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	logger.Info("Logger initialized successfully")

	//initializing config
	logger.Info("initializing config ")
	configName := "config"
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}
	initConfig(configName, "config", logger)
	logger.Info("initializing config completed")

	// initailizing database connection
	logger.Info("initializing database connect")
	pgxPool, gormDB := initDB(viper.GetString("db.url"), logger)
	logger.Info("database connection initialized")

	// initializing persistence layer which is responsible to communicate with the database and module layer
	// which is used as middleware between database and module layer of the application

	logger.Info("initializing persistence layer ")
	persistenceDB := persistencedb.New(pgxPool, logger)
	logger.Info("done initializing persistence layer")

	// initializing module layer
	// this layer is used to make logical operation of the program
	// create casbin enforcer here

	locker := make(map[uuid.UUID]*sync.Mutex)
	logger.Info("initializing module layer")
	adapter, err := gormadapter.NewAdapterByDB(gormDB)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin adapter: %v", err)
	}
	enforcer, err := casbin.NewEnforcer("./config/RBAC.conf", adapter)
	if err != nil {
		log.Fatal(err)
	}
	enforcer.LoadPolicy()
	lgr := InitEnhancedLogger()
	defer lgr.Close()
	// initializing platform
	logger.Info("initializing platform layer")
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Platform initialization panicked", zap.Any("panic", r))
			panic(r)
		}
	}()
	platformInstance := platform.InitPlatform(context.Background(), lgr)
	logger.Info("done initializing platform layer")

	// Initialize ClickHouse client
	logger.Info("initializing ClickHouse client")
	clickhouseConfig := clickhouse.ClickHouseConfig{
		Host:     viper.GetString("clickhouse.host"),
		Port:     viper.GetInt("clickhouse.port"),
		Database: viper.GetString("clickhouse.database"),
		Username: viper.GetString("clickhouse.username"),
		Password: viper.GetString("clickhouse.password"),
		Timeout:  viper.GetDuration("clickhouse.timeout"),
	}
	logger.Info("ClickHouse config",
		zap.String("host", clickhouseConfig.Host),
		zap.Int("port", clickhouseConfig.Port),
		zap.String("database", clickhouseConfig.Database),
		zap.String("username", clickhouseConfig.Username))

	clickhouseClient, err := clickhouse.NewClickHouseClient(clickhouseConfig, logger)
	if err != nil {
		logger.Error("Failed to initialize ClickHouse client", zap.Error(err))
		// Continue without ClickHouse for now
		clickhouseClient = nil
	} else {
		logger.Info("ClickHouse client initialized successfully")

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := clickhouseClient.HealthCheck(ctx); err != nil {
			logger.Error("ClickHouse health check failed", zap.Error(err))
			clickhouseClient = nil
		} else {
			logger.Info("ClickHouse health check passed")
		}
	}

	// Initialize userWS first
	userBalanceWs := utils.InitUserws(logger, nil, nil) // Will be updated after persistence is created

	// Now initialize persistence with Redis, userWS, and ClickHouse
	var redisOTP *redis.RedisOTP
	if realRedis, ok := platformInstance.Redis.(*redis.RedisOTP); ok {
		redisOTP = realRedis
	} else {
		logger.Warn("Using mock Redis client for persistence initialization")
		redisOTP = nil
	}
	persistence := initPersistence(&persistenceDB, logger, gormDB, redisOTP, userBalanceWs, clickhouseClient)

	// Update userWS with the actual balance storage and Redis client
	userBalanceWs = utils.InitUserws(logger, persistence.Balance, platformInstance.Redis)

	// Initialize email services
	logger.Info("initializing email services")
	// Using smtp configuration from config.yaml
	envSmtpPassword := os.Getenv("SMTP_PASSWORD")
	envSmtpUsername := os.Getenv("SMTP_USERNAME")

	// Read from smtp configuration
	smtpPasswordRaw := viper.GetString("smtp.password")
	// Remove any potential leading/trailing quotes from YAML parsing, but keep internal spaces
	smtpPassword := strings.Trim(smtpPasswordRaw, `"'`)
	// Use smtp.username as username (email address)
	smtpUsername := strings.TrimSpace(viper.GetString("smtp.username"))
	if smtpUsername == "" {
		// Fallback to smtp.from if username is not set
		smtpUsername = strings.TrimSpace(viper.GetString("smtp.from"))
	}
	smtpHost := viper.GetString("smtp.host")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := viper.GetInt("smtp.port")
	if smtpPort == 0 {
		smtpPort = 465
	}
	smtpFrom := strings.TrimSpace(viper.GetString("smtp.from"))
	if smtpFrom == "" {
		smtpFrom = smtpUsername
	}
	smtpFromName := viper.GetString("smtp.from_name")
	smtpUseTLS := viper.GetBool("smtp.use_tls")

	// Log all potential sources to diagnose server issues
	logger.Info("Loading SMTP configuration from smtp (via viper)",
		zap.String("config_file_used", viper.ConfigFileUsed()),
		zap.String("config_source", "viper.GetString('smtp.*') from config.yaml"),
		zap.String("password_source", "viper.GetString('smtp.password')"),
		zap.String("username_source", "viper.GetString('smtp.username') or 'smtp.from'"),
		zap.String("env_SMTP_PASSWORD_set", fmt.Sprintf("%v", envSmtpPassword != "")),
		zap.String("env_SMTP_USERNAME_set", fmt.Sprintf("%v", envSmtpUsername != "")),
		zap.String("env_SMTP_PASSWORD_length", fmt.Sprintf("%d", len(envSmtpPassword))),
		zap.String("env_SMTP_USERNAME_length", fmt.Sprintf("%d", len(envSmtpUsername))),
		zap.String("host", smtpHost),
		zap.Int("port", smtpPort),
		zap.String("username", smtpUsername),
		zap.String("username_raw", smtpUsername),
		zap.Int("username_length", len(smtpUsername)),
		zap.Bool("password_set", smtpPassword != ""),
		zap.Int("password_length", len(smtpPassword)),
		zap.String("password_preview", func() string {
			if len(smtpPassword) > 4 {
				return smtpPassword[:4] + "..."
			}
			return "***"
		}()),
		zap.String("password_raw", smtpPassword),
		zap.String("from", smtpFrom),
		zap.String("from_name", smtpFromName),
		zap.Bool("use_tls", smtpUseTLS))
	smtpConfig := emailModule.SMTPConfig{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUsername,
		Password: smtpPassword,
		From:     smtpFrom,
		FromName: smtpFromName,
		UseTLS:   smtpUseTLS,
	}
	emailService, err := emailModule.NewEmailService(smtpConfig, logger)
	if err != nil {
		logger.Error("Failed to initialize email service", zap.Error(err))
		emailService = nil
	} else {
		logger.Info("Email service initialized successfully")
	}

	// Update GrooveStorage with the proper userWS and analytics integration
	var analyticsIntegration *analyticsStorage.AnalyticsIntegration
	var dailyReportService analyticsModule.DailyReportService
	var dailyReportCronjobService analyticsModule.DailyReportCronjobService
	var alertCronjobService alertModule.AlertCronjobService
	if clickhouseClient != nil {
		analyticsStorageInstance := analyticsStorage.NewAnalyticsStorage(clickhouseClient, logger)
		syncService := analyticsModule.NewSyncService(analyticsStorageInstance, persistence.Groove, logger)
		realtimeSyncService := analyticsModule.NewRealtimeSyncService(syncService, analyticsStorageInstance, logger)
		analyticsIntegration = analyticsStorage.NewAnalyticsIntegration(realtimeSyncService, logger)

		// Initialize daily report email service
		if emailService != nil {
			dailyReportEmailService := emailModule.NewDailyReportEmailService(emailService, logger)
			dailyReportService = analyticsModule.NewDailyReportService(analyticsStorageInstance, dailyReportEmailService, logger)

			// Initialize daily report cronjob service
			dailyReportCronjobService = analyticsModule.NewDailyReportCronjobService(logger, dailyReportService)

			logger.Info("Daily report email service initialized successfully")
			logger.Info("Daily report cronjob service initialized successfully")
		}

		logger.Info("Analytics integration initialized successfully")
	} else {
		logger.Warn("ClickHouse client not available, analytics integration disabled")
	}

	// Initialize alert service and cronjob
	if emailService != nil {
		// Create email service adapter for alert service
		// The alert service needs an AlertEmailSender interface
		alertEmailSender := &alertEmailServiceAdapter{
			emailService: emailService,
		}

		alertService := alertModule.NewAlertService(
			persistence.Alert,
			persistence.AlertEmailGroups,
			alertEmailSender,
			logger,
		)
		alertCronjobService = alertModule.NewAlertCronjobService(alertService, logger)
		logger.Info("Alert service initialized successfully")
		logger.Info("Alert cronjob service initialized successfully")
	}
	persistence.Groove = groove.NewGrooveStorage(&persistenceDB, userBalanceWs, analyticsIntegration, logger)

	module := initModule(persistence, logger, locker, enforcer, userBalanceWs, platformInstance.Kafka, redisOTP, platformInstance.Pisi)

	// Start cashback Kafka consumer for real-time bet processing
	if module.CashbackKafkaConsumer != nil {
		logger.Info("Starting cashback Kafka consumer")
		if err := module.CashbackKafkaConsumer.StartConsumer(context.Background()); err != nil {
			logger.Error("Failed to start cashback Kafka consumer", zap.Error(err))
		} else {
			logger.Info("Cashback Kafka consumer started successfully")
		}

		// Start expired cashback processing job
		go module.CashbackKafkaConsumer.ProcessExpiredCashbackJob(context.Background())
		logger.Info("Expired cashback processing job started")
	}

	// Start rakeback scheduler for automatic schedule activation/deactivation
	if module.Cashback != nil {
		logger.Info("Starting rakeback scheduler")
		rakebackScheduler := module.Cashback.GetRakebackScheduler(persistence.Cashback, logger)
		rakebackScheduler.Start(context.Background())
		logger.Info("Rakeback scheduler started successfully - checking every 1 minute")
	}

	// Start daily report cronjob service
	if dailyReportCronjobService != nil {
		logger.Info("Starting daily report cronjob service")
		if err := dailyReportCronjobService.StartScheduler(context.Background()); err != nil {
			logger.Error("Failed to start daily report cronjob service", zap.Error(err))
		} else {
			logger.Info("Daily report cronjob service started successfully")
			logger.Info("Daily reports will be sent automatically at 23:59 UTC to configured recipients")
		}
	}

	// Start alert cronjob service
	if alertCronjobService != nil {
		logger.Info("Starting alert cronjob service")
		if err := alertCronjobService.StartScheduler(context.Background()); err != nil {
			logger.Error("Failed to start alert cronjob service", zap.Error(err))
		} else {
			logger.Info("Alert cronjob service started successfully")
			logger.Info("Alerts will be checked automatically every minute")
		}
	}

	logger.Info("done initializing module layer")

	// initializing handler layer
	// which is the layer responsible to handle http layer and validate user
	logger.Info("initializing handler layer ")
	handler := initHandler(module, persistence, logger, userBalanceWs, dailyReportService, dailyReportCronjobService)
	logger.Info("done initializing handler layer")

	logger.Info("initializing http server")
	server := gin.New()
	server.Use(middleware.GinLogger(*logger))
	server.Use(middleware.CORS())
	server.Use(middleware.PanicRecovery(logger))
	server.Use(middleware.ErrorHandler())
	ginsrv := server.Group("")

	ginsrv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("done initializing server")

	// initializing route which handle route endpoints
	logger.Info("initializing route")
	initRoute(ginsrv, handler, module, logger, enforcer, persistence)
	logger.Info("done initializing route")

	logger.Info("Server configuration",
		zap.String("host", viper.GetString("app.host")),
		zap.Int("port", viper.GetInt("app.port")),
	)

	logger.Info("initializing server")
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", viper.GetString("app.host"), viper.GetInt("app.port")),
		Handler:           server,
		ReadHeaderTimeout: viper.GetDuration("app.timeout"),
		IdleTimeout:       30 * time.Minute,
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT)
		<-sigint
		log.Fatal("HTTP server Shutdown")

	}()
	logger.Info(fmt.Sprintf("http server listening on port : %d", viper.GetInt("app.port")))

	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Could not start HTTP server: %s", err))
	}

}

// alertEmailServiceAdapter adapts EmailService to AlertEmailSender interface
type alertEmailServiceAdapter struct {
	emailService emailModule.EmailService
}

func (a *alertEmailServiceAdapter) SendAlertEmail(ctx context.Context, to []string, alertConfig *dto.AlertConfiguration, trigger *dto.AlertTrigger) error {
	return a.emailService.SendAlertEmail(ctx, to, alertConfig, trigger)
}
