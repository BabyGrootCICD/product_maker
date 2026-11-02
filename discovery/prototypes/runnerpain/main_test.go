package runnerpain

import (
	"testing"
)

func TestAnalyzeIssueSlow(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "Workflow is slow",
		Body:  "My workflow takes too long to run",
	}
	points := detector.AnalyzeIssue(issue)
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Category != "Performance" {
		t.Fatalf("expected 'Performance', got %s", points[0].Category)
	}
}

func TestAnalyzeIssueFlaky(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "Tests are flaky",
		Body:  "Intermittent test failures",
	}
	points := detector.AnalyzeIssue(issue)
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Category != "Reliability" {
		t.Fatalf("expected 'Reliability', got %s", points[0].Category)
	}
}

func TestAnalyzeIssueMultiple(t *testing.T) {
	detector := NewDetector()
	issue := Issue{
		Title: "Slow and flaky",
		Body:  "Workflow is slow and tests are flaky",
	}
	points := detector.AnalyzeIssue(issue)
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
}

func TestFormatPainPoint(t *testing.T) {
	detector := NewDetector()
	point := PainPoint{
		Category:  "Performance",
		Title:     "Slow execution",
		Severity:  "Medium",
		Frequency: 5,
	}
	result := detector.FormatPainPoint(point)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestAggregatePoints(t *testing.T) {
	points := []PainPoint{
		{Category: "Performance", Frequency: 3},
		{Category: "Performance", Frequency: 2},
		{Category: "Reliability", Frequency: 1},
	}
	agg := AggregatePoints(points)
	if agg["Performance"] != 5 {
		t.Fatalf("expected 5, got %d", agg["Performance"])
	}
	if agg["Reliability"] != 1 {
		t.Fatalf("expected 1, got %d", agg["Reliability"])
	}
}

func TestSortByFrequency(t *testing.T) {
	points := []PainPoint{
		{Title: "Low", Frequency: 1},
		{Title: "High", Frequency: 10},
		{Title: "Medium", Frequency: 5},
	}
	sorted := SortByFrequency(points)
	if sorted[0].Title != "High" {
		t.Fatalf("expected 'High' first, got %s", sorted[0].Title)
	}
}

func TestFormatReport(t *testing.T) {
	points := []PainPoint{
		{Title: "Test", Severity: "High", Frequency: 5},
	}
	result := FormatReport(points)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestFormatJSON(t *testing.T) {
	points := []PainPoint{
		{Title: "Test"},
	}
	result := FormatJSON(points)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkAnalyzeIssue(b *testing.B) {
	detector := NewDetector()
	issue := Issue{
		Title: "Slow workflow",
		Body:  "My workflow is slow and flaky",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.AnalyzeIssue(issue)
	}
}
