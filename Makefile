# =============================================================================
#  Unified E2E POC — Makefile
#  Usage: make <target>
#  Run `make help` for a full list of commands.
# =============================================================================

.DEFAULT_GOAL := help

# ── Config ────────────────────────────────────────────────────────────────────
CP_PORT  ?= 4000
UI_PORT  ?= 5173
AGENT_ID ?= docker-agent-001

# Resolve Wails binary for agent-ui targets (handles asdf GOBIN or standard GOPATH/bin)
GOBIN_DIR := $(shell go env GOBIN)
GOPATH_DIR := $(shell go env GOPATH)
WAILS := $(shell if [ -n "$(GOBIN_DIR)" ] && [ -x "$(GOBIN_DIR)/wails" ]; then echo "$(GOBIN_DIR)/wails"; elif [ -x "$(GOPATH_DIR)/bin/wails" ]; then echo "$(GOPATH_DIR)/bin/wails"; else echo "wails"; fi)


# Detect if running in CI
CI ?= false

# Colour helpers
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

.PHONY: help setup build build-go build-docker \
        up down agent logs \
        e2e e2e-open e2e-scan e2e-smoke e2e-docker \
        ci ci-docker clean nuke

# ── Help ──────────────────────────────────────────────────────────────────────
help: ## Show this help message
	@printf "\n$(CYAN)Unified E2E POC — Available Commands$(RESET)\n\n"
	@printf "$(YELLOW)LOCAL (no Docker)$(RESET)\n"
	@printf "  make setup           — one-time setup (downloads rclone, builds Go, installs npm)\n"
	@printf "  make build-go        — build Go binaries → ./bin/\n"
	@printf "  make up              — start control-plane + web-ui locally\n"
	@printf "  make agent           — start agent locally (set AGENT_ID=xyz)\n"
	@printf "  make e2e             — run all Cypress tests locally (services must be running)\n"
	@printf "  make e2e-open        — open Cypress interactive UI\n"
	@printf "  make e2e-scan        — run @scan tagged tests only\n"
	@printf "  make e2e-smoke       — run @smoke tagged tests only\n"
	@printf "\n$(YELLOW)AGENT DESKTOP UI (Wails)$(RESET)\n"
	@printf "  make agent-ui-install       — install frontend deps\n"
	@printf "  make agent-ui-dev           — Vite-only dev (no wails CLI needed)\n"
	@printf "  make agent-ui-wails-dev     — full Wails hot-reload desktop app\n"
	@printf "  make agent-ui-build         — build React frontend → agent-ui/frontend/dist/\n"
	@printf "  make agent-ui-wails-build   — build native .app bundle (requires wails CLI)\n"
	@printf "$(YELLOW)DOCKER$(RESET)\n"
	@printf "  make build       — build all Docker images\n"
	@printf "  make up-docker   — start control-plane + web-ui in Docker\n"
	@printf "  make e2e-docker  — run e2e tests in Docker (full pipeline)\n"
	@printf "  make ci-docker   — full CI: build → start → e2e in Docker\n"
	@printf "  make down        — stop and remove all Docker containers\n"
	@printf "  make logs        — tail logs from all Docker containers\n"
	@printf "\n$(YELLOW)UTILITIES$(RESET)\n"
	@printf "  make clean       — remove binaries + Docker artefacts\n"
	@printf "  make nuke        — remove everything including volumes (⚠ data loss)\n"
	@printf "\n"

# ── Local: Setup ──────────────────────────────────────────────────────────────
setup: ## One-time project setup
	@printf "$(CYAN)Running setup...$(RESET)\n"
	bash scripts/setup.sh

# ── Local: Build Go ───────────────────────────────────────────────────────────
build-go: ## Build Go binaries to ./bin/
	@printf "$(CYAN)Building Go binaries...$(RESET)\n"
	@mkdir -p bin
	cd agent && go build -o ../bin/go-agent .
	cd control-plane && go build -o ../bin/control-plane .
	@printf "$(GREEN)✅ Go binaries ready in ./bin/$(RESET)\n"

# ── Local: Start services ─────────────────────────────────────────────────────
up: build-go ## Build Go + start control-plane & web-ui locally
	@printf "$(CYAN)Cleaning environment...$(RESET)\n"
	@bash scripts/clean.sh
	@printf "$(CYAN)Starting services locally...$(RESET)\n"
	npm run start:all

agent: ## Start agent locally (AGENT_ID=xxx make agent)
	@[ -f bin/go-agent ] || (echo "Run 'make build-go' first" && exit 1)
	bin/go-agent start --id $(AGENT_ID) --cp-url ws://localhost:$(CP_PORT)/ws

# ── Local: E2E ────────────────────────────────────────────────────────────────
e2e: ## Run all Cypress tests (services must already be running)
	npm run e2e

e2e-open: ## Open Cypress interactive UI
	npm run e2e:open

e2e-scan: ## Run @scan tagged tests only
	npm run e2e:scan

e2e-smoke: ## Run @smoke tagged tests only
	npm run e2e:smoke

# ── Agent Desktop UI (Wails) ──────────────────────────────────────────────────
agent-ui-install: ## Install agent-ui frontend npm dependencies
	cd agent-ui/frontend && npm install

