# Unified E2E Automation POC

> **Full-stack e2e automation** — Cypress browser tests interacting with a real Go agent, Go control plane, SQLite database, and React web UI, all wired together via `cy.task`.

---

## Architecture

```
Cypress (browser)
   │  cy.task           REST / polling
   │  ──────────────── Web UI (React :5173)
   │                        │
   │  cy.task spawns        │ REST + WebSocket
   ▼                        ▼
Go Agent CLI ──WS──▶ Control Plane (Go :4000)
(rclone/walk)             │ SQLite data.db
                          ▼
                    cy.task queryDb → direct DB assertions
```

---

## What's Inside

| Directory | Language | Purpose |
|---|---|---|
| `agent/` | Go | CLI agent: connects to CP via WS, runs rclone scans |
| `control-plane/` | Go | REST API + WebSocket server + SQLite state |
| `agent-ui/` | Go + Wails v2 + React | Desktop app to manage & start the agent |
| `web-ui/` | React + Vite | Browser UI for creating agents and running scans |
| `e2e/` | Cypress 13 + Node | Unified e2e automation suite |
| `bin/` | — | Compiled Go binaries (generated, not committed) |

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.22+ | `go version` |
| Node.js | 20.11.0 | Via nvm: `nvm use` |
| rclone | any | Downloaded to `./bin/` by setup script |

---

## Quick Start

### 1. Setup (one-time)

```bash
# Installs node deps + builds Go binaries + downloads rclone to ./bin/
bash scripts/setup.sh
```

### 2. Start services

```bash
npm run start:all
# Control Plane → http://localhost:4000
# Web UI        → http://localhost:5173
```

### 3. Start the agent (in a new terminal)

```bash
# Option A: Agent Manager desktop app (Wails)
cd agent-ui && go run github.com/wailsapp/wails/v2/cmd/wails@v2.9.2 dev

# Option B: Agent CLI directly
bin/go-agent start --id my-agent-001 --cp-url ws://localhost:4000/ws
```

### 4. Run E2E tests

```bash
# All tests (headless)
npm run e2e

# Only scan tests
npm run e2e:scan        # --env grep=@scan

# Only smoke tests
npm run e2e:smoke       # --env grep=@smoke

# Open Cypress UI (interactive)
npm run e2e:open
```

---

## Test Tags

| Tag | Description |
|---|---|
| `@agent` | Agent create + connect/disconnect lifecycle |
| `@scan` | Full scan flow (UI + API + DB assertions) |
| `@smoke` | Critical path: scan success + DB row count |

Run by tag: `cd e2e && npx cypress run --env grep=@smoke`

---

## The Scan Flow (POC scope)

```
1. User creates agent on Web UI → stored in SQLite agents table
2. Agent binary starts → WebSocket connection to Control Plane
3. Agent status → "online" (DB + UI polling)
4. User triggers Scan from Web UI with a source path
5. Control Plane sends RUN_SCAN to agent via WS
6. Agent runs rclone lsjson (or Go native walker) on the path
7. Agent streams SCAN_PROGRESS → Control Plane → DB update
8. Agent sends SCAN_COMPLETE with full tree → DB stores scan_tree rows
9. UI polls GET /scans/:id → shows "success" badge
10. User browses the scan tree at /scans/:id/tree
```

---

## Configuration Files

Each component has its own `config.yaml`:

| File | Port/Config |
|---|---|
| `control-plane/config.yaml` | Port 4000, SQLite path |
| `agent/config.yaml` | Control plane WS URL, rclone binary path |
| `agent-ui/` | Stored in `~/.config/unified-e2e-poc/agent-ui.yaml` |

---

## CI (Jenkins)

```bash
# The Jenkinsfile handles:
# 1. go build → ./bin/
# 2. npm install
# 3. Start control-plane + web-ui in background
# 4. cypress run → JUnit results
# 5. Kill all background processes in post{}
```

JUnit results are published from `e2e/cypress/results/*.xml`.

---

## Project Decisions

- **SQLite (no Docker)** — zero external dependencies for state. `control-plane/data/data.db` is auto-created on startup.
- **Pure Go SQLite** (`modernc.org/sqlite`) — no CGO, no C toolchain required.
- **rclone with Go fallback** — `agent/internal/scan/scanner.go` uses rclone if `./bin/rclone` exists, otherwise falls back to `filepath.WalkDir`.
- **Polling for Web UI** — AgentList polls every 3s, ScanProgress every 1.5s — simple and reliable without WS complexity in the browser.
- **cy.task as the bridge** — Node.js process inside Cypress handles agent spawn/kill, fixture generation, and direct DB assertions. This is what makes it truly "unified".
