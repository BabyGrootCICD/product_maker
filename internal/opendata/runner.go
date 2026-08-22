package opendata

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/opendata/pcc"
	"github.com/BabyGrootCICD/product_maker/internal/opendata/samgov"
)

type Runner struct {
	PCC    *pcc.Client
	SAMGov *samgov.Client
}

func NewRunner(cfg *config.Config, samAPIKey string) *Runner {
	return &Runner{
		PCC:    pcc.NewClient(cfg.OpenData.PCC.BaseURL),
		SAMGov: samgov.NewClient(cfg.OpenData.SAMGov.BaseURL, samAPIKey),
	}
}

func (r *Runner) Run(ctx context.Context, cfg *config.Config) ([]domain.Insight, []string) {
	var insights []domain.Insight
	var skips []string

	if cfg.OpenData.PCC.Enabled {
		pccInsights, err := r.PCC.Run(ctx, cfg.OpenData.PCC)
		if err != nil {
			slog.Warn("pcc skipped", "err", err)
			skips = append(skips, fmt.Sprintf("PCC skipped: %v", err))
		} else {
			insights = append(insights, pccInsights...)
		}
	}

	if cfg.OpenData.SAMGov.Enabled {
		samInsights, err := r.SAMGov.Run(ctx, cfg.OpenData.SAMGov)
		if err != nil {
			slog.Warn("samgov skipped", "err", err)
			skips = append(skips, fmt.Sprintf("SAM.gov skipped: %v", err))
		} else {
			insights = append(insights, samInsights...)
		}
	}

	return insights, skips
}
