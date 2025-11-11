package utils

import (
	"context"

	"net/smtp"

	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
)

func SendEmail(ctx context.Context, emailReq dto.EmailReq) error {
	smtpPassword := viper.GetString("smtp.password")
	smtpFrom := viper.GetString("smtp.from")
	// Use smtp.username for authentication, fallback to smtp.from if not set
	smtpUsername := viper.GetString("smtp.username")
	if smtpUsername == "" {
		smtpUsername = smtpFrom
	}
	smtpHost := viper.GetString("smtp.host")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := viper.GetString("smtp.port")
	if smtpPort == "" {
		smtpPort = "587"
	}
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpFrom, emailReq.To, emailReq.Body)
}
