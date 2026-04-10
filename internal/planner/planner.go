package planner

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/stizzfer36-del/UCLAW/internal/agents"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/ids"
	"github.com/stizzfer36-del/UCLAW/internal/missions"
	"github.com/stizzfer36-del/UCLAW/internal/sqlitepy"
)

type PlanResult struct { Lead agents.Profile `json:"lead"`; DevTeam []agents.Profile `json:"dev_team"`; VerifyTeam []agents.Profile `json:"verify_team"`; MissionID string `json:"mission_id"` }
func Orchestrate(ctx context.Context, cfg config.Config, repoRoot, missionID, leadName string) (PlanResult, error) {
	lead, err := agents.Spawn(ctx, cfg, repoRoot, agents.SpawnRequest{Name:leadName,Role:"lead",Provider:"mock",Capabilities:[]string{"agent_spawn", "read_file"},AllowedPaths:[]string{cfg.Home}}); if err != nil { return PlanResult{}, err }
	dev := []agents.Profile{}; for _, name := range []string{"dev-a", "dev-b"} { profile, err := agents.Spawn(ctx, cfg, repoRoot, agents.SpawnRequest{Name:name,Role:"dev",Provider:"mock",Capabilities:[]string{"write_file", "read_file", "memory_write"},AllowedPaths:[]string{cfg.Home}}); if err != nil { return PlanResult{}, err }; dev = append(dev, profile) }
	verify := []agents.Profile{}; for _, name := range []string{"verify-a", "verify-b"} { profile, err := agents.Spawn(ctx, cfg, repoRoot, agents.SpawnRequest{Name:name,Role:"verify",Provider:"mock",Capabilities:[]string{"artifact_sign", "read_file"},AllowedPaths:[]string{cfg.Home}}); if err != nil { return PlanResult{}, err }; verify = append(verify, profile) }
	for _, profile := range append(dev, verify...) { if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO timeline_events(id, mission_id, event_type, message, created_at) VALUES (?, ?, ?, ?, ?)`, ids.New("tl"), missionID, "planner", fmt.Sprintf("%s assigned to %s", profile.Role, profile.ID), time.Now().UTC().Format(time.RFC3339)); err != nil { return PlanResult{}, err } }
	return PlanResult{Lead:lead,DevTeam:dev,VerifyTeam:verify,MissionID:missionID}, nil }
func DefaultMatrix(cfg config.Config, missionID string) error { _, err := missions.SaveMatrix(cfg, missionID, []missions.MatrixEntry{{ID:"unit-all",Type:"unit",Target:".",Command:"GOCACHE=/tmp/uclaw-gocache go test ./...",Required:true},{ID:"citation-check",Type:"citation_review",Target:"docs/",Required:true,MinCompleteness:80}}); return err }
func Workspace(cfg config.Config, missionID string) string { return filepath.Join(cfg.MissionsPath, missionID) }
