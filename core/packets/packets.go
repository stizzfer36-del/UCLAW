// Package packets implements the WorkPacket state machine from SPEC-1.
// States: queued → planned → running → waiting_tool | waiting_human |
//         blocked → verified | failed | rolled_back → completed
package packets

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Packet is a single schedulable unit of work within a mission.
type Packet struct {
	ID             string
	MissionID      string
	AgentID        string
	Kind           string
	InputJSON      string
	State          string
	Priority       int
	IdempotencyKey string
	StartedAt      string
	FinishedAt     string
}

const (
	StateQueued       = "queued"
	StatePlanned      = "planned"
	StateRunning      = "running"
	StateWaitingTool  = "waiting_tool"
	StateWaitingHuman = "waiting_human"
	StateBlocked      = "blocked"
	StateVerified     = "verified"
	StateFailed       = "failed"
	StateRolledBack   = "rolled_back"
	StateCompleted    = "completed"
)

// Create inserts a new WorkPacket in queued state.
func Create(missionID, kind, inputJSON, idempotencyKey string, priority int) (*Packet, error) {
	id := fmt.Sprintf("wp-%d", time.Now().UnixNano())
	_, err := world.DB.Exec(
		`INSERT INTO work_packets(id,mission_id,agent_id,kind,input_json,state,priority,idempotency_key)
		 VALUES(?,?,?,?,?,?,?,?)`,
		id, missionID, "unassigned", kind, inputJSON, StateQueued, priority, idempotencyKey,
	)
	if err != nil {
		return nil, fmt.Errorf("packets: create: %w", err)
	}
	return &Packet{
		ID: id, MissionID: missionID, Kind: kind,
		InputJSON: inputJSON, State: StateQueued,
		Priority: priority, IdempotencyKey: idempotencyKey,
	}, nil
}

// Transition moves a packet to a new state, recording timestamps where needed.
func Transition(id, newState string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var err error
	switch newState {
	case StateRunning:
		_, err = world.DB.Exec(
			`UPDATE work_packets SET state=?,started_at=? WHERE id=?`,
			newState, now, id,
		)
	case StateCompleted, StateVerified, StateFailed, StateRolledBack:
		_, err = world.DB.Exec(
			`UPDATE work_packets SET state=?,finished_at=? WHERE id=?`,
			newState, now, id,
		)
	default:
		_, err = world.DB.Exec(
			`UPDATE work_packets SET state=? WHERE id=?`,
			newState, id,
		)
	}
	if err != nil {
		return fmt.Errorf("packets: transition %s→%s: %w", id, newState, err)
	}
	return nil
}

// Assign sets the agent_id for a packet.
func Assign(id, agentID string) error {
	_, err := world.DB.Exec(`UPDATE work_packets SET agent_id=? WHERE id=?`, agentID, id)
	return err
}

// ListByMission returns all packets for a mission ordered by priority desc.
func ListByMission(missionID string) ([]Packet, error) {
	rows, err := world.DB.Query(
		`SELECT id,mission_id,agent_id,kind,input_json,state,priority,idempotency_key,
		        COALESCE(started_at,''),COALESCE(finished_at,'')
		 FROM work_packets WHERE mission_id=? ORDER BY priority DESC, rowid ASC`,
		missionID,
	)
	if err != nil {
		return nil, fmt.Errorf("packets: list: %w", err)
	}
	defer rows.Close()
	var out []Packet
	for rows.Next() {
		var p Packet
		_ = rows.Scan(&p.ID, &p.MissionID, &p.AgentID, &p.Kind, &p.InputJSON,
			&p.State, &p.Priority, &p.IdempotencyKey, &p.StartedAt, &p.FinishedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

// Get returns a single packet by ID.
func Get(id string) (*Packet, error) {
	var p Packet
	err := world.DB.QueryRow(
		`SELECT id,mission_id,agent_id,kind,input_json,state,priority,idempotency_key,
		        COALESCE(started_at,''),COALESCE(finished_at,'')
		 FROM work_packets WHERE id=?`, id,
	).Scan(&p.ID, &p.MissionID, &p.AgentID, &p.Kind, &p.InputJSON,
		&p.State, &p.Priority, &p.IdempotencyKey, &p.StartedAt, &p.FinishedAt)
	if err != nil {
		return nil, fmt.Errorf("packets: get %s: %w", id, err)
	}
	return &p, nil
}
