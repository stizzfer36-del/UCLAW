package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/UCLAW/internal/audit"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/testingx"
	"github.com/stizzfer36-del/UCLAW/internal/world"
)

func setup(t *testing.T) (config.Config, string) {
	t.Helper()
	testingx.TempHome(t)
	cfg, err := config.Load()
	if err != nil { t.Fatal(err) }
	if err := config.EnsureLayout(cfg); err != nil { t.Fatal(err) }
	if _, _, err := world.Init(context.Background(), cfg); err != nil { t.Fatal(err) }
	wd, err := os.Getwd(); if err != nil { t.Fatal(err) }
	return cfg, filepath.Clean(filepath.Join(wd, "..", ".."))
}
func TestSpawnRejectsInvalidCapability(t *testing.T) { cfg, repoRoot := setup(t); _, err := Spawn(context.Background(), cfg, repoRoot, SpawnRequest{Name:"bad-agent",Capabilities:[]string{"not_a_real_tool"}}); if err == nil || !strings.Contains(err.Error(), "not registered") { t.Fatalf("expected invalid capability rejection, got %v", err) } }
func TestHighRiskToolRequiresApproval(t *testing.T) { cfg, repoRoot := setup(t); agent, err := Spawn(context.Background(), cfg, repoRoot, SpawnRequest{Name:"risk-agent",Capabilities:[]string{"delete_file"},AllowedPaths:[]string{cfg.Home}}); if err != nil { t.Fatal(err) }; result, err := CallTool(context.Background(), cfg, repoRoot, agent.ID, "delete_file", filepath.Join(cfg.Home, "x.txt"), "", nil); if err != nil { t.Fatal(err) }; if result.Status != "approval_required" || result.RequestID == "" { t.Fatalf("expected approval request, got %+v", result) } }
func TestPathOutsideWhitelistRejectedAndAudited(t *testing.T) { cfg, repoRoot := setup(t); agent, err := Spawn(context.Background(), cfg, repoRoot, SpawnRequest{Name:"scoped-agent",Capabilities:[]string{"read_file"},AllowedPaths:[]string{cfg.Home}}); if err != nil { t.Fatal(err) }; outsideDir := t.TempDir(); outsidePath := filepath.Join(outsideDir, "outside.txt"); if err := os.WriteFile(outsidePath, []byte("x"), 0o644); err != nil { t.Fatal(err) }; _, err = CallTool(context.Background(), cfg, repoRoot, agent.ID, "read_file", outsidePath, "", nil); if err == nil || !strings.Contains(err.Error(), "outside the path whitelist") { t.Fatalf("expected whitelist rejection, got %v", err) }; body, err := os.ReadFile(cfg.AuditPath); if err != nil { t.Fatal(err) }; if !strings.Contains(string(body), "tool_rejected") { t.Fatalf("expected rejection in audit log, got %s", string(body)) } }
func TestAuditIntegrity(t *testing.T) { cfg, repoRoot := setup(t); agent, err := Spawn(context.Background(), cfg, repoRoot, SpawnRequest{Name:"audited-agent",Capabilities:[]string{"read_file"},AllowedPaths:[]string{cfg.Home}}); if err != nil { t.Fatal(err) }; filePath := filepath.Join(cfg.Home, "allowed.txt"); if err := os.WriteFile(filePath, []byte("ok"), 0o644); err != nil { t.Fatal(err) }; if _, err := CallTool(context.Background(), cfg, repoRoot, agent.ID, "read_file", filePath, "", nil); err != nil { t.Fatal(err) }; if err := audit.Verify(cfg.AuditPath); err != nil { t.Fatal(err) } }
