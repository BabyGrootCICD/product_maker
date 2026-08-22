package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/converge"
	"github.com/BabyGrootCICD/product_maker/internal/crm/ghissues"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	ghclient "github.com/BabyGrootCICD/product_maker/internal/issues/github"
	"github.com/BabyGrootCICD/product_maker/internal/opendata"
	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
)

type Env struct {
	GitHubToken   string
	GitHubRepo    string // owner/repo
	SAMAPIKey     string
	XAIAPIKey     string
	StepSummary   string
	DryRun        bool
}

type Orchestrator struct {
	Config   *config.Config
	Repos    *config.ReposConfig
	Env      Env
}

func (o *Orchestrator) RunOpenData(ctx context.Context) ([]domain.Insight, []string, error) {
	runner := opendata.NewRunner(o.Config, o.Env.SAMAPIKey)
	insights, skips := runner.Run(ctx, o.Config)
	return insights, skips, nil
}

func (o *Orchestrator) RunIssues(ctx context.Context) ([]domain.Insight, error) {
	if o.Env.GitHubToken == "" {
		return nil, fmt.Errorf("issues: missing GITHUB_TOKEN")
	}
	client := ghclient.NewClient(o.Env.GitHubToken)
	return client.Run(ctx, o.Config, o.Repos)
}

func (o *Orchestrator) RunConverge(openData, issues []domain.Insight, skips []string) domain.Briefing {
	return converge.Build(openData, issues, o.Config.Converge, skips)
}

func (o *Orchestrator) WriteArtifacts(briefing domain.Briefing, openData, issues []domain.Insight) error {
	dir := o.Config.OutputDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "opendata.json"), openData); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "issues.json"), issues); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "briefing.json"), briefing); err != nil {
		return err
	}
	digest := RenderDigest(briefing)
	return os.WriteFile(filepath.Join(dir, "digest.md"), []byte(digest), 0o644)
}

func (o *Orchestrator) PublishCRM(ctx context.Context, briefing domain.Briefing) error {
	if o.Env.DryRun || o.Env.GitHubToken == "" || o.Env.GitHubRepo == "" {
		slog.Info("crm skipped", "dry_run", o.Env.DryRun)
		return nil
	}
	owner, repo, err := splitRepo(o.Env.GitHubRepo)
	if err != nil {
		return err
	}
	crm := ghissues.NewClient(o.Env.GitHubToken, owner, repo)

	week := isoWeek(time.Now())
	digestTitle := fmt.Sprintf("Weekly discovery digest — %s", week)
	digestBody := RenderDigest(briefing)
	labels := []string{o.Config.CRM.DigestLabel, o.Config.CRM.ReviewLabel}
	if _, err := crm.CreateDigest(ctx, digestTitle, digestBody, labels); err != nil {
		return fmt.Errorf("create digest: %w", err)
	}

	if err := o.upsertSignals(ctx, crm, briefing); err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) Synthesize(ctx context.Context, briefing domain.Briefing) error {
	if o.Env.XAIAPIKey == "" {
		slog.Info("llm skipped", "reason", "missing XAI_API_KEY")
		return nil
	}
	if o.Env.DryRun || o.Env.GitHubToken == "" || o.Env.GitHubRepo == "" {
		return nil
	}

	owner, repo, err := splitRepo(o.Env.GitHubRepo)
	if err != nil {
		return err
	}
	crm := ghissues.NewClient(o.Env.GitHubToken, owner, repo)
	llm := xai.NewClient(o.Config.Qual.XAIBaseURL, o.Env.XAIAPIKey, o.Config.Qual.XAIModel)

	count := 0
	for _, overlap := range briefing.Overlaps {
		if count >= o.Config.Qual.MaxOutreachPerRun {
			break
		}
		if err := o.draftOutreach(ctx, crm, llm, overlap.Issue); err == nil {
			count++
		}
	}
	unmatchedOD, unmatchedIS := converge.UnmatchedTop(briefing, o.Config.Converge.UnmatchedTopN)
	for _, ins := range unmatchedIS {
		if count >= o.Config.Qual.MaxOutreachPerRun {
			break
		}
		if err := o.draftOutreach(ctx, crm, llm, ins); err == nil {
			count++
		}
	}
	_ = unmatchedOD
	return nil
}

func (o *Orchestrator) draftOutreach(ctx context.Context, crm *ghissues.Client, llm *xai.Client, ins domain.Insight) error {
	draft, err := llm.GenerateOutreach(ctx, ins.Title, ins.Body)
	if err != nil {
		slog.Warn("outreach skipped", "fingerprint", ins.Fingerprint, "err", err)
		return err
	}
	body := ghissues.FingerprintComment(ins.Fingerprint) + "\n\n" +
		fmt.Sprintf("Signal from **%s** pipeline.\n\n- URL: %s\n- Score: %.1f\n\n%s",
			ins.Pipeline, ins.URL, ins.Score, ins.Summary)
	body = ghissues.AppendOutreachDraft(body, draft)
	labels := []string{o.Config.CRM.SignalLabel, "pipeline/issues", o.Config.CRM.ReviewLabel}
	_, err = crm.UpsertSignal(ctx, "[signal] "+ins.Title, body, ins.Fingerprint, labels)
	return err
}

