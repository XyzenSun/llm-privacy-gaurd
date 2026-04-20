package repository

import (
	"crypto/rand"
	"encoding/hex"
)

func newSimpleID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
