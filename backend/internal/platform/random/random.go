package random

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// String returns a URL-safe base64 encoded random string of the given byte length.
func String(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byte length must be positive")
	}

	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
