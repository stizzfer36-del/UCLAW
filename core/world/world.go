// Package world manages the UCLAW world-state graph backed by SQLite.
package world

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB is the shared world-state handle.
var DB *sql.DB

// Open initialises (or connects to) the world-state database.
func Open(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return fmt.Errorf("world: mkdir: %w", err)
	}
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("world: open: %w", err)
	}
	DB = db
	return applySchema()
}

func applySchema() error {
	_, err := DB.Exec(schemaSQL)
	return err
}

// EnsureWorld upserts the root world row.
func EnsureWorld(id, name, vaultPath string) error {
	_, err := DB.Exec(
		`INSERT OR IGNORE INTO world(id,name,created_at,vault_path) VALUES(?,?,?,?)`,
		id, name, time.Now().UTC().Format(time.RFC3339), vaultPath,
	)
	return err
}
