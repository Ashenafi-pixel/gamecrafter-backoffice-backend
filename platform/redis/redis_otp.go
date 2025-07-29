package redis

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/platform/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewRedisOTP creates a new RedisOTP client
func NewRedisOTP(redisAddr, redisPassword string, db int, keyPrefix string, otpTTL time.Duration, attemptLimit int, logger logger.Logger) (*RedisOTP, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		err = errors.ErrRedisClientError.Wrap(err, "failed to connect to Redis")
		logger.Error(ctx, "failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	privateKey, err := loadRSAKeysFromFiles(logger)
	if err != nil {
		logger.Info(ctx, "Failed to load existing RSA keys from files, generating new ones", zap.Error(err))
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			err = errors.ErrRedisClientError.Wrap(err, "failed to generate RSA key pair")
			logger.Error(ctx, "failed to generate RSA key pair", zap.Error(err))
			return nil, err
		}
		if err := storeRSAKeysToFiles(privateKey, logger); err != nil {
			logger.Error(ctx, "Failed to store RSA keys to files", zap.Error(err))
		}
	} else {
		logger.Info(ctx, "Successfully loaded existing RSA keys from project files")
	}

	logger.Info(ctx, "Redis OTP client initialized successfully",
		zap.String("redis_addr", redisAddr),
		zap.String("key_prefix", keyPrefix),
		zap.Duration("otp_ttl", otpTTL),
	)

	return &RedisOTP{
		redisClient:  rdb,
		privateKey:   privateKey,
		publicKey:    &privateKey.PublicKey,
		keyPrefix:    keyPrefix,
		otpTTL:       otpTTL,
		logger:       logger,
		AttemptLimit: attemptLimit,
	}, nil
}

// NewRedisOTPFromConfig creates a new RedisOTP client using viper config
func NewRedisOTPFromConfig(logger logger.Logger) (*RedisOTP, error) {
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")
	keyPrefix := viper.GetString("redis.key_prefix")
	otpTTL := viper.GetDuration("redis.ttl")
	if otpTTL == 0 {
		otpTTL = 5 * time.Second // default if not set
	}
	attemptLimit := viper.GetInt("auth.otp_attempt_limit")
	if attemptLimit == 0 {
		attemptLimit = 4 // default fallback
	}
	return NewRedisOTP(redisAddr, redisPassword, db, keyPrefix, otpTTL, attemptLimit, logger)
}

// SaveOTP encrypts and saves the OTP in Redis, mapped by phone number
func (r *RedisOTP) SaveOTP(ctx context.Context, phone, otp string) error {
	redisKey := r.otpKey(phone)
	r.logger.Info(ctx, "Saving OTP", zap.String("phone", phone), zap.String("otp", otp), zap.String("ttl", r.otpTTL.String()), zap.String("redis_key", redisKey))
	encryptedOTP, err := r.encryptOTP(otp)
	if err != nil {
		r.logger.Error(ctx, "encryptOTP failed", zap.Error(err))
		return err
	}
	mapping := OTPMapping{
		EncryptedOTP: encryptedOTP,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(r.otpTTL),
	}
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		r.logger.Error(ctx, "marshal OTPMapping failed", zap.Error(err))
		return err
	}
	r.logger.Info(ctx, "About to SET in Redis", zap.String("redis_key", redisKey), zap.String("mapping_json", string(mappingJSON)), zap.Duration("ttl", r.otpTTL))
	err = r.redisClient.Set(ctx, redisKey, mappingJSON, r.otpTTL).Err()
	if err != nil {
		r.logger.Error(ctx, "Redis SET failed", zap.Error(err))
		return err
	}
	// Reset attempt count when new OTP is set
	attemptKey := redisKey + ":attempts"
	err = r.redisClient.Set(ctx, attemptKey, 0, r.otpTTL).Err()
	if err != nil {
		r.logger.Error(ctx, "Redis SET attempt count failed", zap.Error(err))
		return err
	}
	// Confirm value is in Redis
	val, getErr := r.redisClient.Get(ctx, redisKey).Result()
	if getErr != nil {
		getErr = errors.ErrRedisClientError.Wrap(getErr, "failed to get OTP from Redis")
		r.logger.Error(ctx, "Redis GET after SET failed", zap.Error(getErr))
		return getErr
	}
	r.logger.Info(ctx, "Redis GET after SET succeeded", zap.String("value", val))
	return nil
}

