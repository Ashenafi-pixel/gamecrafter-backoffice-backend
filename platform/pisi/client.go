package pisi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform/logger"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"go.uber.org/zap"
)

type PisiClient interface {
	Login(ctx context.Context) (*LoginResponse, error)
	SendBulkSMS(ctx context.Context, req SendBulkSMSRequest) (string, error)
}

// NewPisiClient creates a new client from viper config
func NewPisiClient(baseURL, password, vaspid string, timeout time.Duration, retryCount int, retryDelay time.Duration, senderID string, logger logger.Logger) PisiClient {
	return &pisiClient{
		baseURL:    baseURL,
		password:   password,
		vaspid:     vaspid,
		timeout:    timeout,
		retryCount: retryCount,
		retryDelay: retryDelay,
		logger:     logger,
		SenderID:   senderID,
	}
}

// Login authenticates and stores the token
func (c *pisiClient) Login(ctx context.Context) (*LoginResponse, error) {
	url := c.baseURL + "Authentication/login"
	body := LoginRequest{
		Password: c.password,
		Vaspid:   c.vaspid,
	}
	resp, err := utils.SendPostHttpRequest(url, body, map[string]string{}, c.timeout)
	if err != nil {
		err = errors.ErrPISISMSError.Wrap(err, "failed to login")
		c.logger.Error(ctx, "failed to login", zap.Error(err))
		return nil, err
	}
	// Marshal/unmarshal to struct
	b, _ := json.Marshal(resp)
	var loginResp LoginResponse
	if err := json.Unmarshal(b, &loginResp); err != nil {
		err = errors.ErrPISISMSError.Wrap(err, "failed to unmarshal login response")
		c.logger.Error(ctx, "failed to unmarshal login response", zap.Error(err))
		return nil, err
	}
	if loginResp.PisiAuthorizationToken != "" {
		c.token = loginResp.PisiAuthorizationToken
	}
	return &loginResp, nil
}

// SendBulkSMS sends an OTP SMS using the token
type SendBulkSMSResponse struct {
	Message string `json:"message"`
}

func (c *pisiClient) SendBulkSMS(ctx context.Context, req SendBulkSMSRequest) (string, error) {
	url := c.baseURL + "Sms/sendbulksms"

	// Login with client before sending SMS
	c.logger.Info(ctx, "logging in before sending SMS")
	loginResp, err := c.Login(ctx)
	if err != nil {
		err = errors.ErrPISISMSError.Wrap(err, "failed to login")
		c.logger.Error(ctx, "failed to login", zap.Error(err))
		return "", err
	}
	c.token = loginResp.PisiAuthorizationToken
	c.logger.Info(ctx, "logged in successfully", zap.String("token", c.token))

	// Send SMS
	c.logger.Info(ctx, "sending SMS")
	var lastErr error
	for i := 0; i < c.retryCount; i++ {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		writer.WriteField("Message", req.Message)
		writer.WriteField("Recipients", req.Recipients)
		writer.WriteField("senderId", c.SenderID)
		writer.Close()

		httpReq, err := http.NewRequest(http.MethodPost, url, &buf)
		if err != nil {
			lastErr = errors.ErrPISISMSError.Wrap(err, "failed to create SMS request")
			c.logger.Error(ctx, "failed to create SMS request", zap.Error(err))
			continue
		}
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())
		httpReq.Header.Set("vaspid", c.vaspid)
		httpReq.Header.Set("pisi-authorization-token", "Bearer "+c.token)

		client := &http.Client{Timeout: c.timeout}
		resp, err := client.Do(httpReq)
		if err != nil {
			lastErr = errors.ErrPISISMSError.Wrap(err, "failed to send SMS request")
			c.logger.Error(ctx, "failed to send SMS request", zap.Error(err))
			time.Sleep(c.retryDelay)
			continue
		}
		defer resp.Body.Close()

		var respBody bytes.Buffer
		_, err = respBody.ReadFrom(resp.Body)
		if err != nil {
			lastErr = errors.ErrPISISMSError.Wrap(err, "failed to read SMS response body")
			c.logger.Error(ctx, "failed to read SMS response body", zap.Error(err))
			time.Sleep(c.retryDelay)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = errors.ErrPISISMSError.Wrap(fmt.Errorf("status: %d, body: %s", resp.StatusCode, respBody.String()), "SMS send failed")
			c.logger.Error(ctx, "SMS send failed", zap.Error(lastErr))
			time.Sleep(c.retryDelay)
			continue
		}
		c.logger.Info(ctx, "SMS sent successfully", zap.String("response", respBody.String()))
		return respBody.String(), nil
	}
	c.logger.Error(ctx, "failed to send SMS", zap.Error(lastErr))
	return "", lastErr
}
