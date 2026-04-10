package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/stizzfer36-del/UCLAW/internal/agents"
	"github.com/stizzfer36-del/UCLAW/internal/artifacts"
	"github.com/stizzfer36-del/UCLAW/internal/audit"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/hardening"
	"github.com/stizzfer36-del/UCLAW/internal/ipc"
	"github.com/stizzfer36-del/UCLAW/internal/memory"
	"github.com/stizzfer36-del/UCLAW/internal/missions"
	"github.com/stizzfer36-del/UCLAW/internal/observability"
	"github.com/stizzfer36-del/UCLAW/internal/planner"
	"github.com/stizzfer36-del/UCLAW/internal/voice"
	"github.com/stizzfer36-del/UCLAW/internal/world"
)

func Main(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if err := Execute(ctx, args, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx = audit.WithNewTrace(ctx)

	if len(args) == 0 {
		return runDesktop(ctx, cfg, nil, stdout)
	}

	switch args[0] {
	case "init":
		return runInit(ctx, cfg, stdout)
	case "world":
		return runWorld(ctx, cfg, args[1:], stdout)
	case "agent":
		return runAgent(ctx, cfg, args[1:], stdout)
	case "approve":
		return runApprove(ctx, cfg, args[1:], stdout)
	case "audit":
		return runAudit(cfg, args[1:], stdout)
	case "memory":
		return runMemory(ctx, cfg, args[1:], stdout)
	case "artifact":
		return runArtifact(ctx, cfg, args[1:], stdout)
	case "mission":
		return runMission(ctx, cfg, args[1:], stdout)
	case "plan":
		return runPlan(ctx, cfg, args[1:], stdout)
	case "review":
		return runReview(ctx, cfg, args[1:], stdout)
	case "override":
		return runOverride(ctx, cfg, args[1:], stdout)
	case "status":
		return runStatus(ctx, cfg, stdout)
	case "health":
		return runHealth(ctx, cfg, stdout)
	case "budget":
		return runBudget(ctx, cfg, args[1:], stdout)
	case "pause":
		return runPause(ctx, cfg, args[1:], stdout)
	case "resume":
		return runResume(ctx, cfg, args[1:], stdout)
	case "policy":
		return runPolicy(ctx, cfg, args[1:], stdout)
	case "desktop":
		return runDesktop(ctx, cfg, args[1:], stdout)
	case "voice":
		return runVoice(ctx, cfg, args[1:], stdout)
	case "sync":
		return runSync(ctx, cfg, args[1:], stdout)
	case "plugin":
		return runPlugin(ctx, cfg, args[1:], stdout)
	case "docs":
		return runDocs(ctx, cfg, args[1:], stdout)
	case "secrets":
		return runSecrets(cfg, args[1:], stdout)
	default:
		return printUsage(stdout)
	}
}

func runInit(ctx context.Context, cfg config.Config, stdout io.Writer) error {
	if err := config.EnsureLayout(cfg); err != nil {
		return err
	}
	summary, created, err := world.Init(ctx, cfg)
	if err != nil {
		return err
	}
	if err := audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          "system",
		Action:           "uclaw_init",
		Target:           cfg.DBPath,
		Outcome:          outcome(created),
		ApprovalRequired: false,
	}); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "world_id=%s\n", summary.ID)
	return nil
}

func runWorld(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing world subcommand")
	}
	switch args[0] {
	case "inspect":
		fs := flag.NewFlagSet("world inspect", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		nodeID := fs.String("node", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		value, err := world.Inspect(ctx, cfg.DBPath, *nodeID)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(value)
	case "socket":
		return runSocket(ctx, cfg, args[1:], stdout)
	default:
		return fmt.Errorf("unknown world subcommand %q", args[0])
	}
}

func runSocket(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing socket action")
	}
	switch args[0] {
	case "serve":
		if err := config.EnsureLayout(cfg); err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer cancel()
		fmt.Fprintf(stdout, "socket=%s\n", ipc.FormatEndpoint(cfg.SocketPath))
		return ipc.Serve(ctx, cfg.SocketPath)
	case "ping":
		resp, err := ipc.Ping(ctx, cfg.SocketPath)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(resp)
	default:
		return fmt.Errorf("unknown socket action %q", args[0])
	}
}

