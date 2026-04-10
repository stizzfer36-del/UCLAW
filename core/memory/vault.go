// Package memory implements the UCLAW knowledge-graph vault backed by SQLite.
package memory

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Node is a knowledge-graph entry.
type Node struct {
	ID        string
	MissionID string
	AgentID   string
	Kind      string // fact | decision | code | doc | link
	Content   string
	Tags      string // comma-separated
	CreatedAt string
}

// InitSchema creates the memory tables if they don't exist.
// Uses plain SQLite tables with LIKE search for maximum compatibility
// (avoids fts5 which is not available in all SQLite builds).
func InitSchema() error {
	_, err := world.DB.Exec(`
		CREATE TABLE IF NOT EXISTS memory_nodes (
			id TEXT PRIMARY KEY,
			mission_id TEXT,
			agent_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			content TEXT NOT NULL,
			tags TEXT,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_memory_kind ON memory_nodes(kind);
		CREATE INDEX IF NOT EXISTS idx_memory_mission ON memory_nodes(mission_id);
	`)
	return err
}

// Write persists a new memory node.
func Write(n *Node) error {
	if n.CreatedAt == "" {
		n.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := world.DB.Exec(
		`INSERT OR REPLACE INTO memory_nodes(id,mission_id,agent_id,kind,content,tags,created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		n.ID, n.MissionID, n.AgentID, n.Kind, n.Content, n.Tags, n.CreatedAt,
	)
	return err
}

// Search performs a LIKE-based full-text search over memory content and tags.
func Search(query string, limit int) ([]Node, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := world.DB.Query(
		`SELECT id, mission_id, agent_id, kind, content, tags, created_at
		 FROM memory_nodes
		 WHERE content LIKE ? OR tags LIKE ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		pattern, pattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("memory: search: %w", err)
	}
	defer rows.Close()
	var out []Node
	for rows.Next() {
		var n Node
		_ = rows.Scan(&n.ID, &n.MissionID, &n.AgentID, &n.Kind, &n.Content, &n.Tags, &n.CreatedAt)
		out = append(out, n)
	}
	return out, rows.Err()
}
