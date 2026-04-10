package world

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/stizzfer36-del/UCLAW/internal/sqlitepy"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migration struct {
	Name string
	Up   string
	Down string
}

func LoadMigrations() ([]Migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	grouped := map[string]*Migration{}
	for _, entry := range entries {
		name := entry.Name()
		content, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, err
		}
		key := strings.TrimSuffix(strings.TrimSuffix(name, "_up.sql"), "_down.sql")
		mig := grouped[key]
		if mig == nil {
			mig = &Migration{Name: key}
			grouped[key] = mig
		}
		if strings.HasSuffix(name, "_up.sql") {
			mig.Up = string(content)
		}
		if strings.HasSuffix(name, "_down.sql") {
			mig.Down = string(content)
		}
	}

	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]Migration, 0, len(names))
	for _, name := range names {
		mig := grouped[name]
		if mig.Up == "" || mig.Down == "" {
			return nil, fmt.Errorf("migration %s missing up or down script", name)
		}
		out = append(out, *mig)
	}
	return out, nil
}

func MigrateUp(ctx context.Context, dbPath string) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}
	if err := sqlitepy.Exec(ctx, dbPath, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	name TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`); err != nil {
		return err
	}
	for _, mig := range migrations {
		rows, err := sqlitepy.Query(ctx, dbPath, `SELECT name FROM schema_migrations WHERE name = ?`, mig.Name)
		if err != nil {
			return err
		}
		if len(rows) > 0 {
			continue
		}
		if err := sqlitepy.Exec(ctx, dbPath, mig.Up); err != nil {
			return err
		}
		if err := sqlitepy.ExecParams(ctx, dbPath, `INSERT INTO schema_migrations(name) VALUES (?)`, mig.Name); err != nil {
			return err
		}
	}
	return nil
}

func MigrateDown(ctx context.Context, dbPath string) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}
	sort.SliceStable(migrations, func(i, j int) bool {
		return migrations[i].Name > migrations[j].Name
	})
	for _, mig := range migrations {
		if err := sqlitepy.Exec(ctx, dbPath, mig.Down); err != nil {
			return err
		}
	}
	return sqlitepy.Exec(ctx, dbPath, `DROP TABLE IF EXISTS schema_migrations;`)
}
