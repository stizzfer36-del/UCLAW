package planner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/missions"
	"github.com/stizzfer36-del/UCLAW/internal/testingx"
	"github.com/stizzfer36-del/UCLAW/internal/world"
)

func setup(t *testing.T) (config.Config, string, missions.Mission) {
	t.Helper(); testingx.TempHome(t); cfg, _ := config.Load(); _ = config.EnsureLayout(cfg); _, _, _ = world.Init(context.Background(), cfg); wd, _ := os.Getwd(); repoRoot := filepath.Clean(filepath.Join(wd, "..", "..")); m, _ := missions.Start(context.Background(), cfg, "plan", "user", "lead"); return cfg, repoRoot, m }
func TestPlannerOrchestratesTeams(t *testing.T) { cfg, repoRoot, m := setup(t); result, err := Orchestrate(context.Background(), cfg, repoRoot, m.ID, "lead"); if err != nil { t.Fatal(err) }; if len(result.DevTeam) < 2 || len(result.VerifyTeam) < 2 { t.Fatalf("expected dev and verify teams, got %+v", result) } }
func TestPlannerWritesDefaultMatrix(t *testing.T) { cfg, _, m := setup(t); if err := DefaultMatrix(cfg, m.ID); err != nil { t.Fatal(err) }; entries, err := missions.LoadMatrix(cfg, m.ID); if err != nil { t.Fatal(err) }; if len(entries) < 2 { t.Fatal("expected default matrix entries") } }
