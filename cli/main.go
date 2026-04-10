// uclaw — canonical CLI control surface for the UCLAW sovereign engineering habitat.
package main

import (
	"fmt"
	"os"
	"time"

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
	rootCmd.AddCommand(approveCmd)
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
		ipc.Register("agents.list", func(_ map[string]string) (interface{}, error) {
			return agents.List(), nil
		})
		ipc.Register("audit.tail", func(p map[string]string) (interface{}, error) {
			return observability.Tail(50)
		})
		ipc.Register("artifacts.list", func(p map[string]string) (interface{}, error) {
			return artifacts.List(p["mission_id"])
		})
		fmt.Println("[uclaw] daemon started — listening on", defaultSocketPath)
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
			fmt.Printf("  %-16s  %-10s  %-20s  %s\n", a.ID, a.Role, a.Provider+"/"+a.Model, a.Mission)
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

var missionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all missions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		rows, err := world.DB.Query(
			`SELECT id, title, status, created_by, created_at FROM missions ORDER BY created_at DESC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		fmt.Printf("%-20s  %-30s  %-10s  %-16s  %s\n", "ID", "TITLE", "STATUS", "CREATED_BY", "CREATED_AT")
		for rows.Next() {
			var id, title, status, createdBy, createdAt string
			_ = rows.Scan(&id, &title, &status, &createdBy, &createdAt)
			fmt.Printf("%-20s  %-30s  %-10s  %-16s  %s\n", id, title, status, createdBy, createdAt)
		}
		return rows.Err()
	},
}

var missionCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		title := args[0]
		id := fmt.Sprintf("m-%d", time.Now().UnixMilli())
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := world.DB.Exec(
			`INSERT INTO missions(id,room_id,title,status,created_by,created_at,updated_at)
			 VALUES(?,?,?,?,?,?,?)`,
			id, "room-default", title, "active", "cli", now, now,
		)
		if err != nil {
			return err
		}
		fmt.Printf("mission created: %s\n", id)
		return nil
	},
}

func init() {
	missionCmd.AddCommand(missionListCmd)
	missionCmd.AddCommand(missionCreateCmd)
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
			preview := n.Content
			if len(preview) > 80 {
				preview = preview[:80]
			}
			fmt.Printf("[%s] %s: %s\n", n.Kind, n.ID, preview)
		}
		return nil
	},
}

func init() {
	memoryCmd.AddCommand(memorySearchCmd)
}

// ── artifact ─────────────────────────────────────────────────────────────────

var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Manage artifacts",
}

var artifactListCmd = &cobra.Command{
	Use:   "list <mission_id>",
	Short: "List artifacts for a mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		list, err := artifacts.List(args[0])
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("no artifacts found for mission", args[0])
			return nil
		}
		fmt.Printf("%-20s  %-20s  %-12s  %-16s  %s\n", "ID", "TITLE", "STATUS", "TRUST", "PATH")
		for _, a := range list {
			fmt.Printf("%-20s  %-20s  %-12s  %-16s  %s\n", a.ID, a.Title, a.VerificationStatus, a.TrustLevel, a.Path)
		}
		return nil
	},
}

var artifactVerifyCmd = &cobra.Command{
	Use:   "verify <artifact_id>",
	Short: "Mark an artifact as verified",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		if err := artifacts.SetStatus(args[0], "verified"); err != nil {
			return err
		}
		fmt.Printf("artifact %s marked verified\n", args[0])
		return nil
	},
}

func init() {
	artifactCmd.AddCommand(artifactListCmd)
	artifactCmd.AddCommand(artifactVerifyCmd)
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

// ── approve ───────────────────────────────────────────────────────────────────

var approveCmd = &cobra.Command{
	Use:   "approve <request_id>",
	Short: "Approve a pending tool-call request",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := world.Open(defaultDBPath); err != nil {
			return err
		}
		now := time.Now().UTC().Format(time.RFC3339)
		res, err := world.DB.Exec(
			`UPDATE approval_requests SET status='approved', decided_at=?, decided_by='cli' WHERE id=? AND status='pending'`,
			now, args[0],
		)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("request %s not found or already decided", args[0])
		}
		fmt.Printf("approved: %s\n", args[0])
		return nil
	},
}