agent-ui-dev: agent-ui-install ## Start Vite dev server for agent-ui frontend (no wails CLI required)
	@printf "$(CYAN)Starting Vite dev server for agent-ui...$(RESET)\n"
	@printf "$(YELLOW)Note: Go bindings are mocked. StartAgent/StopAgent will error — run agent-ui-wails-dev for full functionality.$(RESET)\n"
	cd agent-ui/frontend && npm run dev

agent-ui-build: agent-ui-install ## Build agent-ui React frontend → frontend/dist/
	@printf "$(CYAN)Building agent-ui frontend...$(RESET)\n"
	cd agent-ui/frontend && npm run build
	@printf "$(GREEN)✅ Frontend built → agent-ui/frontend/dist/$(RESET)\n"

agent-ui-go-build: agent-ui-build ## Compile agent-ui Go (after building frontend)
	@printf "$(CYAN)Compiling agent-ui Go binary...$(RESET)\n"
	cd agent-ui && go build -o ../bin/agent-manager .
	@printf "$(GREEN)✅ agent-manager built → ./bin/agent-manager$(RESET)\n"

agent-ui-wails-dev: ## Run agent-ui in Wails hot-reload dev mode
	@bash scripts/clean.sh
	@command -v $(WAILS) >/dev/null 2>&1 || go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd agent-ui && $(WAILS) dev

agent-ui-wails-build: ## Build native .app bundle with Wails (macOS)
	@command -v $(WAILS) >/dev/null 2>&1 || go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd agent-ui && $(WAILS) build
	@printf "$(GREEN)✅ Native app built → agent-ui/build/bin/$(RESET)\n"


# ── Docker: Build ─────────────────────────────────────────────────────────────
build: ## Build all Docker images
	@printf "$(CYAN)Building Docker images...$(RESET)\n"
	docker compose build
	docker compose --profile e2e build
	@printf "$(GREEN)✅ All Docker images built$(RESET)\n"

build-docker: build ## Alias for build

# ── Docker: Start ─────────────────────────────────────────────────────────────
up-docker: ## Start control-plane + web-ui in Docker (detached)
	@printf "$(CYAN)Starting services in Docker...$(RESET)\n"
	docker compose up --build -d
	@printf "$(GREEN)✅ Services running:$(RESET)\n"
	@printf "   Control Plane → http://localhost:$(CP_PORT)\n"
	@printf "   Web UI        → http://localhost:$(UI_PORT)\n"
	@docker compose ps

down: ## Stop all Docker services
	docker compose --profile e2e --profile agent down

logs: ## Tail Docker logs
	docker compose logs -f

# ── Docker: E2E ───────────────────────────────────────────────────────────────
e2e-docker: ## Run Cypress e2e tests in Docker (starts services first)
	@printf "$(CYAN)Running e2e tests in Docker...$(RESET)\n"
	docker compose --profile e2e up \
	  --build \
	  --abort-on-container-exit \
	  --exit-code-from e2e
	@printf "$(GREEN)✅ E2E tests complete$(RESET)\n"

e2e-docker-scan: ## Run @scan tests in Docker
	CYPRESS_grep="@scan" $(MAKE) e2e-docker

e2e-docker-smoke: ## Run @smoke tests in Docker
	CYPRESS_grep="@smoke" $(MAKE) e2e-docker

# ── Full CI Pipeline ──────────────────────────────────────────────────────────
ci: build-go ## Local CI: build Go + npm install + run e2e
	@printf "$(CYAN)Running local CI pipeline...$(RESET)\n"
	@bash scripts/clean.sh
	npm run install:all
	(cd control-plane && ../bin/control-plane &) && \
	  npm run dev --prefix web-ui & \
	  sleep 5 && \
	  npm run e2e; \
	  EXIT=$$?; \
	  pkill go-agent 2>/dev/null; pkill -f "vite" 2>/dev/null; \
	  exit $$EXIT

ci-docker: ## Docker CI: full pipeline — build images → start services → run e2e → clean up
	@printf "$(CYAN)Running full Docker CI pipeline...$(RESET)\n"
	$(MAKE) e2e-docker
	@printf "$(GREEN)✅ CI pipeline complete$(RESET)\n"
	$(MAKE) down

# ── Utilities ─────────────────────────────────────────────────────────────────
clean: ## Remove Go binaries and Docker artefacts (keeps volumes)
	rm -rf bin/go-agent bin/control-plane web-ui/dist e2e/cypress/results
	docker compose down --remove-orphans 2>/dev/null || true
	docker image prune -f 2>/dev/null || true
	@printf "$(GREEN)✅ Clean complete$(RESET)\n"

nuke: ## ⚠ Remove EVERYTHING including Docker volumes (data will be lost!)
	docker compose --profile e2e --profile agent down -v --remove-orphans 2>/dev/null || true
	docker volume prune -f 2>/dev/null || true
	rm -rf bin/ control-plane/data/ web-ui/dist e2e/cypress/results e2e/cypress/screenshots e2e/cypress/videos
	@printf "$(GREEN)✅ Full nuke complete$(RESET)\n"
