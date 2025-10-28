package hashutil

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

// HashArgs encodes the given arguments into a deterministic SHA-256 hash.
// It’s used to generate unique cache keys for function inputs.
//
// If encoding fails (e.g., unsupported type), the function falls back to
// using fmt.Sprintf("%v") to ensure consistent but less unique keys.
func HashArgs(args ...any) string {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(args)
	if err != nil {
		// Fallback: format-based hashing (less reliable)
		return fallbackHash(args...)
	}

	sum := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", sum)
}

// fallbackHash provides a weaker but always-safe hash representation.
func fallbackHash(args ...any) string {
	data := fmt.Sprintf("%#v", args)
	sum := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", sum)
}
