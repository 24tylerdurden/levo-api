# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install SQLite dependencies and build tools
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Install SQLite runtime and wget for health checks
RUN apk add --no-cache sqlite wget

WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Create directories for data and storage
RUN mkdir -p /app/data /app/storage

# Set proper permissions
RUN chmod +x /app/main

EXPOSE 8080

CMD ["./main"]