package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains active WebSocket connections from agents.
type Hub struct {
	mu     sync.RWMutex
	agents map[string]*websocket.Conn // agentId → conn
}

func NewHub() *Hub {
	return &Hub{
		agents: make(map[string]*websocket.Conn),
	}
}

// Register adds or replaces an agent connection.
func (h *Hub) Register(agentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Close existing connection if replacing
	if existing, ok := h.agents[agentID]; ok {
		existing.Close()
	}
	h.agents[agentID] = conn
	log.Printf("[hub] agent %s registered", agentID)
}

// Unregister removes an agent connection.
func (h *Hub) Unregister(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.agents, agentID)
	log.Printf("[hub] agent %s unregistered", agentID)
}

// SendToAgent sends a JSON message to a specific agent.
func (h *Hub) SendToAgent(agentID string, msg any) error {
	h.mu.RLock()
	conn, ok := h.agents[agentID]
	h.mu.RUnlock()
	if !ok {
		return ErrAgentNotConnected
	}
	return conn.WriteJSON(msg)
}

// IsConnected returns true if the agent has an active WS connection.
func (h *Hub) IsConnected(agentID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[agentID]
	return ok
}

// ConnectedAgents returns the list of currently connected agent IDs.
func (h *Hub) ConnectedAgents() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.agents))
	for id := range h.agents {
		ids = append(ids, id)
	}
	return ids
}

// errors
var ErrAgentNotConnected = &hubError{"agent not connected"}

type hubError struct{ msg string }

func (e *hubError) Error() string { return e.msg }
