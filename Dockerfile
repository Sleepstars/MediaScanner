FROM golang:1.24.1-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o mediascanner cmd/mediascanner/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/mediascanner .
COPY --from=builder /app/config.example.yaml /root/config.yaml

RUN mkdir -p /media/input /media/output

# Expose port if needed
# EXPOSE 8080

ENV LLM_API_KEY=""
ENV TMDB_API_KEY=""
ENV TVDB_API_KEY=""
ENV BANGUMI_API_KEY=""

ENTRYPOINT ["./mediascanner", "-config", "config.yaml"]
