// Package main acts as the entry point for the Control Plane binary.
// It initializes the configuration, connects to the SQLite database,
// constructs the HTTP/WebSocket server, and begins listening for connections.
package main

import (
	"log"

	"control-plane/internal/config"
	"control-plane/internal/db"
	"control-plane/internal/server"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Init(cfg.DB.Path)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer database.Close()

	srv := server.New(cfg, database)
	log.Printf("Control Plane listening on :%d", cfg.Server.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
