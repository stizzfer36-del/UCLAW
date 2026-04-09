package world

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stizzfer36-del/uclaw/internal/sqlitepy"
)

func TestMigrateUpDownClean(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "world.db")
	ctx := context.Background()

	if err := MigrateUp(ctx, dbPath); err != nil {
		t.Fatal(err)
	}
	rows, err := sqlitepy.Query(ctx, dbPath, `SELECT name FROM sqlite_master WHERE type='table' AND name='world'`)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected world table after up migration, got %d", len(rows))
	}

	if err := MigrateDown(ctx, dbPath); err != nil {
		t.Fatal(err)
	}
	rows, err = sqlitepy.Query(ctx, dbPath, `SELECT name FROM sqlite_master WHERE type='table' AND name='world'`)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected world table removed after down migration, got %d", len(rows))
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatal(err)
	}
}
