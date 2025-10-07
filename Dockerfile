# Build stage
FROM golang:1.24-bullseye AS builder


# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


WORKDIR /app


# Copy go mod files first for better caching
COPY go.mod go.sum ./


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


WORKDIR /app


# Copy required files
COPY --from=builder /app/tucanbit .
COPY --from=builder /app/config ./config
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/internal/constant/query/schemas ./internal/constant/query/schemas


# Add wait-for-it script
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh .
RUN chmod +x wait-for-it.sh


EXPOSE 8080


CMD ["./tucanbit"]