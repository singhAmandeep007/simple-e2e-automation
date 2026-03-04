package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"
)

// AgentConfig holds persisted configuration for the desktop app.
type AgentConfig struct {
	AgentID        string `yaml:"agent_id" json:"agentId"`
	ControlPlaneWS string `yaml:"control_plane_ws" json:"controlPlaneWs"`
	AgentBinPath   string `yaml:"agent_bin_path" json:"agentBinPath"`
}

// App is the Wails bound struct — all public methods are callable from the frontend.
type App struct {
	ctx     context.Context
	mu      sync.Mutex
	cmd     *exec.Cmd
	status  string
	cfg     AgentConfig
	cfgPath string
}

func newApp() *App {
	return &App{status: "offline"}
}

// startup is called when the app launches.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfgPath, err := configPath()
	if err == nil {
		a.cfgPath = cfgPath
		_ = a.loadConfig()
	}
}

// ── Config ────────────────────────────────────────────────────────────────────

// GetConfig returns the current saved configuration.
func (a *App) GetConfig() AgentConfig {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

// SaveConfig persists the configuration to disk.
func (a *App) SaveConfig(cfg AgentConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg = cfg
	return a.saveConfig()
}

func (a *App) loadConfig() error {
	data, err := os.ReadFile(a.cfgPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &a.cfg)
}

func (a *App) saveConfig() error {
	if a.cfgPath == "" {
		return fmt.Errorf("no config path")
	}
	_ = os.MkdirAll(filepath.Dir(a.cfgPath), 0o755)
	data, err := yaml.Marshal(a.cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(a.cfgPath, data, 0o644)
}

// ── Agent lifecycle ────────────────────────────────────────────────────────────

// GetStatus returns the current agent connection status.
func (a *App) GetStatus() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

// StartAgent spawns the go-agent binary and monitors it.
func (a *App) StartAgent(agentID, cpURL string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd != nil && a.cmd.Process != nil {
		return fmt.Errorf("agent already running")
	}

	binPath := a.cfg.AgentBinPath
	if binPath == "" {
		// 1. Try sibling of the running Wails app binary
		exe, _ := os.Executable()
		// 2. Fallback: look in ../bin/ relative to the repo (for `wails dev` mode)
		repoRoot := filepath.Join(filepath.Dir(exe), "..")
		candidates := []string{
			filepath.Join(filepath.Dir(exe), "go-agent"),
			filepath.Join(repoRoot, "bin", "go-agent"),
			"../bin/go-agent",
		}
		for _, c := range candidates {
			if abs, err := filepath.Abs(c); err == nil {
				if _, err := os.Stat(abs); err == nil {
					binPath = abs
					break
				}
			}
		}
	}

	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("agent binary not found at %s — run scripts/setup.sh first", binPath)
	}

	a.cmd = exec.Command(binPath, "start", "--id", agentID, "--cp-url", cpURL)
	stdout, err := a.cmd.StdoutPipe()
	if err != nil {
		a.cmd = nil
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	stderr, err := a.cmd.StderrPipe()
	if err != nil {
		a.cmd = nil
		return fmt.Errorf("creating stderr pipe: %w", err)
	}

	if err := a.cmd.Start(); err != nil {
		a.cmd = nil
		return fmt.Errorf("starting agent: %w", err)
	}

	a.status = "connecting"
	runtime.EventsEmit(a.ctx, "agent:status", "connecting")
	log.Printf("[app] agent process started (pid=%d)", a.cmd.Process.Pid)

	// Stream stdout and stderr concurrently
	go a.scanLog(stdout)
	go a.scanLog(stderr)

	go func() {
		a.cmd.Wait() //nolint:errcheck
		a.mu.Lock()
		a.cmd = nil
		a.status = "offline"
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "agent:status", "offline")
		log.Printf("[app] agent process exited")
	}()

	return nil
}

// StopAgent kills the running agent process.
func (a *App) StopAgent() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd == nil || a.cmd.Process == nil {
		return nil
	}
	if err := a.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("killing agent: %w", err)
	}
	a.cmd = nil
	a.status = "offline"
	runtime.EventsEmit(a.ctx, "agent:status", "offline")
	return nil
}

// SetStatus allows the frontend to update status (e.g. after WS REGISTERED event).
func (a *App) SetStatus(status string) {
	a.mu.Lock()
	a.status = status
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, "agent:status", status)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "unified-e2e-poc", "agent-ui.yaml"), nil
}

func (a *App) scanLog(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line) // echo to app terminal

		if line != "" {
			a.mu.Lock()
			if strings.Contains(line, "[ws] connected") {
				a.status = "online"
				runtime.EventsEmit(a.ctx, "agent:status", "online")
			} else if strings.Contains(line, "[ws] disconnected") {
				a.status = "connecting"
				runtime.EventsEmit(a.ctx, "agent:status", "connecting")
			}
			a.mu.Unlock()
		}
	}
}
