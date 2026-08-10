package device

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}

	id := base64.RawURLEncoding.EncodeToString(bytes[:])
	return "dev_" + strings.ToLower(id)
}
