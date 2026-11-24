// Package runnerpain implements analysis of runner pain points from stale issues.
// It provides a pain point detector and analyzer for common GitHub Actions
// runner problems and their solutions.
package runnerpain

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PainPoint represents a common runner pain point.
type PainPoint struct {
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Frequency   int      `json:"frequency"`
	Severity    string   `json:"severity"`
	Solutions   []string `json:"solutions,omitempty"`
}

// Issue represents a GitHub issue for analysis.
type Issue struct {
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Labels   []string `json:"labels"`
	Comments int      `json:"comments"`
}

// Detector detects runner pain points.
type Detector struct{}

// NewDetector creates a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// AnalyzeIssue analyzes an issue for pain points.
func (d *Detector) AnalyzeIssue(issue Issue) []PainPoint {
	var points []PainPoint
	body := strings.ToLower(issue.Body)
	title := strings.ToLower(issue.Title)

	if strings.Contains(body, "slow") || strings.Contains(title, "slow") {
		points = append(points, PainPoint{
			Category:  "Performance",
			Title:     "Slow execution",
			Severity:  "Medium",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "flaky") || strings.Contains(title, "flaky") {
		points = append(points, PainPoint{
			Category:  "Reliability",
			Title:     "Flaky tests",
			Severity:  "High",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "cost") || strings.Contains(title, "cost") {
		points = append(points, PainPoint{
			Category:  "Cost",
			Title:     "High costs",
			Severity:  "Medium",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "secret") || strings.Contains(title, "secret") {
		points = append(points, PainPoint{
			Category:  "Security",
			Title:     "Secret management",
			Severity:  "High",
			Frequency: 1,
		})
	}
	return points
}

// FormatPainPoint formats a pain point as a string.
func (d *Detector) FormatPainPoint(point PainPoint) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s\n", point.Severity, point.Title))
	sb.WriteString(fmt.Sprintf("Category: %s\n", point.Category))
	sb.WriteString(fmt.Sprintf("Frequency: %d\n", point.Frequency))
	if len(point.Solutions) > 0 {
		sb.WriteString("Solutions:\n")
		for _, sol := range point.Solutions {
			sb.WriteString(fmt.Sprintf("  - %s\n", sol))
		}
	}
	return sb.String()
}

// AggregatePoints aggregates pain points by category.
func AggregatePoints(points []PainPoint) map[string]int {
	agg := make(map[string]int)
	for _, p := range points {
		agg[p.Category] += p.Frequency
	}
	return agg
}

// SortByFrequency sorts pain points by frequency.
func SortByFrequency(points []PainPoint) []PainPoint {
	sorted := make([]PainPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Frequency > sorted[j].Frequency
	})
	return sorted
}

// FormatReport formats a report of pain points.
func FormatReport(points []PainPoint) string {
	var sb strings.Builder
	sb.WriteString("Runner Pain Points Report\n")
	sb.WriteString("========================\n\n")
	for _, p := range points {
		sb.WriteString(fmt.Sprintf("- [%s] %s (Frequency: %d)\n", p.Severity, p.Title, p.Frequency))
	}
	return sb.String()
}

// FormatJSON formats pain points as JSON.
func FormatJSON(points []PainPoint) string {
	b, _ := json.MarshalIndent(points, "", "  ")
	return string(b)
}
