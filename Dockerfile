# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (sqlite3)
RUN apk add --no-cache build-base

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies:
# - ca-certificates: for HTTPS/S3
# - ffmpeg: for video processing
# - sqlite: CLI tool for database inspection (optional but requested)
RUN apk --no-cache add ca-certificates ffmpeg sqlite

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/app ./app

# Create the assets directory permissions
RUN mkdir -p assets && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the application port
EXPOSE 8091

# Run the application
CMD ["./main"]
