package score

import (
	"testing"
)

func TestFilterAndScore(t *testing.T) {
	cfg := Config{
		MinComments: 10,
		MinThumbsUp: 5,
		MaxComments: 500,
		LabelBonus:  map[string]int{"wontfix": 5, "help wanted": 3},
	}
	candidates := []IssueCandidate{
		{Owner: "actions", Name: "runner", Theme: "devtools", Number: 1, Title: "Cache issue", URL: "https://github.com/a/b/issues/1", Comments: 12, ThumbsUp: 6, Labels: []string{"wontfix"}},
		{Owner: "actions", Name: "runner", Theme: "devtools", Number: 2, Title: "Low", URL: "https://github.com/a/b/issues/2", Comments: 2, ThumbsUp: 1},
	}
	out := FilterAndScore(candidates, cfg)
	if len(out) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(out))
	}
	want := 12.0 + 2*6.0 + 5.0
	if out[0].Score != want {
		t.Fatalf("expected score %v, got %v", want, out[0].Score)
	}
}

func TestMaxCommentsFilter(t *testing.T) {
	cfg := Config{MinComments: 10, MinThumbsUp: 5, MaxComments: 100}
	candidates := []IssueCandidate{
		{Owner: "k", Name: "k", Theme: "devtools", Number: 99, Title: "Mega", Comments: 1000, ThumbsUp: 50},
	}
	out := FilterAndScore(candidates, cfg)
	if len(out) != 0 {
		t.Fatalf("expected mega issue filtered out")
	}
}

func BenchmarkFilterAndScore(b *testing.B) {
	cfg := Config{
		MinComments: 10,
		MinThumbsUp: 5,
		MaxComments: 500,
		LabelBonus:  map[string]int{"wontfix": 5, "help wanted": 3},
	}
	candidates := make([]IssueCandidate, 200)
	for i := range candidates {
		candidates[i] = IssueCandidate{
			Owner:     "owner",
			Name:      "repo",
			Theme:     "devtools",
			Number:    i,
			Title:     "Issue",
			URL:       "https://github.com/owner/repo/issues/1",
			Comments:  i + 5,
			ThumbsUp:  i / 2,
			Labels:    []string{"wontfix"},
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterAndScore(candidates, cfg)
	}
}
