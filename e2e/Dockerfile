# ── Stage 1: Build go-agent from source ───────────────────────────────────────
# Build context MUST be the project root: docker build -f e2e/Dockerfile .
FROM golang:1.23-alpine AS go-builder

RUN apk add --no-cache ca-certificates

WORKDIR /agent
COPY agent/go.mod agent/go.sum ./
RUN go mod download

COPY agent/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go-agent .

# ── Stage 2: Cypress runner with go-agent bundled ────────────────────────────
FROM cypress/included:13.16.1

USER root

# Build tools for native modules (better-sqlite3 requires node-gyp)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    python3 make g++ libsqlite3-dev && \
    rm -rf /var/lib/apt/lists/*

# Copy go-agent binary
COPY --from=go-builder /go-agent /usr/local/bin/go-agent
RUN chmod +x /usr/local/bin/go-agent

# Switch back to cypress user for security
USER 1000

WORKDIR /e2e

# Install Node dependencies first (layer caching)
COPY e2e/package*.json ./
RUN npm ci

# Copy test sources
COPY e2e/ .

# Default: run all tests
CMD ["npm", "run", "cy:ci"]
