package initiator

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB(dbUrl string, log *zap.Logger) (*pgxpool.Pool, *gorm.DB) {
	log.Info("Attempting to connect to database", zap.String("connection_string", dbUrl))
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Error("unable to parse pgxpool config string")
		log.Fatal(err.Error())
	}

	idleConnTimeout := viper.GetDuration("database.idle_conn_timeout")
	if idleConnTimeout == 0 {
		idleConnTimeout = 4 * time.Minute
	}

	config.MaxConnIdleTime = idleConnTimeout
	config.ConnConfig.PreferSimpleProtocol = true
	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	//add gorm db
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	return conn, db
}
