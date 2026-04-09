package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/uclaw/internal/testingx"
)

func TestInitIsIdempotent(t *testing.T) {
	home := testingx.TempHome(t)
	var stdout, stderr bytes.Buffer

	if err := Execute(context.Background(), []string{"init"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	first := strings.TrimSpace(stdout.String())
	if !strings.HasPrefix(first, "world_id=") {
		t.Fatalf("unexpected output %q", first)
	}

	stdout.Reset()
	stderr.Reset()

	if err := Execute(context.Background(), []string{"init"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	second := strings.TrimSpace(stdout.String())
	if first != second {
		t.Fatalf("expected stable world id, first=%q second=%q", first, second)
	}

	if _, err := os.Stat(filepath.Join(home, "world.db")); err != nil {
		t.Fatal(err)
	}
}

func TestCommandUsesSharedTraceID(t *testing.T) {
	home := testingx.TempHome(t)
	var stdout, stderr bytes.Buffer

	if err := Execute(context.Background(), []string{"init"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	initialAudit, err := os.ReadFile(filepath.Join(home, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	initialLines := len(strings.Split(strings.TrimSpace(string(initialAudit)), "\n"))

	if err := Execute(context.Background(), []string{"voice", "--text", "start mission audit"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(home, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	lines = lines[initialLines:]
	if len(lines) < 2 {
		t.Fatalf("expected multiple audit events for voice command, got %d", len(lines))
	}
	traceIDs := map[string]bool{}
	for _, line := range lines {
		var event struct {
			TraceID string `json:"trace_id"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event.TraceID == "" {
			t.Fatal("expected trace_id on audit event")
		}
		traceIDs[event.TraceID] = true
	}
	if len(traceIDs) != 1 {
		t.Fatalf("expected one shared trace id, got %v", traceIDs)
	}
}

func TestInitDoesNotLeakSecrets(t *testing.T) {
	home := testingx.TempHome(t)
	var stdout, stderr bytes.Buffer

	if err := Execute(context.Background(), []string{"init"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	secretValues := []string{"secret-anthropic", "secret-openrouter"}
	combined := stdout.String() + stderr.String()
	for _, value := range secretValues {
		if strings.Contains(combined, value) {
			t.Fatalf("secret leaked to stdout/stderr: %s", value)
		}
	}

	auditBody, err := os.ReadFile(filepath.Join(home, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range secretValues {
		if strings.Contains(string(auditBody), value) {
			t.Fatalf("secret leaked to audit log: %s", value)
		}
	}
}
