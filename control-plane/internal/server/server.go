// Package server configures the Gin HTTP handler and core WebSocket upgrader
// that constitute the Control Plane API layer.
package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"control-plane/internal/config"
	"control-plane/internal/db"
	"control-plane/internal/handlers"
	ws "control-plane/internal/ws"
)

// Server aggregates the shared dependencies (Config, Database, WebSocket Hub)
// and handles routing of all API endpoints and WebSocket upgrade requests.
type Server struct {
	cfg      *config.Config
	db       *db.DB
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func New(cfg *config.Config, database *db.DB) *Server {
	return &Server{
		cfg: cfg,
		db:  database,
		hub: ws.NewHub(),
		upgrader: websocket.Upgrader{
			// Allow all origins for local POC development
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) Run() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware — reflects the request Origin if it is in the allowlist.
	// Browsers require exactly one matching origin in the response header.
	allowedOrigins := make(map[string]bool, len(s.cfg.CORS.AllowOrigins))
	for _, o := range s.cfg.CORS.AllowOrigins {
		allowedOrigins[o] = true
	}
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(s.cfg.CORS.AllowOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// WebSocket endpoint (agents connect here)
	r.GET("/ws", s.handleWS)

	// REST API
	agentH := handlers.NewAgentHandler(s.db, s.hub)
	scanH := handlers.NewScanHandler(s.db, s.hub)

	api := r.Group("/")
	{
		api.POST("/agents", agentH.Create)
		api.GET("/agents", agentH.List)
		api.GET("/agents/:id", agentH.Get)

		api.POST("/agents/:id/scan", scanH.Start)
		api.GET("/scans/:id", scanH.Get)
		api.GET("/scans/:id/tree", scanH.Tree)
	}

	return r.Run(fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port))
}

func (s *Server) handleWS(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// Block the HTTP handler so the TCP connection is kept alive
	ws.HandleConnection(conn, s.hub, s.db)
}
