// Package agent_framework adapts the Microsoft AutoGen / Agent Framework.
// AutoGen: https://github.com/microsoft/autogen
//
// This adapter calls a Python runner that uses the autogen SDK.
// Install: pip install autogen-agentchat autogen-ext
package agent_framework

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
	runnerPath = filepath.Join(filepath.Dir(file), "runner.py")
}

// RunOrchestration executes an AutoGen group-chat orchestration.
func RunOrchestration(task, model string, agents []string) (string, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"task":   task,
		"model":  model,
		"agents": agents,
	})
	cmd := exec.Command("python3", runnerPath, string(payload))
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("agent_framework: %w: %s", err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}
