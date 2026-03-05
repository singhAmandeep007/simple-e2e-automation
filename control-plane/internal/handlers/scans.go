package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"control-plane/internal/db"
	"control-plane/internal/ws"
)

// ScanHandler handles REST operations on the /scans and /agents/:id/scan resources.
type ScanHandler struct {
	db  *db.DB
	hub *ws.Hub
}

// NewScanHandler constructs a ScanHandler with the given database and WS hub.
func NewScanHandler(database *db.DB, hub *ws.Hub) *ScanHandler {
	return &ScanHandler{db: database, hub: hub}
}

// Scan is the REST response shape for a scan record.
type Scan struct {
	ID           string `json:"id"`
	AgentID      string `json:"agentId"`
	SourcePath   string `json:"sourcePath"`
	Status       string `json:"status"`
	TotalFiles   int    `json:"totalFiles"`
	TotalFolders int    `json:"totalFolders"`
	Error        string `json:"error,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// TreeNode is a single file/directory entry returned by GET /scans/:id/tree.
type TreeNode struct {
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// POST /agents/:id/scan — start a scan
func (h *ScanHandler) Start(c *gin.Context) {
	agentID := c.Param("id")

	var body struct {
		SourcePath string `json:"sourcePath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure agent exists
	var exists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM agents WHERE id = ?`, agentID).Scan(&exists); err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	// Check agent is connected
	if !h.hub.IsConnected(agentID) {
		c.JSON(http.StatusConflict, gin.H{"error": "agent is not connected"})
		return
	}

	scanID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := h.db.Exec(`INSERT INTO scans (id, agent_id, source_path, status, created_at, updated_at) VALUES (?, ?, ?, 'pending', ?, ?)`,
		scanID, agentID, body.SourcePath, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send RUN_SCAN to agent via WebSocket
	payload, _ := json.Marshal(map[string]string{
		"scanId":     scanID,
		"sourcePath": body.SourcePath,
	})
	if err := h.hub.SendToAgent(agentID, ws.Message{
		Type: "RUN_SCAN",
		Data: payload,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send scan command to agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"scanId":     scanID,
		"agentId":    agentID,
		"sourcePath": body.SourcePath,
		"status":     "pending",
	})
}

// GET /scans/:id — get scan status + stats
func (h *ScanHandler) Get(c *gin.Context) {
	scanID := c.Param("id")
	var s Scan
	var errVal sql.NullString
	err := h.db.QueryRow(`
		SELECT id, agent_id, source_path, status, total_files, total_folders, error, created_at, updated_at
		FROM scans WHERE id = ?`, scanID).
		Scan(&s.ID, &s.AgentID, &s.SourcePath, &s.Status, &s.TotalFiles, &s.TotalFolders, &errVal, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if errVal.Valid {
		s.Error = errVal.String
	}
	c.JSON(http.StatusOK, s)
}

// GET /scans/:id/tree — get the full folder tree for a completed scan
func (h *ScanHandler) Tree(c *gin.Context) {
	scanID := c.Param("id")

	// Verify scan exists and is complete
	var status string
	if err := h.db.QueryRow(`SELECT status FROM scans WHERE id = ?`, scanID).Scan(&status); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	rows, err := h.db.Query(`
		SELECT path, is_dir, size, mod_time FROM scan_tree WHERE scan_id = ? ORDER BY path ASC`, scanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	tree := []TreeNode{}
	for rows.Next() {
		var n TreeNode
		var isDir int
		if err := rows.Scan(&n.Path, &isDir, &n.Size, &n.ModTime); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		n.IsDir = isDir == 1
		tree = append(tree, n)
	}

	c.Header("Content-Type", "application/json")
	enc := json.NewEncoder(c.Writer)
	_ = enc.Encode(gin.H{"scanId": scanID, "status": status, "tree": tree})
}
