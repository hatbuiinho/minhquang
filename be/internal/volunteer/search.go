package volunteer

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func normalizeSearchText(value string) string {
	var normalized strings.Builder
	for _, character := range norm.NFD.String(strings.ToLower(value)) {
		if unicode.Is(unicode.Mn, character) {
			continue
		}
		if character == 'đ' {
			character = 'd'
		}
		normalized.WriteRune(character)
	}
	return normalized.String()
}
