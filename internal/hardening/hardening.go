package hardening

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/stizzfer36-del/uclaw/internal/audit"
	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/missions"
	"github.com/stizzfer36-del/uclaw/internal/policies"
)

func TightenPolicy(ctx context.Context, cfg config.Config, repoRoot, toolName string) error {
	path := filepath.Join(repoRoot, "core", "policies", "tools.yaml")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(body), "\n")
	for i := range lines {
		if strings.TrimSpace(lines[i]) == "- name: "+toolName {
			for j := i + 1; j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "- name:"); j++ {
				if strings.Contains(lines[j], "requires_approval: false") {
					lines[j] = strings.Replace(lines[j], "false", "true", 1)
				}
				if strings.Contains(lines[j], "risk_level: low") {
					lines[j] = strings.Replace(lines[j], "low", "medium", 1)
				}
				if strings.Contains(lines[j], "risk_level: medium") {
					lines[j] = strings.Replace(lines[j], "medium", "high", 1)
				}
			}
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "policy_tighten",
		Target:           toolName,
		Outcome:          "success",
		ApprovalRequired: true,
	})
}

func FuzzBlock(ctx context.Context, cfg config.Config, repoRoot string) error {
	reg, err := policies.LoadRegistry(repoRoot)
	if err != nil {
		return err
	}
	for _, tool := range []string{"deploy", "delete_file", "shell_exec"} {
		t, ok := reg.Tool(tool)
		if !ok || !t.RequiresApproval {
			return fmt.Errorf("tool %s should require approval", tool)
		}
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "verifier",
		Action:           "policy_fuzz_verify",
		Target:           "tool_registry",
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func Recover(ctx context.Context, cfg config.Config, missionID string) error {
	list, err := missions.ListCheckpoints(ctx, cfg, missionID)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return fmt.Errorf("no checkpoints for mission %s", missionID)
	}
	if err := missions.RestoreCheckpoint(ctx, cfg, list[len(list)-1].ID); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "mission_recover",
		Target:           missionID,
		MissionID:        missionID,
		Outcome:          "success",
		ApprovalRequired: true,
	})
}

func SyncExport(ctx context.Context, cfg config.Config, name string) (string, error) {
	path, err := exportPeerPackage(cfg, name)
	if err != nil {
		return "", err
	}
	if err := audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "sync_export",
		Target:           name,
		Outcome:          "success",
		ApprovalRequired: false,
	}); err != nil {
		return "", err
	}
	return path, nil
}

func SyncImport(ctx context.Context, cfg config.Config, path string) error {
	if err := importPeerPackage(cfg, path); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "sync_import",
		Target:           path,
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func SyncPull(ctx context.Context, cfg config.Config, name, endpoint string) error {
	if err := pullPeerPackage(ctx, cfg, name, endpoint); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "sync_pull",
		Target:           endpoint,
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func SyncPush(ctx context.Context, cfg config.Config, name, endpoint string) error {
	if err := pushPeerPackage(ctx, cfg, name, endpoint); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "sync_push",
		Target:           endpoint,
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func SyncServe(ctx context.Context, cfg config.Config, listen string, log io.Writer) error {
	if err := servePeerTransport(ctx, cfg, listen, log); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "sync_serve",
		Target:           listen,
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func ScaffoldPlugin(ctx context.Context, cfg config.Config, name string) (string, error) {
	root := filepath.Join(cfg.Home, "plugins", name, ".codex-plugin")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	manifest := filepath.Join(root, "plugin.json")
	body := []byte(fmt.Sprintf("{\"name\":\"%s\",\"version\":\"0.1.0\"}\n", name))
	if err := os.WriteFile(manifest, body, 0o644); err != nil {
		return "", err
	}
	if err := audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "plugin_scaffold",
		Target:           name,
		Outcome:          "success",
		ApprovalRequired: false,
	}); err != nil {
		return "", err
	}
	return manifest, nil
}

func GenerateDocs(ctx context.Context, cfg config.Config) (string, error) {
	path := filepath.Join(cfg.Home, "site", "index.html")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	body := []byte("<html><body><h1>UCLAW Docs</h1></body></html>")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	if err := audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "operator",
		Action:           "docs_generate",
		Target:           path,
		Outcome:          "success",
		ApprovalRequired: false,
	}); err != nil {
		return "", err
	}
	return path, nil
}

func SecretScan(ctx context.Context, cfg config.Config, root string) error {
	var hits []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if filepath.Base(path) == ".env" {
			return nil
		}
		if strings.HasPrefix(rel, "internal/testingx/") || strings.HasPrefix(rel, "compliance/evidence/") {
			return nil
		}
		if strings.HasSuffix(rel, "_test.go") || rel == "internal/hardening/hardening.go" || rel == "scripts/run_compliance_review.sh" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(body)
		if strings.Contains(content, "secret-anthropic") || strings.Contains(content, "secret-openrouter") {
			hits = append(hits, path)
		}
		return nil
	})
	if len(hits) > 0 {
		return fmt.Errorf("secrets leaked in %v", hits)
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "verifier",
		Action:           "secret_scan",
		Target:           root,
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func fileHash(path string) string {
	body, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
