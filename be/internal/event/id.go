package event

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func newID() string {
	return newPrefixedID("evt")
}

func newRuleID() string {
	return newPrefixedID("rr")
}

func newReminderJobID() string {
	return newPrefixedID("rj")
}

func newPrefixedID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(err)
	}

	id := base64.RawURLEncoding.EncodeToString(bytes[:])
	return prefix + "_" + strings.ToLower(id)
}
