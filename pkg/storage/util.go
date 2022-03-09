package storage

import (
	"crypto/sha1"
	"encoding/hex"
)

func hash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))

	return hex.EncodeToString(h.Sum(nil))
}
