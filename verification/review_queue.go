// Package verification manages the artifact review queue.
package verification

import (
	"fmt"

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
	// TODO: store reason in artifact_checks
	return artifacts.SetStatus(artifactID, "failed")
}
