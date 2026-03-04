.PHONY: dev build clean install-wails

# Wails binary — resolve from GOBIN or GOPATH/bin (handles asdf, custom GOPATH, etc.)
GOBIN_DIR := $(shell go env GOBIN)
GOPATH_DIR := $(shell go env GOPATH)
WAILS := $(shell if [ -n "$(GOBIN_DIR)" ] && [ -x "$(GOBIN_DIR)/wails" ]; then echo "$(GOBIN_DIR)/wails"; elif [ -x "$(GOPATH_DIR)/bin/wails" ]; then echo "$(GOPATH_DIR)/bin/wails"; else echo "wails"; fi)

# Install Wails CLI if not present
install-wails:
	@command -v $(WAILS) >/dev/null 2>&1 || go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "✅ Wails CLI: $(WAILS)"
	@$(WAILS) version

# Development with live reload
dev: install-wails
	$(WAILS) dev

# Production build
build: install-wails
	$(WAILS) build

# Clean build artifacts
clean:
	rm -rf build/
