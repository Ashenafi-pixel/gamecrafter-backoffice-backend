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
	// Required for Supabase (and other transaction-mode poolers): avoid "prepared statement does not exist"
	config.ConnConfig.PreferSimpleProtocol = true
	log.Info("Database pool configured with PreferSimpleProtocol=true for pooler compatibility")
	config.ConnConfig.PreferSimpleProtocol = true
	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// GORM: use simple protocol so Supabase/pooler does not hit "prepared statement does not exist"
	gormDialector := postgres.New(postgres.Config{
		DSN:                  dbUrl,
		PreferSimpleProtocol: true,
	})
	db, err := gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	return conn, db
}
