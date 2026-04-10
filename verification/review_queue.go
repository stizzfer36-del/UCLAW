// Package verification manages the artifact review queue.
package verification

import (
	"fmt"
	"time"

	"github.com/stizzfer36-del/UCLAW/core/artifacts"
	"github.com/stizzfer36-del/UCLAW/core/world"
)

// PendingItems returns all artifacts awaiting verification.
func PendingItems() ([]artifacts.Artifact, error) {
	rows, err := world.DB.Query(
		`SELECT id,mission_id,title,type,path,origin_agent,trust_level,verification_status,git_ref,created_at
		 FROM artifacts
		 WHERE verification_status IN ('pending','in-review')
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("review_queue: %w", err)
	}
	defer rows.Close()
	var out []artifacts.Artifact
	for rows.Next() {
		var a artifacts.Artifact
		_ = rows.Scan(&a.ID, &a.MissionID, &a.Title, &a.Type, &a.Path, &a.OriginAgent,
			&a.TrustLevel, &a.VerificationStatus, &a.GitRef, &a.CreatedAt)
		out = append(out, a)
	}
	return out, rows.Err()
}

// Approve marks an artifact as verified.
func Approve(artifactID string) error {
	return artifacts.SetStatus(artifactID, "verified")
}

// Reject marks an artifact as failed.
func Reject(artifactID, reason string) error {
	if reason == "" {
		reason = "rejected without reason"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	checkID := fmt.Sprintf("chk-%d", time.Now().UTC().UnixNano())
	if _, err := world.DB.Exec(
		`INSERT INTO artifact_checks(id, artifact_id, check_type, status, run_by, run_at, details)
		 VALUES(?,?,?,?,?,?,?)`,
		checkID, artifactID, "human_review", "failed", "review_queue", now, reason,
	); err != nil {
		return fmt.Errorf("review_queue: log rejection reason: %w", err)
	}
	return artifacts.SetStatus(artifactID, "failed")
}
