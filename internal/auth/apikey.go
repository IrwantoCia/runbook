package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateAPIKey() (string, string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}
	key := "rb_" + hex.EncodeToString(bytes)
	return key, HashToken(key), key[:11], nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
