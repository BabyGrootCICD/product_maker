package modoverride

import (
	"testing"
)

func TestMergeSimple(t *testing.T) {
	merger := NewMerger()
	base := Config{
		Variables: map[string]interface{}{
			"name": "original",
			"port": 8080,
		},
	}
	override := Config{
		Variables: map[string]interface{}{
			"name": "overridden",
		},
	}
	result := merger.Merge(base, override)
	if result.Variables["name"] != "overridden" {
		t.Fatalf("expected 'overridden', got %v", result.Variables["name"])
	}
	if result.Variables["port"] != 8080 {
		t.Fatalf("expected 8080, got %v", result.Variables["port"])
	}
}

func TestMergeDeep(t *testing.T) {
	merger := NewMerger()
	base := Config{
		Variables: map[string]interface{}{
			"config": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	override := Config{
		Variables: map[string]interface{}{
			"config": map[string]interface{}{
				"key2": "new-value2",
				"key3": "value3",
			},
		},
	}
	result := merger.Merge(base, override)
	config, ok := result.Variables["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config to be a map")
	}
	if config["key1"] != "value1" {
		t.Fatalf("expected 'value1', got %v", config["key1"])
	}
	if config["key2"] != "new-value2" {
		t.Fatalf("expected 'new-value2', got %v", config["key2"])
	}
	if config["key3"] != "value3" {
		t.Fatalf("expected 'value3', got %v", config["key3"])
	}
}

func TestMergeProviders(t *testing.T) {
	merger := NewMerger()
	base := Config{
		Providers: map[string]string{
			"aws": "us-east-1",
			"azurerm": "westus",
		},
	}
	override := Config{
		Providers: map[string]string{
			"aws": "us-west-2",
		},
	}
	result := merger.Merge(base, override)
	if result.Providers["aws"] != "us-west-2" {
		t.Fatalf("expected 'us-west-2', got %v", result.Providers["aws"])
	}
	if result.Providers["azurerm"] != "westus" {
		t.Fatalf("expected 'westus', got %v", result.Providers["azurerm"])
	}
}

func TestMergeDependsOn(t *testing.T) {
	merger := NewMerger()
	base := Config{
		DependsOn: []string{"resource.a", "resource.b"},
	}
	override := Config{
		DependsOn: []string{"resource.b", "resource.c"},
	}
	result := merger.Merge(base, override)
	if len(result.DependsOn) != 3 {
		t.Fatalf("expected 3 depends_on, got %d", len(result.DependsOn))
	}
}

func TestComputeDiff(t *testing.T) {
	merger := NewMerger()
	base := Config{
		Variables: map[string]interface{}{
			"name": "original",
			"port": 8080,
		},
	}
	override := Config{
		Variables: map[string]interface{}{
			"name": "overridden",
			"host": "localhost",
		},
	}
	merged := merger.Merge(base, override)
	diff := merger.ComputeDiff(base, merged)
	if len(diff.Modified) != 1 {
		t.Fatalf("expected 1 modified, got %d", len(diff.Modified))
	}
	if len(diff.Added) != 1 {
		t.Fatalf("expected 1 added, got %d", len(diff.Added))
	}
}

func TestValidate(t *testing.T) {
	merger := NewMerger()
	base := Config{
		Variables: map[string]interface{}{
			"name": "string",
			"port": 8080,
		},
		Providers: map[string]string{
			"aws": "us-east-1",
		},
	}
	override := Config{
		Variables: map[string]interface{}{
			"name": 123, // Type mismatch
		},
		Providers: map[string]string{
			"gcp": "us-central1", // Unknown provider
		},
	}
	issues := merger.Validate(base, override)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %v", len(issues), issues)
	}
}

func TestDiffString(t *testing.T) {
	diff := Diff{
		Added:    map[string]interface{}{"key": "value"},
		Modified: map[string]interface{}{"old": "new"},
	}
	s := diff.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestDiffToJSON(t *testing.T) {
	diff := Diff{
		Added: map[string]interface{}{"key": "value"},
	}
	json := diff.ToJSON()
	if json == "" {
		t.Fatal("expected non-empty JSON")
	}
}

func BenchmarkMerge(b *testing.B) {
	merger := NewMerger()
	base := Config{
		Variables: map[string]interface{}{
			"name": "original",
			"port": 8080,
			"config": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		Providers: map[string]string{
			"aws": "us-east-1",
		},
	}
	override := Config{
		Variables: map[string]interface{}{
			"name": "overridden",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger.Merge(base, override)
	}
}
