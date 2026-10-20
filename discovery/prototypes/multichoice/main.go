// Package multichoice implements multi-choice input type for GitHub Actions workflows.
// It provides an input handler that can handle multiple selections from a list
// of options and return them as a JSON array.
package multichoice

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Input represents a multi-choice input definition.
type Input struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Options      []string `json:"options"`
	Default      []string `json:"default,omitempty"`
	Required     bool     `json:"required"`
	MinSelection int      `json:"min_selection,omitempty"`
	MaxSelection int      `json:"max_selection,omitempty"`
}

// Selection represents a user's selection from a multi-choice input.
type Selection struct {
	InputName string   `json:"input_name"`
	Choices   []string `json:"choices"`
}

// Handler handles multi-choice input selections.
type Handler struct{}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// ValidateSelection validates a user's selection against the input definition.
func (h *Handler) ValidateSelection(input Input, selection Selection) []string {
	var issues []string

	// Check required
	if input.Required && len(selection.Choices) == 0 {
		issues = append(issues, fmt.Sprintf("input %s is required", input.Name))
	}

	// Check min selection
	if input.MinSelection > 0 && len(selection.Choices) < input.MinSelection {
		issues = append(issues, fmt.Sprintf("input %s requires at least %d selections", input.Name, input.MinSelection))
	}

	// Check max selection
	if input.MaxSelection > 0 && len(selection.Choices) > input.MaxSelection {
		issues = append(issues, fmt.Sprintf("input %s allows at most %d selections", input.Name, input.MaxSelection))
	}

	// Check all choices are valid options
	validOptions := make(map[string]bool)
	for _, opt := range input.Options {
		validOptions[opt] = true
	}
	for _, choice := range selection.Choices {
		if !validOptions[choice] {
			issues = append(issues, fmt.Sprintf("invalid choice %q for input %s", choice, input.Name))
		}
	}

	return issues
}

// FormatSelection formats a selection as a JSON array string.
func (h *Handler) FormatSelection(selection Selection) string {
	b, _ := json.Marshal(selection.Choices)
	return string(b)
}

// FormatSelectionWithLabels formats a selection with labels for display.
func (h *Handler) FormatSelectionWithLabels(input Input, selection Selection) string {
	if len(selection.Choices) == 0 {
		return "None"
	}
	return strings.Join(selection.Choices, ", ")
}

// ParseInput parses an input definition from a map.
func ParseInput(configMap map[string]interface{}) Input {
	input := Input{}
	if name, ok := configMap["name"].(string); ok {
		input.Name = name
	}
	if desc, ok := configMap["description"].(string); ok {
		input.Description = desc
	}
	if opts, ok := configMap["options"].([]interface{}); ok {
		for _, opt := range opts {
			if s, ok := opt.(string); ok {
				input.Options = append(input.Options, s)
			}
		}
	}
	if def, ok := configMap["default"].([]interface{}); ok {
		for _, d := range def {
			if s, ok := d.(string); ok {
				input.Default = append(input.Default, s)
			}
		}
	}
	if req, ok := configMap["required"].(bool); ok {
		input.Required = req
	}
	if min, ok := configMap["min_selection"].(float64); ok {
		input.MinSelection = int(min)
	}
	if max, ok := configMap["max_selection"].(float64); ok {
		input.MaxSelection = int(max)
	}
	return input
}

// FormatInput returns a formatted string representation of an input definition.
func FormatInput(input Input) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("name: %s\n", input.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", input.Description))
	sb.WriteString("type: multi-choice\n")
	sb.WriteString("options:\n")
	for _, opt := range input.Options {
		sb.WriteString(fmt.Sprintf("  - %s\n", opt))
	}
	if len(input.Default) > 0 {
		sb.WriteString(fmt.Sprintf("default: %s\n", strings.Join(input.Default, ", ")))
	}
	if input.Required {
		sb.WriteString("required: true\n")
	}
	if input.MinSelection > 0 {
		sb.WriteString(fmt.Sprintf("min_selection: %d\n", input.MinSelection))
	}
	if input.MaxSelection > 0 {
		sb.WriteString(fmt.Sprintf("max_selection: %d\n", input.MaxSelection))
	}
	return sb.String()
}

// GetDefaultSelection returns the default selection for an input.
func GetDefaultSelection(input Input) Selection {
	return Selection{
		InputName: input.Name,
		Choices:   input.Default,
	}
}

// GetSortedOptions returns the options sorted alphabetically.
func GetSortedOptions(input Input) []string {
	sorted := make([]string, len(input.Options))
	copy(sorted, input.Options)
	sort.Strings(sorted)
	return sorted
}
