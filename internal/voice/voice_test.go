package voice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/testingx"
	"github.com/stizzfer36-del/UCLAW/internal/world"
)

func setup(t *testing.T) config.Config { t.Helper(); testingx.TempHome(t); cfg, _ := config.Load(); _ = config.EnsureLayout(cfg); _, _, _ = world.Init(context.Background(), cfg); return cfg }
func TestVoiceCommandAccuracy(t *testing.T) { cfg := setup(t); for i := 0; i < 10; i++ { res, err := Dispatch(context.Background(), cfg, "start mission voice-"+string(rune('a'+i)), false, false); if err != nil || res.Action != "mission_start" { t.Fatalf("expected mission_start, got %+v err=%v", res, err) } } }
func TestVoiceNoExternalByDefault(t *testing.T) { cfg := setup(t); _, _ = Dispatch(context.Background(), cfg, "start mission local", false, false); body, _ := os.ReadFile(cfg.AuditPath); if strings.Contains(string(body), "voice_external_opt_in") { t.Fatal("unexpected external opt-in audit") } }
func TestVoiceAuditLogged(t *testing.T) { cfg := setup(t); _, _ = Dispatch(context.Background(), cfg, "start mission audit", true, false); body, _ := os.ReadFile(cfg.AuditPath); if !strings.Contains(string(body), "voice_dispatch") { t.Fatal("expected voice dispatch audit event") } }
func TestVoiceDispatchFromFile(t *testing.T) { cfg := setup(t); path := filepath.Join(t.TempDir(), "voice.txt"); if err := os.WriteFile(path, []byte("start mission file based"), 0o644); err != nil { t.Fatal(err) }; res, err := DispatchFromFile(context.Background(), cfg, path, false, false); if err != nil || res.Action != "mission_start" { t.Fatalf("expected mission_start from file, got %+v err=%v", res, err) } }
func TestVoiceLiveDispatch(t *testing.T) { cfg := setup(t); captureCmd := "printf 'RIFFtest' > {output}"; sttCmd := "printf 'start mission live path'"; res, err := DispatchLive(context.Background(), cfg, LiveOptions{CaptureCommand:captureCmd, TranscribeCmd:sttCmd}); if err != nil { t.Fatal(err) }; if res.Action != "mission_start" || !strings.HasPrefix(res.AudioPath, cfg.VoicePath) { t.Fatalf("expected live mission_start with audio path, got %+v", res) } }
func TestVoiceTranscriptOnly(t *testing.T) { cfg := setup(t); audioPath := filepath.Join(t.TempDir(), "capture.wav"); if err := os.WriteFile(audioPath, []byte("wave"), 0o644); err != nil { t.Fatal(err) }; res, err := DispatchLive(context.Background(), cfg, LiveOptions{AudioPath:audioPath, TranscribeCmd:"printf 'start mission transcript only'", TranscriptOnly:true}); if err != nil { t.Fatal(err) }; if res.Action != "transcript" || res.Transcript == "" { t.Fatalf("expected transcript-only response, got %+v", res) } }
