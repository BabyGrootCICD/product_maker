// Package backendvars implements variable interpolation for Terraform backend configuration.
// It provides an interpolator that can resolve variable references in backend
// configuration blocks during initialization.
package backendvars

import (
	"fmt"
	"regexp"
	"strings"
)

// BackendConfig represents a Terraform backend configuration.
type BackendConfig struct {
	Type   string
	Config map[string]interface{}
}

// VariableSet represents a set of variables for interpolation.
type VariableSet struct {
	Variables map[string]interface{}
	Locals    map[string]interface{}
	Terraform map[string]interface{}
}

// NewVariableSet creates a new VariableSet.
func NewVariableSet() *VariableSet {
	return &VariableSet{
		Variables: make(map[string]interface{}),
		Locals:    make(map[string]interface{}),
		Terraform: make(map[string]interface{}),
	}
}

// Interpolator handles variable interpolation in backend configurations.
type Interpolator struct {
	varRe *regexp.Regexp
}

// NewInterpolator creates a new Interpolator.
func NewInterpolator() *Interpolator {
	return &Interpolator{
		varRe: regexp.MustCompile(`\$\{(var|local|terraform)\.([a-zA-Z_][a-zA-Z0-9_]*)\}`),
	}
}

// Interpolate resolves variable references in a backend configuration.
func (i *Interpolator) Interpolate(config BackendConfig, vars VariableSet) BackendConfig {
	result := BackendConfig{
		Type:   config.Type,
		Config: make(map[string]interface{}),
	}
	for k, v := range config.Config {
		result.Config[k] = i.interpolateValue(v, vars)
	}
	return result
}

// interpolateValue resolves variable references in a value.
func (i *Interpolator) interpolateValue(value interface{}, vars VariableSet) interface{} {
	switch v := value.(type) {
	case string:
		return i.interpolateString(v, vars)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = i.interpolateValue(val, vars)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for idx, val := range v {
			result[idx] = i.interpolateValue(val, vars)
		}
		return result
	default:
		return value
	}
}

// interpolateString resolves variable references in a string.
func (i *Interpolator) interpolateString(s string, vars VariableSet) string {
	return i.varRe.ReplaceAllStringFunc(s, func(match string) string {
		// Remove ${ and } from match
		inner := match[2 : len(match)-1]
		parts := strings.SplitN(inner, ".", 2)
		if len(parts) != 2 {
			return match
		}
		prefix, name := parts[0], parts[1]
		var val interface{}
		switch prefix {
		case "var":
			val = vars.Variables[name]
		case "local":
			val = vars.Locals[name]
		case "terraform":
			val = vars.Terraform[name]
		}
		if val == nil {
			return match
		}
		return fmt.Sprintf("%v", val)
	})
}

// Validate checks if a backend configuration has unresolved references.
func (i *Interpolator) Validate(config BackendConfig) []string {
	var issues []string
	for k, v := range config.Config {
		if s, ok := v.(string); ok {
			if i.varRe.MatchString(s) {
				issues = append(issues, fmt.Sprintf("unresolved reference in %s: %s", k, s))
			}
		}
	}
	return issues
}

// ParseConfig parses a backend configuration from a map.
func ParseConfig(configMap map[string]interface{}) BackendConfig {
	config := BackendConfig{
		Config: make(map[string]interface{}),
	}
	if t, ok := configMap["type"].(string); ok {
		config.Type = t
	}
	for k, v := range configMap {
		if k != "type" {
			config.Config[k] = v
		}
	}
	return config
}

// FormatConfig returns a formatted string representation of a backend configuration.
func FormatConfig(config BackendConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("backend %q {\n", config.Type))
	for k, v := range config.Config {
		sb.WriteString(fmt.Sprintf("  %s = %v\n", k, v))
	}
	sb.WriteString("}")
	return sb.String()
}
