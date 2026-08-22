package pipeline

import (
	"strings"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

func TestRenderDigestIncludesSections(t *testing.T) {
	b := domain.Briefing{
		RankedOpenData: []domain.Insight{{Title: "PCC alert", Score: 40, Source: "pcc", Summary: "rate=0.5"}},
		RankedIssues:   []domain.Insight{{Title: "Cache bug", Score: 30, Source: "github", URL: "https://github.com/a/b/issues/1", Summary: "pain"}},
		Overlaps: []domain.Overlap{{
			OpenData: domain.Insight{Title: "PCC alert"},
			Issue:    domain.Insight{Title: "Cache bug", URL: "https://github.com/a/b/issues/1"},
			Reason:   "same theme: gov-procurement",
		}},
		Skips: []string{"SAM.gov skipped: missing key"},
	}
	md := RenderDigest(b)
	for _, want := range []string{"Quant A", "Quant B", "Overlaps", "SAM.gov skipped"} {
		if !strings.Contains(md, want) {
			t.Fatalf("digest missing %q:\n%s", want, md)
		}
	}
}
