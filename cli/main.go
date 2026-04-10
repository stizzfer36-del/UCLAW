// uclaw — UCLAW sovereign engineering habitat CLI.
// SPEC-1 Phase 1: full WorkPacket state machine, mission journal, checkpoints, audit.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/stizzfer36-del/UCLAW/core/agents"
	"github.com/stizzfer36-del/UCLAW/core/artifacts"
	"github.com/stizzfer36-del/UCLAW/core/audit"
	"github.com/stizzfer36-del/UCLAW/core/checkpoints"
	"github.com/stizzfer36-del/UCLAW/core/ipc"
	"github.com/stizzfer36-del/UCLAW/core/journal"
	"github.com/stizzfer36-del/UCLAW/core/memory"
	"github.com/stizzfer36-del/UCLAW/core/observability"
	"github.com/stizzfer36-del/UCLAW/core/packets"
	"github.com/stizzfer36-del/UCLAW/core/world"
)

const (
	defaultDBPath     = ".uclaw/world.db"
	defaultSocketPath = ".uclaw/uclaw.sock"
	defaultPolicyPath = "core/policies/tools.yaml"
	defaultVaultPath  = ".uclaw/vault"
	defaultCPDir      = ".uclaw/checkpoints"
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
	rootCmd.AddCommand(packetCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(journalCmd)
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// openDB opens the world database and seeds the default hierarchy.
func openDB() error {
	if err := world.Open(defaultDBPath); err != nil {
		return err
	}
	return world.EnsureWorld("w1", "default", defaultVaultPath)
}

// ── daemon ───────────────────────────────────────────────────────────────────

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the UCLAW runtime daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		if err := memory.InitSchema(); err != nil {
			return err
		}
		if err := agents.LoadPolicies(defaultPolicyPath); err != nil {
			return err
		}
		_ = audit.Log("", "", "system", "daemon", "daemon.start", `{}`)
		ipc.Register("agents.list", func(_ map[string]string) (interface{}, error) {
			return agents.List(), nil
		})
		ipc.Register("audit.tail", func(p map[string]string) (interface{}, error) {
			return audit.Tail(50)
		})
		ipc.Register("artifacts.list", func(p map[string]string) (interface{}, error) {
			return artifacts.List(p["mission_id"])
		})
		ipc.Register("packets.list", func(p map[string]string) (interface{}, error) {
			return packets.ListByMission(p["mission_id"])
		})
		fmt.Println("[uclaw] daemon started — listening on", defaultSocketPath)
		return ipc.ListenAndServe(defaultSocketPath)
	},
}

// ── serve (daemon + HTTP web UI) ─────────────────────────────────────────────

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start daemon and serve the web UI at http://localhost:8080",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		if err := memory.InitSchema(); err != nil {
			return err
		}
		_ = audit.Log("", "", "system", "daemon", "serve.start", `{}`)
		if err := agents.LoadPolicies(defaultPolicyPath); err != nil {
			return err
		}
		go func() {
			_ = ipc.ListenAndServe(defaultSocketPath)
		}()
		fmt.Println("[uclaw] serving web UI at http://localhost:8080")
		fmt.Println("[uclaw] open Chrome → http://localhost:8080")
		return serveHTTP(":8080", "desktop/src")
	},
}

// ── world ────────────────────────────────────────────────────────────────────

var worldCmd = &cobra.Command{
	Use:   "world",
	Short: "Inspect the world state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		fmt.Println("World state: .uclaw/world.db")
		// Print mission count
		var count int
		_ = world.DB.QueryRow(`SELECT COUNT(*) FROM missions`).Scan(&count)
		fmt.Printf("Missions: %d\n", count)
		var packetCount int
		_ = world.DB.QueryRow(`SELECT COUNT(*) FROM work_packets`).Scan(&packetCount)
		fmt.Printf("WorkPackets: %d\n", packetCount)
		var auditCount int
		_ = world.DB.QueryRow(`SELECT COUNT(*) FROM audit_events`).Scan(&auditCount)
		fmt.Printf("Audit events: %d\n", auditCount)
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

func init() { agentCmd.AddCommand(agentListCmd) }

// ── mission ──────────────────────────────────────────────────────────────────

var missionCmd = &cobra.Command{
	Use:   "mission",
	Short: "Manage missions",
}

var missionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all missions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		rows, err := world.DB.Query(
			`SELECT id, title, state, created_by, created_at FROM missions ORDER BY created_at DESC`,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		fmt.Printf("%-22s  %-30s  %-12s  %-16s  %s\n", "ID", "TITLE", "STATE", "CREATED_BY", "CREATED_AT")
		for rows.Next() {
			var id, title, state, createdBy, createdAt string
			_ = rows.Scan(&id, &title, &state, &createdBy, &createdAt)
			fmt.Printf("%-22s  %-30s  %-12s  %-16s  %s\n", id, title, state, createdBy, createdAt)
		}
		return rows.Err()
	},
}

var missionCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		title := args[0]
		id := fmt.Sprintf("m-%d", time.Now().UnixMilli())
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := world.DB.Exec(
			`INSERT INTO missions(id,room_id,title,goal,state,created_by,created_at,updated_at)
			 VALUES(?,?,?,?,?,?,?,?)`,
			id, "room-default", title, title, "active", "cli", now, now,
		)
		if err != nil {
			return err
		}
		// Journal the creation event
		_, _ = journal.Append(id, "", "mission.created", fmt.Sprintf(`{"title":%q}`, title), "create-"+id)
		_ = audit.Log("", id, "human", "cli", "mission.created", fmt.Sprintf(`{"title":%q}`, title))
		fmt.Printf("mission created: %s\n", id)
		return nil
	},
}

var missionStatusCmd = &cobra.Command{
	Use:   "status <mission_id>",
	Short: "Show detailed mission status including work packets",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		mid := args[0]
		var title, state, createdAt string
		err := world.DB.QueryRow(`SELECT title,state,created_at FROM missions WHERE id=?`, mid).
			Scan(&title, &state, &createdAt)
		if err != nil {
			return fmt.Errorf("mission %s not found", mid)
		}
		fmt.Printf("Mission: %s\nTitle:   %s\nState:   %s\nCreated: %s\n\n", mid, title, state, createdAt)
		pkts, err := packets.ListByMission(mid)
		if err != nil {
			return err
		}
		if len(pkts) == 0 {
			fmt.Println("No work packets yet. Use: uclaw packet add <mission_id> <kind>")
			return nil
		}
		fmt.Printf("%-24s  %-12s  %-14s  %s\n", "PACKET_ID", "KIND", "STATE", "AGENT")
		for _, p := range pkts {
			fmt.Printf("%-24s  %-12s  %-14s  %s\n", p.ID, p.Kind, p.State, p.AgentID)
		}
		return nil
	},
}

func init() {
	missionCmd.AddCommand(missionListCmd)
	missionCmd.AddCommand(missionCreateCmd)
	missionCmd.AddCommand(missionStatusCmd)
}

// ── packet ───────────────────────────────────────────────────────────────────

var packetCmd = &cobra.Command{
	Use:   "packet",
	Short: "Manage work packets (schedulable units within a mission)",
}

var packetAddCmd = &cobra.Command{
	Use:   "add <mission_id> <kind>",
	Short: "Add a work packet to a mission",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		mid, kind := args[0], args[1]
		key := fmt.Sprintf("%s-%s-%d", mid, kind, time.Now().UnixNano())
		p, err := packets.Create(mid, kind, `{}`, key, 0)
		if err != nil {
			return err
		}
		_, _ = journal.Append(mid, p.ID, "packet.created", fmt.Sprintf(`{"kind":%q}`, kind), "create-"+p.ID)
		_ = audit.Log("", mid, "human", "cli", "packet.created", fmt.Sprintf(`{"packet_id":%q,"kind":%q}`, p.ID, kind))
		fmt.Printf("packet created: %s (state: %s)\n", p.ID, p.State)
		return nil
	},
}

var packetMoveCmd = &cobra.Command{
	Use:   "move <packet_id> <new_state>",
	Short: "Transition a work packet to a new state",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		pid, newState := args[0], args[1]
		p, err := packets.Get(pid)
		if err != nil {
			return err
		}
		oldState := p.State
		// Journal before side effect (no-turns-missed rule)
		_, _ = journal.Append(p.MissionID, pid, "packet.transition",
			fmt.Sprintf(`{"from":%q,"to":%q}`, oldState, newState),
			fmt.Sprintf("transition-%s-%s", pid, newState))
		if err := packets.Transition(pid, newState); err != nil {
			return err
		}
		_ = audit.Log("", p.MissionID, "human", "cli", "packet.transition",
			fmt.Sprintf(`{"packet_id":%q,"from":%q,"to":%q}`, pid, oldState, newState))
		fmt.Printf("packet %s: %s → %s\n", pid, oldState, newState)
		return nil
	},
}

