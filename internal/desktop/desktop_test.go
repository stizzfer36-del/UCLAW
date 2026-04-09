package desktop_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/uclaw/internal/agents"
	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/missions"
	"github.com/stizzfer36-del/uclaw/internal/observability"
	"github.com/stizzfer36-del/uclaw/internal/testingx"
	"github.com/stizzfer36-del/uclaw/internal/world"
)

func setup(t *testing.T) config.Config {
	t.Helper(); testingx.TempHome(t); cfg, err := config.Load(); if err != nil { t.Fatal(err) }; if err := config.EnsureLayout(cfg); err != nil { t.Fatal(err) }; if _, _, err := world.Init(context.Background(), cfg); err != nil { t.Fatal(err) }; return cfg }
func TestTUIRendersAtWidths(t *testing.T) { state := observability.State{}; if len(observability.RenderTUI(state, 80)) == 0 || len(observability.RenderTUI(state, 220)) == 0 { t.Fatal("expected non-empty tui render") } }
func TestDesktopMatchesCLIState(t *testing.T) { cfg := setup(t); state, err := observability.Status(context.Background(), cfg); if err != nil { t.Fatal(err) }; path := filepath.Join(t.TempDir(), "desktop.html"); out, err := observability.RenderHTML(state, path); if err != nil { t.Fatal(err) }; body, err := os.ReadFile(out); if err != nil { t.Fatal(err) }; if !strings.Contains(string(body), "\"missions\":") { t.Fatal("expected serialized state in desktop html") } }
func TestApprovalFlowVisibleInDesktop(t *testing.T) { cfg := setup(t); wd, _ := os.Getwd(); repoRoot := filepath.Clean(filepath.Join(wd, "..", "..")); agent, err := agents.Spawn(context.Background(), cfg, repoRoot, agents.SpawnRequest{Name:"risk", Role:"dev", Capabilities:[]string{"delete_file"}, AllowedPaths:[]string{cfg.Home}}); if err != nil { t.Fatal(err) }; _, err = agents.CallTool(context.Background(), cfg, repoRoot, agent.ID, "delete_file", filepath.Join(cfg.Home, "x.txt"), "", nil); if err != nil { t.Fatal(err) }; state, err := observability.Status(context.Background(), cfg); if err != nil { t.Fatal(err) }; if len(state.Approvals) == 0 { t.Fatal("expected pending approval in desktop state") } }
func TestBudgetAccuracy(t *testing.T) { cfg := setup(t); if err := missions.RecordUsage(context.Background(), cfg, "m1", "a1", "mock", 100, 1.25); err != nil { t.Fatal(err) }; budget, err := observability.Budget(context.Background(), cfg); if err != nil { t.Fatal(err) }; if budget["tokens"].(int) != 100 { t.Fatalf("expected exact budget tokens, got %+v", budget) } }
func TestElectronScaffoldBuildsAssets(t *testing.T) { wd, _ := os.Getwd(); repoRoot := filepath.Clean(filepath.Join(wd, "..", "..")); cmd := exec.Command("node", "desktop/build.js"); cmd.Dir = repoRoot; if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("desktop build failed: %v\n%s", err, string(out)) }; for _, path := range []string{filepath.Join(repoRoot, "desktop", "dist", "index.html"), filepath.Join(repoRoot, "desktop", "dist", "state.json"), filepath.Join(repoRoot, "desktop", "electron", "main.js"), filepath.Join(repoRoot, "desktop", "electron", "preload.js")} { if _, err := os.Stat(path); err != nil { t.Fatalf("expected generated desktop asset %s: %v", path, err) } } }
