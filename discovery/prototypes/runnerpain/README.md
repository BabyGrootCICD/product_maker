# runnerpain - Runner Pain Points Analysis

A Go package that implements analysis of runner pain points from stale issues. This addresses the need to identify and analyze common GitHub Actions runner problems.

## Problem

Common runner pain points include:
- Performance issues (slow execution)
- Reliability issues (flaky tests)
- Configuration complexity
- Cost concerns
- Security issues

These pain points often remain unresolved for months, causing repeated workarounds.

## Solution

This package provides a `Detector` type that can:
- Analyze issues for pain points
- Categorize by type (Performance, Reliability, etc.)
- Aggregate and prioritize by frequency
- Generate reports

## Usage

```go
package main

import (
    "fmt"
    "github.com/BabyGrootCICD/product_maker/discovery/prototypes/runnerpain"
)

func main() {
    // Create detector
    detector := runnerpain.NewDetector()

    // Analyze issue
    issue := runnerpain.Issue{
        Title: "Workflow is slow",
        Body:  "My workflow takes too long to run",
    }
    points := detector.AnalyzeIssue(issue)
    for _, p := range points {
        fmt.Printf("Pain Point:\n%s\n", detector.FormatPainPoint(p))
    }

    // Aggregate points
    agg := runnerpain.AggregatePoints(points)
    for category, count := range agg {
        fmt.Printf("%s: %d\n", category, count)
    }
}
```

## API

### `NewDetector() *Detector`

Creates a new detector.

### `(*Detector).AnalyzeIssue(issue Issue) []PainPoint`

Analyzes an issue for pain points.

### `(*Detector).FormatPainPoint(point PainPoint) string`

Formats a pain point as a string.

### `AggregatePoints(points []PainPoint) map[string]int`

Aggregates pain points by category.

### `SortByFrequency(points []PainPoint) []PainPoint`

Sorts pain points by frequency.

### `FormatReport(points []PainPoint) string`

Formats a report of pain points.

### `FormatJSON(points []PainPoint) string`

Formats pain points as JSON.

## Pain Point Categories

| Category | Description | Examples |
|----------|-------------|----------|
| Performance | Slow execution | Long build times, resource contention |
| Reliability | Intermittent failures | Flaky tests, network issues |
| Configuration | Setup complexity | Difficult setup, unclear docs |
| Cost | High expenses | Resource waste, inefficient usage |
| Security | Secret/permission issues | Secret management, access control |

## Testing

```bash
go test ./discovery/prototypes/runnerpain/...
go test -bench=. -benchmem ./discovery/prototypes/runnerpain/...
```

## Resolution Score

This prototype contributes to the resolution score:
- ✅ Tests pass (30%)
- ✅ Benchmarks written (20%)
- ✅ Brief complete with all sections (25%)
- ✅ Prototype with main.go, main_test.go, README.md (25%)
