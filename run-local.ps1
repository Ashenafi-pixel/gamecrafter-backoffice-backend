# PowerShell script to run the application locally with correct environment variables
# This script sets Kafka and other service addresses to localhost for local development

Write-Host "Setting up environment variables for local development..." -ForegroundColor Cyan

# Set Kafka configuration to use localhost
$env:KAFKA_BOOTSTRAP_SERVERS = "localhost:9092"
$env:KAFKA_BOOTSTRAP_SERVER = "localhost:9092"
$env:KAFKA_TOPIC = "bet_transactions"
$env:KAFKA_TOPICS = "events,bet_transactions"
$env:KAFKA_CLUSTER_API_KEY = ""
$env:KAFKA_CLUSTER_API_SECRET = ""
$env:KAFKA_SECURITY_PROTOCOL = "PLAINTEXT"
$env:KAFKA_MECHANISMS = "PLAIN"
$env:KAFKA_ACKS = "all"

# Set Redis configuration to use localhost with mapped port
$env:REDIS_ADDR = "localhost:63790"
$env:REDIS_PASSWORD = ""
$env:REDIS_DB = "0"
$env:REDIS_KEY_PREFIX = "game_crafter:"
$env:REDIS_TTL = "5m"
$env:REDIS_ATTEMPTS = "3"

# Set ClickHouse configuration to use localhost
$env:CLICKHOUSE_HOST = "localhost"
$env:CLICKHOUSE_PORT = "8123"
$env:CLICKHOUSE_DATABASE = "game_crafter_analytics"
$env:CLICKHOUSE_USERNAME = "game_crafter"
$env:CLICKHOUSE_PASSWORD = "game_crafter_clickhouse_password"
$env:CLICKHOUSE_TIMEOUT = "30s"

# Set Database configuration
$env:DB_URL = "postgres://game_crafter_user:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/game_crafter?sslmode=disable"
$env:POSTGRES_PASSWORD = "5kj0YmV5FKKpU9D50B7yH5A"

# Set Application configuration
$env:APP_HOST = "0.0.0.0"
$env:APP_PORT = "8094"
$env:JWT_SECRET = "tokensecrethere"
$env:DB_CONNECT_RETRIES = "10"
$env:DB_CONNECT_TIMEOUT = "5s"

# Set Pisi configuration
$env:PISI_BASE_URL = "http://pisi-service:8080"
$env:PISI_PASSWORD = "pisi_password"
$env:PISI_VASPID = "pisi_vaspid"
$env:PISI_TIMEOUT = "10s"
$env:PISI_RETRY_COUNT = "3"
$env:PISI_RETRY_DELAY = "5s"
$env:PISI_SENDER_ID = "game_crafter"

$env:CONFIG_NAME = "config"

Write-Host "Environment variables set!" -ForegroundColor Green
Write-Host "KAFKA_BOOTSTRAP_SERVERS = $env:KAFKA_BOOTSTRAP_SERVERS" -ForegroundColor Yellow
Write-Host "REDIS_ADDR = $env:REDIS_ADDR" -ForegroundColor Yellow
Write-Host "CLICKHOUSE_HOST = $env:CLICKHOUSE_HOST" -ForegroundColor Yellow
Write-Host ""
Write-Host "Starting application..." -ForegroundColor Cyan
Write-Host ""

# Run the application
go run cmd/main.go



