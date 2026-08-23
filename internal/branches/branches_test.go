package branches

import (
	"strings"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/tasks"
)

func TestShortHashStable(t *testing.T) {
	a := ShortHash("issues:o/r#1")
	b := ShortHash("issues:o/r#1")
	if a != b || len(a) != 8 {
		t.Fatalf("hash %s %s", a, b)
	}
}

func TestBranchName(t *testing.T) {
	name := BranchName("2026-W34", "issues:o/r#1")
	if !strings.HasPrefix(name, "discovery/2026-W34/") {
		t.Fatalf("name=%s", name)
	}
}

func TestIssueBranchName(t *testing.T) {
	name := IssueBranchName(1, "inverse-targeting")
	if name != "issue/1-inverse-targeting" {
		t.Fatalf("name=%s", name)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"feature request: inverse targeting / exclude", "feature-request-inverse-targeting-exclud"},
		{"Please support something like allow-failure for a given job", "please-support-something-like-allow-fail"},
		{"A method to override configuration and meta arguments within a module", "a-method-to-override-configuration-and-m"},
		{"  spaces  and  spaces  ", "spaces-and-spaces"},
		{"UPPERCASE", "uppercase"},
		{"special!@#$%^&*()chars", "specialchars"},
	}
	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRenderBrief(t *testing.T) {
	md := RenderBrief(tasks.Task{
		Fingerprint: "fp1",
		Title:       "T",
		Pipeline:    "issues",
		Priority:    1.2,
		Score:       10,
		Reward:      2,
		Difficulty:  3,
		Risk:        2,
		AxesSource:  "xai",
		URL:         "https://x",
		Summary:     "s",
		Seen:        "2026-W34",
	}, "## Path 1\nApproach: do thing")
	if !strings.Contains(md, "fingerprint:fp1") || !strings.Contains(md, "Recommended solution paths") {
		t.Fatalf("brief=%s", md)
	}
}
