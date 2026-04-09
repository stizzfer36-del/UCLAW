// Package artifacts manages artifact registration and verification lifecycle.
package artifacts

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/world"
)

// Artifact mirrors the artifacts table.
type Artifact struct {
	ID                 string
	MissionID          string
	Title              string
	Type               string
	Path               string
	OriginAgent        string
	TrustLevel         string
	VerificationStatus string
	GitRef             string
	CreatedAt          string
}

// Register inserts a new artifact record.
func Register(a *Artifact) error {
	if a.CreatedAt == "" {
		a.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if a.VerificationStatus == "" {
		a.VerificationStatus = "pending"
	}
	if a.TrustLevel == "" {
		a.TrustLevel = "unprovenanced"
	}
	_, err := world.DB.Exec(
		`INSERT INTO artifacts(id,mission_id,title,type,path,origin_agent,trust_level,verification_status,git_ref,created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.MissionID, a.Title, a.Type, a.Path, a.OriginAgent,
		a.TrustLevel, a.VerificationStatus, a.GitRef, a.CreatedAt,
	)
	return err
}

// SetStatus updates verification_status for an artifact.
func SetStatus(id, status string) error {
	res, err := world.DB.Exec(
		`UPDATE artifacts SET verification_status=? WHERE id=?`, status, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("artifacts: %s not found", id)
	}
	return nil
}

// List returns all artifacts for a mission.
func List(missionID string) ([]Artifact, error) {
	rows, err := world.DB.Query(
		`SELECT id,mission_id,title,type,path,origin_agent,trust_level,verification_status,git_ref,created_at
		 FROM artifacts WHERE mission_id=? ORDER BY created_at DESC`,
		missionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Artifact
	for rows.Next() {
		var a Artifact
		_ = rows.Scan(&a.ID, &a.MissionID, &a.Title, &a.Type, &a.Path, &a.OriginAgent,
			&a.TrustLevel, &a.VerificationStatus, &a.GitRef, &a.CreatedAt)
		out = append(out, a)
	}
	return out, rows.Err()
}
