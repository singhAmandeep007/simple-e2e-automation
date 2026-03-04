#!/usr/bin/env bash
# scripts/start-all.sh — Start control plane + web UI for development
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR/.."

if [ ! -f "$ROOT/bin/control-plane" ]; then
  echo "❌ bin/control-plane not found. Run: bash scripts/setup.sh"
  exit 1
fi

echo "Starting Control Plane on :4000 and Web UI on :5173..."
echo "Press Ctrl+C to stop both."
echo ""

# Start CP from its own dir so config.yaml and data/ are found correctly
(cd "$ROOT/control-plane" && "$ROOT/bin/control-plane") &
CP_PID=$!
echo "Control Plane started (pid=$CP_PID)"

# Start web UI
source ~/.nvm/nvm.sh 2>/dev/null && nvm use 2>/dev/null
npm run dev --prefix "$ROOT/web-ui" &
UI_PID=$!
echo "Web UI started (pid=$UI_PID)"

trap 'kill $CP_PID $UI_PID 2>/dev/null; echo Stopped.' EXIT INT TERM
wait
