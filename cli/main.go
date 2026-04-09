// uclaw — canonical CLI control surface for the UCLAW sovereign engineering habitat.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stizzfer36-del/UCLAW/core/agents"
	"github.com/stizzfer36-del/UCLAW/core/artifacts"
	"github.com/stizzfer36-del/UCLAW/core/ipc"
	"github.com/stizzfer36-del/UCLAW/core/memory"
	"github.com/stizzfer36-del/UCLAW/core/observability"
	"github.com/stizzfer36-del/UCLAW/core/world"
)

const (
	defaultDBPath     = ".uclaw/world.db"
	defaultSocketPath = ".uclaw/uclaw.sock"
	defaultPolicyPath = "core/policies/tools.yaml"
)

var rootCmd = &cobra.Command{
	Use:   "uclaw",
	Short: "UCLAW — Sovereign Engineering Habitat",
	Long:  `UCLAW is the place an engineer opens first. Type uclaw --help to explore.`,
}

func init() {
	rootCmd.AddCommand(worldCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(missionCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(artifactCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(daemonCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ── daemon ──────────────────────────────────────────────────────────────────

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the UCLAW runtime daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		if err := world.EnsureWorld("w1", "default", ".uclaw/vault"); err != nil {
			return err
		}
		if err := memory.InitSchema(); err != nil {
			return err
		}
		if err := agents.LoadPolicies(defaultPolicyPath); err != nil {
			return err
		}
		// Register IPC handlers
		ipc.Register("agents.list", func(_ map[string]string) (interface{}, error) {
			return agents.List(), nil
		})
		ipc.Register("audit.tail", func(p map[string]string) (interface{}, error) {
			return observability.Tail(50)
		})
		ipc.Register("artifacts.list", func(p map[string]string) (interface{}, error) {
			return artifacts.List(p["mission_id"])
		})
		return ipc.ListenAndServe(defaultSocketPath)
	},
}

// ── world ────────────────────────────────────────────────────────────────────

var worldCmd = &cobra.Command{
	Use:   "world",
	Short: "Inspect the world state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		fmt.Println("World state: .uclaw/world.db")
		return nil
	},
}

// ── agent ────────────────────────────────────────────────────────────────────

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List running agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, a := range agents.List() {
			fmt.Printf("  %s\t%s/%s\t%s\n", a.ID, a.Provider, a.Model, a.Role)
		}
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
}

// ── mission ──────────────────────────────────────────────────────────────────

var missionCmd = &cobra.Command{
	Use:   "mission",
	Short: "Manage missions",
}

// ── memory ───────────────────────────────────────────────────────────────────

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Query the knowledge-graph vault",
}

var memorySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Full-text search the memory vault",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		if err := memory.InitSchema(); err != nil {
			return err
		}
		nodes, err := memory.Search(args[0], 20)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			fmt.Printf("[%s] %s: %s\n", n.Kind, n.ID, n.Content[:min(80, len(n.Content))])
		}
		return nil
	},
}

func init() {
	memoryCmd.AddCommand(memorySearchCmd)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── artifact ─────────────────────────────────────────────────────────────────

var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Manage artifacts",
}

// ── audit ─────────────────────────────────────────────────────────────────────

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit trail",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		events, err := observability.Tail(50)
		if err != nil {
			return err
		}
		for _, e := range events {
			fmt.Printf("%s  %-12s  %-20s  %s\n", e.Timestamp, e.AgentID, e.Action, e.Outcome)
		}
		return nil
	},
}
