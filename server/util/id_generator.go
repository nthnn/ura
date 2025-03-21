package util

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func GenerateRandomIdentifier(bits int) (string, error) {
	if bits <= 0 {
		return "", errors.New("bits must be greater than zero")
	}

	n := (bits + 7) / 8
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
