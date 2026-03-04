package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// DB wraps sql.DB with convenience methods.
type DB struct {
	*sql.DB
}

// Init opens (or creates) the SQLite database and applies the schema.
func Init(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	// Enable WAL mode for concurrent readers
	if _, err := sqldb.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enabling WAL: %w", err)
	}
	if _, err := sqldb.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	// Apply schema
	if _, err := sqldb.Exec(schema); err != nil {
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return &DB{sqldb}, nil
}
