// Package selfupdate implements `linkai update`: it checks the latest published
// version, detects how the CLI was installed, and delegates the actual upgrade
// to the matching package manager or install script rather than re-implementing
// binary download/replace/rollback logic.
package selfupdate

import (
	"strconv"
	"strings"
)

// Version is a parsed semantic version (major.minor.patch), ignoring any
// pre-release / build metadata suffix. A nil result means the string could not
// be parsed as a version.
type Version struct {
	Major, Minor, Patch int
}

// ParseVersion parses "1.2.3", "v1.2.3", "1.2.3-rc1" → {1,2,3}. Returns nil on
// a string that has no numeric major component.
func ParseVersion(s string) *Version {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if s == "" {
		return nil
	}
	// Drop pre-release / build metadata (e.g. "-rc1", "+build").
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 {
		return nil
	}
	nums := make([]int, 3)
	for i := 0; i < 3; i++ {
		if i >= len(parts) {
			nums[i] = 0
			continue
		}
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return &Version{Major: nums[0], Minor: nums[1], Patch: nums[2]}
}

// Compare returns -1 if a < b, 0 if equal, +1 if a > b.
func (a *Version) Compare(b *Version) int {
	switch {
	case a.Major != b.Major:
		return sign(a.Major - b.Major)
	case a.Minor != b.Minor:
		return sign(a.Minor - b.Minor)
	default:
		return sign(a.Patch - b.Patch)
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

// IsNewer reports whether latest is a strictly newer version than current.
// Unparseable versions are treated conservatively: if current is "dev" or
// otherwise unparseable, any parseable latest counts as newer.
func IsNewer(latest, current string) bool {
	lv := ParseVersion(latest)
	if lv == nil {
		return false
	}
	cv := ParseVersion(current)
	if cv == nil {
		return true
	}
	return lv.Compare(cv) > 0
}

// Normalize strips a leading "v" so versions from different sources compare and
// display consistently.
func Normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	return strings.TrimPrefix(s, "V")
}
