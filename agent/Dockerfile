# ── Stage 1: Build Go Agent ───────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go-agent .

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
# Minimal Alpine image — the Go agent uses the native filepath.WalkDir
# implementation, so no external scanning binary is needed.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/go-agent .
COPY config.yaml .

# Scan source data can be mounted here
VOLUME ["/scan-source"]

ENTRYPOINT ["./go-agent", "start"]
