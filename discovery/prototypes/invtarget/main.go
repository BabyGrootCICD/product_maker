// Package invtarget implements inverse targeting (exclude) for Terraform resources.
// It provides a pattern matcher and filter engine that can exclude resources
// from Terraform plan/apply operations using glob, regex, or exact matching.
package invtarget

import (
	"fmt"
	"regexp"
	"strings"
)

// Filter represents an exclusion filter for Terraform resources.
type Filter struct {
	exact   []string
	glob    []string
	regex   []*regexp.Regexp
}

// NewFilter creates a new Filter from a list of exclusion patterns.
// Patterns are classified automatically:
//   - Exact match: "aws_instance.example"
//   - Glob match: "aws_instance.*", "module.legacy.*"
//   - Regex match: "/^aws_(instance|ebs)/"
func NewFilter(patterns []string) (*Filter, error) {
	f := &Filter{}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "/") && strings.HasSuffix(p, "/") {
			// Regex pattern
			inner := strings.TrimPrefix(strings.TrimSuffix(p, "/"), "/")
			re, err := regexp.Compile(inner)
			if err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", p, err)
			}
			f.regex = append(f.regex, re)
		} else if strings.ContainsAny(p, "*?[") {
			// Glob pattern - convert to regex
			regexStr := globToRegex(p)
			re, err := regexp.Compile("^" + regexStr + "$")
			if err != nil {
				return nil, fmt.Errorf("invalid glob %q: %w", p, err)
			}
			f.glob = append(f.glob, regexStr)
			f.regex = append(f.regex, re)
		} else {
			// Exact match
			f.exact = append(f.exact, p)
		}
	}
	return f, nil
}

// Matches returns true if the given resource address matches any exclusion pattern.
func (f *Filter) Matches(address string) bool {
	// Check exact matches
	for _, e := range f.exact {
		if address == e {
			return true
		}
	}
	// Check glob/regex matches
	for _, re := range f.regex {
		if re.MatchString(address) {
			return true
		}
	}
	return false
}

// FilterResources returns only resources that are NOT excluded.
func (f *Filter) FilterResources(resources []string) []string {
	var result []string
	for _, r := range resources {
		if !f.Matches(r) {
			result = append(result, r)
		}
	}
	return result
}

// String returns a human-readable description of the filter.
func (f *Filter) String() string {
	parts := []string{}
	if len(f.exact) > 0 {
		parts = append(parts, fmt.Sprintf("exact=%v", f.exact))
	}
	if len(f.glob) > 0 {
		parts = append(parts, fmt.Sprintf("glob=%v", f.glob))
	}
	if len(f.regex) > 0 {
		parts = append(parts, fmt.Sprintf("regex=%d patterns", len(f.regex)))
	}
	return fmt.Sprintf("Filter{%s}", strings.Join(parts, ", "))
}

// Count returns the total number of exclusion patterns.
func (f *Filter) Count() int {
	return len(f.exact) + len(f.glob) + len(f.regex)
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

// Plan represents a Terraform plan with resources to potentially exclude.
type Plan struct {
	Resources []Resource
}

// Resource represents a single Terraform resource in a plan.
type Resource struct {
	Address string
	Type    string
	Name    string
	Module  string
}

// ApplyFilter applies an exclusion filter to a Plan and returns a new Plan
// with excluded resources removed.
func ApplyFilter(plan Plan, filter *Filter) Plan {
	result := Plan{}
	for _, r := range plan.Resources {
		if !filter.Matches(r.Address) {
			result.Resources = append(result.Resources, r)
		}
	}
	return result
}

// ExcludedCount returns the number of resources that would be excluded.
func ExcludedCount(plan Plan, filter *Filter) int {
	count := 0
	for _, r := range plan.Resources {
		if filter.Matches(r.Address) {
			count++
		}
	}
	return count
}