func runSecrets(cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) < 2 || args[0] != "get" {
		return errors.New("usage: uclaw secrets get <key>")
	}
	fmt.Fprintln(stdout, config.Secret(cfg, args[1]))
	return nil
}

func runAgent(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing agent subcommand")
	}
	repoRoot, err := repoRoot()
	if err != nil {
		return err
	}
	switch args[0] {
	case "spawn":
		nameArg, flagArgs := splitLeadingPositionals(args[1:], 1)
		fs := flag.NewFlagSet("agent spawn", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		teamID := fs.String("team", "", "")
		role := fs.String("role", "dev", "")
		provider := fs.String("provider", "mock", "")
		handbook := fs.String("handbook", "", "")
		capabilitiesCSV := fs.String("capabilities", "", "")
		allowPathCSV := fs.String("allow-paths", "", "")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}
		if len(nameArg) == 0 {
			return errors.New("usage: uclaw agent spawn <name> [flags]")
		}
		profile, err := agents.Spawn(ctx, cfg, repoRoot, agents.SpawnRequest{
			Name:         nameArg[0],
			Role:         *role,
			TeamID:       *teamID,
			Provider:     *provider,
			Capabilities: splitCSV(*capabilitiesCSV),
			AllowedPaths: normalizePaths(splitCSV(*allowPathCSV)),
			HandbookPath: *handbook,
		})
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(profile)
	case "list":
		profiles, err := agents.List(ctx, cfg)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(profiles)
	case "inspect":
		if len(args) < 2 {
			return errors.New("usage: uclaw agent inspect <id>")
		}
		profile, handbook, err := agents.Inspect(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(map[string]interface{}{
			"profile":  profile,
			"handbook": handbook,
		})
	case "retire":
		if len(args) < 2 {
			return errors.New("usage: uclaw agent retire <id>")
		}
		return agents.Retire(ctx, cfg, args[1])
	case "pause":
		if len(args) < 2 {
			return errors.New("usage: uclaw agent pause <id>")
		}
		return agents.Pause(ctx, cfg, args[1])
	case "resume":
		if len(args) < 2 {
			return errors.New("usage: uclaw agent resume <id>")
		}
		return agents.Resume(ctx, cfg, args[1])
	case "call-tool":
		positional, flagArgs := splitLeadingPositionals(args[1:], 2)
		fs := flag.NewFlagSet("agent call-tool", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		target := fs.String("target", "", "")
		missionID := fs.String("mission", "", "")
		commandCSV := fs.String("command", "", "")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}
		if len(positional) < 2 {
			return errors.New("usage: uclaw agent call-tool <agent_id> <tool_name> [--target path]")
		}
		result, err := agents.CallTool(ctx, cfg, repoRoot, positional[0], positional[1], *target, *missionID, splitCSV(*commandCSV))
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(result)
	default:
		return fmt.Errorf("unknown agent subcommand %q", args[0])
	}
}

func runApprove(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: uclaw approve <request_id>")
	}
	if err := agents.Approve(ctx, cfg, args[0], "operator"); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "approved")
	return err
}

func runAudit(cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 1 && args[0] == "verify" {
		if err := audit.Verify(cfg.AuditPath); err != nil {
			return err
		}
		_, err := fmt.Fprintln(stdout, "audit-ok")
		return err
	}
	body, err := os.ReadFile(cfg.AuditPath)
	if err != nil {
		return err
	}
	_, err = stdout.Write(body)
	return err
}

