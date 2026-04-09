// Package praisonai wraps the PraisonAI multi-agent framework.
// PraisonAI: https://github.com/MervinPraison/PraisonAI
//
// Expected: pip install praisonai  →  praisonai binary on PATH
package praisonai

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunCrew executes a PraisonAI crew YAML file and returns output.
func RunCrew(crewYAMLPath string) (string, error) {
	cmd := exec.Command("praisonai", "--file", crewYAMLPath)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("praisonai: run %s: %w: %s", crewYAMLPath, err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}
