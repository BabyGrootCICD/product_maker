package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
)

// AxesRater rates difficulty/risk for a task.
type AxesRater interface {
	RateTaskAxes(ctx context.Context, title, summary, pipeline, url string, metrics map[string]float64) (xai.TaskAxes, error)
}

// EnrichAxes rates topN tasks with xAI; others keep heuristic axes.
func EnrichAxes(ctx context.Context, list []Task, rater AxesRater, topN int) {
	if topN <= 0 {
		topN = 15
	}
	if rater == nil {
		return
	}
	n := topN
	if n > len(list) {
		n = len(list)
	}
	for i := 0; i < n; i++ {
		t := &list[i]
		axes, err := rater.RateTaskAxes(ctx, t.Title, t.Summary, t.Pipeline, t.URL, t.Metrics)
		if err != nil {
			slog.Warn("xai axes fallback", "fingerprint", t.Fingerprint, "err", err)
			continue
		}
		d, r := ClampAxes(axes.Difficulty, axes.Risk)
		t.Difficulty = d
		t.Risk = r
		t.AxesSource = AxesSourceXAI
		if axes.Rationale != "" {
			t.Rationale = axes.Rationale
		}
	}
	Prioritize(list)
}

// WriteFile merges briefing into path and writes markdown.
func WriteFile(ctx context.Context, path string, briefing domain.Briefing, rater AxesRater, axesTopN int, week string) ([]Task, error) {
	existingMD, _ := os.ReadFile(path)
	existing, notes := Parse(string(existingMD))

	incoming := FromBriefing(briefing, week)
	EnrichAxes(ctx, incoming, rater, axesTopN)
	merged := Merge(existing, incoming, week)

	md := Render(merged, week, notes)
	if err := os.WriteFile(path, []byte(md), 0o644); err != nil {
		return nil, fmt.Errorf("write tasks.md: %w", err)
	}
	return merged, nil
}
