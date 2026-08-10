package account

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func newPrefixedID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}

	id := base64.RawURLEncoding.EncodeToString(bytes[:])
	return prefix + "_" + strings.ToLower(id)
}
