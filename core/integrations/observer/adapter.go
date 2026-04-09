// Package observer wraps the Observer local multi-agent orchestrator.
// Observer: local orchestration for local models with screen-watching and actions.
// Ref: https://github.com/jina-ai/node-DeepResearch (similar pattern)
//
// This adapter calls Observer's REST API (default: http://localhost:7878).
package observer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config holds Observer server details.
type Config struct {
	BaseURL string // e.g. http://localhost:7878
}

type taskReq struct {
	Instruction string `json:"instruction"`
	Model       string `json:"model"`
}

type taskResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// RunTask sends an instruction to Observer and returns the task ID.
func (c *Config) RunTask(instruction, model string) (string, error) {
	body, _ := json.Marshal(taskReq{Instruction: instruction, Model: model})
	resp, err := http.Post(c.BaseURL+"/task", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("observer: run task: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var r taskResp
	_ = json.Unmarshal(raw, &r)
	return r.ID, nil
}
