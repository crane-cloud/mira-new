# syntax=docker/dockerfile:1

# Stage 1: Build the Go application
FROM golang:1.21 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary with subcommands (image-builder, api-server)
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/mira ./main.go

# Stage 2: Create the final image
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
  ca-certificates \
  git \
  curl \
  docker.io \
  && rm -rf /var/lib/apt/lists/*

# Create a non-root user (optional)
RUN useradd -m appuser

# Set working directory
WORKDIR /home/appuser

# Copy the built binary
COPY --from=builder /bin/mira /usr/local/bin/mira

# Make binary executable
RUN chmod +x /usr/local/bin/mira

# Run as non-root user
USER appuser

# Default command (can be overridden at runtime)
ENTRYPOINT ["mira"]
