// Package pccalert implements analysis of PCC alert system stale issues.
// It provides a problem detector and analyzer for common PCC alert
// problems and their solutions.
package pccalert

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Problem represents a common PCC alert problem.
type Problem struct {
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

// Detector detects PCC alert problems.
type Detector struct{}

// NewDetector creates a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// AnalyzeIssue analyzes an issue for problems.
func (d *Detector) AnalyzeIssue(issue Issue) []Problem {
	var problems []Problem
	body := strings.ToLower(issue.Body)
	title := strings.ToLower(issue.Title)

	if strings.Contains(body, "fatigue") || strings.Contains(title, "fatigue") {
		problems = append(problems, Problem{
			Category:  "Alert Fatigue",
			Title:     "Too many alerts",
			Severity:  "Medium",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "false positive") || strings.Contains(title, "false positive") {
		problems = append(problems, Problem{
			Category:  "False Positives",
			Title:     "Inaccurate alerts",
			Severity:  "High",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "missing") || strings.Contains(title, "missing") {
		problems = append(problems, Problem{
			Category:  "Missing Alerts",
			Title:     "Undetected problems",
			Severity:  "High",
			Frequency: 1,
		})
	}
	if strings.Contains(body, "notification") || strings.Contains(title, "notification") {
		problems = append(problems, Problem{
			Category:  "Notification Issues",
			Title:     "Alert delivery problems",
			Severity:  "Medium",
			Frequency: 1,
		})
	}
	return problems
}

// FormatProblem formats a problem as a string.
func (d *Detector) FormatProblem(problem Problem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s\n", problem.Severity, problem.Title))
	sb.WriteString(fmt.Sprintf("Category: %s\n", problem.Category))
	sb.WriteString(fmt.Sprintf("Frequency: %d\n", problem.Frequency))
	if len(problem.Solutions) > 0 {
		sb.WriteString("Solutions:\n")
		for _, sol := range problem.Solutions {
			sb.WriteString(fmt.Sprintf("  - %s\n", sol))
		}
	}
	return sb.String()
}

// AggregateProblems aggregates problems by category.
func AggregateProblems(problems []Problem) map[string]int {
	agg := make(map[string]int)
	for _, p := range problems {
		agg[p.Category] += p.Frequency
	}
	return agg
}

// SortByFrequency sorts problems by frequency.
func SortByFrequency(problems []Problem) []Problem {
	sorted := make([]Problem, len(problems))
	copy(sorted, problems)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Frequency > sorted[j].Frequency
	})
	return sorted
}

// FormatReport formats a report of problems.
func FormatReport(problems []Problem) string {
	var sb strings.Builder
	sb.WriteString("PCC Alert Problems Report\n")
	sb.WriteString("========================\n\n")
	for _, p := range problems {
		sb.WriteString(fmt.Sprintf("- [%s] %s (Frequency: %d)\n", p.Severity, p.Title, p.Frequency))
	}
	return sb.String()
}

// FormatJSON formats problems as JSON.
func FormatJSON(problems []Problem) string {
	b, _ := json.MarshalIndent(problems, "", "  ")
	return string(b)
}
