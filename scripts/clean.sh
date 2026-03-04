#!/bin/bash
set -e

echo "🧹 Cleaning up old processes..."
pkill -f "go-agent" || true
pkill -f "control-plane" || true
pkill -f "wails" || true
pkill -f "vite" || true

echo "🗑️  Deleting SQLite database..."
rm -rf control-plane/data/data.db

echo "✅ Environment cleaned."
