package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stizzfer36-del/UCLAW/internal/artifacts"
	"github.com/stizzfer36-del/UCLAW/internal/audit"
	"github.com/stizzfer36-del/UCLAW/internal/config"
	"github.com/stizzfer36-del/UCLAW/internal/ids"
	"github.com/stizzfer36-del/UCLAW/internal/policies"
	"github.com/stizzfer36-del/UCLAW/internal/providers"
	"github.com/stizzfer36-del/UCLAW/internal/sqlitepy"
)

type Profile struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	TeamID       string   `json:"team_id"`
	MemberID     string   `json:"member_id"`
	Provider     string   `json:"provider"`
	Status       string   `json:"status"`
	HandbookPath string   `json:"handbook_path"`
	Capabilities []string `json:"capabilities"`
	AllowedPaths []string `json:"allowed_paths"`
}

type SpawnRequest struct {
	Name         string
	Role         string
	TeamID       string
	Provider     string
	Capabilities []string
	AllowedPaths []string
	HandbookPath string
}

type ToolResult struct {
	Status      string `json:"status"`
	RequestID   string `json:"request_id,omitempty"`
	Output      string `json:"output,omitempty"`
	ApprovalMsg string `json:"approval_message,omitempty"`
}

func Spawn(ctx context.Context, cfg config.Config, repoRoot string, req SpawnRequest) (Profile, error) {
	registry, err := policies.LoadRegistry(repoRoot)
	if err != nil {
		return Profile{}, err
	}
	for _, capability := range req.Capabilities {
		if _, ok := registry.Tool(capability); !ok {
			return Profile{}, fmt.Errorf("capability %q is not registered", capability)
		}
	}
	if req.TeamID == "" {
		req.TeamID, err = defaultTeamID(ctx, cfg.DBPath)
		if err != nil {
			return Profile{}, err
		}
	}
	if req.Provider == "" {
		req.Provider = "mock"
	}
	if _, err := providers.New(req.Provider, ""); err != nil {
		return Profile{}, err
	}

	profile := Profile{
		ID:           ids.New("agent"),
		Name:         req.Name,
		Role:         firstNonEmpty(req.Role, "dev"),
		TeamID:       req.TeamID,
		MemberID:     ids.New("member"),
		Provider:     req.Provider,
		Status:       "active",
		Capabilities: req.Capabilities,
		AllowedPaths: req.AllowedPaths,
	}
	if err := os.MkdirAll(filepath.Join(cfg.AgentsPath, profile.ID), 0o755); err != nil {
		return Profile{}, err
	}
	if req.HandbookPath == "" {
		req.HandbookPath = filepath.Join(cfg.AgentsPath, profile.ID, "handbook.md")
		if err := os.WriteFile(req.HandbookPath, []byte(defaultHandbook(profile.Name, profile.Role)), 0o644); err != nil {
			return Profile{}, err
		}
	}
	profile.HandbookPath = req.HandbookPath

	now := time.Now().UTC().Format(time.RFC3339)
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO members(id, team_id, name, type, handbook_path, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		profile.MemberID, profile.TeamID, profile.Name, "agent", profile.HandbookPath, now); err != nil {
		return Profile{}, err
	}
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO agent_profiles(id, member_id, role, provider, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		profile.ID, profile.MemberID, profile.Role, profile.Provider, profile.Status, now); err != nil {
		return Profile{}, err
	}
	for _, capability := range profile.Capabilities {
		if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO capabilities(id, member_id, tool_name, granted_at) VALUES (?, ?, ?, ?)`,
			ids.New("cap"), profile.MemberID, capability, now); err != nil {
			return Profile{}, err
		}
	}
	for _, allowedPath := range profile.AllowedPaths {
		if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO path_whitelists(id, member_id, path) VALUES (?, ?, ?)`,
			ids.New("path"), profile.MemberID, allowedPath); err != nil {
			return Profile{}, err
		}
	}
	if err := writeProfileFile(cfg, profile); err != nil {
		return Profile{}, err
	}
	if err := audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          profile.ID,
		Action:           "agent_spawn",
		Target:           profile.Name,
		Outcome:          "success",
		ApprovalRequired: false,
	}); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func List(ctx context.Context, cfg config.Config) ([]Profile, error) {
	rows, err := sqlitepy.Query(ctx, cfg.DBPath, `
SELECT ap.id, m.name, ap.role, m.team_id, ap.member_id, ap.provider, ap.status, m.handbook_path
FROM agent_profiles ap
JOIN members m ON m.id = ap.member_id
ORDER BY m.created_at ASC`)
	if err != nil {
		return nil, err
	}
	profiles := make([]Profile, 0, len(rows))
	for _, row := range rows {
		profile := Profile{
			ID:           str(row["id"]),
			Name:         str(row["name"]),
			Role:         str(row["role"]),
			TeamID:       str(row["team_id"]),
			MemberID:     str(row["member_id"]),
			Provider:     str(row["provider"]),
			Status:       str(row["status"]),
			HandbookPath: str(row["handbook_path"]),
		}
		caps, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT tool_name FROM capabilities WHERE member_id = ? ORDER BY tool_name`, profile.MemberID)
		if err != nil {
			return nil, err
		}
		for _, capRow := range caps {
			profile.Capabilities = append(profile.Capabilities, str(capRow["tool_name"]))
		}
		paths, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT path FROM path_whitelists WHERE member_id = ? ORDER BY path`, profile.MemberID)
		if err != nil {
			return nil, err
		}
		for _, pathRow := range paths {
			profile.AllowedPaths = append(profile.AllowedPaths, str(pathRow["path"]))
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func Inspect(ctx context.Context, cfg config.Config, agentID string) (Profile, Handbook, error) {
	profiles, err := List(ctx, cfg)
	if err != nil {
		return Profile{}, Handbook{}, err
	}
	for _, profile := range profiles {
		if profile.ID == agentID {
			handbook, err := LoadHandbook(profile.HandbookPath)
			return profile, handbook, err
		}
	}
	return Profile{}, Handbook{}, fmt.Errorf("agent %s not found", agentID)
}

func Retire(ctx context.Context, cfg config.Config, agentID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE agent_profiles SET status = ?, retired_at = ? WHERE id = ?`, "retired", now, agentID); err != nil {
		return err
	}
	return audit.Write(ctx, cfg.AuditPath, audit.Event{
		AgentID:          agentID,
		Action:           "agent_retire",
		Outcome:          "success",
		ApprovalRequired: false,
	})
}

func Pause(ctx context.Context, cfg config.Config, agentID string) error {
	return sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE agent_profiles SET status = ? WHERE id = ?`, "paused", agentID)
}

func Resume(ctx context.Context, cfg config.Config, agentID string) error {
	return sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE agent_profiles SET status = ? WHERE id = ?`, "active", agentID)
}

func CallTool(ctx context.Context, cfg config.Config, repoRoot, agentID, toolName, target, missionID string, command []string) (ToolResult, error) {
	registry, err := policies.LoadRegistry(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	tool, ok := registry.Tool(toolName)
	if !ok {
		return ToolResult{}, fmt.Errorf("tool %s not found", toolName)
	}
	profile, _, err := Inspect(ctx, cfg, agentID)
	if err != nil {
		return ToolResult{}, err
	}
	if !contains(profile.Capabilities, toolName) {
		return ToolResult{}, fmt.Errorf("agent %s does not have capability %s", agentID, toolName)
	}
	if tool.RequiresRole != "" && profile.Role != tool.RequiresRole {
		return ToolResult{}, fmt.Errorf("agent role %s cannot use %s", profile.Role, toolName)
	}
	if tool.PathWhitelistRequired && target != "" && !allowed(profile.AllowedPaths, target) {
		_ = audit.Write(ctx, cfg.AuditPath, audit.Event{
			AgentID:          agentID,
			Action:           "tool_rejected",
			Target:           target,
			Outcome:          "rejected",
			Tool:             toolName,
			ApprovalRequired: false,
		})
		return ToolResult{}, fmt.Errorf("target %s is outside the path whitelist", target)
	}
	if tool.RequiresApproval {
		approved, requestID, err := approvalState(ctx, cfg.DBPath, agentID, toolName, target)
		if err != nil {
			return ToolResult{}, err
		}
		if !approved {
			if requestID == "" {
				requestID = ids.New("approval")
				now := time.Now().UTC().Format(time.RFC3339)
				if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO approval_requests(id, agent_id, tool_name, target, status, requested_at) VALUES (?, ?, ?, ?, ?, ?)`,
					requestID, agentID, toolName, target, "pending", now); err != nil {
					return ToolResult{}, err
				}
			}
			_ = audit.Write(ctx, cfg.AuditPath, audit.Event{
				AgentID:          agentID,
				Action:           "tool_blocked",
				Target:           target,
				Outcome:          "blocked",
				Tool:             toolName,
				ApprovalRequired: true,
			})
			return ToolResult{Status:"approval_required",RequestID:requestID,ApprovalMsg:"tool requires approval"}, nil
		}
	}
	if toolName == "shell_exec" && len(tool.CommandWhitelist) > 0 {
		commandLine := strings.Join(command, " ")
		matched := false
		for _, allowedCommand := range tool.CommandWhitelist {
			if strings.HasPrefix(commandLine, allowedCommand) {
				matched = true
				break
			}
		}
		if !matched {
			_ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:agentID,Action:"tool_rejected",Target:commandLine,Outcome:"rejected",Tool:toolName,ApprovalRequired:true})
			return ToolResult{}, fmt.Errorf("command %q is outside the command whitelist", commandLine)
		}
	}
	result, err := executeTool(toolName, target, command)
	outcome := "success"
	if err != nil { outcome = "failed" }
	_ = audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:agentID,Action:"tool_call",Target:target,Outcome:outcome,Tool:toolName,ApprovalRequired:tool.RequiresApproval})
	if err != nil { return ToolResult{}, err }
	if (toolName == "write_file" || toolName == "append_file") && target != "" { _ = artifacts.AutoCreateOnWrite(ctx, cfg, repoRoot, agentID, missionID, target) }
	return ToolResult{Status:"ok", Output:result}, nil
}

