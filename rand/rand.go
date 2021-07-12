package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const TokenBytes = 16

func Token() string {
	b := make([]byte, TokenBytes)
	rand.Read(b) // NB: there's no error checking...
	return base64.URLEncoding.EncodeToString(b)
}
