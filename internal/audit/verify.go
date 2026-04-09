package audit

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func Verify(path string) error {
	f, err := os.Open(path); if err != nil { return err }
	defer f.Close(); var prevLine string; scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text(); var event Event; if err := json.Unmarshal([]byte(line), &event); err != nil { return err }
		expected := ""; if prevLine != "" { sum := sha256.Sum256([]byte(prevLine)); expected = hex.EncodeToString(sum[:]) }
		if event.PrevEventHash != expected { return fmt.Errorf("audit chain mismatch for event %s", event.EventID) }
		prevLine = line
	}
	return scanner.Err()
}
