// Package providereach implements provider forwarding to modules in for_each for Terraform.
// It provides a provider mapper that can assign providers to module instances
// when using for_each syntax.
package providereach

import (
	"fmt"
	"sort"
	"strings"
)

// Provider represents a Terraform provider reference.
type Provider struct {
	Name  string
	Alias string
}

// ModuleInstance represents a single module instance created from for_each.
type ModuleInstance struct {
	Key       string
	Config    map[string]interface{}
	Providers map[string]Provider
}

// ProviderMapper handles mapping providers to module instances.
type ProviderMapper struct{}

// NewProviderMapper creates a new ProviderMapper.
func NewProviderMapper() *ProviderMapper {
	return &ProviderMapper{}
}

// MapProviders maps providers to module instances based on a provider map.
func (m *ProviderMapper) MapProviders(instances []ModuleInstance, providerMap map[string]Provider) []ModuleInstance {
	result := make([]ModuleInstance, len(instances))
	copy(result, instances)

	for i := range result {
		result[i].Providers = make(map[string]Provider)
		for name, provider := range providerMap {
			result[i].Providers[name] = provider
		}
	}
	return result
}

// MapProvidersDynamic maps providers dynamically based on instance key.
func (m *ProviderMapper) MapProvidersDynamic(instances []ModuleInstance, providerPrefix string, providerAliases map[string]string) []ModuleInstance {
	result := make([]ModuleInstance, len(instances))
	copy(result, instances)

	for i := range result {
		result[i].Providers = make(map[string]Provider)
		alias, exists := providerAliases[result[i].Key]
		if !exists {
			alias = result[i].Key
		}
		result[i].Providers[providerPrefix] = Provider{
			Name:  providerPrefix,
			Alias: alias,
		}
	}
	return result
}

// MapProvidersFromTemplate maps providers using a template with each.key placeholder.
func (m *ProviderMapper) MapProvidersFromTemplate(instances []ModuleInstance, template map[string]string) []ModuleInstance {
	result := make([]ModuleInstance, len(instances))
	copy(result, instances)

	for i := range result {
		result[i].Providers = make(map[string]Provider)
		for name, aliasTemplate := range template {
			alias := strings.ReplaceAll(aliasTemplate, "{each.key}", result[i].Key)
			result[i].Providers[name] = Provider{
				Name:  name,
				Alias: alias,
			}
		}
	}
	return result
}

// Validate checks if provider assignments are valid.
func (m *ProviderMapper) Validate(instances []ModuleInstance) []string {
	var issues []string
	for _, inst := range instances {
		for name, provider := range inst.Providers {
			if provider.Name == "" {
				issues = append(issues, fmt.Sprintf("instance %s: provider %s has empty name", inst.Key, name))
			}
		}
	}
	return issues
}

// Diff shows the differences between two sets of module instances.
type Diff struct {
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
	Modified []string `json:"modified,omitempty"`
}

// ComputeDiff computes the differences between two sets of module instances.
func (m *ProviderMapper) ComputeDiff(old, new []ModuleInstance) Diff {
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)
	for _, inst := range old {
		oldMap[inst.Key] = true
	}
	for _, inst := range new {
		newMap[inst.Key] = true
	}

	diff := Diff{}
	for key := range newMap {
		if !oldMap[key] {
			diff.Added = append(diff.Added, key)
		}
	}
	for key := range oldMap {
		if !newMap[key] {
			diff.Removed = append(diff.Removed, key)
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
