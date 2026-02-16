package initiator

import (
	"fmt"
	"os"
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

	// Log google.smtp configuration to check if environment variables are overriding
	googleSmtpPassword := viper.GetString("google.smtp.password")
	googleSmtpFrom := viper.GetString("google.smtp.from")
	envGoogleSmtpPassword := os.Getenv("GOOGLE_SMTP_PASSWORD")
	envGoogleSmtpFrom := os.Getenv("GOOGLE_SMTP_FROM")

	log.Info("Google SMTP configuration check",
		zap.String("config_file_used", viper.ConfigFileUsed()),
		zap.String("google.smtp.password_set", fmt.Sprintf("%v", googleSmtpPassword != "")),
		zap.String("google.smtp.from", googleSmtpFrom),
		zap.String("google.smtp.password_length", fmt.Sprintf("%d", len(googleSmtpPassword))),
		zap.String("GOOGLE_SMTP_PASSWORD_env_set", fmt.Sprintf("%v", envGoogleSmtpPassword != "")),
		zap.String("GOOGLE_SMTP_FROM_env_set", fmt.Sprintf("%v", envGoogleSmtpFrom != "")),
		zap.String("GOOGLE_SMTP_PASSWORD_env_length", fmt.Sprintf("%d", len(envGoogleSmtpPassword))),
		zap.String("GOOGLE_SMTP_FROM_env_value", envGoogleSmtpFrom),
		zap.Bool("env_override_detected", envGoogleSmtpPassword != "" || envGoogleSmtpFrom != ""),
		zap.Bool("password_matches_env", envGoogleSmtpPassword != "" && googleSmtpPassword == envGoogleSmtpPassword),
		zap.Bool("username_matches_env", envGoogleSmtpFrom != "" && googleSmtpFrom == envGoogleSmtpFrom))
}
