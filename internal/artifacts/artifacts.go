package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/stizzfer36-del/uclaw/internal/audit"
	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/ids"
	"github.com/stizzfer36-del/uclaw/internal/sqlitepy"
)

type Source struct { Type string `json:"type"`; Ref string `json:"ref"`; CitedBy string `json:"cited_by"` }
type Check struct { Type string `json:"type"`; Status string `json:"status"`; RunBy string `json:"run_by"`; RunAt string `json:"run_at"`; Details string `json:"details"` }
type Artifact struct {
	ID string `json:"id"`; Title string `json:"title"`; Type string `json:"type"`; Path string `json:"path"`; OriginAgent string `json:"origin_agent"`; MissionID string `json:"mission_id"`; CreatedAt string `json:"created_at"`; VerificationStatus string `json:"verification_status"`; GitRef string `json:"git_ref"`; TrustLevel string `json:"trust_level"`; ClaimCount int `json:"claim_count"`; Sources []Source `json:"sources"`; Checks []Check `json:"checks"`; SignOffChain []string `json:"sign_off_chain"` }
type CreateRequest struct { Title string; Type string; Path string; OriginAgent string; MissionID string; ClaimCount int; Sources []Source }
func Create(ctx context.Context, cfg config.Config, repoRoot string, req CreateRequest) (Artifact, error) {
	if req.ClaimCount <= 0 { req.ClaimCount = 1 }
	now := time.Now().UTC().Format(time.RFC3339)
	artifact := Artifact{ID:ids.New("art"),Title:req.Title,Type:req.Type,Path:req.Path,OriginAgent:req.OriginAgent,MissionID:req.MissionID,CreatedAt:now,VerificationStatus:"pending",GitRef:currentGitRef(repoRoot),ClaimCount:req.ClaimCount,Sources:req.Sources}
	artifact.TrustLevel = TrustLevel(artifact.ClaimCount, artifact.Sources)
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO artifacts(id, title, type, path, origin_agent, mission_id, created_at, verification_status, git_ref, trust_level, claim_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, artifact.ID, artifact.Title, artifact.Type, artifact.Path, artifact.OriginAgent, artifact.MissionID, artifact.CreatedAt, artifact.VerificationStatus, artifact.GitRef, artifact.TrustLevel, artifact.ClaimCount); err != nil { return Artifact{}, err }
	for _, source := range artifact.Sources { if err := AddSource(ctx, cfg, artifact.ID, source); err != nil { return Artifact{}, err } }
	_ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:req.OriginAgent,Action:"artifact_create",Target:req.Path,MissionID:req.MissionID,Outcome:"success",ApprovalRequired:false})
	return Show(ctx, cfg, artifact.ID)
}
func AutoCreateOnWrite(ctx context.Context, cfg config.Config, repoRoot, agentID, missionID, path string) error { _, err := Create(ctx, cfg, repoRoot, CreateRequest{Title:filepath.Base(path),Type:"code",Path:path,OriginAgent:agentID,MissionID:defaultMissionID(missionID),ClaimCount:1}); return err }
func AddSource(ctx context.Context, cfg config.Config, artifactID string, source Source) error { return sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO artifact_sources(id, artifact_id, source_type, source_ref, cited_by) VALUES (?, ?, ?, ?, ?)`, ids.New("src"), artifactID, source.Type, source.Ref, source.CitedBy) }
func List(ctx context.Context, cfg config.Config, missionID string) ([]Artifact, error) {
	sql := `SELECT id FROM artifacts`; args := []interface{}{}
	if missionID != "" { sql += ` WHERE mission_id = ?`; args = append(args, missionID) }
	sql += ` ORDER BY created_at ASC`
	rows, err := sqlitepy.Query(ctx, cfg.DBPath, sql, args...); if err != nil { return nil, err }
	out := make([]Artifact, 0, len(rows)); for _, row := range rows { art, err := Show(ctx, cfg, fmt.Sprintf("%v", row["id"])); if err != nil { return nil, err }; out = append(out, art) }; return out, nil }
func Show(ctx context.Context, cfg config.Config, artifactID string) (Artifact, error) {
	rows, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT * FROM artifacts WHERE id = ?`, artifactID); if err != nil { return Artifact{}, err }; if len(rows) == 0 { return Artifact{}, fmt.Errorf("artifact %s not found", artifactID) }
	row := rows[0]; artifact := Artifact{ID:str(row["id"]),Title:str(row["title"]),Type:str(row["type"]),Path:str(row["path"]),OriginAgent:str(row["origin_agent"]),MissionID:str(row["mission_id"]),CreatedAt:str(row["created_at"]),VerificationStatus:str(row["verification_status"]),GitRef:str(row["git_ref"]),TrustLevel:str(row["trust_level"]),ClaimCount:intVal(row["claim_count"])}
	srcRows, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT source_type, source_ref, cited_by FROM artifact_sources WHERE artifact_id = ? ORDER BY id`, artifactID); for _, srcRow := range srcRows { artifact.Sources = append(artifact.Sources, Source{Type:str(srcRow["source_type"]), Ref:str(srcRow["source_ref"]), CitedBy:str(srcRow["cited_by"])}) }
	checkRows, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT check_type, status, run_by, run_at, details FROM artifact_checks WHERE artifact_id = ? ORDER BY id`, artifactID); for _, checkRow := range checkRows { artifact.Checks = append(artifact.Checks, Check{Type:str(checkRow["check_type"]), Status:str(checkRow["status"]), RunBy:str(checkRow["run_by"]), RunAt:str(checkRow["run_at"]), Details:str(checkRow["details"])}) }
	signRows, _ := sqlitepy.Query(ctx, cfg.DBPath, `SELECT signed_by FROM artifact_signoffs WHERE artifact_id = ? ORDER BY signed_at`, artifactID); for _, signRow := range signRows { artifact.SignOffChain = append(artifact.SignOffChain, str(signRow["signed_by"])) }
	return artifact, nil
}
func RunChecks(ctx context.Context, cfg config.Config, artifactID, verifierID string, testCommand string, workspace string) ([]Check, error) {
	artifact, err := Show(ctx, cfg, artifactID); if err != nil { return nil, err }
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE artifacts SET verification_status = ? WHERE id = ?`, "in-review", artifactID); err != nil { return nil, err }
	checks := []Check{runTestCheck(testCommand, verifierID), runCitationCheck(artifact, verifierID), runPolicyCheck(artifact, verifierID, workspace)}
	allPassed := true
	for _, check := range checks { if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO artifact_checks(id, artifact_id, check_type, status, run_by, run_at, details) VALUES (?, ?, ?, ?, ?, ?, ?)`, ids.New("chk"), artifactID, check.Type, check.Status, check.RunBy, check.RunAt, check.Details); err != nil { return nil, err }; if check.Status != "passed" { allPassed = false } }
	status := "verified"; if !allPassed { status = "failed"; if err := queueReview(ctx, cfg, artifactID, "verification check failed"); err != nil { return nil, err } }
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE artifacts SET verification_status = ?, trust_level = ? WHERE id = ?`, status, TrustLevel(artifact.ClaimCount, artifact.Sources), artifactID); err != nil { return nil, err }
	_ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:verifierID,Action:"artifact_verify",Target:artifactID,MissionID:artifact.MissionID,Outcome:status,ApprovalRequired:false})
	return checks, nil
}
func Reverify(ctx context.Context, cfg config.Config, artifactID string) error { if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `DELETE FROM artifact_checks WHERE artifact_id = ?`, artifactID); err != nil { return err }; if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE artifacts SET verification_status = ? WHERE id = ?`, "pending", artifactID); err != nil { return err }; return audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"system",Action:"artifact_reverify_required",Target:artifactID,Outcome:"success",ApprovalRequired:false}) }
func Sign(ctx context.Context, cfg config.Config, artifactID, signerID string) error { artifact, err := Show(ctx, cfg, artifactID); if err != nil { return err }; if artifact.OriginAgent == signerID { return fmt.Errorf("builder and verifier must differ") }; if artifact.VerificationStatus != "verified" { return fmt.Errorf("artifact %s is not verified", artifactID) }; if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO artifact_signoffs(id, artifact_id, signed_by, signed_at) VALUES (?, ?, ?, ?)`, ids.New("sig"), artifactID, signerID, time.Now().UTC().Format(time.RFC3339)); err != nil { return err }; return audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:signerID,Action:"artifact_signoff",Target:artifactID,MissionID:artifact.MissionID,Outcome:"success",ApprovalRequired:false}) }
func Flag(ctx context.Context, cfg config.Config, artifactID, reason string) error { if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE artifacts SET verification_status = ? WHERE id = ?`, "disputed", artifactID); err != nil { return err }; if err := queueReview(ctx, cfg, artifactID, reason); err != nil { return err }; return audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"system",Action:"artifact_flag",Target:artifactID,Outcome:"success",ApprovalRequired:false}) }
func Revert(ctx context.Context, cfg config.Config, repoRoot, artifactID, sha string) error {
	artifact, err := Show(ctx, cfg, artifactID); if err != nil { return err }
	relPath := artifact.Path; if filepath.IsAbs(relPath) { if rel, err := filepath.Rel(repoRoot, relPath); err == nil { relPath = rel } }
	cmd := exec.CommandContext(ctx, "git", "show", sha+":"+relPath); cmd.Dir = repoRoot; body, err := cmd.Output(); if err != nil { return err }
	if err := os.WriteFile(artifact.Path, body, 0o644); err != nil { return err }
	return audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"system",Action:"artifact_revert",Target:artifact.Path,MissionID:artifact.MissionID,Outcome:"success",ApprovalRequired:false})
}
func TrustLevel(claimCount int, sources []Source) string { if claimCount <= 0 || len(sources) == 0 { return "unprovenanced" }; pct := len(sources) * 100 / claimCount; switch { case pct >= 80: return "provenanced"; case pct >= 40: return "partially-provenanced"; default: return "unprovenanced" } }
func currentGitRef(repoRoot string) string { cmd := exec.Command("git", "rev-parse", "HEAD"); cmd.Dir = repoRoot; body, err := cmd.Output(); if err != nil { cmd = exec.Command("git", "branch", "--show-current"); cmd.Dir = repoRoot; body, err = cmd.Output(); if err != nil { return "" } }; return strings.TrimSpace(string(body)) }
func runTestCheck(command, verifierID string) Check { now := time.Now().UTC().Format(time.RFC3339); if strings.TrimSpace(command) == "" { return Check{Type:"test",Status:"failed",RunBy:verifierID,RunAt:now,Details:"no test command provided"} }; cmd := exec.Command("bash", "-lc", command); cmd.Env = append(os.Environ(), "GOCACHE=/tmp/uclaw-gocache"); body, err := cmd.CombinedOutput(); if err != nil { return Check{Type:"test",Status:"failed",RunBy:verifierID,RunAt:now,Details:strings.TrimSpace(string(body))} }; return Check{Type:"test",Status:"passed",RunBy:verifierID,RunAt:now,Details:strings.TrimSpace(string(body))} }
func runCitationCheck(artifact Artifact, verifierID string) Check { now := time.Now().UTC().Format(time.RFC3339); level := TrustLevel(artifact.ClaimCount, artifact.Sources); status := "failed"; if level == "provenanced" { status = "passed" }; details, _ := json.Marshal(map[string]interface{}{"trust_level": level, "sources": len(artifact.Sources), "claim_count": artifact.ClaimCount}); return Check{Type:"citation_review",Status:status,RunBy:verifierID,RunAt:now,Details:string(details)} }
func runPolicyCheck(artifact Artifact, verifierID, workspace string) Check { now := time.Now().UTC().Format(time.RFC3339); path := filepath.Clean(artifact.Path); root := filepath.Clean(workspace); status := "passed"; details := "path within workspace"; if root != "" && !(path == root || strings.HasPrefix(path, root+string(os.PathSeparator))) { status = "failed"; details = "artifact path outside workspace" }; return Check{Type:"policy_compliance",Status:status,RunBy:verifierID,RunAt:now,Details:details} }
func queueReview(ctx context.Context, cfg config.Config, artifactID, reason string) error { return sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO review_queue(id, artifact_id, reason, status, created_at) VALUES (?, ?, ?, ?, ?)`, ids.New("rq"), artifactID, reason, "open", time.Now().UTC().Format(time.RFC3339)) }
func defaultMissionID(missionID string) string { if missionID == "" { return "mission-unscoped" }; return missionID }
func ResolveTestCommand(cfg config.Config, artifactID string) (string, error) { rows, err := sqlitepy.Query(context.Background(), cfg.DBPath, `SELECT mission_id FROM artifacts WHERE id = ?`, artifactID); if err != nil { return "", err }; if len(rows) == 0 { return "", fmt.Errorf("artifact %s not found", artifactID) }; path := filepath.Join(cfg.MissionsPath, str(rows[0]["mission_id"]), "test_matrix.yaml"); body, err := os.ReadFile(path); if err != nil { return "", err }; for _, line := range strings.Split(string(body), "\n") { trim := strings.TrimSpace(line); if strings.HasPrefix(trim, "command:") { return strings.TrimSpace(strings.TrimPrefix(trim, "command:")), nil } }; return "", fmt.Errorf("no test command found for artifact %s", artifactID) }
func str(v interface{}) string { return fmt.Sprintf("%v", v) }
func intVal(v interface{}) int { var n int; fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &n); return n }
