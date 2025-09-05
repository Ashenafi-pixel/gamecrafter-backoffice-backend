# Database Migration Script
# Install golang-migrate for database migrations
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration directory
mkdir -p /opt/tucanbit/migrations

# Copy migration files
cp -r /home/ashenafi-alemu/Downloads/tucanbit/migrations/* /opt/tucanbit/migrations/

# Run migrations
cd /opt/tucanbit
export PATH=$PATH:/home/ubuntu/go/bin
migrate -path migrations -database 'postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5432/tucanbit?sslmode=disable' up

# Verify database
psql -h localhost -U tucanbit -d tucanbit -c '\dt'
