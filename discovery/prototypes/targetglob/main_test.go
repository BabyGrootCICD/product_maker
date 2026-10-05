package targetglob

import (
	"testing"
)

func TestParseTargetExact(t *testing.T) {
	target, err := ParseTarget("aws_instance.example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Type != PatternExact {
		t.Fatalf("expected exact, got %d", target.Type)
	}
	if !target.Matches("aws_instance.example") {
		t.Fatal("expected match")
	}
	if target.Matches("aws_instance.other") {
		t.Fatal("should not match")
	}
}

func TestParseTargetGlob(t *testing.T) {
	target, err := ParseTarget("aws_instance.*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Type != PatternGlob {
		t.Fatalf("expected glob, got %d", target.Type)
	}
	if !target.Matches("aws_instance.example") {
		t.Fatal("expected match")
	}
	if !target.Matches("aws_instance.another") {
		t.Fatal("expected match")
	}
	if target.Matches("aws_s3_bucket.data") {
		t.Fatal("should not match")
	}
}

func TestParseTargetRegex(t *testing.T) {
	target, err := ParseTarget("/^aws_(instance|ebs)/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Type != PatternRegex {
		t.Fatalf("expected regex, got %d", target.Type)
	}
	if !target.Matches("aws_instance.example") {
		t.Fatal("expected match")
	}
	if !target.Matches("aws_ebs_volume.data") {
		t.Fatal("expected match")
	}
	if target.Matches("aws_s3_bucket.data") {
		t.Fatal("should not match")
	}
}

func TestParseTargetInvalid(t *testing.T) {
	_, err := ParseTarget("/invalid[regex/")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestMatcher(t *testing.T) {
	matcher, err := NewMatcher([]string{"aws_instance.*", "module.legacy.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matcher.MatchesAny("aws_instance.example") {
		t.Fatal("expected match")
	}
	if !matcher.MatchesAny("module.legacy.old") {
		t.Fatal("expected match")
	}
	if matcher.MatchesAny("aws_s3_bucket.data") {
		t.Fatal("should not match")
	}
}

func TestExpandTargets(t *testing.T) {
	matcher, err := NewMatcher([]string{"aws_instance.*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resources := []string{
		"aws_instance.example",
		"aws_instance.other",
		"aws_s3_bucket.data",
	}
	expanded := matcher.ExpandTargets(resources)
	if len(expanded) != 2 {
		t.Fatalf("expected 2 expanded, got %d", len(expanded))
	}
}

func TestValidateTargets(t *testing.T) {
	issues := ValidateTargets([]string{"aws_instance.*", "/^aws_/"})
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
	issues = ValidateTargets([]string{"/invalid[regex/"})
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
}

func BenchmarkExpandTargets(b *testing.B) {
	matcher, _ := NewMatcher([]string{"aws_instance.*", "module.legacy.*", "/^test_/"})
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
		matcher.ExpandTargets(resources)
	}
}
