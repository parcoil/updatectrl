# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY src/ .
RUN go build -o updatectrl main.go

# Runtime stage
FROM alpine:latest

# Install docker cli and git
RUN apk add --no-cache docker-cli git

# Copy binary
COPY --from=builder /app/updatectrl /usr/local/bin/updatectrl

# Create config directory
RUN mkdir -p /etc/updatectrl

# Set working directory
WORKDIR /etc/updatectrl

# Default command
CMD ["updatectrl", "watch"]