// IncrementOTPAttempt increments the OTP attempt count and returns the new value
func (r *RedisOTP) IncrementOTPAttempt(ctx context.Context, phone string) (int, error) {
	redisKey := r.otpKey(phone) + ":attempts"
	count, err := r.redisClient.Incr(ctx, redisKey).Result()
	if err != nil {
		r.logger.Error(ctx, "Redis INCR attempt count failed", zap.Error(err))
		return 0, err
	}
	// Set expiry to match OTP expiry if not already set
	if ttl, _ := r.redisClient.TTL(ctx, redisKey).Result(); ttl < 0 {
		// fallback to 5 minutes if OTP key is missing
		r.redisClient.Expire(ctx, redisKey, 5*time.Minute)
	}
	return int(count), nil
}

// GetOTPAttemptCount returns the current OTP attempt count
func (r *RedisOTP) GetOTPAttemptCount(ctx context.Context, phone string) (int, error) {
	redisKey := r.otpKey(phone) + ":attempts"
	count, err := r.redisClient.Get(ctx, redisKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		r.logger.Error(ctx, "Redis GET attempt count failed", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// ResetOTPAttempts resets the OTP attempt count (used when new OTP is sent)
func (r *RedisOTP) ResetOTPAttempts(ctx context.Context, phone string) error {
	redisKey := r.otpKey(phone) + ":attempts"
	return r.redisClient.Set(ctx, redisKey, 0, r.otpTTL).Err()
}

// VerifyAndRemoveOTP checks the OTP for a phone, removes it if valid, and returns true if valid
func (r *RedisOTP) VerifyAndRemoveOTP(ctx context.Context, phone, otp string) (bool, error) {
	redisKey := r.otpKey(phone)
	// Check attempt count first
	attempts, err := r.GetOTPAttemptCount(ctx, phone)
	if err != nil {
		return false, err
	}
	if attempts >= r.AttemptLimit {
		err = errors.ErrAcessError.New("Too many invalid attempts. Please request a new OTP.")
		r.logger.Warn(ctx, "OTP attempts exceeded, blocking further attempts", zap.String("phone", phone))
		// Return error first, then clear OTP and attempts asynchronously
		r.redisClient.Del(ctx, redisKey)
		r.redisClient.Del(ctx, redisKey+":attempts")
		return false, err
	}

	mappingJSON, err := r.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		r.logger.Error(ctx, "Redis GET failed", zap.Error(err))
		_, err = r.IncrementOTPAttempt(ctx, phone)
		if err != nil {
			err = errors.ErrAcessError.New("unable to count login attempt")
		}
		err = errors.ErrAcessError.New("otp expired or invalid")
		return false, err
	}
	r.logger.Info(ctx, "Redis GET succeeded", zap.String("mapping_json", mappingJSON))
	var mapping OTPMapping
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		r.logger.Error(ctx, "Unmarshal OTPMapping failed", zap.Error(err))
		_, err = r.IncrementOTPAttempt(ctx, phone)
		if err != nil {
			err = errors.ErrAcessError.New("unable to count login attempt")
		}
		return false, err
	}
	if time.Now().After(mapping.ExpiresAt) {
		r.logger.Warn(ctx, "OTP expired", zap.Time("expires_at", mapping.ExpiresAt))
		r.redisClient.Del(ctx, redisKey)
		_, err = r.IncrementOTPAttempt(ctx, phone)
		if err != nil {
			err = errors.ErrAcessError.New("unable to count login attempt")
		}
		return false, nil // Expired
	}
	decryptedOTP, err := r.decryptOTP(mapping.EncryptedOTP)
	if err != nil {
		r.logger.Error(ctx, "decryptOTP failed", zap.Error(err))
		_, err = r.IncrementOTPAttempt(ctx, phone)
		if err != nil {
			err = errors.ErrAcessError.New("unable to count login attempt")
		}
		return false, err
	}
	if decryptedOTP != otp {
		_, err = r.IncrementOTPAttempt(ctx, phone)
		if err != nil {
			err = errors.ErrRedisClientError.New("unable to count login attempt")
		}
		err = errors.ErrAcessError.New("otp expired or invalid")
		r.logger.Warn(ctx, "OTP does not match", zap.String("expected", decryptedOTP), zap.String("got", otp))
		return false, err
	}
	// Remove OTP after successful verification
	r.logger.Info(ctx, "OTP verified, deleting from Redis", zap.String("redis_key", redisKey))
	r.redisClient.Del(ctx, redisKey)
	return true, nil
}

// encryptOTP encrypts the OTP using RSA public key
func (r *RedisOTP) encryptOTP(otp string) (string, error) {
	otpBytes := []byte(otp)
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, otpBytes)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to encrypt OTP using RSA")
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// decryptOTP decrypts the OTP using RSA private key
func (r *RedisOTP) decryptOTP(encryptedOTP string) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedOTP)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to decode encrypted OTP from base64")
	}
	decryptedBytes, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, encryptedBytes)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to decrypt OTP using RSA")
	}
	return string(decryptedBytes), nil
}

