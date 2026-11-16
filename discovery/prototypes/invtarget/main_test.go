package invtarget

import (
	"testing"
)

func TestNewFilterExact(t *testing.T) {
	f, err := NewFilter([]string{"aws_instance.example", "aws_s3_bucket.data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Count() != 2 {
		t.Fatalf("expected 2 patterns, got %d", f.Count())
	}
	if !f.Matches("aws_instance.example") {
		t.Fatal("expected exact match")
	}
	if f.Matches("aws_instance.other") {
		t.Fatal("should not match different resource")
	}
}

func TestNewFilterGlob(t *testing.T) {
	f, err := NewFilter([]string{"aws_instance.*", "module.legacy.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Matches("aws_instance.example") {
		t.Fatal("expected glob match")
	}
	if !f.Matches("aws_instance.another") {
		t.Fatal("expected glob match for second instance")
	}
	if f.Matches("aws_s3_bucket.data") {
		t.Fatal("should not match different type")
	}
}

func TestNewFilterRegex(t *testing.T) {
	f, err := NewFilter([]string{"/^aws_(instance|ebs)/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Matches("aws_instance.example") {
		t.Fatal("expected regex match")
	}
	if !f.Matches("aws_ebs_volume.data") {
		t.Fatal("expected regex match for ebs")
	}
	if f.Matches("aws_s3_bucket.data") {
		t.Fatal("should not match different type")
	}
}

func TestNewFilterInvalidRegex(t *testing.T) {
	_, err := NewFilter([]string{"/invalid[regex/"})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestFilterResources(t *testing.T) {
	f, err := NewFilter([]string{"aws_instance.example", "module.legacy.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resources := []string{
		"aws_instance.example",
		"aws_instance.other",
		"module.legacy.old",
		"aws_s3_bucket.data",
	}
	result := f.FilterResources(resources)
	if len(result) != 2 {
		t.Fatalf("expected 2 resources, got %d: %v", len(result), result)
	}
	if result[0] != "aws_instance.other" {
		t.Fatalf("expected aws_instance.other, got %s", result[0])
	}
	if result[1] != "aws_s3_bucket.data" {
		t.Fatalf("expected aws_s3_bucket.data, got %s", result[1])
	}
}

func TestApplyFilter(t *testing.T) {
	f, err := NewFilter([]string{"aws_instance.example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plan := Plan{
		Resources: []Resource{
			{Address: "aws_instance.example", Type: "aws_instance", Name: "example"},
			{Address: "aws_instance.other", Type: "aws_instance", Name: "other"},
		},
	}
	result := ApplyFilter(plan, f)
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}
	if result.Resources[0].Address != "aws_instance.other" {
		t.Fatalf("expected aws_instance.other, got %s", result.Resources[0].Address)
	}
}

func TestExcludedCount(t *testing.T) {
	f, err := NewFilter([]string{"aws_instance.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plan := Plan{
		Resources: []Resource{
			{Address: "aws_instance.a"},
			{Address: "aws_instance.b"},
			{Address: "aws_s3_bucket.c"},
		},
	}
	count := ExcludedCount(plan, f)
	if count != 2 {
		t.Fatalf("expected 2 excluded, got %d", count)
	}
}

func TestString(t *testing.T) {
	f, err := NewFilter([]string{"exact.res", "glob.*", "/regex/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := f.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkFilterResources(b *testing.B) {
	f, _ := NewFilter([]string{"aws_instance.*", "module.legacy.*", "/^test_/"})
	resources := make([]string, 1000)
	for i := range resources {
		if i%3 == 0 {
			resources[i] = "aws_instance." + string(rune('a'+i%26))
		} else if i%3 == 1 {
			resources[i] = "module.legacy." + string(rune('a'+i%26))
		} else {
			resources[i] = "test_" + string(rune('a'+i%26))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.FilterResources(resources)
	}
}
