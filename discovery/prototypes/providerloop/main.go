// Package providerloop implements provider instantiation with for_each for Terraform.
// It provides a provider expander that can create multiple provider instances
// from a configuration map using for_each syntax.
package providerloop

import (
	"fmt"
	"sort"
	"strings"
)

// Provider represents a Terraform provider configuration.
type Provider struct {
	Name   string
	Config map[string]interface{}
}

// ProviderInstance represents a single provider instance created from for_each.
type ProviderInstance struct {
	Alias      string
	Provider   Provider
	EachKey    string
	EachValue  interface{}
}

// Expander handles expanding provider configurations with for_each.
type Expander struct{}

// NewExpander creates a new Expander.
func NewExpander() *Expander {
	return &Expander{}
}

// Expand takes a provider with a for_each map and returns expanded instances.
func (e *Expander) Expand(provider Provider, forEach map[string]interface{}) []ProviderInstance {
	var instances []ProviderInstance
	for key, value := range forEach {
		instance := ProviderInstance{
			Alias:     fmt.Sprintf("%s_%s", provider.Name, key),
			Provider:  provider,
			EachKey:   key,
			EachValue: value,
		}
		instances = append(instances, instance)
	}
	// Sort by alias for deterministic output
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Alias < instances[j].Alias
	})
	return instances
}

// ExpandWithConfig expands a provider with per-instance configuration overrides.
func (e *Expander) ExpandWithConfig(provider Provider, configs map[string]map[string]interface{}) []ProviderInstance {
	var instances []ProviderInstance
	for key, config := range configs {
		mergedConfig := make(map[string]interface{})
		for k, v := range provider.Config {
			mergedConfig[k] = v
		}
		for k, v := range config {
			mergedConfig[k] = v
		}
		instance := ProviderInstance{
			Alias:     fmt.Sprintf("%s_%s", provider.Name, key),
			Provider:  Provider{Name: provider.Name, Config: mergedConfig},
			EachKey:   key,
			EachValue: config,
		}
		instances = append(instances, instance)
	}
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Alias < instances[j].Alias
	})
	return instances
}

// Validate checks if a for_each map is valid for provider expansion.
func (e *Expander) Validate(forEach map[string]interface{}) []string {
	var issues []string
	if len(forEach) == 0 {
		issues = append(issues, "for_each map is empty")
	}
	for key := range forEach {
		if key == "" {
			issues = append(issues, "for_each contains empty key")
		}
		if strings.Contains(key, " ") {
			issues = append(issues, fmt.Sprintf("for_each key contains space: %s", key))
		}
	}
	return issues
}

// Diff shows the differences between two sets of provider instances.
type Diff struct {
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
	Modified []string `json:"modified,omitempty"`
}

// ComputeDiff computes the differences between two sets of provider instances.
func (e *Expander) ComputeDiff(old, new []ProviderInstance) Diff {
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)
	for _, p := range old {
		oldMap[p.Alias] = true
	}
	for _, p := range new {
		newMap[p.Alias] = true
	}

	diff := Diff{}
	for alias := range newMap {
		if !oldMap[alias] {
			diff.Added = append(diff.Added, alias)
		}
	}
	for alias := range oldMap {
		if !newMap[alias] {
			diff.Removed = append(diff.Removed, alias)
		}
	}

	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	sort.Strings(diff.Modified)
	return diff
}

// String returns a human-readable representation of the diff.
func (d Diff) String() string {
	var sb strings.Builder
	if len(d.Added) > 0 {
		sb.WriteString("Added:\n")
		for _, a := range d.Added {
			sb.WriteString(fmt.Sprintf("  + %s\n", a))
		}
	}
	if len(d.Removed) > 0 {
		sb.WriteString("Removed:\n")
		for _, r := range d.Removed {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}
	if len(d.Modified) > 0 {
		sb.WriteString("Modified:\n")
		for _, m := range d.Modified {
			sb.WriteString(fmt.Sprintf("  ~ %s\n", m))
		}
	}
	if sb.Len() == 0 {
		return "No changes"
	}
	return sb.String()
}
