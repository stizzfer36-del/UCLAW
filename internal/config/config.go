package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Home string; WorldName string; VaultPath string; DBPath string; GraphDBPath string; AuditPath string; SocketPath string; AgentsPath string; ChecksPath string; MissionsPath string; DesktopPath string; PeersPath string; VoicePath string; BudgetLimit string; Secrets map[string]string }
func Load() (Config, error) {
	home := strings.TrimSpace(os.Getenv("UCLAW_HOME")); if home == "" { userHome, err := os.UserHomeDir(); if err != nil { return Config{}, err }; home = filepath.Join(userHome, ".uclaw") }
	secrets := map[string]string{}; envPath := filepath.Join(home, ".env")
	if _, err := os.Stat(envPath); err == nil { if err := loadEnvFile(envPath, secrets); err != nil { return Config{}, err } } else if !errors.Is(err, os.ErrNotExist) { return Config{}, err }
	cfg := Config{Home:home,WorldName:firstNonEmpty(os.Getenv("UCLAW_WORLD_NAME"), secrets["UCLAW_WORLD_NAME"], "my-world"),VaultPath:expandPath(home, firstNonEmpty(os.Getenv("UCLAW_VAULT_PATH"), secrets["UCLAW_VAULT_PATH"], filepath.Join(home, "vault"))),DBPath:expandPath(home, firstNonEmpty(os.Getenv("UCLAW_DB_PATH"), secrets["UCLAW_DB_PATH"], filepath.Join(home, "world.db"))),GraphDBPath:filepath.Join(expandPath(home, firstNonEmpty(os.Getenv("UCLAW_VAULT_PATH"), secrets["UCLAW_VAULT_PATH"], filepath.Join(home, "vault"))), "graph.db"),AuditPath:expandPath(home, firstNonEmpty(os.Getenv("UCLAW_AUDIT_PATH"), secrets["UCLAW_AUDIT_PATH"], filepath.Join(home, "audit.jsonl"))),SocketPath:filepath.Join(home, "world.sock"),AgentsPath:filepath.Join(home, "agents"),ChecksPath:filepath.Join(home, "checkpoints"),MissionsPath:filepath.Join(home, "missions"),DesktopPath:filepath.Join(home, "desktop"),PeersPath:filepath.Join(home, "peers"),VoicePath:filepath.Join(home, "voice"),BudgetLimit:firstNonEmpty(os.Getenv("UCLAW_BUDGET_LIMIT"), secrets["UCLAW_BUDGET_LIMIT"], "10.00"),Secrets:secrets}
	return cfg, nil
}
func EnsureLayout(cfg Config) error { paths := []string{cfg.Home,cfg.VaultPath,filepath.Join(cfg.VaultPath, "decisions"),filepath.Join(cfg.VaultPath, "prompts"),filepath.Join(cfg.VaultPath, "sources"),filepath.Join(cfg.VaultPath, "logs"),filepath.Join(cfg.VaultPath, "notes"),filepath.Join(cfg.VaultPath, "research"),cfg.ChecksPath,cfg.AgentsPath,cfg.MissionsPath,cfg.DesktopPath,cfg.PeersPath,cfg.VoicePath,filepath.Dir(cfg.DBPath),filepath.Dir(cfg.AuditPath)}; for _, path := range paths { if err := os.MkdirAll(path, 0o755); err != nil { return err } }; return nil }
func Secret(cfg Config, key string) string { if value := os.Getenv(key); value != "" { return value }; return cfg.Secrets[key] }
func loadEnvFile(path string, secrets map[string]string) error { f, err := os.Open(path); if err != nil { return err }; defer f.Close(); scanner := bufio.NewScanner(f); for scanner.Scan() { line := strings.TrimSpace(scanner.Text()); if line == "" || strings.HasPrefix(line, "#") { continue }; parts := strings.SplitN(line, "=", 2); if len(parts) != 2 { continue }; secrets[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1]) }; return scanner.Err() }
func expandPath(home, path string) string { if strings.HasPrefix(path, "~/.uclaw") { return filepath.Join(home, strings.TrimPrefix(path, "~/.uclaw/")) }; if strings.HasPrefix(path, "~") { userHome, _ := os.UserHomeDir(); return filepath.Join(userHome, strings.TrimPrefix(path, "~/")) }; return path }
func firstNonEmpty(values ...string) string { for _, value := range values { if strings.TrimSpace(value) != "" { return value } }; return "" }
