package world

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/stizzfer36-del/uclaw/internal/config"
	"github.com/stizzfer36-del/uclaw/internal/ids"
	"github.com/stizzfer36-del/uclaw/internal/sqlitepy"
)

type Summary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	VaultPath string `json:"vault_path"`
}

func Init(ctx context.Context, cfg config.Config) (Summary, bool, error) {
	if err := os.MkdirAll(cfg.Home, 0o755); err != nil {
		return Summary{}, false, err
	}
	if err := MigrateUp(ctx, cfg.DBPath); err != nil {
		return Summary{}, false, err
	}

	rows, err := sqlitepy.Query(ctx, cfg.DBPath, `SELECT id, name, created_at, vault_path FROM world LIMIT 1`)
	if err != nil {
		return Summary{}, false, err
	}
	if len(rows) > 0 {
		return mapSummary(rows[0]), false, nil
	}

	summary := Summary{
		ID:        ids.New("world"),
		Name:      cfg.WorldName,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		VaultPath: cfg.VaultPath,
	}
	if err := sqlitepy.ExecParams(ctx, cfg.DBPath, `INSERT INTO world(id, name, created_at, vault_path) VALUES (?, ?, ?, ?)`,
		summary.ID, summary.Name, summary.CreatedAt, summary.VaultPath); err != nil {
		return Summary{}, false, err
	}

	officeID := ids.New("office")
	teamID := ids.New("team")
	memberID := ids.New("member")
	machineID := ids.New("machine")
	roomID := ids.New("room")
	now := summary.CreatedAt

	ops := []struct {
		sql    string
		params []interface{}
	}{
		{`INSERT INTO offices(id, world_id, name, description, created_at) VALUES (?, ?, ?, ?, ?)`, []interface{}{officeID, summary.ID, "Default Office", "Initial local office", now}},
		{`INSERT INTO teams(id, office_id, name, role, lead_agent_id, created_at) VALUES (?, ?, ?, ?, ?, ?)`, []interface{}{teamID, officeID, "Foundry", "lead", "", now}},
		{`INSERT INTO members(id, team_id, name, type, handbook_path, created_at) VALUES (?, ?, ?, ?, ?, ?)`, []interface{}{memberID, teamID, "operator", "human", "", now}},
		{`INSERT INTO machines(id, member_id, hostname, os) VALUES (?, ?, ?, ?)`, []interface{}{machineID, memberID, hostname(), runtimeOS()}},
		{`INSERT INTO rooms(id, machine_id, name, type) VALUES (?, ?, ?, ?)`, []interface{}{roomID, machineID, "Primary Room", "workspace"}},
	}
	for _, op := range ops {
		if err := sqlitepy.ExecParams(ctx, cfg.DBPath, op.sql, op.params...); err != nil {
			return Summary{}, false, err
		}
	}

	return summary, true, nil
}

func Inspect(ctx context.Context, dbPath, nodeID string) (map[string]interface{}, error) {
	if nodeID == "" {
		rows, err := sqlitepy.Query(ctx, dbPath, `SELECT id, name, created_at, vault_path FROM world LIMIT 1`)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, fmt.Errorf("world not initialized")
		}
		return rows[0], nil
	}

	tables := []string{"world", "offices", "teams", "members", "machines", "rooms", "missions"}
	for _, table := range tables {
		rows, err := sqlitepy.Query(ctx, dbPath, fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, table), nodeID)
		if err != nil {
			return nil, err
		}
		if len(rows) > 0 {
			rows[0]["table"] = table
			return rows[0], nil
		}
	}
	return nil, fmt.Errorf("node %s not found", nodeID)
}

func mapSummary(row map[string]interface{}) Summary {
	return Summary{
		ID:        stringValue(row["id"]),
		Name:      stringValue(row["name"]),
		CreatedAt: stringValue(row["created_at"]),
		VaultPath: stringValue(row["vault_path"]),
	}
}

func stringValue(v interface{}) string {
	switch value := v.(type) {
	case string:
		return value
	default:
		return fmt.Sprintf("%v", value)
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return name
}

func runtimeOS() string {
	return runtime.GOOS
}
