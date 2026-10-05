package providerloop

import (
	"testing"
)

func TestExpand(t *testing.T) {
	expander := NewExpander()
	provider := Provider{Name: "aws", Config: map[string]interface{}{"region": "us-east-1"}}
	forEach := map[string]interface{}{
		"us-east": "us-east-1",
		"us-west": "us-west-2",
	}
	instances := expander.Expand(provider, forEach)
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Alias != "aws_us-east" {
		t.Fatalf("expected alias 'aws_us-east', got %s", instances[0].Alias)
	}
}

func TestExpandWithConfig(t *testing.T) {
	expander := NewExpander()
	provider := Provider{Name: "aws", Config: map[string]interface{}{"region": "us-east-1"}}
	configs := map[string]map[string]interface{}{
		"us-east": {"region": "us-east-1", "zone": "a"},
		"us-west": {"region": "us-west-2", "zone": "b"},
	}
	instances := expander.ExpandWithConfig(provider, configs)
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Provider.Config["zone"] != "a" {
		t.Fatalf("expected zone 'a', got %v", instances[0].Provider.Config["zone"])
	}
}

func TestValidate(t *testing.T) {
	expander := NewExpander()
	valid := map[string]interface{}{
		"us-east": "us-east-1",
		"us-west": "us-west-2",
	}
	issues := expander.Validate(valid)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}

	invalid := map[string]interface{}{
		"": "empty-key",
	}
	issues = expander.Validate(invalid)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestComputeDiff(t *testing.T) {
	expander := NewExpander()
	old := []ProviderInstance{
		{Alias: "aws_us-east"},
		{Alias: "aws_us-west"},
	}
	new := []ProviderInstance{
		{Alias: "aws_us-east"},
		{Alias: "aws_eu-west"},
	}
	diff := expander.ComputeDiff(old, new)
	if len(diff.Added) != 1 || diff.Added[0] != "aws_eu-west" {
		t.Fatalf("expected 1 added, got %v", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0] != "aws_us-west" {
		t.Fatalf("expected 1 removed, got %v", diff.Removed)
	}
}

func TestDiffString(t *testing.T) {
	diff := Diff{
		Added:   []string{"aws_eu-west"},
		Removed: []string{"aws_us-west"},
	}
	s := diff.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkExpand(b *testing.B) {
	expander := NewExpander()
	provider := Provider{Name: "aws", Config: map[string]interface{}{"region": "us-east-1"}}
	forEach := map[string]interface{}{
		"us-east": "us-east-1",
		"us-west": "us-west-2",
		"eu-west": "eu-west-1",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expander.Expand(provider, forEach)
	}
}
