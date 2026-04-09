// policy.go — tool-call gate enforcing tools.yaml rules.
package agents

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ToolEntry mirrors a single entry in core/policies/tools.yaml.
type ToolEntry struct {
	Name                string   `yaml:"name"`
	Description         string   `yaml:"description"`
	RiskLevel           string   `yaml:"risk_level"`
	RequiresApproval    bool     `yaml:"requires_approval"`
	RequiresHuman       bool     `yaml:"requires_human"`
	PathWhitelistRequired bool   `yaml:"path_whitelist_required"`
	CommandWhitelist    []string `yaml:"command_whitelist"`
	RequiresRole        string   `yaml:"requires_role"`
	AuditAlways         bool     `yaml:"audit_always"`
}

type policyFile struct {
	Tools []ToolEntry `yaml:"tools"`
}

var toolRegistry map[string]ToolEntry

// LoadPolicies reads core/policies/tools.yaml into memory.
func LoadPolicies(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("policy: read %s: %w", path, err)
	}
	var pf policyFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return fmt.Errorf("policy: parse: %w", err)
	}
	toolRegistry = make(map[string]ToolEntry, len(pf.Tools))
	for _, t := range pf.Tools {
		toolRegistry[t.Name] = t
	}
	return nil
}

// CheckTool returns an error if agent is not allowed to call toolName.
func CheckTool(agent *Agent, toolName string, args map[string]string) error {
	t, ok := toolRegistry[toolName]
	if !ok {
		return fmt.Errorf("policy: unknown tool %q", toolName)
	}
	if t.RequiresRole != "" && string(agent.Role) != t.RequiresRole {
		return fmt.Errorf("policy: tool %q requires role %s, agent has %s", toolName, t.RequiresRole, agent.Role)
	}
	if t.RequiresApproval {
		return fmt.Errorf("policy: tool %q requires human approval — run: uclaw approve %s", toolName, toolName)
	}
	if t.CommandWhitelist != nil {
		cmd, ok := args["command"]
		if !ok {
			return fmt.Errorf("policy: tool %q requires 'command' arg", toolName)
		}
		allowed := false
		for _, w := range t.CommandWhitelist {
			if strings.HasPrefix(cmd, w) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("policy: command %q not whitelisted for tool %q", cmd, toolName)
		}
	}
	return nil
}
