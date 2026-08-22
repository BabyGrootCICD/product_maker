package converge

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

func Build(openData, issues []domain.Insight, cfg config.ConvergeConfig, skips []string) domain.Briefing {
	rankedOD := rankTop(openData, cfg.TopN)
	rankedIS := rankTop(issues, cfg.TopN)
	overlaps := findOverlaps(rankedOD, rankedIS, cfg.JaccardThreshold)

	matchedOD := map[string]bool{}
	matchedIS := map[string]bool{}
	for _, o := range overlaps {
		matchedOD[o.OpenData.Fingerprint] = true
		matchedIS[o.Issue.Fingerprint] = true
	}

	return domain.Briefing{
		RankedOpenData: rankedOD,
		RankedIssues:   rankedIS,
		Overlaps:       overlaps,
		Skips:          skips,
	}
}

func rankTop(items []domain.Insight, n int) []domain.Insight {
	sorted := append([]domain.Insight(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

func findOverlaps(openData, issues []domain.Insight, threshold float64) []domain.Overlap {
	var out []domain.Overlap
	for _, od := range openData {
		for _, is := range issues {
			reason, ok := matchReason(od, is, threshold)
			if ok {
				out = append(out, domain.Overlap{
					OpenData: od,
					Issue:    is,
					Reason:   reason,
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].OpenData.Score+out[i].Issue.Score > out[j].OpenData.Score+out[j].Issue.Score
	})
	return out
}

func matchReason(od, is domain.Insight, threshold float64) (string, bool) {
	if od.Theme != "" && od.Theme == is.Theme {
		return fmt.Sprintf("same theme: %s", od.Theme), true
	}
	j := jaccard(od.Keywords, is.Keywords)
	if j >= threshold {
		return fmt.Sprintf("keyword jaccard=%.2f", j), true
	}
	return "", false
}

func jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := map[string]struct{}{}
	for _, k := range a {
		setA[strings.ToLower(k)] = struct{}{}
	}
	intersect := 0
	union := map[string]struct{}{}
	for k := range setA {
		union[k] = struct{}{}
	}
	for _, k := range b {
		k = strings.ToLower(k)
		union[k] = struct{}{}
		if _, ok := setA[k]; ok {
			intersect++
		}
	}
	if len(union) == 0 {
		return 0
	}
	return float64(intersect) / float64(len(union))
}

// UnmatchedTop returns insights not in any overlap, limited per pipeline.
func UnmatchedTop(b domain.Briefing, n int) (openData, issues []domain.Insight) {
	matchedOD := map[string]bool{}
	matchedIS := map[string]bool{}
	for _, o := range b.Overlaps {
		matchedOD[o.OpenData.Fingerprint] = true
		matchedIS[o.Issue.Fingerprint] = true
	}
	for _, od := range b.RankedOpenData {
		if matchedOD[od.Fingerprint] {
			continue
		}
		openData = append(openData, od)
		if len(openData) >= n {
			break
		}
	}
	for _, is := range b.RankedIssues {
		if matchedIS[is.Fingerprint] {
			continue
		}
		issues = append(issues, is)
		if len(issues) >= n {
			break
		}
	}
	return openData, issues
}
