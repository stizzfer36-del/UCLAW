package audit

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stizzfer36-del/uclaw/internal/ids"
)

type Event struct { EventID string `json:"event_id"`; TraceID string `json:"trace_id,omitempty"`; Timestamp string `json:"timestamp"`; AgentID string `json:"agent_id"`; Action string `json:"action"`; Target string `json:"target,omitempty"`; Outcome string `json:"outcome"`; MissionID string `json:"mission_id,omitempty"`; Tool string `json:"tool,omitempty"`; ApprovalRequired bool `json:"approval_required"`; ApprovalGranted *bool `json:"approval_granted"`; PrevEventHash string `json:"prev_event_hash,omitempty"` }
type traceKey struct{}
func WithTrace(ctx context.Context, traceID string) context.Context { return context.WithValue(ctx, traceKey{}, traceID) }
func WithNewTrace(ctx context.Context) context.Context { return WithTrace(ctx, ids.New("trace")) }
func TraceID(ctx context.Context) string { if ctx == nil { return "" }; value, _ := ctx.Value(traceKey{}).(string); return value }
func Write(ctx context.Context, path string, event Event) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
	prevHash, err := lastHash(path); if err != nil { return err }
	event.EventID = ids.New("evt"); if event.TraceID == "" { event.TraceID = TraceID(ctx) }; if event.TraceID == "" { event.TraceID = ids.New("trace") }; event.Timestamp = time.Now().UTC().Format(time.RFC3339); event.PrevEventHash = prevHash
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); if err != nil { return err }
	defer f.Close(); enc := json.NewEncoder(f); return enc.Encode(event)
}
func lastHash(path string) (string, error) {
	f, err := os.Open(path); if err != nil { if os.IsNotExist(err) { return "", nil }; return "", err }
	defer f.Close(); var last string; scanner := bufio.NewScanner(f); for scanner.Scan() { line := strings.TrimSpace(scanner.Text()); if line != "" { last = line } }; if err := scanner.Err(); err != nil { return "", err }; if last == "" { return "", nil }; sum := sha256.Sum256([]byte(last)); return hex.EncodeToString(sum[:]), nil }
