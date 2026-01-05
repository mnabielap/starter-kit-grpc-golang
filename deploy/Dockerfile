# --- Stage 1: Builder ---
FROM golang:1.25.4-alpine AS builder

# Install git (required for fetching dependencies)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary (crucial for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# --- Stage 2: Runner ---
FROM alpine:latest

WORKDIR /usr/src/app

# Install ca-certificates (required for HTTPS outgoing calls or SMTP)
RUN apk --no-cache add ca-certificates

# Create directory for media/uploads (as requested)
RUN mkdir -p /usr/src/app/media

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy entrypoint script
COPY deploy/entrypoint.sh .
# If entrypoint.sh is in root, use: COPY entrypoint.sh .

# Make entrypoint executable
RUN chmod +x entrypoint.sh

# Change ownership to non-root user
RUN chown -R appuser:appgroup /usr/src/app

# Switch to non-root user
USER appuser

# Expose ports (50051 for gRPC, 8080 for Gateway)
EXPOSE 50051
EXPOSE 8080

# Use Entrypoint
ENTRYPOINT ["./entrypoint.sh"]

# Default Command
CMD ["./server"]