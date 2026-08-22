package score

import (
	"fmt"
	"strings"

	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

type IssueCandidate struct {
	Owner       string
	Name        string
	Theme       string
	Number      int
	Title       string
	URL         string
	Comments    int
	ThumbsUp    int
	Labels      []string
	Body        string
}

type Config struct {
	MinComments int
	MinThumbsUp int
	MaxComments int
	LabelBonus  map[string]int
}

func FilterAndScore(candidates []IssueCandidate, cfg Config) []domain.Insight {
	var out []domain.Insight
	for _, c := range candidates {
		if c.Comments < cfg.MinComments || c.ThumbsUp <= cfg.MinThumbsUp {
			continue
		}
		if cfg.MaxComments > 0 && c.Comments > cfg.MaxComments {
			continue
		}
		pain := float64(c.Comments) + 2*float64(c.ThumbsUp) + labelBonus(c.Labels, cfg.LabelBonus)
		fp := fmtFingerprint(c.Owner, c.Name, c.Number)
		out = append(out, domain.Insight{
			Fingerprint: fp,
			Pipeline:    domain.PipelineIssues,
			Source:      "github",
			Kind:        domain.KindPainPoint,
			Theme:       c.Theme,
			Title:       c.Title,
			URL:         c.URL,
			Score:       pain,
			Metrics: map[string]float64{
				"comments":   float64(c.Comments),
				"thumbs_up":  float64(c.ThumbsUp),
				"pain_score": pain,
			},
			Keywords: append(tokenize(c.Title), c.Labels...),
			Summary:  fmt.Sprintf("repo=%s/%s comments=%d thumbs=%d", c.Owner, c.Name, c.Comments, c.ThumbsUp),
			Body:     c.Body,
		})
	}
	return out
}

func fmtFingerprint(owner, name string, number int) string {
	return "issues:" + owner + "/" + name + "#" + itoa(number)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func labelBonus(labels []string, bonus map[string]int) float64 {
	var total float64
	for _, l := range labels {
		key := strings.ToLower(l)
		if v, ok := bonus[key]; ok {
			total += float64(v)
		}
	}
	return total
}

func tokenize(s string) []string {
	parts := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return r == ' ' || r == '/' || r == '-' || r == ':' || r == ','
	})
	var out []string
	for _, p := range parts {
		if len(p) >= 3 {
			out = append(out, p)
		}
	}
	return out
}
