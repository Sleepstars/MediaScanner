# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24.1-alpine AS builder

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

# Build arguments for multi-arch support
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "Building on $BUILDPLATFORM for $TARGETPLATFORM"

# Set up GOOS and GOARCH based on TARGETPLATFORM
RUN if [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      export GOOS=linux GOARCH=amd64 ; \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      export GOOS=linux GOARCH=arm64 ; \
    else \
      export GOOS=linux GOARCH=amd64 ; \
    fi && \
    CGO_ENABLED=0 go build -a -installsuffix cgo -o mediascanner cmd/mediascanner/main.go

# Final stage
FROM --platform=$TARGETPLATFORM alpine:latest

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
