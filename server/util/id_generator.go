package util

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomIdentifier(bits int) (string, error) {
	n := bits / 8
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
