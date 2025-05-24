package workflow

import (
	"strconv"
	"strings"
)

func isAsciiLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isAsciiDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isSpaceDashDot(r rune) bool {
	return r == ' ' || r == '-' || r == '.'
}

func SanitizeString(input string) string {
	if len(input) == 0 {
		return ""
	}

	var sb strings.Builder
	isFirstChar := true

	for _, r := range input {
		if isFirstChar {
			isFirstChar = false
			if isAsciiLetter(r) {
				sb.WriteRune(r)
			} else {
				// First char is not an ASCII letter.
				// Prepend "x" to its codepoint. 'x' is an ASCII letter.
				sb.WriteString("x")
				sb.WriteString(strconv.Itoa(int(r)))
			}
		} else {
			if isAsciiLetter(r) || isAsciiDigit(r) || r == '_' {
				sb.WriteRune(r)
			} else if isSpaceDashDot(r) {
				sb.WriteString("_")
			} else {
				// Not an ASCII letter, ASCII digit, or underscore.
				// Prepend "x" to its codepoint.
				sb.WriteString("x")
				sb.WriteString(strconv.Itoa(int(r)))
			}
		}
	}
	return sb.String()
}
