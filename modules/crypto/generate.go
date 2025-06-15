package crypto

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

func RandomBytes(length int64) ([]byte, error) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	return buf, err
}

func HashString(stringToHash string) string {
	h := sha256.New()

	h.Write([]byte(stringToHash))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
