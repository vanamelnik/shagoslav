package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const TokenBytes = 8

func Token() (string, error) {
	b := make([]byte, TokenBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
