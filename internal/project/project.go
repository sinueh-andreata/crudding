package project

import (
	"strings"
)

func ProjectSlug(name string) string {
	var slug strings.Builder
	lastHyphen := false
	for _, character := range strings.ToLower(name) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			slug.WriteRune(character)
			lastHyphen = false
		} else if !lastHyphen && slug.Len() > 0 {
			slug.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(slug.String(), "-")
}