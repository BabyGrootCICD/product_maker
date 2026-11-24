package multichoice

import (
	"testing"
)

func TestValidateSelectionValid(t *testing.T) {
	handler := NewHandler()
	input := Input{
		Name:    "targets",
		Options: []string{"production", "staging", "development"},
		Required: true,
	}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"production", "staging"},
	}
	issues := handler.ValidateSelection(input, selection)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

func TestValidateSelectionRequired(t *testing.T) {
	handler := NewHandler()
	input := Input{
		Name:     "targets",
		Options:  []string{"production", "staging"},
		Required: true,
	}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{},
	}
	issues := handler.ValidateSelection(input, selection)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestValidateSelectionInvalidChoice(t *testing.T) {
	handler := NewHandler()
	input := Input{
		Name:    "targets",
		Options: []string{"production", "staging"},
	}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"invalid"},
	}
	issues := handler.ValidateSelection(input, selection)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestValidateSelectionMinMax(t *testing.T) {
	handler := NewHandler()
	input := Input{
		Name:         "targets",
		Options:      []string{"a", "b", "c", "d"},
		MinSelection: 2,
		MaxSelection: 3,
	}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"a"},
	}
	issues := handler.ValidateSelection(input, selection)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestFormatSelection(t *testing.T) {
	handler := NewHandler()
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"production", "staging"},
	}
	result := handler.FormatSelection(selection)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestFormatSelectionWithLabels(t *testing.T) {
	handler := NewHandler()
	input := Input{Name: "targets"}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"production", "staging"},
	}
	result := handler.FormatSelectionWithLabels(input, selection)
	if result != "production, staging" {
		t.Fatalf("expected 'production, staging', got %s", result)
	}
}

func TestParseInput(t *testing.T) {
	configMap := map[string]interface{}{
		"name":        "targets",
		"description": "Select targets",
		"options":     []interface{}{"production", "staging"},
		"required":    true,
	}
	input := ParseInput(configMap)
	if input.Name != "targets" {
		t.Fatalf("expected 'targets', got %s", input.Name)
	}
	if len(input.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(input.Options))
	}
}

func TestFormatInput(t *testing.T) {
	input := Input{
		Name:        "targets",
		Description: "Select targets",
		Options:     []string{"production", "staging"},
		Required:    true,
	}
	result := FormatInput(input)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestGetDefaultSelection(t *testing.T) {
	input := Input{
		Name:    "targets",
		Default: []string{"staging"},
	}
	selection := GetDefaultSelection(input)
	if len(selection.Choices) != 1 || selection.Choices[0] != "staging" {
		t.Fatalf("expected [staging], got %v", selection.Choices)
	}
}

func TestGetSortedOptions(t *testing.T) {
	input := Input{
		Options: []string{"c", "a", "b"},
	}
	sorted := GetSortedOptions(input)
	if sorted[0] != "a" || sorted[1] != "b" || sorted[2] != "c" {
		t.Fatalf("expected [a b c], got %v", sorted)
	}
}

func BenchmarkValidateSelection(b *testing.B) {
	handler := NewHandler()
	input := Input{
		Name:    "targets",
		Options: []string{"a", "b", "c", "d", "e"},
	}
	selection := Selection{
		InputName: "targets",
		Choices:   []string{"a", "b", "c"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ValidateSelection(input, selection)
	}
}
