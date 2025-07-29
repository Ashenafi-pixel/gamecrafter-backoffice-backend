package redis

import (
	"context"
	"os"
	"testing"

	"github.com/joshjones612/egyptkingcrash/platform/logger"
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

// getTestLogger returns a logger.Logger using zap.NewDevelopment for testing
func getTestLogger() logger.Logger {
	l, _ := zap.NewDevelopment()
	return logger.New(l)
}

func TestRedisOTP_SaveOTP(t *testing.T) {
	log := getTestLogger()
	otpClient, err := NewRedisOTPFromConfig(log)
	assert.NoError(t, err)
	ctx := context.Background()
	phone := "2349067382888"
	otp := "3456"
	err = otpClient.SaveOTP(ctx, phone, otp)
	assert.NoError(t, err)
}
func TestRedisOTP_Verify(t *testing.T) {
	log := getTestLogger()

	// Clean up keys before and after
	os.Remove("config/rsa_private_key.pem")
	os.Remove("config/rsa_public_key.pem")

	otpClient, err := NewRedisOTPFromConfig(log)
	assert.NoError(t, err)
	ctx := context.Background()

	phone := "2349067382888"
	otp := "3456"

	// Save OTP for phone
	err = otpClient.SaveOTP(ctx, phone, otp)
	assert.NoError(t, err)

	// Should verify successfully
	verified, err := otpClient.VerifyAndRemoveOTP(ctx, phone, otp)
	assert.NoError(t, err)
	assert.True(t, verified)
}
