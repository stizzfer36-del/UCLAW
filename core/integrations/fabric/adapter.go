// Package fabric wraps the Fabric CLI as a UCLAW capability.
// Fabric repo: https://github.com/danielmiessler/fabric
//
// Expected binary: fabric (on PATH)
// Install: go install github.com/danielmiessler/fabric@latest
package fabric

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunPattern executes `fabric --pattern <name>` with content piped to stdin.
func RunPattern(pattern, content string) (string, error) {
	cmd := exec.Command("fabric", "--pattern", pattern)
	cmd.Stdin = strings.NewReader(content)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("fabric: run pattern %q: %w: %s", pattern, err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// ListPatterns returns all available Fabric patterns.
func ListPatterns() ([]string, error) {
	out, err := exec.Command("fabric", "--list").Output()
	if err != nil {
		return nil, fmt.Errorf("fabric: list: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var patterns []string
	for _, l := range lines {
		if l != "" {
			patterns = append(patterns, strings.TrimSpace(l))
		}
	}
	return patterns, nil
}
