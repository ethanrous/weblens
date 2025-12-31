// Package cryptography provides cryptographic utilities for generating random values and hashing.
package cryptography

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
)

// RandomInt returns a crypto random integer between 0 and limit, inclusive
func RandomInt(limit int64) (int64, error) {
	rInt, err := rand.Int(rand.Reader, big.NewInt(limit))
	if err != nil {
		return 0, err
	}

	return rInt.Int64(), nil
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString generates a cryptographically secure random string of the specified length.
func RandomString(length int64) (string, error) {
	buf := make([]byte, length)

	limit := int64(len(chars))
	for i := range buf {
		num, err := RandomInt(limit)
		if err != nil {
			return "", err
		}

		buf[i] = chars[num]
	}

	return string(buf), nil
}

// RandomBytes generates cryptographically secure random bytes of the specified length.
func RandomBytes(length int64) ([]byte, error) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)

	return buf, err
}

// HashString computes the SHA-256 hash of a string and returns it as a base64-encoded string.
func HashString(stringToHash string) string {
	h := sha256.New()

	h.Write([]byte(stringToHash))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
