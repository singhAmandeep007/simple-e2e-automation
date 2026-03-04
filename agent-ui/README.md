# Agent Desktop UI (Wails v2)

A native macOS/Windows/Linux desktop application wrapping the `go-agent` CLI. 
Provides a GUI for configuring the agent ID and Control Plane WebSocket URL, and allows starting/stopping the background agent process with live connection status monitoring.

## Prerequisites

- **Go 1.22+**
- **Node.js 20+**
- **Wails v2 system dependencies**
  - macOS: `xcode-select --install` (Xcode CLI tools)
  - Linux: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`
  - Windows: WebView2 runtime

> Note: The Wails CLI is **auto-installed** by the Makefile if it is missing from your system.

## Quick Start (Wails Dev Mode)

The easiest way to run the UI during development is using the Wails CLI, which provides hot-reloading for the React frontend and auto-compiles the Go backend:

```bash
cd agent-ui

# Start the desktop app in development mode
make dev
```

For a production binary (creates a macOS `.app` bundle, Windows `.exe`, or Linux binary depending on your OS):

```bash
make build
# Output will be located in agent-ui/build/bin/
```

## How It Works

1. **Frontend**: React + Vite SPA located in `frontend/`. Connects to Wails-generated JS bindings mapping directly to Go methods.
2. **Backend**: `app.go` provides methods for checking status, starting the `go-agent` binary via `os/exec`, and saving configuration to the user's `~/.config/unified-e2e-poc/agent-ui.yaml` file.
3. **Bindings**: `wails dev` or `wails build` auto-generates bridge code in `frontend/src/wailsjs/` so the React frontend can trivially call Go functions like `GetStatus()` or `StartAgent()`.
