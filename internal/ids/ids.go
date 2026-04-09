package ids

import (
	"crypto/rand"
	"encoding/hex"
)

func New(prefix string) string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}
