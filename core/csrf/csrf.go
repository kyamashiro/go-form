package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateCSRFToken() (string, error) {
	// Generate 32 random bytes
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the token in a URL-safe base64 format
	return base64.URLEncoding.EncodeToString(token), nil
}
