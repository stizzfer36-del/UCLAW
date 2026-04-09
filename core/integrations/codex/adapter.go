// Package codex wraps the OpenAI Codex CLI for local code editing.
// Codex CLI: https://github.com/openai/codex
//
// Expected binary: codex (on PATH)
// Install: npm install -g @openai/codex
package codex

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Edit applies a natural-language instruction to a file.
func Edit(filePath, instruction string) (string, error) {
	cmd := exec.Command("codex", "edit", filePath, "--instruction", instruction, "--approval-mode", "full-auto")
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("codex: edit %s: %w: %s", filePath, err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// Build runs codex build in the given directory.
func Build(dir string) (string, error) {
	cmd := exec.Command("codex", "build")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("codex: build in %s: %w: %s", dir, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
