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
