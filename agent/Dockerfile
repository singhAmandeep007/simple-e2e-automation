# ── Stage 1: Build Go Agent ───────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go-agent .

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl unzip

WORKDIR /app

# Download rclone at image build time (pinned version)
ARG RCLONE_VERSION=v1.68.2
RUN ARCH=$(uname -m | sed s/x86_64/amd64/ | sed s/aarch64/arm64/) && \
    curl -fsSL "https://downloads.rclone.org/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-${ARCH}.zip" \
    -o /tmp/rclone.zip && \
    unzip -q /tmp/rclone.zip -d /tmp && \
    mv /tmp/rclone-${RCLONE_VERSION}-linux-${ARCH}/rclone /usr/local/bin/rclone && \
    chmod +x /usr/local/bin/rclone && \
    rm -rf /tmp/rclone*

COPY --from=builder /app/go-agent .
COPY config.yaml .

# Override rclone bin_path in config to point to installed rclone
ENV RCLONE_BIN=/usr/local/bin/rclone

# Scan source data can be mounted here
VOLUME ["/scan-source"]

ENTRYPOINT ["./go-agent", "start"]
