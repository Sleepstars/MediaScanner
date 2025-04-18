# Build stage
# 注意：这个 Dockerfile 将在原生架构上构建，不需要指定平台
FROM golang:1.24.1-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o mediascanner cmd/mediascanner/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/mediascanner .
COPY --from=builder /app/config.example.yaml /root/config.yaml

# Create directories for media files
RUN mkdir -p /media/input /media/output

# Expose port if needed
# EXPOSE 8080

# Set environment variables
ENV LLM_API_KEY=""
ENV TMDB_API_KEY=""
ENV TVDB_API_KEY=""
ENV BANGUMI_API_KEY=""

# Command to run
ENTRYPOINT ["./mediascanner", "-config", "config.yaml"]
