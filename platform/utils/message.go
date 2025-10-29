package utils

import (
	"context"

	"net/smtp"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/spf13/viper"
)

func SendEmail(ctx context.Context, emailReq dto.EmailReq) error {
	smtpPassword := viper.GetString("google.smtp.password")
	smtpFrom := viper.GetString("google.smtp.from")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", smtpFrom, smtpPassword, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpFrom, emailReq.To, emailReq.Body)
}