// loadRSAKeysFromFiles loads RSA keys from project files
func loadRSAKeysFromFiles(logger logger.Logger) (*rsa.PrivateKey, error) {
	privateKeyPath := "config/rsa_private_key.pem"
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ErrRedisClientError.New("RSA private key file not found")
		}
		return nil, errors.ErrRedisClientError.Wrap(err, "failed to read RSA private key file")
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.ErrRedisClientError.New("failed to decode private key PEM from file")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.ErrRedisClientError.Wrap(err, "failed to parse private key from file")
	}
	logger.Debug(context.Background(), "Successfully loaded RSA private key from file", zap.String("path", privateKeyPath))
	return privateKey, nil
}

// storeRSAKeysToFiles stores RSA keys to project files
func storeRSAKeysToFiles(privateKey *rsa.PrivateKey, logger logger.Logger) error {
	privateKeyPath := "config/rsa_private_key.pem"
	publicKeyPath := "config/rsa_public_key.pem"
	if err := os.MkdirAll("config", 0755); err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to create config directory")
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to marshal public key")
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	if err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to write private key to file")
	}
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	if err != nil {
		return errors.ErrRedisClientError.Wrap(err, "failed to write public key to file")
	}
	logger.Info(context.Background(), "Successfully stored RSA keys to project files",
		zap.String("private_key_path", privateKeyPath),
		zap.String("public_key_path", publicKeyPath),
	)
	return nil
}

// otpKey returns the Redis key for a given phone number
func (r *RedisOTP) otpKey(phone string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, phone)
}

func (r *RedisOTP) GetOTP(ctx context.Context, phone string) (string, error) {
	redisKey := r.otpKey(phone)
	mappingJSON, err := r.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.ErrRedisClientError.New("OTP not found")
		}
		return "", errors.ErrRedisClientError.Wrap(err, "failed to get OTP from Redis")
	}
	var mapping OTPMapping
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to unmarshal OTP mapping")
	}
	if time.Now().After(mapping.ExpiresAt) {
		return "", errors.ErrRedisClientError.New("OTP expired")
	}
	decryptedOTP, err := r.decryptOTP(mapping.EncryptedOTP)
	if err != nil {
		return "", errors.ErrRedisClientError.Wrap(err, "failed to decrypt OTP")
	}
	return decryptedOTP, nil
}
