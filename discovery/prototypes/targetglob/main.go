// Package targetglob implements glob pattern matching for Terraform resource targets.
// It provides a pattern matcher that can match resource addresses using glob,
// regex, or exact patterns for the -target flag.
package targetglob

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// PatternType represents the type of pattern.
type PatternType int

const (
	// PatternExact is an exact match pattern.
	PatternExact PatternType = iota
	// PatternGlob is a glob pattern with wildcards.
	PatternGlob
	// PatternRegex is a regex pattern (wrapped in /).
	PatternRegex
)

// Target represents a parsed target pattern.
type Target struct {
	Raw     string
	Type    PatternType
	regex   *regexp.Regexp
}

// ParseTarget parses a target string and returns a Target.
func ParseTarget(raw string) (*Target, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty target")
	}

	// Check for regex pattern (wrapped in /)
	if strings.HasPrefix(raw, "/") && strings.HasSuffix(raw, "/") {
		inner := strings.TrimPrefix(strings.TrimSuffix(raw, "/"), "/")
		re, err := regexp.Compile(inner)
		if err != nil {
			return nil, fmt.Errorf("invalid regex %q: %w", raw, err)
		}
		return &Target{Raw: raw, Type: PatternRegex, regex: re}, nil
	}

	// Check for glob pattern (contains *, ?, [)
	if strings.ContainsAny(raw, "*?[") {
		regexStr := globToRegex(raw)
		re, err := regexp.Compile("^" + regexStr + "$")
		if err != nil {
			return nil, fmt.Errorf("invalid glob %q: %w", raw, err)
		}
		return &Target{Raw: raw, Type: PatternGlob, regex: re}, nil
	}

	// Exact match
	return &Target{Raw: raw, Type: PatternExact}, nil
}

// Matches returns true if the target matches the given resource address.
func (t *Target) Matches(address string) bool {
	switch t.Type {
	case PatternExact:
		return address == t.Raw
	case PatternGlob, PatternRegex:
		return t.regex.MatchString(address)
	default:
		return false
	}
}

// String returns a string representation of the target.
func (t *Target) String() string {
	return t.Raw
}

// Matcher handles matching multiple targets against resource addresses.
type Matcher struct {
	targets []*Target
}

// NewMatcher creates a new Matcher from a list of target strings.
func NewMatcher(targets []string) (*Matcher, error) {
	m := &Matcher{}
	for _, raw := range targets {
		t, err := ParseTarget(raw)
		if err != nil {
			return nil, err
		}
		m.targets = append(m.targets, t)
	}
	return m, nil
}

// Match returns all targets that match the given resource address.
func (m *Matcher) Match(address string) []*Target {
	var matched []*Target
	for _, t := range m.targets {
		if t.Matches(address) {
			matched = append(matched, t)
		}
	}
	return matched
}

// MatchesAny returns true if any target matches the address.
func (m *Matcher) MatchesAny(address string) bool {
	return len(m.Match(address)) > 0
}

// ExpandTargets expands all targets against a list of resource addresses
// and returns the matched addresses.
func (m *Matcher) ExpandTargets(resources []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, r := range resources {
		if m.MatchesAny(r) && !seen[r] {
			result = append(result, r)
			seen[r] = true
		}
	}
	sort.Strings(result)
	return result
}

// Count returns the number of targets.
func (m *Matcher) Count() int {
	return len(m.targets)
}

// globToRegex converts a simple glob pattern to a regex string.
func globToRegex(glob string) string {
	var result strings.Builder
	for _, c := range glob {
		switch c {
		case '*':
			result.WriteString(".*")
		case '?':
			result.WriteString(".")
		case '[':
			result.WriteString("[")
		case ']':
			result.WriteString("]")
		case '.':
			result.WriteString("\\.")
		default:
			result.WriteRune(c)
		}
	}
	return result.String()
}

// ValidateTarget checks if a target string is valid.
func ValidateTarget(raw string) error {
	_, err := ParseTarget(raw)
	return err
}

// ValidateTargets checks if multiple target strings are valid.
func ValidateTargets(targets []string) []string {
	var issues []string
	for _, raw := range targets {
		if err := ValidateTarget(raw); err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", raw, err))
		}
	}
	return issues
}
