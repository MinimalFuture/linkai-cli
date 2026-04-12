// Package validate provides input validation and output sanitization utilities.
package validate

import (
	"fmt"
	"strings"
	"unicode"
)

// RejectControlChars returns an error if s contains C0 control characters
// (except \n, \r, \t) or dangerous Unicode characters (Bidi overrides, zero-width).
func RejectControlChars(field, s string) error {
	for i, r := range s {
		if r != '\n' && r != '\r' && r != '\t' && unicode.Is(unicode.Cc, r) {
			return fmt.Errorf("%s contains control character at position %d (U+%04X)", field, i, r)
		}
		if isDangerousUnicode(r) {
			return fmt.Errorf("%s contains dangerous Unicode character at position %d (U+%04X)", field, i, r)
		}
	}
	return nil
}

// SanitizeOutput strips ANSI escape sequences, C0 control characters
// (except \n, \t), and dangerous Unicode from terminal output.
func SanitizeOutput(s string) string {
	s = stripANSI(s)

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case unicode.Is(unicode.Cc, r):
			// strip C0 control chars
		case isDangerousUnicode(r):
			// strip Bidi overrides, zero-width, etc.
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isDangerousUnicode checks for Bidi overrides, zero-width, and line/paragraph separators.
func isDangerousUnicode(r rune) bool {
	switch {
	case r >= 0x202A && r <= 0x202E: // Bidi embedding/override
		return true
	case r >= 0x2066 && r <= 0x2069: // Bidi isolate
		return true
	case r >= 0x200B && r <= 0x200D: // zero-width space/joiner
		return true
	case r == 0xFEFF: // BOM / zero-width no-break space
		return true
	case r == 0x2028 || r == 0x2029: // line/paragraph separator
		return true
	}
	return false
}

// stripANSI removes ANSI escape sequences (\x1b[...m, \x1b[...H, etc.).
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			// skip ESC [ ... <final byte>
			j := i + 2
			for j < len(s) && s[j] >= 0x20 && s[j] <= 0x3F {
				j++
			}
			if j < len(s) && s[j] >= 0x40 && s[j] <= 0x7E {
				j++ // skip final byte
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
