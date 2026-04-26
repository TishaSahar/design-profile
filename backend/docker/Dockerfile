# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install swag for OpenAPI doc generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs
RUN swag init -g cmd/main/main.go -o docs

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/main/main.go

# ── Runtime stage ───────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server         ./server
COPY --from=builder /app/migrations     ./migrations

# Config and keys are mounted at runtime (not baked into the image)
# docker run -v /path/to/config:/app/config ...

EXPOSE 8080

ENTRYPOINT ["./server", "-config", "./config/config.yaml"]
