// Package checkpoints implements restart-safe mission checkpointing from SPEC-1.
// A checkpoint records the journal offset so the runtime can replay and resume.
package checkpoints

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Checkpoint is a point-in-time snapshot of a mission.
type Checkpoint struct {
	ID            string
	MissionID     string
	Reason        string
	JournalOffset int64
	SnapshotPath  string
	CreatedAt     string
}

// Create saves a checkpoint for a mission at the given journal offset.
// snapshotDir is where the snapshot file will be written.
func Create(missionID, reason string, journalOffset int64, snapshotDir string) (*Checkpoint, error) {
	if err := os.MkdirAll(snapshotDir, 0700); err != nil {
		return nil, fmt.Errorf("checkpoints: mkdir: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := fmt.Sprintf("cp-%d", time.Now().UnixNano())
	snapshotPath := filepath.Join(snapshotDir, id+".json")

	// Write a minimal snapshot marker — full state serialization added in Phase 2.
	marker := fmt.Sprintf(`{"id":%q,"mission_id":%q,"journal_offset":%d,"created_at":%q}\n`,
		id, missionID, journalOffset, now)
	if err := os.WriteFile(snapshotPath, []byte(marker), 0600); err != nil {
		return nil, fmt.Errorf("checkpoints: write snapshot: %w", err)
	}

	_, err := world.DB.Exec(
		`INSERT INTO checkpoints(id,mission_id,reason,journal_offset,snapshot_path,created_at)
		 VALUES(?,?,?,?,?,?)`,
		id, missionID, reason, journalOffset, snapshotPath, now,
	)
	if err != nil {
		return nil, fmt.Errorf("checkpoints: insert: %w", err)
	}
	return &Checkpoint{
		ID: id, MissionID: missionID, Reason: reason,
		JournalOffset: journalOffset, SnapshotPath: snapshotPath, CreatedAt: now,
	}, nil
}

// Latest returns the most recent checkpoint for a mission, or nil if none.
func Latest(missionID string) (*Checkpoint, error) {
	var c Checkpoint
	err := world.DB.QueryRow(
		`SELECT id,mission_id,reason,journal_offset,snapshot_path,created_at
		 FROM checkpoints WHERE mission_id=? ORDER BY rowid DESC LIMIT 1`,
		missionID,
	).Scan(&c.ID, &c.MissionID, &c.Reason, &c.JournalOffset, &c.SnapshotPath, &c.CreatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("checkpoints: latest: %w", err)
	}
	return &c, nil
}

// List returns all checkpoints for a mission, oldest first.
func List(missionID string) ([]Checkpoint, error) {
	rows, err := world.DB.Query(
		`SELECT id,mission_id,reason,journal_offset,snapshot_path,created_at
		 FROM checkpoints WHERE mission_id=? ORDER BY rowid ASC`,
		missionID,
	)
	if err != nil {
		return nil, fmt.Errorf("checkpoints: list: %w", err)
	}
	defer rows.Close()
	var out []Checkpoint
	for rows.Next() {
		var c Checkpoint
		_ = rows.Scan(&c.ID, &c.MissionID, &c.Reason, &c.JournalOffset, &c.SnapshotPath, &c.CreatedAt)
		out = append(out, c)
	}
	return out, rows.Err()
}
