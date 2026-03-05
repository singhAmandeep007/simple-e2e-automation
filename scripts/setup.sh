#!/usr/bin/env bash
# scripts/setup.sh — Project bootstrap. Run once before starting development.
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR/.."
BIN="$ROOT/bin"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Unified E2E POC — Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Create bin directory
mkdir -p "$BIN"

# ── Check Go ──────────────────────────────────────────────────────────────────
if ! command -v go &>/dev/null; then
  echo "❌ Go not found. Install from https://go.dev/dl/"
  exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go: $GO_VERSION"

# ── Build Go binaries into ./bin ─────────────────────────────────────────────
echo "🔨 Building agent → ./bin/go-agent ..."
(cd "$ROOT/agent" && go mod download && go build -o "$BIN/go-agent" .)
echo "✅ go-agent built"

echo "🔨 Building control-plane → ./bin/control-plane ..."
(cd "$ROOT/control-plane" && go mod download && go build -o "$BIN/control-plane" .)
echo "✅ control-plane built"

# ── Node deps ─────────────────────────────────────────────────────────────────
echo "📦 Installing npm dependencies ..."
(cd "$ROOT" && npm install)
(cd "$ROOT/web-ui" && npm install)
(cd "$ROOT/e2e" && npm install)
(cd "$ROOT/agent-ui/frontend" && npm install)
echo "✅ Node deps installed"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Setup complete!"
echo ""
echo "  Start services:  npm run start:all"
echo "  Run e2e tests:   npm run e2e"
echo "  Open Cypress:    npm run e2e:open"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
