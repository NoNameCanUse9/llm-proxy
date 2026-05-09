# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled (since we use pure go sqlite driver)
RUN CGO_ENABLED=0 GOOS=linux go build -o llm-proxy ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/llm-proxy .
COPY --from=builder /app/configs/config.example.toml ./config.example.toml

# Create data directory
RUN mkdir -p data

EXPOSE 8080

CMD ["./llm-proxy"]
