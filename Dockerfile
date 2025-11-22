# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o updatectl main.go

# Runtime stage
FROM alpine:latest

# Install docker cli and git
RUN apk add --no-cache docker-cli git

# Copy binary
COPY --from=builder /app/updatectl /usr/local/bin/updatectl

# Create config directory
RUN mkdir -p /etc/updatectl

# Set working directory
WORKDIR /etc/updatectl

# Default command
CMD ["updatectl", "watch"]