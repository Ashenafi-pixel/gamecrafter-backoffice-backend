package initiator

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initConfig(name, path string, log *zap.Logger) {
	// Load .env file if it exists
	if err := godotenv.Load(".env"); err != nil {
		log.Info("No .env file found, using default configuration")
	}

	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	// Enable environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind JWT_SECRET environment variable to the expected config keys
	viper.BindEnv("app.jwt_secret", "JWT_SECRET")
	viper.BindEnv("auth.jwt_secret", "JWT_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Sprintf("unable to start config %v ", err))
	}

	// Debug: Log the JWT and SMTP configuration values
	log.Info("Configuration loaded",
		zap.String("app.jwt_secret", viper.GetString("app.jwt_secret")),
		zap.String("auth.jwt_secret", viper.GetString("auth.jwt_secret")),
		zap.String("smtp.host", viper.GetString("smtp.host")),
		zap.Int("smtp.port", viper.GetInt("smtp.port")),
		zap.String("smtp.username", viper.GetString("smtp.username")),
		zap.String("smtp.from", viper.GetString("smtp.from")),
		zap.Bool("smtp.use_tls", viper.GetBool("smtp.use_tls")))
}
