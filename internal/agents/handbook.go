package agents

import (
	"bufio"
	"os"
	"strings"
)

type Handbook struct { Role string; CitationRules []string; ReasoningStyle []string; LogFormat []string; EscalationRules []string }
func LoadHandbook(path string) (Handbook, error) {
	f, err := os.Open(path); if err != nil { return Handbook{}, err }
	defer f.Close()
	var hb Handbook; section := ""; scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") { section = strings.TrimSpace(strings.TrimPrefix(line, "## ")); continue }
		if line == "" { continue }
		switch section {
		case "Role": hb.Role = line
		case "Citation Rules": hb.CitationRules = appendItem(hb.CitationRules, line)
		case "Reasoning Style": hb.ReasoningStyle = appendItem(hb.ReasoningStyle, line)
		case "Log Format": hb.LogFormat = appendItem(hb.LogFormat, line)
		case "Escalation Rules": hb.EscalationRules = appendItem(hb.EscalationRules, line)
		}
	}
	return hb, scanner.Err()
}
func appendItem(dst []string, line string) []string { line = strings.TrimSpace(strings.TrimPrefix(line, "- ")); if line == "" { return dst }; return append(dst, line) }
