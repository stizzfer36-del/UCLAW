// Package voltagent wraps the VoltAgent TypeScript agent platform.
// VoltAgent: https://github.com/voltagent/voltagent
//
// This adapter calls a lightweight Node.js runner: core/integrations/voltagent/runner.js
package voltagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var runnerPath string

func init() {
	_, file, _, _ := runtime.Caller(0)
	runnerPath = filepath.Join(filepath.Dir(file), "runner.js")
}

type spawnResp struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// Spawn dispatches a task to VoltAgent and returns the task ID.
func Spawn(task, model string, tools []string) (string, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"action": "spawn",
		"task":   task,
		"model":  model,
		"tools":  tools,
	})
	cmd := exec.Command("node", runnerPath, string(payload))
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("voltagent: spawn: %w: %s", err, errBuf.String())
	}
	var r spawnResp
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &r); err != nil {
		return "", fmt.Errorf("voltagent: decode: %w (out: %s)", err, out.String())
	}
	return r.TaskID, nil
}

// Status returns the current status string for a task.
func Status(taskID string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"action": "status", "task_id": taskID})
	cmd := exec.Command("node", runnerPath, string(payload))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("voltagent: status: %w", err)
	}
	var r struct{ Status string `json:"status"` }
	_ = json.Unmarshal(bytes.TrimSpace(out), &r)
	return strings.TrimSpace(r.Status), nil
}
