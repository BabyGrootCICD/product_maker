package backendvars

import (
	"testing"
)

func TestInterpolateVariables(t *testing.T) {
	interpolator := NewInterpolator()
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "${var.bucket}",
			"key":    "${var.key}",
			"region": "us-east-1",
		},
	}
	vars := NewVariableSet()
	vars.Variables["bucket"] = "my-bucket"
	vars.Variables["key"] = "terraform/state"

	result := interpolator.Interpolate(config, *vars)
	if result.Config["bucket"] != "my-bucket" {
		t.Fatalf("expected 'my-bucket', got %v", result.Config["bucket"])
	}
	if result.Config["key"] != "terraform/state" {
		t.Fatalf("expected 'terraform/state', got %v", result.Config["key"])
	}
	if result.Config["region"] != "us-east-1" {
		t.Fatalf("expected 'us-east-1', got %v", result.Config["region"])
	}
}

func TestInterpolateLocals(t *testing.T) {
	interpolator := NewInterpolator()
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "${local.bucket}",
		},
	}
	vars := NewVariableSet()
	vars.Locals["bucket"] = "local-bucket"

	result := interpolator.Interpolate(config, *vars)
	if result.Config["bucket"] != "local-bucket" {
		t.Fatalf("expected 'local-bucket', got %v", result.Config["bucket"])
	}
}

func TestInterpolateNested(t *testing.T) {
	interpolator := NewInterpolator()
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "${var.nested_key}",
			},
		},
	}
	vars := NewVariableSet()
	vars.Variables["nested_key"] = "nested-value"

	result := interpolator.Interpolate(config, *vars)
	nested, ok := result.Config["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested map")
	}
	if nested["key"] != "nested-value" {
		t.Fatalf("expected 'nested-value', got %v", nested["key"])
	}
}

func TestValidate(t *testing.T) {
	interpolator := NewInterpolator()
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "resolved",
			"key":    "${var.unresolved}",
		},
	}
	issues := interpolator.Validate(config)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestParseConfig(t *testing.T) {
	configMap := map[string]interface{}{
		"type":   "s3",
		"bucket": "my-bucket",
		"key":    "state",
	}
	config := ParseConfig(configMap)
	if config.Type != "s3" {
		t.Fatalf("expected 's3', got %s", config.Type)
	}
	if config.Config["bucket"] != "my-bucket" {
		t.Fatalf("expected 'my-bucket', got %v", config.Config["bucket"])
	}
}

func TestFormatConfig(t *testing.T) {
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "my-bucket",
		},
	}
	s := FormatConfig(config)
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkInterpolate(b *testing.B) {
	interpolator := NewInterpolator()
	config := BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "${var.bucket}",
			"key":    "${var.key}",
			"region": "us-east-1",
		},
	}
	vars := NewVariableSet()
	vars.Variables["bucket"] = "my-bucket"
	vars.Variables["key"] = "state"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interpolator.Interpolate(config, *vars)
	}
}
