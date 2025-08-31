package initiator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/platform"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	lgr := InitLogger()
	// initializing platform
	logger.Info("initializing platform layer")
	platformInstance := platform.InitPlatform(context.Background(), lgr)
	logger.Info("done initializing platform layer")
	
	// Now initialize persistence with Redis
	persistence := initPersistence(&persistenceDB, logger, gormDB, platformInstance.Redis.(*redis.RedisOTP))
	userBalanceWs := utils.InitUserws(logger, persistence.Balance)
	
	module := initModule(persistence, logger, locker, enforcer, userBalanceWs, platformInstance.Kafka, platformInstance.Redis.(*redis.RedisOTP), platformInstance.Pisi)

	logger.Info("done initializing module layer")

	// initializing handler layer
	// which is the layer responsible to handle http layer and validate user
	logger.Info("initializing handler layer ")
	handler := initHandler(module, logger, userBalanceWs)
	logger.Info("done initializing handler layer")

	logger.Info("initializing http server")
	server := gin.New()
	server.Use(middleware.GinLogger(*logger))
	server.Use(middleware.CORS())
	server.Use(middleware.ErrorHandler())
	ginsrv := server.Group("")

	ginsrv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("done initializing server")

	// initializing route which handle route endpoints
	logger.Info("initializing route")
	initRoute(ginsrv, handler, module, logger, enforcer)
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
