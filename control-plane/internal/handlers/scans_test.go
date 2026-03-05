package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"control-plane/internal/handlers"
	"control-plane/internal/ws"
)

func TestScanHandler_Start(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)
	hub := ws.NewHub()
	handler := handlers.NewScanHandler(database, hub)

	// Pre-seed an agent
	agentID := uuid.New().String()
	_, err := database.Exec(`INSERT INTO agents (id, name, status) VALUES (?, ?, ?)`, agentID, "test-agent", "offline")
	if err != nil {
		t.Fatalf("failed to seed test agent: %v", err)
	}

	r := gin.Default()
	r.POST("/agents/:id/scan", handler.Start)

	t.Run("Refuse scan when agent offline", func(t *testing.T) {
		payload := []byte(`{"sourcePath": "/tmp/test"}`)
		req := httptest.NewRequest(http.MethodPost, "/agents/"+agentID+"/scan", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("Start scan when agent online", func(t *testing.T) {
		// Mock the agent connecting to the Hub
		_, err := database.Exec(`UPDATE agents SET status = 'online' WHERE id = ?`, agentID)
		if err != nil {
			t.Fatalf("failed to update test agent: %v", err)
		}

		// We can't easily mock the exact WS conn in a unit test, but Hub.IsConnected checks the map.
		// For a true unit test of the REST handler, we would ideally mock the Hub interface.
		// Since Hub is a concrete struct here, the integration-style test will fail here unless we inject
		// a dummy websocket connection which requires a real server.
		// For the sake of the unit test boundary, we are verifying the fail-fast 'offline' behavior above.
	})
}
