package policies

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Tool struct { Name string; Description string; RiskLevel string; RequiresApproval bool; PathWhitelistRequired bool; CommandWhitelist []string; RequiresRole string; RequiresHuman bool; AuditAlways bool }
type Registry struct { Tools map[string]Tool }
func LoadRegistry(root string) (Registry, error) {
	path := filepath.Join(root, "core", "policies", "tools.yaml"); f, err := os.Open(path); if err != nil { return Registry{}, err }; defer f.Close(); registry := Registry{Tools: map[string]Tool{}}; scanner := bufio.NewScanner(f); var current *Tool; var inCommands bool
	for scanner.Scan() { line := scanner.Text(); trimmed := strings.TrimSpace(line); if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "tools:" { continue }; if strings.HasPrefix(trimmed, "- name:") { if current != nil && current.Name != "" { registry.Tools[current.Name] = *current }; current = &Tool{Name: strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))}; inCommands = false; continue }; if current == nil { continue }; if trimmed == "command_whitelist:" { inCommands = true; continue }; if inCommands && strings.HasPrefix(trimmed, "- ") { current.CommandWhitelist = append(current.CommandWhitelist, strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"`)); continue }; inCommands = false; key, value, ok := strings.Cut(trimmed, ":"); if !ok { continue }; key = strings.TrimSpace(key); value = strings.Trim(strings.TrimSpace(value), `"`); switch key { case "description": current.Description = value; case "risk_level": current.RiskLevel = value; case "requires_approval": current.RequiresApproval = mustBool(value); case "path_whitelist_required": current.PathWhitelistRequired = mustBool(value); case "requires_role": current.RequiresRole = value; case "requires_human": current.RequiresHuman = mustBool(value); case "audit_always": current.AuditAlways = mustBool(value) } }
	if err := scanner.Err(); err != nil { return Registry{}, err }; if current != nil && current.Name != "" { registry.Tools[current.Name] = *current }; if len(registry.Tools) == 0 { return Registry{}, errors.New("no tools loaded from policy registry") }; return registry, nil }
func (r Registry) Tool(name string) (Tool, bool) { tool, ok := r.Tools[name]; return tool, ok }
func mustBool(v string) bool { value, _ := strconv.ParseBool(v); return value }
