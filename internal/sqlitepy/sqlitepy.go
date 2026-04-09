package sqlitepy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type request struct { SQL string `json:"sql"`; Params []interface{} `json:"params,omitempty"` }
type response struct { Rows []map[string]interface{} `json:"rows"` }
const helper = `
import json
import sqlite3
import sys

mode = sys.argv[1]
db_path = sys.argv[2]
payload = json.loads(sys.stdin.read() or "{}")
conn = sqlite3.connect(db_path)
conn.row_factory = sqlite3.Row
conn.execute("PRAGMA foreign_keys = ON")
try:
    if mode == "exec":
        if payload.get("params"):
            conn.execute(payload["sql"], payload["params"])
        else:
            conn.executescript(payload["sql"])
        conn.commit()
        print("{}")
    elif mode == "query":
        cur = conn.execute(payload["sql"], payload.get("params", []))
        rows = [dict(row) for row in cur.fetchall()]
        print(json.dumps({"rows": rows}))
    else:
        raise ValueError("unknown mode")
finally:
    conn.close()
`
func Exec(ctx context.Context, dbPath, sql string) error { _, err := run(ctx, "exec", dbPath, request{SQL: sql}); return err }
func ExecParams(ctx context.Context, dbPath, sql string, params ...interface{}) error { _, err := run(ctx, "exec", dbPath, request{SQL: sql, Params: params}); return err }
func Query(ctx context.Context, dbPath, sql string, params ...interface{}) ([]map[string]interface{}, error) { body, err := run(ctx, "query", dbPath, request{SQL: sql, Params: params}); if err != nil { return nil, err }; var resp response; if err := json.Unmarshal(body, &resp); err != nil { return nil, err }; return resp.Rows, nil }
func run(ctx context.Context, mode, dbPath string, req request) ([]byte, error) { body, err := json.Marshal(req); if err != nil { return nil, err }; cmd := exec.CommandContext(ctx, "python3", "-c", helper, mode, dbPath); cmd.Stdin = bytes.NewReader(body); var stdout, stderr bytes.Buffer; cmd.Stdout = &stdout; cmd.Stderr = &stderr; if err := cmd.Run(); err != nil { return nil, fmt.Errorf("sqlite helper failed: %w: %s", err, stderr.String()) }; return stdout.Bytes(), nil }
