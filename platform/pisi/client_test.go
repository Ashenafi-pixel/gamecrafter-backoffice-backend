package pisi

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tucanbit/platform/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../config")
	_ = viper.ReadInConfig()
	os.Exit(m.Run())
}

func TestPisiClient_Login(t *testing.T) {
	client := NewPisiClient(
		viper.GetString("pisi.base_url"),
		viper.GetString("pisi.password"),
		viper.GetString("pisi.vaspid"),
		10*time.Second,
		viper.GetInt("pisi.retry_count"),
		time.Duration(viper.GetInt("pisi.retry_delay"))*time.Millisecond,
		"Pisi",
		logger.New(zap.NewNop()),
	)
	ctx := context.Background()
	resp, err := client.Login(ctx)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.PisiAuthorizationToken)
}

func TestPisiClient_SendBulkSMS(t *testing.T) {
	client := NewPisiClient(
		viper.GetString("pisi.base_url"),
		viper.GetString("pisi.password"),
		viper.GetString("pisi.vaspid"),
		10*time.Second,
		viper.GetInt("pisi.retry_count"),
		time.Duration(viper.GetInt("pisi.retry_delay"))*time.Millisecond,
		"Pisi",
		logger.New(zap.NewNop()),
	)
	ctx := context.Background()
	smsReq := SendBulkSMSRequest{
		Message:    "0987",
		Recipients: "2349067382888",
		SenderId:   "Pisi",
	}
	resp, err := client.SendBulkSMS(ctx, smsReq)
	assert.NoError(t, err)
	assert.Contains(t, resp, "Sms transaction ID")
}
