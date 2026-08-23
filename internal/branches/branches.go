package branches

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
	"github.com/BabyGrootCICD/product_maker/internal/tasks"
)

type SolutionGenerator interface {
	GenerateSolutionPaths(ctx context.Context, title, summary, pipeline, url string, priority float64) (string, error)
}

type Options struct {
	TasksPath   string
	TopN        int
	DryRun      bool
	Week        string
	WorkDir     string // repo root
	IncludeStale bool
	AutoCreatePR bool
}

func ShortHash(fingerprint string) string {
	sum := sha1.Sum([]byte(fingerprint))
	return hex.EncodeToString(sum[:])[:8]
}

func BranchName(week, fingerprint string) string {
	return fmt.Sprintf("discovery/%s/%s", week, ShortHash(fingerprint))
}

func IssueBranchName(number int, slug string) string {
	return fmt.Sprintf("issue/%d-%s", number, slug)
}

func Slugify(title string) string {
	slug := strings.ToLower(title)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1
	}, slug)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 40 {
		slug = slug[:40]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

func RenderBrief(t tasks.Task, solutions string) string {
	url := t.URL
	if url == "" {
		url = "_n/a_"
	}
	return fmt.Sprintf(`<!-- fingerprint:%s -->
# %s

## Issue / signal

- **pipeline**: %s
- **priority**: %.2f
- **score**: %.1f
- **axes**: reward=%.0f difficulty=%d risk=%d source=%s
- **url**: %s
- **summary**: %s
- **seen**: %s

## Recommended solution paths

%s
`, t.Fingerprint, t.Title, t.Pipeline, t.Priority, t.Score, t.Reward, t.Difficulty, t.Risk, t.AxesSource, url, t.Summary, t.Seen, strings.TrimSpace(solutions))
}

func Run(ctx context.Context, gen SolutionGenerator, opts Options) error {
	if opts.TopN <= 0 {
		opts.TopN = 3
	}
	if opts.Week == "" {
		opts.Week = tasks.ISOWeek(time.Now())
	}
	if opts.WorkDir == "" {
		opts.WorkDir = "."
	}
	if gen == nil {
		return fmt.Errorf("branches: missing xAI client")
	}

	md, err := os.ReadFile(opts.TasksPath)
	if err != nil {
		return fmt.Errorf("read tasks: %w", err)
	}
	list, _ := tasks.Parse(string(md))
	selected := tasks.SelectTop(list, opts.TopN, !opts.IncludeStale)
	if len(selected) == 0 {
		slog.Info("branches: no tasks selected")
		return nil
	}

	origRef, err := gitOutput(opts.WorkDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	origRef = strings.TrimSpace(origRef)

	for i, t := range selected {
		if err := createOrUpdateBranch(ctx, gen, opts, t, i+1); err != nil {
			slog.Warn("branch failed", "fingerprint", t.Fingerprint, "err", err)
			continue
		}
	}

	if err := gitRun(opts.WorkDir, "checkout", origRef); err != nil {
		return fmt.Errorf("restore ref %s: %w", origRef, err)
	}
	return nil
}

func createOrUpdateBranch(ctx context.Context, gen SolutionGenerator, opts Options, t tasks.Task, index int) error {
	solutions, err := gen.GenerateSolutionPaths(ctx, t.Title, t.Summary, t.Pipeline, t.URL, t.Priority)
	if err != nil {
		return err
	}
	brief := RenderBrief(t, solutions)
	hash := ShortHash(t.Fingerprint)

	var branch string
	if opts.IncludeStale || strings.Contains(t.Fingerprint, "opendata:") || t.Pipeline == "issues" {
		slug := Slugify(t.Title)
		branch = IssueBranchName(index, slug)
	} else {
		branch = BranchName(opts.Week, t.Fingerprint)
	}
	relPath := filepath.Join("discovery", "briefs", hash+".md")

	if opts.DryRun {
		slog.Info("branches dry-run", "branch", branch, "file", relPath)
		_ = os.MkdirAll(filepath.Join(opts.WorkDir, "discovery", "briefs"), 0o755)
		return os.WriteFile(filepath.Join(opts.WorkDir, relPath), []byte(brief), 0o644)
	}

	base := "main"
	if err := gitRun(opts.WorkDir, "rev-parse", "--verify", "origin/main"); err != nil {
		base = "master"
	}

	if err := gitRun(opts.WorkDir, "fetch", "origin", base); err != nil {
		slog.Warn("git fetch skipped", "err", err)
	}

	startPoint := "origin/" + base
	if err := gitRun(opts.WorkDir, "rev-parse", "--verify", startPoint); err != nil {
		startPoint = base
	}

	if err := gitRun(opts.WorkDir, "checkout", "-B", branch, startPoint); err != nil {
		return fmt.Errorf("checkout %s: %w", branch, err)
	}

	abs := filepath.Join(opts.WorkDir, relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(abs, []byte(brief), 0o644); err != nil {
		return err
	}
	if err := gitRun(opts.WorkDir, "add", "-f", relPath); err != nil {
		return err
	}
	if err := gitRun(opts.WorkDir, "diff", "--staged", "--quiet"); err == nil {
		slog.Info("branch up to date", "branch", branch)
	} else {
		msg := fmt.Sprintf("docs: discovery brief for %s", hash)
		if err := gitRun(opts.WorkDir, "commit", "-m", msg); err != nil {
			return err
		}
	}
	if err := gitRun(opts.WorkDir, "push", "-u", "origin", "HEAD"); err != nil {
		return fmt.Errorf("push %s: %w", branch, err)
	}
	slog.Info("branch pushed", "branch", branch)

	if opts.AutoCreatePR {
		title := fmt.Sprintf("Solution for: %s", t.Title)
		body := fmt.Sprintf(`## Solution for: %s
- **Original issue**: %s
- **Pipeline**: %s
- **Priority**: %.2f
- **Branch**: %s

### Solution Brief
%s`, t.Title, t.URL, t.Pipeline, t.Priority, branch, brief)
		prArgs := []string{"pr", "create", "--draft", "--title", title, "--body", body, "--base", base}
		cmd := exec.Command("gh", prArgs...)
		cmd.Dir = opts.WorkDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			slog.Warn("auto-create PR failed", "branch", branch, "err", err)
		} else {
			slog.Info("PR created", "branch", branch)
		}
	}

	return nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// Ensure SolutionGenerator matches xai.Client
var _ SolutionGenerator = (*xai.Client)(nil)
