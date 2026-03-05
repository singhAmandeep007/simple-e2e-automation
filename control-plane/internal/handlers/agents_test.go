package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"control-plane/internal/db"
	"control-plane/internal/handlers"
	"control-plane/internal/ws"
)

func setupTestDB(t *testing.T) *db.DB {
	database, err := db.Init(":memory:")
	if err != nil {
		t.Fatalf("failed to init memory db: %v", err)
	}
	return database
}

func TestAgentHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)
	hub := ws.NewHub()
	handler := handlers.NewAgentHandler(database, hub)

	r := gin.Default()
	r.POST("/agents", handler.Create)

	t.Run("Create new agent with random ID", func(t *testing.T) {
		payload := []byte(`{"name": "test-agent"}`)
		req := httptest.NewRequest(http.MethodPost, "/agents", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d - body: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["name"] != "test-agent" {
			t.Errorf("expected name 'test-agent', got %v", resp["name"])
		}
		if resp["status"] != "offline" {
			t.Errorf("expected initial status 'offline', got %v", resp["status"])
		}
	})
}
