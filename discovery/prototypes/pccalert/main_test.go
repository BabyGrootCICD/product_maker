package pccalert

import (
	"testing"
)

func TestAnalyzeIssueFatigue(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "Alert fatigue",
		Body:  "Too many alerts firing",
	}
	problems := detector.AnalyzeIssue(issue)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(problems))
	}
	if problems[0].Category != "Alert Fatigue" {
		t.Fatalf("expected 'Alert Fatigue', got %s", problems[0].Category)
	}
}

func TestAnalyzeIssueFalsePositive(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "False positive alerts",
		Body:  "Getting alerts for non-issues",
	}
	problems := detector.AnalyzeIssue(issue)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(problems))
	}
	if problems[0].Category != "False Positives" {
		t.Fatalf("expected 'False Positives', got %s", problems[0].Category)
	}
}

func TestAnalyzeIssueMultiple(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "Fatigue and false positives",
		Body:  "Too many alerts and false positives",
	}
	problems := detector.AnalyzeIssue(issue)
	if len(problems) != 2 {
		t.Fatalf("expected 2 problems, got %d", len(problems))
	}
}

func TestFormatProblem(t *testing.T) {
	detector := NewDetector()
	problem := Problem{
		Category:  "Alert Fatigue",
		Title:     "Too many alerts",
		Severity:  "Medium",
		Frequency: 5,
	}
	result := detector.FormatProblem(problem)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestAggregateProblems(t *testing.T) {
	problems := []Problem{
		{Category: "Alert Fatigue", Frequency: 3},
		{Category: "Alert Fatigue", Frequency: 2},
		{Category: "False Positives", Frequency: 1},
	}
	agg := AggregateProblems(problems)
	if agg["Alert Fatigue"] != 5 {
		t.Fatalf("expected 5, got %d", agg["Alert Fatigue"])
	}
	if agg["False Positives"] != 1 {
		t.Fatalf("expected 1, got %d", agg["False Positives"])
	}
}

func TestSortByFrequency(t *testing.T) {
	problems := []Problem{
		{Title: "Low", Frequency: 1},
		{Title: "High", Frequency: 10},
		{Title: "Medium", Frequency: 5},
	}
	sorted := SortByFrequency(problems)
	if sorted[0].Title != "High" {
		t.Fatalf("expected 'High' first, got %s", sorted[0].Title)
	}
}

func TestFormatReport(t *testing.T) {
	problems := []Problem{
		{Title: "Test", Severity: "High", Frequency: 5},
	}
	result := FormatReport(problems)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestFormatJSON(t *testing.T) {
	problems := []Problem{
		{Title: "Test"},
	}
	result := FormatJSON(problems)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkAnalyzeIssue(b *testing.B) {
	detector := NewDetector()
	issue := Issue{
		Title: "Fatigue and false positives",
		Body:  "Too many alerts and false positives",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.AnalyzeIssue(issue)
	}
}
