package testingx

import (
	"os"
	"path/filepath"
	"testing"
)

func TempHome(t *testing.T) string {
	t.Helper(); dir := t.TempDir(); home := filepath.Join(dir, ".uclaw"); if err := os.MkdirAll(home, 0o755); err != nil { t.Fatal(err) }; envPath := filepath.Join(home, ".env"); content := "ANTHROPIC_API_KEY=secret-anthropic\nOPENROUTER_API_KEY=secret-openrouter\nUCLAW_WORLD_NAME=test-world\n"; if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil { t.Fatal(err) }; t.Setenv("UCLAW_HOME", home); return home }