func runMemory(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing memory subcommand")
	}
	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("memory create", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		nodeType := fs.String("type", "note", "")
		title := fs.String("title", "", "")
		content := fs.String("content", "", "")
		missionID := fs.String("mission", "", "")
		agentID := fs.String("agent", "", "")
		verified := fs.Bool("verified", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		node, err := memory.CreateNode(ctx, cfg, memory.Node{
			Type:      *nodeType,
			Title:     *title,
			MissionID: *missionID,
			AgentID:   *agentID,
			Content:   *content,
			Verified:  *verified,
		})
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(node)
	case "link":
		fs := flag.NewFlagSet("memory link", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		fromID := fs.String("from", "", "")
		toID := fs.String("to", "", "")
		edgeType := fs.String("edge", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		edge, err := memory.AddEdge(ctx, cfg, *fromID, *toID, *edgeType)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(edge)
	case "query":
		fs := flag.NewFlagSet("memory query", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		nodeType := fs.String("type", "", "")
		edgeType := fs.String("edge", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		nodes, err := memory.Query(ctx, cfg, *nodeType, *edgeType)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(nodes)
	case "graph":
		fs := flag.NewFlagSet("memory graph", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		depth := fs.Int("depth", 1, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		graph, err := memory.Graph(ctx, cfg, *missionID, *depth)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(graph)
	case "search":
		if len(args) < 2 {
			return errors.New("usage: uclaw memory search <query>")
		}
		nodes, err := memory.Search(ctx, cfg, strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(nodes)
	case "audit":
		nodes, err := memory.Unverified(ctx, cfg)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(nodes)
	case "log":
		fs := flag.NewFlagSet("memory log", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		agentID := fs.String("agent", "", "")
		content := fs.String("content", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		node, err := memory.WriteAgentLog(ctx, cfg, *missionID, *agentID, *content)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(node)
	default:
		return fmt.Errorf("unknown memory subcommand %q", args[0])
	}
}

func runArtifact(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing artifact subcommand")
	}
	repoRoot, err := repoRoot()
	if err != nil {
		return err
	}
	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("artifact create", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		title := fs.String("title", "", "")
		artifactType := fs.String("type", "doc", "")
		path := fs.String("path", "", "")
		agentID := fs.String("agent", "", "")
		missionID := fs.String("mission", "", "")
		claimCount := fs.Int("claim-count", 1, "")
		sourcesCSV := fs.String("sources", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		art, err := artifacts.Create(ctx, cfg, repoRoot, artifacts.CreateRequest{
			Title:       *title,
			Type:        *artifactType,
			Path:        *path,
			OriginAgent: *agentID,
			MissionID:   *missionID,
			ClaimCount:  *claimCount,
			Sources:     parseSources(*sourcesCSV, *agentID),
		})
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(art)
	case "list":
		fs := flag.NewFlagSet("artifact list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		list, err := artifacts.List(ctx, cfg, *missionID)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(list)
	case "show":
		if len(args) < 2 {
			return errors.New("usage: uclaw artifact show <id>")
		}
		art, err := artifacts.Show(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(art)
	case "verify":
		fs := flag.NewFlagSet("artifact verify", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		verifierID := fs.String("verifier", "", "")
		testCommand := fs.String("test-command", "", "")
		workspace := fs.String("workspace", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw artifact verify <id>")
		}
		if *testCommand == "" {
			if resolved, err := artifacts.ResolveTestCommand(cfg, args[1]); err == nil {
				*testCommand = resolved
			}
		}
		checks, err := artifacts.RunChecks(ctx, cfg, args[1], *verifierID, *testCommand, *workspace)
		if err != nil {
			return err
		}
		artifact, err := artifacts.Show(ctx, cfg, args[1])
		if err == nil {
			failed := false
			for _, check := range checks {
				if check.Status == "failed" {
					failed = true
				}
			}
			if failed {
				_ = missions.UpdateStatus(ctx, cfg, artifact.MissionID, "blocked")
				_, _ = missions.MaybeGenerateAmendment(ctx, cfg, artifact.OriginAgent)
			} else {
				_ = missions.UpdateStatus(ctx, cfg, artifact.MissionID, "verifying")
			}
		}
		return json.NewEncoder(stdout).Encode(checks)
	case "flag":
		fs := flag.NewFlagSet("artifact flag", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		reason := fs.String("reason", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw artifact flag <id> --reason <reason>")
		}
		return artifacts.Flag(ctx, cfg, args[1], *reason)
	case "sign":
		fs := flag.NewFlagSet("artifact sign", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		signer := fs.String("by", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw artifact sign <id> --by <agent>")
		}
		return artifacts.Sign(ctx, cfg, args[1], *signer)
	case "revert":
		fs := flag.NewFlagSet("artifact revert", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		to := fs.String("to", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw artifact revert <id> --to <sha>")
		}
		return artifacts.Revert(ctx, cfg, repoRoot, args[1], *to)
	default:
		return fmt.Errorf("unknown artifact subcommand %q", args[0])
	}
}

func runMission(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing mission subcommand")
	}
	switch args[0] {
	case "start":
		fs := flag.NewFlagSet("mission start", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		createdBy := fs.String("by", "user", "")
		assignedTo := fs.String("assign", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw mission start <title>")
		}
		m, err := missions.Start(ctx, cfg, args[1], *createdBy, *assignedTo)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(m)
	case "list":
		m, err := missions.List(ctx, cfg)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(m)
	case "status":
		if len(args) < 2 {
			return errors.New("usage: uclaw mission status <id>")
		}
		status, err := missions.Status(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(status)
	case "set-status":
		fs := flag.NewFlagSet("mission set-status", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		value := fs.String("value", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw mission set-status <id> --value <status>")
		}
		return missions.UpdateStatus(ctx, cfg, args[1], *value)
	case "matrix":
		fs := flag.NewFlagSet("mission matrix", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		entriesCSV := fs.String("entries", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw mission matrix <id> --entries ...")
		}
		path, err := missions.SaveMatrix(cfg, args[1], parseMatrixEntries(*entriesCSV))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, path)
		return err
	case "checkpoints":
		if len(args) < 2 {
			return errors.New("usage: uclaw mission checkpoints <id>")
		}
		list, err := missions.ListCheckpoints(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(list)
	case "checkpoint-save":
		fs := flag.NewFlagSet("mission checkpoint-save", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		trigger := fs.String("trigger", "manual", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw mission checkpoint-save <id> --trigger <value>")
		}
		cp, err := missions.SaveCheckpoint(ctx, cfg, args[1], *trigger)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(cp)
	case "replay":
		fs := flag.NewFlagSet("mission replay", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		from := fs.String("from", "", "")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if len(args) < 2 {
			return errors.New("usage: uclaw mission replay <id> --from <checkpoint>")
		}
		status, err := missions.Replay(ctx, cfg, args[1], *from)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(status)
	case "rollback":
		if len(args) < 2 {
			return errors.New("usage: uclaw mission rollback <id>")
		}
		list, err := missions.ListCheckpoints(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return errors.New("no checkpoints for mission")
		}
		return missions.RestoreCheckpoint(ctx, cfg, list[len(list)-1].ID)
	case "diff":
		if len(args) < 3 {
			return errors.New("usage: uclaw mission diff <a> <b>")
		}
		diff, err := missions.DiffCheckpoints(ctx, cfg, args[1], args[2])
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(diff)
	default:
		return fmt.Errorf("unknown mission subcommand %q", args[0])
	}
}

func runPlan(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing plan subcommand")
	}
	repoRoot, err := repoRoot()
	if err != nil {
		return err
	}
	switch args[0] {
	case "orchestrate":
		fs := flag.NewFlagSet("plan orchestrate", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		lead := fs.String("lead", "planner", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := planner.Orchestrate(ctx, cfg, repoRoot, *missionID, *lead)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(result)
	case "matrix":
		fs := flag.NewFlagSet("plan matrix", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return planner.DefaultMatrix(cfg, *missionID)
	default:
		return fmt.Errorf("unknown plan subcommand %q", args[0])
	}
}

func runReview(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("missing review subcommand")
	}
	switch args[0] {
	case "queue":
		queue, err := missions.ReviewQueue(ctx, cfg)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(queue)
	default:
		return fmt.Errorf("unknown review subcommand %q", args[0])
	}
}

func runOverride(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) < 2 || args[0] != "--artifact" {
		return errors.New("usage: uclaw override --artifact <id>")
	}
	if err := missions.OverrideReview(ctx, cfg, args[1]); err != nil {
		return err
	}
	if err := artifacts.Reverify(ctx, cfg, args[1]); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "override-applied-reverification-required")
	return err
}

func runStatus(ctx context.Context, cfg config.Config, stdout io.Writer) error {
	state, err := observability.Status(ctx, cfg)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(state)
}

func runHealth(ctx context.Context, cfg config.Config, stdout io.Writer) error {
	health, err := observability.Health(ctx, cfg)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(health)
}

func runBudget(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) >= 1 && args[0] == "record" {
		fs := flag.NewFlagSet("budget record", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		missionID := fs.String("mission", "", "")
		agentID := fs.String("agent", "", "")
		provider := fs.String("provider", "mock", "")
		tokens := fs.Int("tokens", 0, "")
		cost := fs.Float64("cost", 0, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return missions.RecordUsage(ctx, cfg, *missionID, *agentID, *provider, *tokens, *cost)
	}
	budget, err := observability.Budget(ctx, cfg)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(budget)
}

func runPause(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 1 && args[0] == "--all" {
		profiles, err := agents.List(ctx, cfg)
		if err != nil {
			return err
		}
		for _, profile := range profiles {
			_ = agents.Pause(ctx, cfg, profile.ID)
		}
		_, err = fmt.Fprintln(stdout, "paused-all")
		return err
	}
	return errors.New("usage: uclaw pause --all")
}

func runResume(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 2 && args[0] == "--agent" {
		if err := agents.Resume(ctx, cfg, args[1]); err != nil {
			return err
		}
		_, err := fmt.Fprintln(stdout, "resumed")
		return err
	}
	return errors.New("usage: uclaw resume --agent <id>")
}

func runPolicy(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 2 && args[0] == "tighten" {
		repoRoot, err := repoRoot()
		if err != nil {
			return err
		}
		if err := hardening.TightenPolicy(ctx, cfg, repoRoot, args[1]); err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, "policy-tightened")
		return err
	}
	if len(args) == 3 && args[0] == "tighten" && args[1] == "--tool" {
		repoRoot, err := repoRoot()
		if err != nil {
			return err
		}
		if err := hardening.TightenPolicy(ctx, cfg, repoRoot, args[2]); err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, "policy-tightened")
		return err
	}
	return errors.New("usage: uclaw policy tighten --tool <name>")
}

func runDesktop(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	state, err := observability.Status(ctx, cfg)
	if err != nil {
		return err
	}
	if len(args) > 0 && args[0] == "build" {
		path := filepath.Join(cfg.DesktopPath, "state.json")
		body, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, path)
		return err
	}
	format := "html"
	width := 120
	if len(args) > 0 && args[0] == "render" {
		fs := flag.NewFlagSet("desktop render", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		f := fs.String("format", "html", "")
		w := fs.Int("width", 120, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		format = *f
		width = *w
	}
	if format == "tui" {
		_, err := fmt.Fprint(stdout, observability.RenderTUI(state, width))
		return err
	}
	path := filepath.Join(cfg.DesktopPath, "index.html")
	out, err := observability.RenderHTML(state, path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, out)
	return err
}

func runVoice(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("voice", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	text := fs.String("text", "", "")
	file := fs.String("file", "", "")
	audio := fs.String("audio", "", "")
	live := fs.Bool("live", false, "")
	tts := fs.Bool("tts", false, "")
	external := fs.Bool("external-stt", false, "")
	transcriptOnly := fs.Bool("transcript-only", false, "")
	seconds := fs.Int("seconds", 5, "")
	device := fs.String("device", "", "")
	captureCommand := fs.String("capture-command", "", "")
	sttCommand := fs.String("stt-command", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var result voice.Result
	var err error
	switch {
	case *live || *audio != "":
		result, err = voice.DispatchLive(ctx, cfg, voice.LiveOptions{
			AudioPath:      *audio,
			CaptureSeconds: *seconds,
			Device:         *device,
			CaptureCommand: *captureCommand,
			TranscribeCmd:  *sttCommand,
			TranscriptOnly: *transcriptOnly,
			TTS:            *tts,
			External:       *external,
		})
	case *file != "":
		result, err = voice.DispatchFromFile(ctx, cfg, *file, *tts, *external)
	default:
		result, err = voice.Dispatch(ctx, cfg, *text, *tts, *external)
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(result)
}

func runSync(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	switch {
	case len(args) == 2 && args[0] == "export":
		path, err := hardening.SyncExport(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, path)
		return err
	case len(args) == 2 && args[0] == "import":
		return hardening.SyncImport(ctx, cfg, args[1])
	case len(args) > 0 && args[0] == "serve":
		fs := flag.NewFlagSet("sync serve", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		listen := fs.String("listen", "127.0.0.1:44144", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return hardening.SyncServe(ctx, cfg, *listen, stdout)
	case len(args) > 0 && args[0] == "pull":
		fs := flag.NewFlagSet("sync pull", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		peer := fs.String("peer", "peer", "")
		from := fs.String("from", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*from) == "" {
			return errors.New("usage: uclaw sync pull --from <url> [--peer <name>]")
		}
		return hardening.SyncPull(ctx, cfg, *peer, *from)
	case len(args) > 0 && args[0] == "push":
		fs := flag.NewFlagSet("sync push", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		peer := fs.String("peer", "peer", "")
		to := fs.String("to", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*to) == "" {
			return errors.New("usage: uclaw sync push --to <url> [--peer <name>]")
		}
		return hardening.SyncPush(ctx, cfg, *peer, *to)
	}
	return errors.New("usage: uclaw sync <export|import> <name|path> | serve [--listen <addr>] | pull --from <url> [--peer <name>] | push --to <url> [--peer <name>]")
}

func runPlugin(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 2 && args[0] == "new" {
		path, err := hardening.ScaffoldPlugin(ctx, cfg, args[1])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, path)
		return err
	}
	return errors.New("usage: uclaw plugin new <name>")
}

func runDocs(ctx context.Context, cfg config.Config, args []string, stdout io.Writer) error {
	if len(args) == 1 && args[0] == "build" {
		path, err := hardening.GenerateDocs(ctx, cfg)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, path)
		return err
	}
	return errors.New("usage: uclaw docs build")
}

func printUsage(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, "usage: uclaw <desktop|status|health|budget|voice|sync|plugin|docs|plan|init|world|agent|approve|audit|memory|artifact|mission|review|override|policy|pause|resume|secrets>")
	return err
}

func outcome(created bool) string {
	if created {
		return "created"
	}
	return "existing"
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func normalizePaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		out = append(out, filepath.Clean(path))
	}
	return out
}

func parseSources(csv, agentID string) []artifacts.Source {
	if csv == "" {
		return nil
	}
	out := []artifacts.Source{}
	for _, raw := range strings.Split(csv, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		sourceType := "url"
		if strings.HasPrefix(raw, "vault:") {
			sourceType = "vault_node"
			raw = strings.TrimPrefix(raw, "vault:")
		}
		out = append(out, artifacts.Source{Type: sourceType, Ref: raw, CitedBy: agentID})
	}
	return out
}

func parseMatrixEntries(csv string) []missions.MatrixEntry {
	if csv == "" {
		return nil
	}
	out := []missions.MatrixEntry{}
	for _, chunk := range strings.Split(csv, ",") {
		parts := strings.Split(chunk, "|")
		entry := missions.MatrixEntry{Required: true}
		if len(parts) > 0 {
			entry.ID = parts[0]
		}
		if len(parts) > 1 {
			entry.Type = parts[1]
		}
		if len(parts) > 2 {
			entry.Target = parts[2]
		}
		if len(parts) > 3 {
			entry.Command = parts[3]
		}
		out = append(out, entry)
	}
	return out
}

func splitLeadingPositionals(args []string, want int) ([]string, []string) {
	positional := []string{}
	rest := []string{}
	for i, arg := range args {
		if len(positional) < want && !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		rest = append(rest, args[i:]...)
		break
	}
	return positional, rest
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("repo root not found")
		}
		dir = parent
	}
}
