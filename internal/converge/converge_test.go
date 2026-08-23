package converge

import (
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

func TestFindOverlapsByTheme(t *testing.T) {
	od := []domain.Insight{{
		Fingerprint: "opendata:pcc:cat:info",
		Theme:       "gov-procurement",
		Score:       50,
		Keywords:    []string{"procurement"},
	}}
	is := []domain.Insight{{
		Fingerprint: "issues:ckan/ckan#1",
		Theme:       "gov-procurement",
		Score:       30,
		Keywords:    []string{"open-data"},
	}}
	b := Build(od, is, config.ConvergeConfig{TopN: 5, JaccardThreshold: 0.5}, nil)
	if len(b.Overlaps) != 1 {
		t.Fatalf("expected 1 overlap, got %d", len(b.Overlaps))
	}
}

func TestUnmatchedTopN(t *testing.T) {
	od := []domain.Insight{
		{Fingerprint: "a", Theme: "gov-procurement", Score: 10},
		{Fingerprint: "b", Theme: "gov-procurement", Score: 9},
	}
	is := []domain.Insight{
		{Fingerprint: "c", Theme: "devtools", Score: 20},
	}
	b := Build(od, is, config.ConvergeConfig{TopN: 5, JaccardThreshold: 0.99}, nil)
	unmatchedOD, unmatchedIS := UnmatchedTop(b, 2)
	if len(unmatchedOD) != 2 {
		t.Fatalf("expected 2 unmatched opendata, got %d", len(unmatchedOD))
	}
	if len(unmatchedIS) != 1 {
		t.Fatalf("expected 1 unmatched issue, got %d", len(unmatchedIS))
	}
}

func BenchmarkBuild(b *testing.B) {
	od := make([]domain.Insight, 50)
	for i := range od {
		od[i] = domain.Insight{
			Fingerprint: "opendata:pcc:" + string(rune('a'+i%26)),
			Theme:       "gov-procurement",
			Score:       float64(i * 10),
			Keywords:    []string{"procurement", "failure"},
		}
	}
	is := make([]domain.Insight, 50)
	for i := range is {
		is[i] = domain.Insight{
			Fingerprint: "issues:repo#" + string(rune('a'+i%26)),
			Theme:       "devtools",
			Score:       float64(i * 8),
			Keywords:    []string{"performance", "cache"},
		}
	}
	cfg := config.ConvergeConfig{TopN: 20, JaccardThreshold: 0.15}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Build(od, is, cfg, nil)
	}
}
