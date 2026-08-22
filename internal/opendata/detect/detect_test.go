package detect

import (
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

func TestFromCategoryStatsThreshold(t *testing.T) {
	stats := []CategoryStats{
		{Category: "資訊服務", TotalCases: 100, FailedCases: 40, FailureRate: 0.4},
		{Category: "勞務類", TotalCases: 50, FailedCases: 30, FailureRate: 0.6},
	}
	cfg := DetectConfig{MinCases: 100, FailureRateThreshold: 0.4, Source: "pcc", Theme: "gov-procurement"}
	insights := FromCategoryStats(stats, cfg)
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Kind != domain.KindOpportunityAlert {
		t.Fatalf("unexpected kind %s", insights[0].Kind)
	}
	if insights[0].Metrics["failure_rate"] != 0.4 {
		t.Fatalf("unexpected failure_rate %v", insights[0].Metrics["failure_rate"])
	}
}

func TestFromNAICSStatsThreshold(t *testing.T) {
	stats := []NAICSStats{{NAICS: "541512", Total: 25, SourcesSoughtOrSpecial: 8, PainProxy: 0.32}}
	cfg := SAMDetectConfig{MinCases: 20, PainProxyThreshold: 0.25, Theme: "gov-procurement"}
	insights := FromNAICSStats(stats, cfg)
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
}
