FROM golang:1.23-bullseye AS builder

# Install migration tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app
COPY . .

# Build the application
RUN go build -o egyptkingcrash cmd/main.go

FROM debian:bullseye-slim

WORKDIR /app

# Copy required files
COPY --from=builder /app/egyptkingcrash .
COPY --from=builder /app/config ./config
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/internal/constant/query/schemas ./internal/constant/query/schemas

# Add wait-for-it script
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh .
RUN chmod +x wait-for-it.sh

EXPOSE 8000

CMD ["./egyptkingcrash"]