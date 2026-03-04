package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"control-plane/internal/db"
	"control-plane/internal/ws"
)

// Agent is the API response shape for an agent.
type Agent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Connected bool   `json:"connected"`
}

type AgentHandler struct {
	db  *db.DB
	hub *ws.Hub
}

func NewAgentHandler(database *db.DB, hub *ws.Hub) *AgentHandler {
	return &AgentHandler{db: database, hub: hub}
}

// POST /agents — create a new agent
func (h *AgentHandler) Create(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		ID   string `json:"id"` // optional — auto-generated if empty
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := body.ID
	if agentID == "" {
		agentID = uuid.New().String()
	}

	_, err := h.db.Exec(`INSERT INTO agents (id, name) VALUES (?, ?)`, agentID, body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, Agent{
		ID:        agentID,
		Name:      body.Name,
		Status:    "offline",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Connected: false,
	})
}

// GET /agents — list all agents
func (h *AgentHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name, status, created_at FROM agents ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	agents := []Agent{}
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.Status, &a.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		a.Connected = h.hub.IsConnected(a.ID)
		// Keep status in sync with live connection
		if a.Connected && a.Status != "online" {
			a.Status = "online"
		}
		agents = append(agents, a)
	}
	c.JSON(http.StatusOK, agents)
}

// GET /agents/:id — get a single agent
func (h *AgentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var a Agent
	err := h.db.QueryRow(`SELECT id, name, status, created_at FROM agents WHERE id = ?`, id).
		Scan(&a.ID, &a.Name, &a.Status, &a.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a.Connected = h.hub.IsConnected(a.ID)
	c.JSON(http.StatusOK, a)
}
