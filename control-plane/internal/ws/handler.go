package ws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"control-plane/internal/db"
)

// Message is a generic WebSocket message envelope.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Incoming message payloads from agent
type RegisterPayload struct {
	AgentID string `json:"agentId"`
}

type ScanProgressPayload struct {
	ScanID       string `json:"scanId"`
	FilesScanned int    `json:"filesScanned"`
}

type ScanCompletePayload struct {
	ScanID string      `json:"scanId"`
	Stats  ScanStats   `json:"stats"`
	Tree   []TreeEntry `json:"tree"`
}

type ScanFailedPayload struct {
	ScanID string `json:"scanId"`
	Error  string `json:"error"`
}

type ScanStats struct {
	TotalFiles   int `json:"totalFiles"`
	TotalFolders int `json:"totalFolders"`
}

type TreeEntry struct {
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// HandleConnection reads messages from a new WS connection.
// It handles registration first; subsequent messages are routed by type.
func HandleConnection(conn *websocket.Conn, hub *Hub, database *db.DB) {
	var agentID string
	defer func() {
		conn.Close()
		if agentID != "" {
			hub.Unregister(agentID)
			_ = setAgentStatus(database, agentID, "offline")
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error for agent %s: %v", agentID, err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad message from %s: %v", agentID, err)
			continue
		}

		switch msg.Type {
		case "REGISTER":
			var p RegisterPayload
			if err := json.Unmarshal(msg.Data, &p); err != nil {
				log.Printf("[ws] bad REGISTER payload: %v", err)
				continue
			}
			agentID = p.AgentID
			hub.Register(agentID, conn)
			if err := setAgentStatus(database, agentID, "online"); err != nil {
				log.Printf("[ws] failed to set agent online: %v", err)
			}
			resp, _ := json.Marshal(Message{Type: "REGISTERED", Data: mustMarshal(map[string]string{"agentId": agentID})})
			conn.WriteMessage(websocket.TextMessage, resp)

		case "HEARTBEAT":
			// Keep connection alive — no-op for now

		case "SCAN_PROGRESS":
			var p ScanProgressPayload
			if err := json.Unmarshal(msg.Data, &p); err != nil {
				log.Printf("[ws] bad SCAN_PROGRESS payload: %v", err)
				continue
			}
			_ = updateScanProgress(database, p.ScanID, p.FilesScanned)

		case "SCAN_COMPLETE":
			var p ScanCompletePayload
			if err := json.Unmarshal(msg.Data, &p); err != nil {
				log.Printf("[ws] bad SCAN_COMPLETE payload: %v", err)
				continue
			}
			if err := completeScan(database, p.ScanID, p.Stats, p.Tree); err != nil {
				log.Printf("[ws] failed to complete scan %s: %v", p.ScanID, err)
			}

		case "SCAN_FAILED":
			var p ScanFailedPayload
			if err := json.Unmarshal(msg.Data, &p); err != nil {
				log.Printf("[ws] bad SCAN_FAILED payload: %v", err)
				continue
			}
			_ = failScan(database, p.ScanID, p.Error)

		default:
			log.Printf("[ws] unknown message type %q from agent %s", msg.Type, agentID)
		}
	}
}

// ── DB helpers ────────────────────────────────────────────────────────────────

func setAgentStatus(database *db.DB, agentID, status string) error {
	_, err := database.Exec(`UPDATE agents SET status = ? WHERE id = ?`, status, agentID)
	return err
}

func updateScanProgress(database *db.DB, scanID string, filesScanned int) error {
	_, err := database.Exec(`
		UPDATE scans SET status = 'running', total_files = ?, updated_at = ? WHERE id = ?`,
		filesScanned, now(), scanID)
	return err
}

func completeScan(database *db.DB, scanID string, stats ScanStats, tree []TreeEntry) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`
		UPDATE scans SET status = 'success', total_files = ?, total_folders = ?, updated_at = ? WHERE id = ?`,
		stats.TotalFiles, stats.TotalFolders, now(), scanID)
	if err != nil {
		return fmt.Errorf("updating scan: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO scan_tree (scan_id, path, is_dir, size, mod_time) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing tree insert: %w", err)
	}
	defer stmt.Close()

	for _, entry := range tree {
		isDir := 0
		if entry.IsDir {
			isDir = 1
		}
		if _, err := stmt.Exec(scanID, entry.Path, isDir, entry.Size, entry.ModTime); err != nil {
			return fmt.Errorf("inserting tree entry %s: %w", entry.Path, err)
		}
	}

	return tx.Commit()
}

func failScan(database *db.DB, scanID, errMsg string) error {
	_, err := database.Exec(`
		UPDATE scans SET status = 'failed', error = ?, updated_at = ? WHERE id = ?`,
		errMsg, now(), scanID)
	return err
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// Ensure sql.ErrNoRows is accessible via import
var _ = sql.ErrNoRows
