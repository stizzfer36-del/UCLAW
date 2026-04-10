// Package world manages the UCLAW world-state graph backed by SQLite.
package world

import (
	_ "embed"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

// EnsureWorld upserts the full default hierarchy:
// world -> office -> team -> member -> machine -> room
// so that missions can be created immediately after init.
func EnsureWorld(id, name, vaultPath string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// world
	_, err := DB.Exec(
		`INSERT OR IGNORE INTO world(id,name,created_at,vault_path) VALUES(?,?,?,?)`,
		id, name, now, vaultPath,
	)
	if err != nil {
		return fmt.Errorf("world: ensure world: %w", err)
	}

	// office
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO offices(id,world_id,name,description,created_at) VALUES(?,?,?,?,?)`,
		"office-default", id, "Default Office", "Auto-created default office", now,
	)
	if err != nil {
		return fmt.Errorf("world: ensure office: %w", err)
	}

	// team
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO teams(id,office_id,name,role,created_at) VALUES(?,?,?,?,?)`,
		"team-default", "office-default", "Default Team", "dev", now,
	)
	if err != nil {
		return fmt.Errorf("world: ensure team: %w", err)
	}

	// member
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO members(id,team_id,name,type,created_at) VALUES(?,?,?,?,?)`,
		"member-local", "team-default", "local", "human", now,
	)
	if err != nil {
		return fmt.Errorf("world: ensure member: %w", err)
	}

	// machine
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO machines(id,member_id,hostname,os) VALUES(?,?,?,?)`,
		"machine-local", "member-local", "penguin", "linux",
	)
	if err != nil {
		return fmt.Errorf("world: ensure machine: %w", err)
	}

	// default room
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO rooms(id,machine_id,name,type) VALUES(?,?,?,?)`,
		"room-default", "machine-local", "Default Room", "mission",
	)
	if err != nil {
		return fmt.Errorf("world: ensure room: %w", err)
	}

	return nil
}