func (o *Orchestrator) upsertSignals(ctx context.Context, crm *ghissues.Client, briefing domain.Briefing) error {
	for _, overlap := range briefing.Overlaps {
		fp := "overlap:" + overlap.OpenData.Fingerprint + "+" + overlap.Issue.Fingerprint
		body := ghissues.FingerprintComment(fp) + "\n\n" +
			fmt.Sprintf("## Overlap (%s)\n\n### Open Data\n- %s\n- %s\n\n### Issue\n- %s\n- %s\n",
				overlap.Reason, overlap.OpenData.Title, overlap.OpenData.Summary, overlap.Issue.Title, overlap.Issue.URL)
		labels := []string{o.Config.CRM.SignalLabel, "pipeline/overlap", o.Config.CRM.ReviewLabel}
		if _, err := crm.UpsertSignal(ctx, "[overlap] "+overlap.Issue.Title, body, fp, labels); err != nil {
			return err
		}
	}

	unmatchedOD, unmatchedIS := converge.UnmatchedTop(briefing, o.Config.Converge.UnmatchedTopN)
	for _, ins := range unmatchedOD {
		body := formatSignalBody(ins)
		labels := []string{o.Config.CRM.SignalLabel, "pipeline/opendata", o.Config.CRM.ReviewLabel}
		if _, err := crm.UpsertSignal(ctx, "[signal] "+ins.Title, body, ins.Fingerprint, labels); err != nil {
			return err
		}
	}
	for _, ins := range unmatchedIS {
		body := formatSignalBody(ins)
		labels := []string{o.Config.CRM.SignalLabel, "pipeline/issues", o.Config.CRM.ReviewLabel}
		if _, err := crm.UpsertSignal(ctx, "[signal] "+ins.Title, body, ins.Fingerprint, labels); err != nil {
			return err
		}
	}
	return nil
}

func formatSignalBody(ins domain.Insight) string {
	urlLine := ins.URL
	if urlLine == "" {
		urlLine = "_n/a_"
	}
	return ghissues.FingerprintComment(ins.Fingerprint) + "\n\n" +
		fmt.Sprintf("Signal from **%s** pipeline (%s).\n\n- URL: %s\n- Score: %.1f\n- Summary: %s\n",
			ins.Pipeline, ins.Source, urlLine, ins.Score, ins.Summary)
}

func (o *Orchestrator) WriteStepSummary(briefing domain.Briefing) {
	if o.Env.StepSummary == "" {
		return
	}
	content := fmt.Sprintf(`## Discovery pipeline summary

| Metric | Count |
| --- | --- |
| Open data insights | %d |
| Issue insights | %d |
| Overlaps | %d |
| Skips | %d |

`, len(briefing.RankedOpenData), len(briefing.RankedIssues), len(briefing.Overlaps), len(briefing.Skips))
	if len(briefing.Skips) > 0 {
		content += "### Skips\n"
		for _, s := range briefing.Skips {
			content += "- " + s + "\n"
		}
	}
	_ = os.WriteFile(o.Env.StepSummary, []byte(content), 0o644)
}

func RenderDigest(b domain.Briefing) string {
	var sb strings.Builder
	sb.WriteString("# Weekly Discovery Digest\n\n")
	sb.WriteString("## Quant A — Open Data insights\n\n")
	sb.WriteString(renderInsightTable(b.RankedOpenData))
	sb.WriteString("\n## Quant B — Issue pain points\n\n")
	sb.WriteString(renderInsightTable(b.RankedIssues))
	sb.WriteString("\n## Overlaps (priority for qualitative interviews)\n\n")
	if len(b.Overlaps) == 0 {
		sb.WriteString("_No overlaps this week (unmatched signals below)._ \n\n")
	} else {
		for _, o := range b.Overlaps {
			sb.WriteString(fmt.Sprintf("- **%s** ↔ [%s](%s) — _%s_\n", o.OpenData.Title, o.Issue.Title, o.Issue.URL, o.Reason))
		}
		sb.WriteString("\n")
	}
	if len(b.Skips) > 0 {
		sb.WriteString("## Skips / errors\n\n")
		for _, s := range b.Skips {
			sb.WriteString("- " + s + "\n")
		}
	}
	return sb.String()
}

func renderInsightTable(items []domain.Insight) string {
	if len(items) == 0 {
		return "_None_\n"
	}
	var sb strings.Builder
	sb.WriteString("| Title | Score | Source | Summary |\n| --- | ---: | --- | --- |\n")
	for _, i := range items {
		title := i.Title
		if i.URL != "" {
			title = fmt.Sprintf("[%s](%s)", i.Title, i.URL)
		}
		sb.WriteString(fmt.Sprintf("| %s | %.1f | %s | %s |\n", title, i.Score, i.Source, i.Summary))
	}
	return sb.String()
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func splitRepo(full string) (owner, repo string, err error) {
	parts := strings.SplitN(full, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo %q", full)
	}
	return parts[0], parts[1], nil
}

func isoWeek(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}
