// Package journal implements the UCLAW mission journal.
// Every state transition is written here BEFORE side effects,
// satisfying the no-turns-missed rule from SPEC-1.
package journal

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Entry is a single journal record.
type Entry struct {
	Offset         int64
	MissionID      string
	PacketID       string
	EventType      string
	PayloadJSON    string
	IdempotencyKey string
	CreatedAt      string
}

// Append writes a journal entry before the caller performs a side effect.
// Returns the new journal offset.
func Append(missionID, packetID, eventType, payloadJSON, idempotencyKey string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := world.DB.Exec(
		`INSERT INTO mission_journal(mission_id,packet_id,event_type,payload_json,idempotency_key,created_at)
		 VALUES(?,?,?,?,?,?)`,
		missionID, packetID, eventType, payloadJSON, idempotencyKey, now,
	)
	if err != nil {
		return 0, fmt.Errorf("journal: append: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// Since returns all journal entries after the given offset for a mission.
func Since(missionID string, afterOffset int64) ([]Entry, error) {
	rows, err := world.DB.Query(
		`SELECT offset,mission_id,packet_id,event_type,payload_json,idempotency_key,created_at
		 FROM mission_journal
		 WHERE mission_id=? AND offset>?
		 ORDER BY offset ASC`,
		missionID, afterOffset,
	)
	if err != nil {
		return nil, fmt.Errorf("journal: since: %w", err)
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		_ = rows.Scan(&e.Offset, &e.MissionID, &e.PacketID, &e.EventType,
			&e.PayloadJSON, &e.IdempotencyKey, &e.CreatedAt)
		out = append(out, e)
	}
	return out, rows.Err()
}

// AlreadyApplied returns true if the given idempotency key is already in the journal.
func AlreadyApplied(idempotencyKey string) (bool, error) {
	var count int
	err := world.DB.QueryRow(
		`SELECT COUNT(*) FROM mission_journal WHERE idempotency_key=? AND idempotency_key!=''`,
		idempotencyKey,
	).Scan(&count)
	return count > 0, err
}
