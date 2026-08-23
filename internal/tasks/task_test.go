package tasks

import (
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

func TestFromBriefingOverlapPreferred(t *testing.T) {
	b := domain.Briefing{
		RankedOpenData: []domain.Insight{{
			Fingerprint: "opendata:pcc:a", Theme: "gov-procurement", Score: 40, Summary: "od",
			Pipeline: domain.PipelineOpenData, Kind: domain.KindOpportunityAlert,
		}},
		RankedIssues: []domain.Insight{{
			Fingerprint: "issues:o/r#1", Theme: "gov-procurement", Score: 30, Title: "Pain",
			URL: "https://example.com/1", Summary: "is", Pipeline: domain.PipelineIssues, Kind: domain.KindPainPoint,
		}},
		Overlaps: []domain.Overlap{{
			OpenData: domain.Insight{Fingerprint: "opendata:pcc:a", Score: 40, Summary: "od"},
			Issue:    domain.Insight{Fingerprint: "issues:o/r#1", Score: 30, Title: "Pain", URL: "https://example.com/1", Summary: "is"},
			Reason:   "same theme: gov-procurement",
		}},
	}
	list := FromBriefing(b, "2026-W34")
	if len(list) != 1 {
		t.Fatalf("expected 1 overlap-only task, got %d %#v", len(list), list)
	}
	if list[0].Pipeline != "overlap" || list[0].Reward != 3 {
		t.Fatalf("unexpected task %+v", list[0])
	}
}

func TestMergeDedupAndStale(t *testing.T) {
	existing := []Task{{Fingerprint: "a", Title: "Old", Priority: 1, Reward: 2, Difficulty: 2, Risk: 1, ScoreNorm: 1}}
	incoming := []Task{{Fingerprint: "b", Title: "New", Priority: 2, Reward: 2, Difficulty: 1, Risk: 1, ScoreNorm: 1, Score: 10}}
	ApplyHeuristicAxes(incoming)
	NormalizeScores(incoming)
	Prioritize(incoming)
	merged := Merge(existing, incoming, "2026-W35")
	if len(merged) != 2 {
		t.Fatalf("expected 2, got %d", len(merged))
	}
	var staleA bool
	for _, m := range merged {
		if m.Fingerprint == "a" && m.Stale {
			staleA = true
		}
		if m.Fingerprint == "b" && m.Stale {
			t.Fatal("new task should not be stale")
		}
	}
	if !staleA {
		t.Fatal("expected fingerprint a to be stale")
	}
}

func TestRenderParseRoundTrip(t *testing.T) {
	list := []Task{{
		Fingerprint: "issues:o/r#1",
		Title:       "Cache bug",
		Pipeline:    "issues",
		URL:         "https://example.com/1",
		Summary:     "pain",
		Score:       40,
		ScoreNorm:   1,
		Reward:      2,
		Difficulty:  3,
		Risk:        2,
		AxesSource:  AxesSourceXAI,
		Rationale:   "complex migration",
		Priority:    1.5,
		Seen:        "2026-W34",
	}}
	Prioritize(list)
	md := Render(list, "2026-W34", "keep me")
	parsed, notes := Parse(md)
	if notes != "keep me" {
		t.Fatalf("notes=%q", notes)
	}
	if len(parsed) != 1 {
		t.Fatalf("parsed %d", len(parsed))
	}
	if parsed[0].Fingerprint != "issues:o/r#1" || parsed[0].Difficulty != 3 || parsed[0].AxesSource != AxesSourceXAI {
		t.Fatalf("parsed %+v", parsed[0])
	}
}

func TestPriorityUsesAxes(t *testing.T) {
	list := []Task{
		{Fingerprint: "hard", Score: 10, ScoreNorm: 1, Reward: 2, Difficulty: 5, Risk: 5},
		{Fingerprint: "easy", Score: 10, ScoreNorm: 1, Reward: 2, Difficulty: 1, Risk: 1},
	}
	Prioritize(list)
	if list[0].Fingerprint != "easy" {
		t.Fatalf("expected easy first, got %s (p0=%v p1=%v)", list[0].Fingerprint, list[0].Priority, list[1].Priority)
	}
}

func BenchmarkParse(b *testing.B) {
	md := Render([]Task{{
		Fingerprint: "issues:o/r#1",
		Title:       "Cache bug",
		Pipeline:    "issues",
		URL:         "https://example.com/1",
		Summary:     "pain",
		Score:       40,
		ScoreNorm:   1,
		Reward:      2,
		Difficulty:  3,
		Risk:        2,
		AxesSource:  AxesSourceXAI,
		Rationale:   "complex migration",
		Priority:    1.5,
		Seen:        "2026-W34",
	}}, "2026-W34", "notes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(md)
	}
}

func BenchmarkRender(b *testing.B) {
	list := []Task{{
		Fingerprint: "issues:o/r#1",
		Title:       "Cache bug",
		Pipeline:    "issues",
		URL:         "https://example.com/1",
		Summary:     "pain",
		Score:       40,
		ScoreNorm:   1,
		Reward:      2,
		Difficulty:  3,
		Risk:        2,
		AxesSource:  AxesSourceXAI,
		Rationale:   "complex migration",
		Priority:    1.5,
		Seen:        "2026-W34",
	}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Render(list, "2026-W34", "notes")
	}
}

func BenchmarkPrioritize(b *testing.B) {
	list := make([]Task, 100)
	for i := range list {
		list[i] = Task{
			Fingerprint: "fp" + string(rune('a'+i%26)),
			Score:       float64(i * 10),
			ScoreNorm:   float64(i) / 100.0,
			Reward:      2,
			Difficulty:  i%5 + 1,
			Risk:        i%3 + 1,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Prioritize(list)
	}
}

func BenchmarkMerge(b *testing.B) {
	existing := make([]Task, 50)
	for i := range existing {
		existing[i] = Task{Fingerprint: "old" + string(rune('a'+i%26)), Title: "Old", Priority: 1}
	}
	incoming := make([]Task, 50)
	for i := range incoming {
		incoming[i] = Task{Fingerprint: "new" + string(rune('a'+i%26)), Title: "New", Score: float64(i)}
	}
	ApplyHeuristicAxes(incoming)
	NormalizeScores(incoming)
	Prioritize(incoming)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Merge(existing, incoming, "2026-W35")
	}
}
