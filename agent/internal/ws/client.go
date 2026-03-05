// Package ws implements the WebSocket client used by the agent to communicate
// with the Control Plane. It handles connection, reconnection with exponential
// backoff, heartbeats, and scan command dispatching.
package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message is the generic WebSocket envelope used by both agent and control plane.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// RunScanPayload is the body of a RUN_SCAN command received from the control plane.
type RunScanPayload struct {
	ScanID     string `json:"scanId"`
	SourcePath string `json:"sourcePath"`
}

// Client manages the WebSocket connection to the control plane.
type Client struct {
	agentID   string
	wsURL     string
	mu        sync.Mutex
	conn      *websocket.Conn
	onRunScan func(scanID, sourcePath string)
}

// NewClient constructs a Client with the given agent ID, control plane WebSocket URL,
// and a callback invoked whenever a RUN_SCAN command is received.
func NewClient(agentID, wsURL string, onRunScan func(scanID, sourcePath string)) *Client {
	return &Client{
		agentID:   agentID,
		wsURL:     wsURL,
		onRunScan: onRunScan,
	}
}

// Connect starts the WS connection loop with exponential backoff reconnect.
func (c *Client) Connect() {
	attempt := 0
	for {
		log.Printf("[ws] connecting to %s (attempt %d)...", c.wsURL, attempt+1)
		conn, _, err := websocket.DefaultDialer.Dial(c.wsURL, nil)
		if err != nil {
			wait := backoff(attempt)
			log.Printf("[ws] connection failed: %v — retrying in %v", err, wait)
			time.Sleep(wait)
			attempt++
			continue
		}

		attempt = 0
		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()

		log.Printf("[ws] connected")

		// Register with control plane
		_ = c.send(Message{
			Type: "REGISTER",
			Data: mustMarshal(map[string]string{"agentId": c.agentID}),
		})

		// Start heartbeat
		stopHB := c.startHeartbeat()

		// Read loop
		c.readLoop()

		stopHB()
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()

		log.Printf("[ws] disconnected — reconnecting...")
		time.Sleep(1 * time.Second)
	}
}

// SendScanProgress sends a SCAN_PROGRESS message.
func (c *Client) SendScanProgress(scanID string, filesScanned int) {
	_ = c.send(Message{
		Type: "SCAN_PROGRESS",
		Data: mustMarshal(map[string]any{"scanId": scanID, "filesScanned": filesScanned}),
	})
}

// SendScanComplete sends a SCAN_COMPLETE message with the full tree.
func (c *Client) SendScanComplete(scanID string, stats, tree any) {
	_ = c.send(Message{
		Type: "SCAN_COMPLETE",
		Data: mustMarshal(map[string]any{
			"scanId": scanID,
			"stats":  stats,
			"tree":   tree,
		}),
	})
}

// SendScanFailed sends a SCAN_FAILED message.
func (c *Client) SendScanFailed(scanID, errMsg string) {
	_ = c.send(Message{
		Type: "SCAN_FAILED",
		Data: mustMarshal(map[string]any{"scanId": scanID, "error": errMsg}),
	})
}

// ── internal ──────────────────────────────────────────────────────────────────

func (c *Client) readLoop() {
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "REGISTERED":
			log.Printf("[ws] registered with control plane")
		case "RUN_SCAN":
			var p RunScanPayload
			if err := json.Unmarshal(msg.Data, &p); err != nil {
				log.Printf("[ws] bad RUN_SCAN payload: %v", err)
				continue
			}
			log.Printf("[ws] received RUN_SCAN: scanId=%s path=%s", p.ScanID, p.SourcePath)
			go c.onRunScan(p.ScanID, p.SourcePath)
		}
	}
}

func (c *Client) send(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteJSON(msg)
}

func (c *Client) startHeartbeat() (stop func()) {
	ticker := time.NewTicker(15 * time.Second)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				c.send(Message{ //nolint:errcheck
					Type: "HEARTBEAT",
					Data: mustMarshal(map[string]string{"agentId": c.agentID}),
				})
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
	return func() { close(done) }
}

func backoff(attempt int) time.Duration {
	wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if wait > 30*time.Second {
		wait = 30 * time.Second
	}
	return wait
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
