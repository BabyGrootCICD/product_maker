package detect

import (
	"fmt"
	"strings"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

type CategoryStats struct {
	Category     string
	TotalCases   int
	FailedCases  int
	FailureRate  float64
}

type DetectConfig struct {
	MinCases             int
	FailureRateThreshold float64
	Source               string
	Theme                string
}

func FromCategoryStats(stats []CategoryStats, cfg DetectConfig) []domain.Insight {
	var out []domain.Insight
	for _, s := range stats {
		if s.TotalCases < cfg.MinCases || s.FailureRate < cfg.FailureRateThreshold {
			continue
		}
		fp := fmt.Sprintf("opendata:%s:cat:%s", cfg.Source, normalizeKey(s.Category))
		out = append(out, domain.Insight{
			Fingerprint: fp,
			Pipeline:    domain.PipelineOpenData,
			Source:      cfg.Source,
			Kind:        domain.KindOpportunityAlert,
			Theme:       cfg.Theme,
			Title:       fmt.Sprintf("[%s] %s 流標率 %.0f%%", cfg.Source, s.Category, s.FailureRate*100),
			URL:         "",
			Score:       s.FailureRate * float64(s.TotalCases),
			Metrics: map[string]float64{
				"total_cases":   float64(s.TotalCases),
				"failed_cases":  float64(s.FailedCases),
				"failure_rate":  s.FailureRate,
			},
			Keywords: tokenize(s.Category),
			Summary:  fmt.Sprintf("category=%s total=%d failed=%d rate=%.2f", s.Category, s.TotalCases, s.FailedCases, s.FailureRate),
		})
	}
	return out
}

type NAICSStats struct {
	NAICS      string
	Total      int
	SourcesSoughtOrSpecial int
	PainProxy  float64
}

type SAMDetectConfig struct {
	MinCases           int
	PainProxyThreshold float64
	Theme              string
}

func FromNAICSStats(stats []NAICSStats, cfg SAMDetectConfig) []domain.Insight {
	var out []domain.Insight
	for _, s := range stats {
		if s.Total < cfg.MinCases || s.PainProxy < cfg.PainProxyThreshold {
			continue
		}
		fp := fmt.Sprintf("opendata:samgov:naics:%s", s.NAICS)
		out = append(out, domain.Insight{
			Fingerprint: fp,
			Pipeline:    domain.PipelineOpenData,
			Source:      "samgov",
			Kind:        domain.KindOpportunityAlert,
			Theme:       cfg.Theme,
			Title:       fmt.Sprintf("[samgov] NAICS %s pain_proxy %.0f%%", s.NAICS, s.PainProxy*100),
			URL:         "https://sam.gov/search/",
			Score:       s.PainProxy * float64(s.Total),
			Metrics: map[string]float64{
				"total":        float64(s.Total),
				"sources_special": float64(s.SourcesSoughtOrSpecial),
				"pain_proxy":   s.PainProxy,
			},
			Keywords: append([]string{s.NAICS, "procurement", "gov"}, tokenize(s.NAICS)...),
			Summary:  fmt.Sprintf("naics=%s total=%d proxy=%.2f", s.NAICS, s.Total, s.PainProxy),
		})
	}
	return out
}

func normalizeKey(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "_")
}

func tokenize(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '/' || r == '-' || r == ','
	})
	var out []string
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) >= 2 {
			out = append(out, p)
		}
	}
	return out
}
