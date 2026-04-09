package hardening

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/uclaw/internal/audit"
	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/missions"
	"github.com/stizzfer36-del/uclaw/internal/testingx"
	"github.com/stizzfer36-del/uclaw/internal/world"
)

func setup(t *testing.T) (config.Config, string) {
	t.Helper()
	testingx.TempHome(t)
	cfg, _ := config.Load()
	_ = config.EnsureLayout(cfg)
	_, _, _ = world.Init(context.Background(), cfg)
	wd, _ := os.Getwd()
	return cfg, filepath.Clean(filepath.Join(wd, "..", ".."))
}

func TestPolicyFuzzBlock(t *testing.T) {
	cfg, repo := setup(t)
	if err := FuzzBlock(context.Background(), cfg, repo); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryPath(t *testing.T) {
	cfg, _ := setup(t)
	m, _ := missions.Start(context.Background(), cfg, "recover", "user", "dev")
	_, _ = missions.SaveCheckpoint(context.Background(), cfg, m.ID, "baseline")
	if err := Recover(context.Background(), cfg, m.ID); err != nil {
		t.Fatal(err)
	}
}

func TestSyncExportImport(t *testing.T) {
	cfg, _ := setup(t)
	if err := os.WriteFile(filepath.Join(cfg.VaultPath, "notes", "shared.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := SyncExport(context.Background(), cfg, "peer")
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var pkg peerPackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		t.Fatal(err)
	}
	found := false
	for i := range pkg.Files {
		if pkg.Files[i].Path == "vault/notes/shared.md" {
			pkg.Files[i].BaseHash = hashString("base\n")
			pkg.Files[i].BaseB64 = base64.StdEncoding.EncodeToString([]byte("base\n"))
			pkg.Files[i].ContentB64 = base64.StdEncoding.EncodeToString([]byte("base\nremote\n"))
			pkg.Files[i].Hash = hashString("base\nremote\n")
			found = true
		}
	}
	if !found {
		t.Fatal("expected shared note in peer package")
	}
	if err := os.WriteFile(path, mustJSON(t, pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.VaultPath, "notes", "shared.md"), []byte("base\nlocal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SyncImport(context.Background(), cfg, path); err != nil {
		t.Fatal(err)
	}
	merged, err := os.ReadFile(filepath.Join(cfg.VaultPath, "notes", "shared.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merged), "local") || !strings.Contains(string(merged), "remote") {
		t.Fatalf("expected merged content, got %q", string(merged))
	}
}

func TestSecretScan(t *testing.T) {
	cfg, _ := setup(t)
	if err := SecretScan(context.Background(), cfg, cfg.Home); err != nil {
		t.Fatal(err)
	}
}

func TestAuditCoverage(t *testing.T) {
	cfg, _ := setup(t)
	_ = audit.Write(context.Background(), cfg.AuditPath, audit.Event{AgentID: "agent", Action: "tool_call", Outcome: "success"})
	body, err := os.ReadFile(cfg.AuditPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "tool_call") {
		t.Fatal("expected audit coverage")
	}
}

func TestSyncConflictRecording(t *testing.T) {
	cfg, _ := setup(t)
	path := filepath.Join(t.TempDir(), "peer.json")
	pkg := peerPackage{
		Peer: "peer",
		Files: []peerFile{
			{
				Path:        "vault/blobs/conflict.bin",
				BaseHash:    hashString("base\n"),
				BaseB64:     base64.StdEncoding.EncodeToString([]byte("base\n")),
				Hash:        hashString("remote\n"),
				ContentB64:  base64.StdEncoding.EncodeToString([]byte("remote\n")),
				ContentType: "binary",
			},
		},
	}
	if err := os.MkdirAll(filepath.Join(cfg.VaultPath, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.VaultPath, "blobs", "conflict.bin"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, mustJSON(t, pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SyncImport(context.Background(), cfg, path); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(cfg.Home, "sync-conflicts", "peer"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected conflict artifacts")
	}
}

func TestSyncTransportHandlerExport(t *testing.T) {
	cfgRemote, _ := setup(t)
	if err := os.WriteFile(filepath.Join(cfgRemote.VaultPath, "notes", "remote.md"), []byte("start mission remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/sync/export?peer=peer", nil)
	rec := httptest.NewRecorder()
	peerTransportHandler(cfgRemote).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	testingx.TempHome(t)
	cfgLocal, _ := config.Load()
	_ = config.EnsureLayout(cfgLocal)
	_, _, _ = world.Init(context.Background(), cfgLocal)

	if err := importPeerPackageBytes(cfgLocal, rec.Body.Bytes()); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(cfgLocal.VaultPath, "notes", "remote.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "start mission remote") {
		t.Fatalf("expected transported note, got %q", string(body))
	}
}

func TestSyncDatabaseConflictRule(t *testing.T) {
	cfg, _ := setup(t)
	path := filepath.Join(t.TempDir(), "peer-db.json")
	pkg := peerPackage{
		Peer: "peer-db",
		Files: []peerFile{
			{
				Path:        "world.db",
				BaseHash:    hashString("base-db"),
				BaseB64:     base64.StdEncoding.EncodeToString([]byte("base-db")),
				Hash:        hashString("remote-db"),
				ContentB64:  base64.StdEncoding.EncodeToString([]byte("remote-db")),
				ContentType: "binary",
			},
		},
	}
	if err := os.WriteFile(cfg.DBPath, []byte("local-db"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, mustJSON(t, pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SyncImport(context.Background(), cfg, path); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(cfg.Home, "sync-conflicts", "peer-db"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected db conflict artifacts")
	}
}

func mustJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return body
}
