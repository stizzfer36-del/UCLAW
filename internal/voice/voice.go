package voice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/stizzfer36-del/UCLAW/internal/audit"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/ids"
	"github.com/stizzfer36-del/UCLAW/internal/missions"
)

type Result struct { Action string `json:"action"`; MissionID string `json:"mission_id,omitempty"`; Feedback string `json:"feedback"`; Transcript string `json:"transcript,omitempty"`; AudioPath string `json:"audio_path,omitempty"` }
type LiveOptions struct { AudioPath string; CaptureSeconds int; Device string; CaptureCommand string; TranscribeCmd string; TranscriptOnly bool; TTS bool; External bool }
func Dispatch(ctx context.Context, cfg config.Config, transcript string, tts bool, external bool) (Result, error) { action := strings.ToLower(strings.TrimSpace(transcript)); if strings.HasPrefix(action, "start mission ") { title := strings.TrimSpace(strings.TrimPrefix(transcript, "start mission ")); m, err := missions.Start(ctx, cfg, title, "voice", "lead"); if err != nil { return Result{}, err }; _ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"voice", Action:"voice_dispatch", Target:title, MissionID:m.ID, Outcome:"success", ApprovalRequired:false}); feedback := "Mission created"; if tts { feedback = "TTS: Mission created" }; return Result{Action:"mission_start", MissionID:m.ID, Feedback:feedback, Transcript:transcript}, nil }; if external { _ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"voice", Action:"voice_external_opt_in", Outcome:"success", ApprovalRequired:true}) }; return Result{Action:"noop", Feedback:"No mission action parsed", Transcript:transcript}, nil }
func DispatchFromFile(ctx context.Context, cfg config.Config, transcriptPath string, tts bool, external bool) (Result, error) { body, err := os.ReadFile(transcriptPath); if err != nil { return Result{}, err }; return Dispatch(ctx, cfg, string(body), tts, external) }
func DispatchLive(ctx context.Context, cfg config.Config, opts LiveOptions) (Result, error) { audioPath := opts.AudioPath; if audioPath == "" { var err error; audioPath, err = CaptureAudio(ctx, cfg, opts); if err != nil { return Result{}, err } }; transcript, err := TranscribeAudio(ctx, cfg, audioPath, opts); if err != nil { return Result{}, err }; if err := audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"voice", Action:"voice_live_transcribe", Target:audioPath, Outcome:"success", ApprovalRequired:opts.External}); err != nil { return Result{}, err }; if opts.TranscriptOnly { return Result{Action:"transcript", Feedback:"Transcript captured", Transcript:transcript, AudioPath:audioPath}, nil }; result, err := Dispatch(ctx, cfg, transcript, opts.TTS, opts.External); if err != nil { return Result{}, err }; result.AudioPath = audioPath; result.Transcript = transcript; return result, nil }
func CaptureAudio(ctx context.Context, cfg config.Config, opts LiveOptions) (string, error) { seconds := opts.CaptureSeconds; if seconds <= 0 { seconds = 5 }; output := filepath.Join(cfg.VoicePath, ids.New("capture")+".wav"); command := strings.TrimSpace(opts.CaptureCommand); if command == "" { device := strings.TrimSpace(opts.Device); if device == "" { device = "default" }; command = fmt.Sprintf("ffmpeg -y -f alsa -i %s -t %d %s", shellEscape(device), seconds, shellEscape(output)) } else { command = expandTemplate(command, output, seconds, opts.Device) }; if err := runShell(ctx, command); err != nil { return "", err }; if _, err := os.Stat(output); err != nil { return "", err }; if err := audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"voice", Action:"voice_live_capture", Target:output, Outcome:"success", ApprovalRequired:false}); err != nil { return "", err }; return output, nil }
func TranscribeAudio(ctx context.Context, cfg config.Config, audioPath string, opts LiveOptions) (string, error) { command := strings.TrimSpace(opts.TranscribeCmd); if command == "" { command = strings.TrimSpace(os.Getenv("UCLAW_STT_COMMAND")) }; if command == "" { command = whisperCommand(audioPath) }; if command == "" { return "", fmt.Errorf("no local STT command configured") }; output, err := runShellOutput(ctx, expandTemplate(command, audioPath, opts.CaptureSeconds, opts.Device)); if err != nil { return "", err }; transcript := strings.TrimSpace(output); if transcript == "" { return "", fmt.Errorf("empty transcript from STT command") }; _ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:"voice", Action:"voice_transcribe", Target:audioPath, Outcome:"success", ApprovalRequired:opts.External}); return transcript, nil }
func whisperCommand(audioPath string) string { txtPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".txt"; return fmt.Sprintf("python3 - <<'PY'\nfrom pathlib import Path\ntext = Path(%q)\nif text.exists():\n    print(text.read_text())\nPY", txtPath) }
func expandTemplate(command, output string, seconds int, device string) string { repl := strings.NewReplacer("{output}", shellEscape(output), "{input}", shellEscape(output), "{seconds}", fmt.Sprintf("%d", seconds), "{device}", shellEscape(device)); return repl.Replace(command) }
func runShell(ctx context.Context, command string) error { cmd := exec.CommandContext(ctx, "bash", "-lc", command); cmd.Stderr = os.Stderr; return cmd.Run() }
func runShellOutput(ctx context.Context, command string) (string, error) { cmd := exec.CommandContext(ctx, "bash", "-lc", command); out, err := cmd.Output(); return string(out), err }
func shellEscape(value string) string { if value == "" { return "''" }; return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'" }
func RecordingPath(cfg config.Config) string { return filepath.Join(cfg.VoicePath, time.Now().UTC().Format("20060102T150405")+".wav") }
