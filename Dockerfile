# Build stage
FROM --platform=linux/amd64 golang:1.24-bullseye AS builder


# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


WORKDIR /app


# Copy go mod files first for better caching
COPY go.mod go.sum ./


# Configure Go module proxy to bypass proxy.golang.org (which is failing with TLS errors)
# and fetch modules directly from the VCS hosts (GitHub, etc.).
ENV GOPROXY=direct


# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download


# Copy the entire source code
COPY . .


# Set Go environment variables
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64


# Build the application
RUN go build -o tucanbit cmd/main.go


FROM debian:bullseye-slim

# Install PostgreSQL client for database operations in entrypoint script
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app


# Copy required files
COPY --from=builder /app/tucanbit .
COPY --from=builder /app/config ./config
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/internal/constant/query/schemas ./internal/constant/query/schemas

# Copy and set up entrypoint script
COPY --from=builder /app/scripts/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Add wait-for-it script
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh .
RUN chmod +x wait-for-it.sh


EXPOSE 8080


CMD ["./tucanbit"]