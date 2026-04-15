package artifacts

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/testingx"
	"github.com/stizzfer36-del/UCLAW/internal/world"
)

func setup(t *testing.T) (config.Config, string) {
	t.Helper(); testingx.TempHome(t); cfg, err := config.Load(); if err != nil { t.Fatal(err) }; if err := config.EnsureLayout(cfg); err != nil { t.Fatal(err) }; if _, _, err := world.Init(context.Background(), cfg); err != nil { t.Fatal(err) }
	repo := t.TempDir(); run := func(args ...string) { cmd := exec.Command(args[0], args[1:]...); cmd.Dir = repo; if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("%v failed: %v %s", args, err, string(out)) } }
	run("git", "init"); run("git", "config", "user.email", "test@example.com"); run("git", "config", "user.name", "Test User"); return cfg, repo }
func TestArtifactNoSourcesIsUnprovenanced(t *testing.T) { cfg, repo := setup(t); path := filepath.Join(repo, "main.go"); if err := os.WriteFile(path, []byte("package main"), 0o644); err != nil { t.Fatal(err) }; art, err := Create(context.Background(), cfg, repo, CreateRequest{Title:"code", Type:"code", Path:path, OriginAgent:"builder", MissionID:"m1"}); if err != nil { t.Fatal(err) }; if art.TrustLevel != "unprovenanced" { t.Fatalf("expected unprovenanced, got %s", art.TrustLevel) } }
func TestCitationCompletenessProvenanced(t *testing.T) { cfg, repo := setup(t); path := filepath.Join(repo, "doc.md"); if err := os.WriteFile(path, []byte("doc"), 0o644); err != nil { t.Fatal(err) }; art, err := Create(context.Background(), cfg, repo, CreateRequest{Title:"doc", Type:"doc", Path:path, OriginAgent:"builder", MissionID:"m1", ClaimCount:5, Sources:[]Source{{Type:"url", Ref:"https://example.com/1", CitedBy:"builder"},{Type:"url", Ref:"https://example.com/2", CitedBy:"builder"},{Type:"url", Ref:"https://example.com/3", CitedBy:"builder"},{Type:"url", Ref:"https://example.com/4", CitedBy:"builder"}}}); if err != nil { t.Fatal(err) }; if art.TrustLevel != "provenanced" { t.Fatalf("expected provenanced, got %s", art.TrustLevel) } }
func TestSignOffByBuilderBlocked(t *testing.T) { cfg, repo := setup(t); path := filepath.Join(repo, "doc.md"); _ = os.WriteFile(path, []byte("doc"), 0o644); art, err := Create(context.Background(), cfg, repo, CreateRequest{Title:"doc", Type:"doc", Path:path, OriginAgent:"builder", MissionID:"m1", ClaimCount:1, Sources:[]Source{{Type:"url", Ref:"https://example.com/1", CitedBy:"builder"}}}); if err != nil { t.Fatal(err) }; if _, err := RunChecks(context.Background(), cfg, art.ID, "verifier", "", repo); err != nil { t.Fatal(err) }; if err := Sign(context.Background(), cfg, art.ID, "builder"); err == nil { t.Fatal("expected builder sign-off rejection") } }
func TestArtifactRevertRestoresGitSHA(t *testing.T) { cfg, repo := setup(t); path := filepath.Join(repo, "main.go"); _ = os.WriteFile(path, []byte("v1"), 0o644); run := func(args ...string) string { cmd := exec.Command(args[0], args[1:]...); cmd.Dir = repo; out, err := cmd.CombinedOutput(); if err != nil { t.Fatalf("%v failed: %v %s", args, err, string(out)) }; return string(out) }; run("git", "add", "main.go"); run("git", "commit", "-m", "v1"); sha := strings.TrimSpace(run("git", "rev-parse", "HEAD")); art, err := Create(context.Background(), cfg, repo, CreateRequest{Title:"code", Type:"code", Path:path, OriginAgent:"builder", MissionID:"m1"}); if err != nil { t.Fatal(err) }; _ = os.WriteFile(path, []byte("v2"), 0o644); if err := Revert(context.Background(), cfg, repo, art.ID, sha); err != nil { t.Fatal(err) }; body, _ := os.ReadFile(path); if string(body) != "v1" { t.Fatalf("expected reverted content, got %s", string(body)) } }
