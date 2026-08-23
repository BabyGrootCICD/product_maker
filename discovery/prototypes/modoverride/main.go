// Package modoverride implements module configuration override for Terraform.
// It provides a deep merge engine that can override module configuration
// and meta arguments using a declarative override mechanism.
package modoverride

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Config represents a Terraform module configuration.
type Config struct {
	Variables map[string]interface{} `json:"variables"`
	Providers map[string]string      `json:"providers"`
	DependsOn []string               `json:"depends_on"`
	Meta      map[string]interface{} `json:"meta"`
}

// Override represents a configuration override.
type Override struct {
	Variables map[string]interface{} `json:"variables,omitempty"`
	Providers map[string]string      `json:"providers,omitempty"`
	DependsOn []string               `json:"depends_on,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// Merger handles merging of configurations with overrides.
type Merger struct{}

// NewMerger creates a new Merger.
func NewMerger() *Merger {
	return &Merger{}
}

// Merge applies an override to a base configuration and returns the merged result.
func (m *Merger) Merge(base, override Config) Config {
	result := Config{
		Variables: m.mergeMaps(base.Variables, override.Variables),
		Providers: m.mergeProviderMaps(base.Providers, override.Providers),
		DependsOn: m.mergeSlices(base.DependsOn, override.DependsOn),
		Meta:      m.mergeMaps(base.Meta, override.Meta),
	}
	return result
}

// mergeMaps deep merges two maps, with override values taking precedence.
func (m *Merger) mergeMaps(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			// Deep merge if both are maps
			if baseMap, ok := baseVal.(map[string]interface{}); ok {
				if overrideMap, ok := v.(map[string]interface{}); ok {
					result[k] = m.mergeMaps(baseMap, overrideMap)
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

// mergeProviderMaps merges provider maps with override taking precedence.
func (m *Merger) mergeProviderMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// mergeSlices merges slices, appending override values to base.
func (m *Merger) mergeSlices(base, override []string) []string {
	result := make([]string, 0, len(base)+len(override))
	result = append(result, base...)
	for _, v := range override {
		// Avoid duplicates
		found := false
		for _, existing := range result {
			if existing == v {
				found = true
				break
			}
		}
		if !found {
			result = append(result, v)
		}
	}
	return result
}

// Diff shows the differences between two configurations.
type Diff struct {
	Added    map[string]interface{} `json:"added,omitempty"`
	Modified map[string]interface{} `json:"modified,omitempty"`
	Removed  map[string]interface{} `json:"removed,omitempty"`
}

// ComputeDiff computes the differences between base and merged configurations.
func (m *Merger) ComputeDiff(base, merged Config) Diff {
	diff := Diff{
		Added:    make(map[string]interface{}),
		Modified: make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
	}

	// Compare variables
	for k, v := range merged.Variables {
		if baseVal, exists := base.Variables[k]; !exists {
			diff.Added["variables."+k] = v
		} else if !reflect.DeepEqual(baseVal, v) {
			diff.Modified["variables."+k] = map[string]interface{}{
				"old": baseVal,
				"new": v,
			}
		}
	}
	for k := range base.Variables {
		if _, exists := merged.Variables[k]; !exists {
			diff.Removed["variables."+k] = base.Variables[k]
		}
	}

	// Compare providers
	for k, v := range merged.Providers {
		if baseVal, exists := base.Providers[k]; !exists {
			diff.Added["providers."+k] = v
		} else if baseVal != v {
			diff.Modified["providers."+k] = map[string]interface{}{
				"old": baseVal,
				"new": v,
			}
		}
	}

	return diff
}

// Validate checks if an override is compatible with a base configuration.
func (m *Merger) Validate(base, override Config) []string {
	var issues []string

	// Check for type mismatches in variables
	for k, v := range override.Variables {
		if baseVal, exists := base.Variables[k]; exists {
			if reflect.TypeOf(baseVal) != reflect.TypeOf(v) {
				issues = append(issues, fmt.Sprintf("type mismatch for variable %s: %T vs %T", k, baseVal, v))
			}
		}
	}

	// Check for invalid providers
	for k := range override.Providers {
		if _, exists := base.Providers[k]; !exists {
			issues = append(issues, fmt.Sprintf("override references unknown provider: %s", k))
		}
	}

	return issues
}

// String returns a human-readable representation of the diff.
func (d Diff) String() string {
	var sb strings.Builder
	if len(d.Added) > 0 {
		sb.WriteString("Added:\n")
		for k, v := range d.Added {
			sb.WriteString(fmt.Sprintf("  + %s: %v\n", k, v))
		}
	}
	if len(d.Modified) > 0 {
		sb.WriteString("Modified:\n")
		for k, v := range d.Modified {
			sb.WriteString(fmt.Sprintf("  ~ %s: %v\n", k, v))
		}
	}
	if len(d.Removed) > 0 {
		sb.WriteString("Removed:\n")
		for k, v := range d.Removed {
			sb.WriteString(fmt.Sprintf("  - %s: %v\n", k, v))
		}
	}
	if sb.Len() == 0 {
		return "No changes"
	}
	return sb.String()
}

// ToJSON returns the diff as JSON.
func (d Diff) ToJSON() string {
	b, _ := json.MarshalIndent(d, "", "  ")
	return string(b)
}
