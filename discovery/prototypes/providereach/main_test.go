package providereach

import (
	"testing"
)

func TestMapProviders(t *testing.T) {
	mapper := NewProviderMapper()
	instances := []ModuleInstance{
		{Key: "us-east", Config: map[string]interface{}{}},
		{Key: "us-west", Config: map[string]interface{}{}},
	}
	providerMap := map[string]Provider{
		"aws": {Name: "aws", Alias: "us-east-1"},
	}
	result := mapper.MapProviders(instances, providerMap)
	if len(result) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(result))
	}
	if result[0].Providers["aws"].Alias != "us-east-1" {
		t.Fatalf("expected alias 'us-east-1', got %s", result[0].Providers["aws"].Alias)
	}
}

func TestMapProvidersDynamic(t *testing.T) {
	mapper := NewProviderMapper()
	instances := []ModuleInstance{
		{Key: "us-east", Config: map[string]interface{}{}},
		{Key: "us-west", Config: map[string]interface{}{}},
	}
	aliases := map[string]string{
		"us-east": "us-east-1",
		"us-west": "us-west-2",
	}
	result := mapper.MapProvidersDynamic(instances, "aws", aliases)
	if result[0].Providers["aws"].Alias != "us-east-1" {
		t.Fatalf("expected alias 'us-east-1', got %s", result[0].Providers["aws"].Alias)
	}
	if result[1].Providers["aws"].Alias != "us-west-2" {
		t.Fatalf("expected alias 'us-west-2', got %s", result[1].Providers["aws"].Alias)
	}
}

func TestMapProvidersFromTemplate(t *testing.T) {
	mapper := NewProviderMapper()
	instances := []ModuleInstance{
		{Key: "us-east", Config: map[string]interface{}{}},
		{Key: "us-west", Config: map[string]interface{}{}},
	}
	template := map[string]string{
		"aws": "aws_{each.key}",
	}
	result := mapper.MapProvidersFromTemplate(instances, template)
	if result[0].Providers["aws"].Alias != "aws_us-east" {
		t.Fatalf("expected alias 'aws_us-east', got %s", result[0].Providers["aws"].Alias)
	}
}

func TestValidate(t *testing.T) {
	mapper := NewProviderMapper()
	instances := []ModuleInstance{
		{Key: "valid", Providers: map[string]Provider{"aws": {Name: "aws"}}},
		{Key: "invalid", Providers: map[string]Provider{"aws": {Name: ""}}},
	}
	issues := mapper.Validate(instances)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func TestComputeDiff(t *testing.T) {
	mapper := NewProviderMapper()
	old := []ModuleInstance{{Key: "us-east"}, {Key: "us-west"}}
	new := []ModuleInstance{{Key: "us-east"}, {Key: "eu-west"}}
	diff := mapper.ComputeDiff(old, new)
	if len(diff.Added) != 1 || diff.Added[0] != "eu-west" {
		t.Fatalf("expected 1 added, got %v", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0] != "us-west" {
		t.Fatalf("expected 1 removed, got %v", diff.Removed)
	}
}

func TestDiffString(t *testing.T) {
	diff := Diff{
		Added:   []string{"eu-west"},
		Removed: []string{"us-west"},
	}
	s := diff.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkMapProviders(b *testing.B) {
	mapper := NewProviderMapper()
	instances := make([]ModuleInstance, 100)
	for i := range instances {
		instances[i] = ModuleInstance{Key: "inst" + string(rune('a'+i%26)), Config: map[string]interface{}{}}
	}
	providerMap := map[string]Provider{"aws": {Name: "aws", Alias: "us-east-1"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.MapProviders(instances, providerMap)
	}
}
