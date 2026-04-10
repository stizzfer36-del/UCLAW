// Package audit implements the UCLAW append-only audit stream from SPEC-1.
package audit

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Event is a single audit record.
type Event struct {
	Seq         int64
	TraceID     string
	MissionID   string
	ActorType   string
	ActorID     string
	EventType   string
	PayloadJSON string
	CreatedAt   string
}

// Log appends an event to the audit stream.
func Log(traceID, missionID, actorType, actorID, eventType, payloadJSON string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := world.DB.Exec(
		`INSERT INTO audit_events(trace_id,mission_id,actor_type,actor_id,event_type,payload_json,created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		traceID, missionID, actorType, actorID, eventType, payloadJSON, now,
	)
	if err != nil {
		return fmt.Errorf("audit: log: %w", err)
	}
	return nil
}

// Tail returns the last n audit events, newest first.
func Tail(n int) ([]Event, error) {
	rows, err := world.DB.Query(
		`SELECT seq,trace_id,COALESCE(mission_id,''),actor_type,actor_id,event_type,payload_json,created_at
		 FROM audit_events ORDER BY seq DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil, fmt.Errorf("audit: tail: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		_ = rows.Scan(&e.Seq, &e.TraceID, &e.MissionID, &e.ActorType,
			&e.ActorID, &e.EventType, &e.PayloadJSON, &e.CreatedAt)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ByMission returns all audit events for a mission, oldest first.
func ByMission(missionID string) ([]Event, error) {
	rows, err := world.DB.Query(
		`SELECT seq,trace_id,COALESCE(mission_id,''),actor_type,actor_id,event_type,payload_json,created_at
		 FROM audit_events WHERE mission_id=? ORDER BY seq ASC`, missionID,
	)
	if err != nil {
		return nil, fmt.Errorf("audit: by_mission: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		_ = rows.Scan(&e.Seq, &e.TraceID, &e.MissionID, &e.ActorType,
			&e.ActorID, &e.EventType, &e.PayloadJSON, &e.CreatedAt)
		out = append(out, e)
	}
	return out, rows.Err()
}
