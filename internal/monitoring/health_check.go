package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a specific component
type ComponentHealth struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
	Details   interface{}  `json:"details,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
	Metrics    map[string]interface{}     `json:"metrics,omitempty"`
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	db          *sql.DB
	redisClient *redis.Client
	metrics     *EnterpriseRegistrationMetrics
	logger      *zap.Logger
	startTime   time.Time
	version     string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, redisClient *redis.Client, metrics *EnterpriseRegistrationMetrics, logger *zap.Logger, version string) *HealthChecker {
	return &HealthChecker{
		db:          db,
		redisClient: redisClient,
		metrics:     metrics,
		logger:      logger,
		startTime:   time.Now(),
		version:     version,
	}
}

// CheckDatabaseHealth checks the database health
func (h *HealthChecker) CheckDatabaseHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	// Test database connection
	err := h.db.PingContext(ctx)
	duration := time.Since(start)

	health := ComponentHealth{
		Timestamp: time.Now(),
	}

	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Database connection failed: %v", err)
		health.Details = map[string]interface{}{
			"error":    err.Error(),
			"duration": duration.String(),
		}

		// Record metrics
		if h.metrics != nil {
			h.metrics.RecordDatabaseError("ping", "database", "connection_failed")
			h.metrics.RecordDatabaseOperationDuration("ping", "database", duration)
		}

		h.logger.Error("Database health check failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	// Check if enterprise_registrations table exists
	var tableExists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'enterprise_registrations'
		)
	`

	err = h.db.QueryRowContext(ctx, query).Scan(&tableExists)
	if err != nil {
		health.Status = HealthStatusDegraded
		health.Message = "Database connected but table check failed"
		health.Details = map[string]interface{}{
			"connection":  "ok",
			"table_check": "failed",
			"error":       err.Error(),
			"duration":    duration.String(),
		}

		if h.metrics != nil {
			h.metrics.RecordDatabaseError("table_check", "enterprise_registrations", "query_failed")
		}

		h.logger.Warn("Database table check failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	if !tableExists {
		health.Status = HealthStatusDegraded
		health.Message = "Database connected but enterprise_registrations table missing"
		health.Details = map[string]interface{}{
			"connection":   "ok",
			"table_exists": false,
			"duration":     duration.String(),
		}

		h.logger.Warn("Enterprise registrations table missing",
			zap.Duration("duration", duration))
		return health
	}

	// Check table statistics
	var count int
	err = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enterprise_registrations").Scan(&count)
	if err != nil {
		health.Status = HealthStatusDegraded
		health.Message = "Database connected but statistics query failed"
		health.Details = map[string]interface{}{
			"connection":   "ok",
			"table_exists": true,
			"stats_query":  "failed",
			"error":        err.Error(),
			"duration":     duration.String(),
		}

		if h.metrics != nil {
			h.metrics.RecordDatabaseError("stats_query", "enterprise_registrations", "query_failed")
		}

		h.logger.Warn("Database statistics query failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	health.Status = HealthStatusHealthy
	health.Message = "Database healthy"
	health.Details = map[string]interface{}{
		"connection":   "ok",
		"table_exists": true,
		"record_count": count,
		"duration":     duration.String(),
	}

	// Record metrics
	if h.metrics != nil {
		h.metrics.RecordDatabaseOperationDuration("health_check", "enterprise_registrations", duration)
	}

	h.logger.Debug("Database health check passed",
		zap.Duration("duration", duration),
		zap.Int("record_count", count))

	return health
}

// CheckRedisHealth checks the Redis health
func (h *HealthChecker) CheckRedisHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	health := ComponentHealth{
		Timestamp: time.Now(),
	}

	if h.redisClient == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Redis client not configured"
		health.Details = map[string]interface{}{
			"error":    "redis client is nil",
			"duration": time.Since(start).String(),
		}

		h.logger.Error("Redis health check failed - client not configured")
		return health
	}

	// Test Redis connection
	err := h.redisClient.Ping(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Redis connection failed: %v", err)
		health.Details = map[string]interface{}{
			"error":    err.Error(),
			"duration": duration.String(),
		}

		h.logger.Error("Redis health check failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	// Test OTP storage functionality
	testKey := "health_check:test"
	testValue := "test_value"

	// Set test value
	err = h.redisClient.Set(ctx, testKey, testValue, time.Minute).Err()
	if err != nil {
		health.Status = HealthStatusDegraded
		health.Message = "Redis connected but write operation failed"
		health.Details = map[string]interface{}{
			"connection": "ok",
			"write_test": "failed",
			"error":      err.Error(),
			"duration":   duration.String(),
		}

		h.logger.Warn("Redis write test failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	// Get test value
	val, err := h.redisClient.Get(ctx, testKey).Result()
	if err != nil {
		health.Status = HealthStatusDegraded
		health.Message = "Redis connected but read operation failed"
		health.Details = map[string]interface{}{
			"connection": "ok",
			"write_test": "ok",
			"read_test":  "failed",
			"error":      err.Error(),
			"duration":   duration.String(),
		}

		h.logger.Warn("Redis read test failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return health
	}

	// Clean up test key
	h.redisClient.Del(ctx, testKey)

	if val != testValue {
		health.Status = HealthStatusDegraded
		health.Message = "Redis connected but data integrity check failed"
		health.Details = map[string]interface{}{
			"connection":     "ok",
			"write_test":     "ok",
			"read_test":      "ok",
			"data_integrity": "failed",
			"expected":       testValue,
			"actual":         val,
			"duration":       duration.String(),
		}

		h.logger.Warn("Redis data integrity check failed",
			zap.String("expected", testValue),
			zap.String("actual", val),
			zap.Duration("duration", duration))
		return health
	}

	health.Status = HealthStatusHealthy
	health.Message = "Redis healthy"
	health.Details = map[string]interface{}{
		"connection":     "ok",
		"write_test":     "ok",
		"read_test":      "ok",
		"data_integrity": "ok",
		"duration":       duration.String(),
	}

	h.logger.Debug("Redis health check passed",
		zap.Duration("duration", duration))

	return health
}

// CheckEmailServiceHealth checks the email service health
func (h *HealthChecker) CheckEmailServiceHealth(ctx context.Context) ComponentHealth {
	health := ComponentHealth{
		Timestamp: time.Now(),
	}

	// For now, we'll just check if the metrics are available
	// In a real implementation, you might want to test actual email sending
	if h.metrics == nil {
		health.Status = HealthStatusDegraded
		health.Message = "Email service metrics not available"
		health.Details = map[string]interface{}{
			"metrics": "not_configured",
		}

		h.logger.Warn("Email service health check - metrics not configured")
		return health
	}

	health.Status = HealthStatusHealthy
	health.Message = "Email service healthy"
	health.Details = map[string]interface{}{
		"metrics":  "available",
		"provider": "smtp",
	}

	h.logger.Debug("Email service health check passed")
	return health
}

// CheckOverallHealth performs a comprehensive health check
func (h *HealthChecker) CheckOverallHealth(ctx context.Context) SystemHealth {
	start := time.Now()

	health := SystemHealth{
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime).String(),
		Version:    h.version,
		Components: make(map[string]ComponentHealth),
	}

	// Check individual components
	dbHealth := h.CheckDatabaseHealth(ctx)
	redisHealth := h.CheckRedisHealth(ctx)
	emailHealth := h.CheckEmailServiceHealth(ctx)

	health.Components["database"] = dbHealth
	health.Components["redis"] = redisHealth
	health.Components["email_service"] = emailHealth

	// Determine overall status
	unhealthyCount := 0
	degradedCount := 0

	for _, component := range health.Components {
		switch component.Status {
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		health.Status = HealthStatusUnhealthy
	} else if degradedCount > 0 {
		health.Status = HealthStatusDegraded
	} else {
		health.Status = HealthStatusHealthy
	}

	// Add metrics summary if available
	if h.metrics != nil {
		health.Metrics = h.metrics.GetMetricsSummary(ctx)
	}

	duration := time.Since(start)
	h.logger.Info("Overall health check completed",
		zap.String("status", string(health.Status)),
		zap.Int("unhealthy_components", unhealthyCount),
		zap.Int("degraded_components", degradedCount),
		zap.Duration("duration", duration))

	return health
}

// HealthCheckHandler handles HTTP health check requests
func (h *HealthChecker) HealthCheckHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if it's a detailed health check
	detailed := c.Query("detailed") == "true"

	if detailed {
		health := h.CheckOverallHealth(ctx)
		c.JSON(http.StatusOK, health)
		return
	}

	// Quick health check
	health := h.CheckOverallHealth(ctx)

	// Return appropriate HTTP status
	status := http.StatusOK
	if health.Status == HealthStatusUnhealthy {
		status = http.StatusServiceUnavailable
	} else if health.Status == HealthStatusDegraded {
		status = http.StatusPartialContent
	}

	c.JSON(status, gin.H{
		"status":    health.Status,
		"timestamp": health.Timestamp,
		"uptime":    health.Uptime,
		"version":   health.Version,
	})
}

// ReadinessHandler handles readiness probe requests
func (h *HealthChecker) ReadinessHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Quick check of critical components
	dbHealth := h.CheckDatabaseHealth(ctx)
	redisHealth := h.CheckRedisHealth(ctx)

	if dbHealth.Status == HealthStatusUnhealthy || redisHealth.Status == HealthStatusUnhealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"message":   "Critical services unavailable",
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"message":   "All critical services available",
		"timestamp": time.Now(),
	})
}

// LivenessHandler handles liveness probe requests
func (h *HealthChecker) LivenessHandler(c *gin.Context) {
	// Liveness check is simple - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"message":   "Service is running",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// MetricsHandler handles metrics endpoint requests
func (h *HealthChecker) MetricsHandler(c *gin.Context) {
	if h.metrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics not available",
		})
		return
	}

	ctx := c.Request.Context()
	summary := h.metrics.GetMetricsSummary(ctx)

	c.JSON(http.StatusOK, summary)
}

// RegisterHealthRoutes registers health check routes
func (h *HealthChecker) RegisterHealthRoutes(router *gin.Engine) {
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("", h.HealthCheckHandler)
		healthGroup.GET("/detailed", h.HealthCheckHandler)
		healthGroup.GET("/ready", h.ReadinessHandler)
		healthGroup.GET("/live", h.LivenessHandler)
		healthGroup.GET("/metrics", h.MetricsHandler)
	}

	h.logger.Info("Health check routes registered",
		zap.Strings("endpoints", []string{
			"/health",
			"/health/detailed",
			"/health/ready",
			"/health/live",
			"/health/metrics",
		}))
}
