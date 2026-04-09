// Package observability handles audit events, budget tracking, and health checks.
package observability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Event records a single auditable action.
type Event struct {
	ID              string
	Timestamp       string
	AgentID         string
	Action          string
	Target          string
	Outcome         string
	MissionID       string
	ToolName        string
	ApprovalRequired bool
	ApprovalGranted  *bool
	PrevEventHash   string
}

var lastHash string

// Emit persists an audit event with a chained hash.
func Emit(e *Event) error {
	e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	e.PrevEventHash = lastHash
	hash := sha256.Sum256([]byte(e.Timestamp + e.AgentID + e.Action + lastHash))
	newHash := hex.EncodeToString(hash[:])
	lastHash = newHash
	e.ID = newHash[:16]

	var approvalGranted interface{}
	if e.ApprovalGranted != nil {
		if *e.ApprovalGranted {
			approvalGranted = 1
		} else {
			approvalGranted = 0
		}
	}

	_, err := world.DB.Exec(
		`INSERT INTO audit_events(id,timestamp,agent_id,action,target,outcome,mission_id,tool_name,
			approval_required,approval_granted,prev_event_hash)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.Timestamp, e.AgentID, e.Action, e.Target, e.Outcome, e.MissionID, e.ToolName,
		boolToInt(e.ApprovalRequired), approvalGranted, e.PrevEventHash,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Tail returns the last n audit events.
func Tail(n int) ([]Event, error) {
	if n <= 0 {
		n = 50
	}
	rows, err := world.DB.Query(
		`SELECT id,timestamp,agent_id,action,target,outcome,mission_id,tool_name,prev_event_hash
		 FROM audit_events ORDER BY timestamp DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil, fmt.Errorf("audit: tail: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		_ = rows.Scan(&e.ID, &e.Timestamp, &e.AgentID, &e.Action, &e.Target, &e.Outcome,
			&e.MissionID, &e.ToolName, &e.PrevEventHash)
		out = append(out, e)
	}
	return out, rows.Err()
}
