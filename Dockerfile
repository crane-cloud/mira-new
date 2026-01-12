# syntax=docker/dockerfile:1

# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Install Air
RUN go install github.com/air-verse/air@latest

# Copy the source code
COPY . .

# Build the Go binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /bin/mira ./main.go

# For development: keep Go available for live reload
FROM golang:1.24-alpine AS development

# Add metadata labels
LABEL maintainer="mira-team" \
      version="1.0" \
      description="MIRA - Container Image Builder (Development)"

# Install runtime dependencies 
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl \
    docker \
    docker-cli \
    && rm -rf /var/cache/apk/*

# Add user to docker group for socket access
RUN addgroup -g 999 docker || true
RUN adduser -D -u 1000 appuser || true  
RUN adduser appuser docker || true

# Set working directory
WORKDIR /app

# Copy Air binary from builder
COPY --from=builder /go/bin/air /usr/local/bin/air


# Copy scripts from builder stage
# COPY --from=builder /app/scripts/start-imagebuilder.sh ./scripts/start-imagebuilder.sh

# Make scripts executable
# RUN chmod +x ./scripts/start-imagebuilder.sh

RUN mkdir -p /app/uploads /app/logs /app/builds /app/tmp /usr/local/crane

EXPOSE 3000

# Switch to non-root user (commented out for Docker socket access)
# USER appuser

CMD ["air"]