func Approve(ctx context.Context, cfg config.Config, requestID, decidedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT status FROM approval_requests WHERE id = ?`, requestID)
	if err != nil { return err }
	if len(rows) == 0 { return fmt.Errorf("approval request %s not found", requestID) }
	if str(rows[0]["status"]) != "pending" { return fmt.Errorf("approval request %s is not pending", requestID) }
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `UPDATE approval_requests SET status = ?, decided_at = ?, decided_by = ? WHERE id = ?`, "approved", now, decidedBy, requestID); err != nil { return err }
	return audit.Write(ctx, cfg.AuditPath, audit.Event{AgentID:decidedBy,Action:"approval_granted",Target:requestID,Outcome:"success",ApprovalRequired:false})
}

func writeProfileFile(cfg config.Config, profile Profile) error {
	body, err := json.MarshalIndent(profile, "", "  ")
	if err != nil { return err }
	return os.WriteFile(filepath.Join(cfg.AgentsPath, profile.ID, "profile.json"), body, 0o644)
}
func defaultTeamID(ctx context.Context, dbPath string) (string, error) {
	rows, err := sqlitepy.Query(ctx, dbPath, `SELECT id FROM teams ORDER BY created_at ASC LIMIT 1`)
	if err != nil { return "", err }
	if len(rows) == 0 { return "", fmt.Errorf("no team exists; run uclaw init first") }
	return str(rows[0]["id"]), nil
}
func defaultHandbook(name, role string) string { return fmt.Sprintf(`# Agent Handbook: %s

## Role
%s

## Citation Rules
- Always cite source URLs inline
- Minimum citation completeness: 80%%
- Flag unverifiable claims with [UNVERIFIED]

## Reasoning Style
- Think step by step before acting
- Prefer reversible actions
- Escalate if confidence < 60%%

## Log Format
- Action: <verb> <object>
- Source: <url or path>
- Confidence: <0-100>
- Timestamp: <ISO8601>

## Escalation Rules
- Tool call requires path outside whitelist -> request approval
- Risky action (delete, deploy, send) -> human review
- Conflicting instructions -> flag to lead agent
`, name, role) }
func executeTool(toolName, target string, command []string) (string, error) {
	switch toolName {
	case "read_file": body, err := os.ReadFile(target); return string(body), err
	case "write_file": return "write permitted", os.WriteFile(target, []byte(strings.Join(command, " ")), 0o644)
	case "append_file":
		f, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); if err != nil { return "", err }
		defer f.Close(); _, err = f.WriteString(strings.Join(command, " ")); return "append permitted", err
	case "shell_exec": return strings.Join(command, " "), nil
	default: return toolName + " permitted", nil
	}
}
func approvalState(ctx context.Context, dbPath, agentID, toolName, target string) (bool, string, error) {
	rows, err := sqlitepy.Query(ctx, dbPath, `SELECT id, status FROM approval_requests WHERE agent_id = ? AND tool_name = ? AND target = ? ORDER BY requested_at DESC LIMIT 1`, agentID, toolName, target)
	if err != nil { return false, "", err }
	if len(rows) == 0 { return false, "", nil }
	status := str(rows[0]["status"])
	return status == "approved", str(rows[0]["id"]), nil
}
func allowed(roots []string, target string) bool { cleanTarget := filepath.Clean(target); for _, root := range roots { cleanRoot := filepath.Clean(root); if cleanTarget == cleanRoot || strings.HasPrefix(cleanTarget, cleanRoot+string(os.PathSeparator)) { return true } }; return false }
func contains(values []string, target string) bool { for _, value := range values { if value == target { return true } }; return false }
func str(v interface{}) string { return fmt.Sprintf("%v", v) }
func firstNonEmpty(values ...string) string { for _, value := range values { if strings.TrimSpace(value) != "" { return value } }; return "" }
