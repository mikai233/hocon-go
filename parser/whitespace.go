package parser

import (
	"unicode"
)

func isWhitespace(r rune) bool {
	if r == '\uFEFF' {
		return true
	}
	return unicode.IsSpace(r)
}

func isHorizontalWhitespace(r rune) bool {
	if r == '\n' || r == '\r' {
		return false
	}
	return isWhitespace(r)
}