var packetListCmd = &cobra.Command{
	Use:   "list <mission_id>",
	Short: "List all work packets for a mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		pkts, err := packets.ListByMission(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%-24s  %-12s  %-14s  %-4s  %s\n", "ID", "KIND", "STATE", "PRI", "AGENT")
		for _, p := range pkts {
			fmt.Printf("%-24s  %-12s  %-14s  %-4d  %s\n", p.ID, p.Kind, p.State, p.Priority, p.AgentID)
		}
		return nil
	},
}

func init() {
	packetCmd.AddCommand(packetAddCmd)
	packetCmd.AddCommand(packetMoveCmd)
	packetCmd.AddCommand(packetListCmd)
}

// ── checkpoint ───────────────────────────────────────────────────────────────

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Manage mission checkpoints (snapshots for rollback/replay)",
}

var checkpointCreateCmd = &cobra.Command{
	Use:   "create <mission_id> <reason>",
	Short: "Create a checkpoint for a mission",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		mid, reason := args[0], args[1]
		// Get latest journal offset
		var offset int64
		_ = world.DB.QueryRow(`SELECT COALESCE(MAX(offset),0) FROM mission_journal WHERE mission_id=?`, mid).Scan(&offset)
		cp, err := checkpoints.Create(mid, reason, offset, defaultCPDir)
		if err != nil {
			return err
		}
		_ = audit.Log("", mid, "human", "cli", "checkpoint.created",
			fmt.Sprintf(`{"checkpoint_id":%q,"journal_offset":%d}`, cp.ID, cp.JournalOffset))
		fmt.Printf("checkpoint created: %s (journal_offset=%d)\n", cp.ID, cp.JournalOffset)
		return nil
	},
}

var checkpointListCmd = &cobra.Command{
	Use:   "list <mission_id>",
	Short: "List checkpoints for a mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		cps, err := checkpoints.List(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%-22s  %-8s  %-25s  %s\n", "ID", "OFFSET", "CREATED_AT", "REASON")
		for _, c := range cps {
			fmt.Printf("%-22s  %-8d  %-25s  %s\n", c.ID, c.JournalOffset, c.CreatedAt, c.Reason)
		}
		return nil
	},
}

func init() {
	checkpointCmd.AddCommand(checkpointCreateCmd)
	checkpointCmd.AddCommand(checkpointListCmd)
}

// ── journal ───────────────────────────────────────────────────────────────────

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Inspect the mission journal (append-only event log)",
}

var journalTailCmd = &cobra.Command{
	Use:   "tail <mission_id> [after_offset]",
	Short: "Show journal entries for a mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		mid := args[0]
		var after int64
		if len(args) > 1 {
			after, _ = strconv.ParseInt(args[1], 10, 64)
		}
		entries, err := journal.Since(mid, after)
		if err != nil {
			return err
		}
		fmt.Printf("%-6s  %-20s  %-18s  %s\n", "OFF", "EVENT_TYPE", "PACKET_ID", "CREATED_AT")
		for _, e := range entries {
			fmt.Printf("%-6d  %-20s  %-18s  %s\n", e.Offset, e.EventType, e.PacketID, e.CreatedAt)
		}
		return nil
	},
}

func init() { journalCmd.AddCommand(journalTailCmd) }

// ── memory ────────────────────────────────────────────────────────────────────

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Query the knowledge-graph vault",
}

var memorySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Full-text search the memory vault",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
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

func init() { memoryCmd.AddCommand(memorySearchCmd) }

// ── artifact ──────────────────────────────────────────────────────────────────

var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Manage artifacts",
}

var artifactListCmd = &cobra.Command{
	Use:   "list <mission_id>",
	Short: "List artifacts for a mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
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

func init() { artifactCmd.AddCommand(artifactListCmd) }

// ── audit ─────────────────────────────────────────────────────────────────────

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit trail",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openDB(); err != nil {
			return err
		}
		events, err := audit.Tail(50)
		if err != nil {
			return err
		}
		fmt.Printf("%-6s  %-26s  %-12s  %-16s  %s\n", "SEQ", "CREATED_AT", "ACTOR", "EVENT_TYPE", "MISSION")
		for _, e := range events {
			fmt.Printf("%-6d  %-26s  %-12s  %-16s  %s\n", e.Seq, e.CreatedAt, e.ActorID, e.EventType, e.MissionID)
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
		if err := openDB(); err != nil {
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
		_ = audit.Log("", "", "human", "cli", "approval.approved",
			fmt.Sprintf(`{"request_id":%q}`, args[0]))
		fmt.Printf("approved: %s\n", args[0])
		return nil
	},
}
