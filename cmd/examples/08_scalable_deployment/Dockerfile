# Multi-stage Dockerfile for Lokstra applications
FROM golang:1.21.6-alpine3.20 AS builder

# Update and install build dependencies
RUN apk update && apk upgrade && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Update and install ca-certificates for HTTPS requests
RUN apk update && apk upgrade && apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership
RUN chown appuser:appgroup main

# Switch to non-root user
USER appuser

# Expose port (will be overridden by docker-compose)
EXPOSE 8080

# Default config file (can be overridden)
ENV CONFIG_FILE=config.yaml

# Run the application
CMD ["./main"]