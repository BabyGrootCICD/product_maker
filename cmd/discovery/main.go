package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/branches"
	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/pipeline"
	"github.com/BabyGrootCICD/product_maker/internal/qual/jtbd"
	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
	"github.com/BabyGrootCICD/product_maker/internal/tasks"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	switch os.Args[1] {
	case "run":
		os.Exit(runCommand(os.Args[2:]))
	case "opendata":
		os.Exit(opendataCommand(os.Args[2:]))
	case "issues":
		os.Exit(issuesCommand(os.Args[2:]))
	case "converge":
		os.Exit(convergeCommand(os.Args[2:]))
	case "tasks":
		os.Exit(tasksCommand(os.Args[2:]))
	case "branches":
		os.Exit(branchesCommand(os.Args[2:]))
	case "synthesize":
		os.Exit(synthesizeCommand(os.Args[2:]))
	case "jtbd":
		os.Exit(jtbdCommand(os.Args[2:]))
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `discovery — 量化雙軌 ➔ 質化收斂

Usage:
  discovery run [--config path] [--dry-run]
  discovery opendata [--config path]
  discovery issues [--config path]
  discovery converge [--config path] [--opendata path] [--issues path]
  discovery tasks [--config path] [--briefing path] [--out path]
  discovery branches [--config path] [--tasks path] [--top N] [--dry-run]
  discovery synthesize [--config path] [--briefing path]
  discovery jtbd --transcript path [--out path] [--config path]

Environment:
  GITHUB_TOKEN      GitHub API token (issues + CRM)
  GITHUB_REPOSITORY owner/repo for CRM output
  SAM_API_KEY       SAM.gov public API key
  XAI_API_KEY       xAI API key (axes / outreach / branches / jtbd)
  GITHUB_STEP_SUMMARY path to write job summary markdown
`)
}

func loadOrchestrator(args []string) (*pipeline.Orchestrator, error) {
	fs := flag.NewFlagSet("cmd", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	dryRun := fs.Bool("dry-run", false, "skip CRM writes")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		return nil, err
	}
	configDir := filepath.Dir(*configPath)
	reposPath := config.ResolveReposPath(cfg, configDir)
	repos, err := config.LoadRepos(reposPath)
	if err != nil {
		return nil, err
	}

	env := pipeline.Env{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubRepo:  firstNonEmpty(os.Getenv("GITHUB_REPOSITORY"), os.Getenv("DISCOVERY_GITHUB_REPO")),
		SAMAPIKey:   os.Getenv("SAM_API_KEY"),
		XAIAPIKey:   os.Getenv("XAI_API_KEY"),
		StepSummary: os.Getenv("GITHUB_STEP_SUMMARY"),
		DryRun:      *dryRun,
	}
	return &pipeline.Orchestrator{Config: cfg, Repos: repos, Env: env}, nil
}

func runCommand(args []string) int {
	orch, err := loadOrchestrator(args)
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	ctx := context.Background()

	openData, skips, err := orch.RunOpenData(ctx)
	if err != nil {
		slog.Error("opendata", "err", err)
	}
	issues, err := orch.RunIssues(ctx)
	if err != nil {
		slog.Error("issues", "err", err)
		skips = append(skips, fmt.Sprintf("Issues skipped: %v", err))
		issues = nil
	}

	briefing := orch.RunConverge(openData, issues, skips)
	if err := orch.WriteArtifacts(briefing, openData, issues); err != nil {
		slog.Error("write artifacts", "err", err)
		return 1
	}
	if err := orch.WriteTasks(ctx, briefing); err != nil {
		slog.Error("tasks", "err", err)
		return 1
	}
	if err := orch.PublishCRM(ctx, briefing); err != nil {
		slog.Error("crm", "err", err)
		return 1
	}
	if err := orch.Synthesize(ctx, briefing); err != nil {
		slog.Warn("synthesize", "err", err)
	}
	orch.WriteStepSummary(briefing)
	slog.Info("done", "opendata", len(openData), "issues", len(issues), "overlaps", len(briefing.Overlaps))
	return 0
}

func opendataCommand(args []string) int {
	orch, err := loadOrchestrator(args)
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	insights, skips, err := orch.RunOpenData(context.Background())
	if err != nil {
		slog.Error("opendata", "err", err)
		return 1
	}
	briefing := orch.RunConverge(insights, nil, skips)
	return writeOnly(orch, briefing, insights, nil)
}

func issuesCommand(args []string) int {
	orch, err := loadOrchestrator(args)
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	insights, err := orch.RunIssues(context.Background())
	if err != nil {
		slog.Error("issues", "err", err)
		return 1
	}
	briefing := orch.RunConverge(nil, insights, nil)
	return writeOnly(orch, briefing, nil, insights)
}

func convergeCommand(args []string) int {
	fs := flag.NewFlagSet("converge", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	openPath := fs.String("opendata", "", "opendata json path")
	issuesPath := fs.String("issues", "", "issues json path")
	_ = fs.Parse(args)

	orch, err := loadOrchestrator([]string{"--config", *configPath})
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	openData, err := loadInsights(*openPath, orch.Config.OutputDir+"/opendata.json")
	if err != nil {
		slog.Error("load opendata", "err", err)
		return 1
	}
	issues, err := loadInsights(*issuesPath, orch.Config.OutputDir+"/issues.json")
	if err != nil {
		slog.Error("load issues", "err", err)
		return 1
	}
	briefing := orch.RunConverge(openData, issues, nil)
	return writeOnly(orch, briefing, openData, issues)
}

func tasksCommand(args []string) int {
	fs := flag.NewFlagSet("tasks", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	briefingPath := fs.String("briefing", "", "briefing json path")
	outPath := fs.String("out", "", "tasks.md output path")
	_ = fs.Parse(args)

	orch, err := loadOrchestrator([]string{"--config", *configPath})
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	if *outPath != "" {
		orch.Config.Tasks.Path = *outPath
	}
	path := *briefingPath
	if path == "" {
		path = orch.Config.OutputDir + "/briefing.json"
	}
	briefing, err := loadBriefing(path)
	if err != nil {
		slog.Error("load briefing", "err", err)
		return 1
	}
	if err := orch.WriteTasks(context.Background(), briefing); err != nil {
		slog.Error("tasks", "err", err)
		return 1
	}
	return 0
}

func branchesCommand(args []string) int {
	fs := flag.NewFlagSet("branches", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	tasksPath := fs.String("tasks", "", "tasks.md path")
	top := fs.Int("top", 0, "top N non-stale tasks")
	dryRun := fs.Bool("dry-run", false, "write briefs locally without git push")
	all := fs.Bool("all", false, "include all tasks (including stale)")
	createPR := fs.Bool("create-pr", false, "auto-create draft PRs via gh")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	path := *tasksPath
	if path == "" {
		path = cfg.Tasks.Path
	}
	n := *top
	if *all {
		n = 999
	} else if n <= 0 {
		n = cfg.Tasks.BranchTopN
	}
	key := os.Getenv("XAI_API_KEY")
	if key == "" && !*dryRun {
		slog.Warn("branches skipped", "reason", "missing XAI_API_KEY")
		return 0
	}
	var gen branches.SolutionGenerator
	if key != "" {
		gen = xai.NewClient(cfg.Qual.XAIBaseURL, key, cfg.Qual.XAIModel)
	} else {
		gen = stubSolutions{}
	}
	err = branches.Run(context.Background(), gen, branches.Options{
		TasksPath:    path,
		TopN:         n,
		DryRun:       *dryRun,
		Week:         tasks.ISOWeek(time.Now()),
		WorkDir:      ".",
		IncludeStale: *all,
		AutoCreatePR: *createPR,
	})
	if err != nil {
		slog.Error("branches", "err", err)
		return 1
	}
	return 0
}

type stubSolutions struct{}

func (stubSolutions) GenerateSolutionPaths(ctx context.Context, title, summary, pipeline, url string, priority float64) (string, error) {
	return fmt.Sprintf("## Path 1 (dry-run stub)\n\nApproach: inspect %s\nWhy: priority %.2f\nSteps: 1) read issue 2) prototype\nRisks: unknown\nDone-when: validated hypothesis\n", url, priority), nil
}

func synthesizeCommand(args []string) int {
	fs := flag.NewFlagSet("synthesize", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	briefingPath := fs.String("briefing", "", "briefing json path")
	_ = fs.Parse(args)

	orch, err := loadOrchestrator([]string{"--config", *configPath})
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	path := *briefingPath
	if path == "" {
		path = orch.Config.OutputDir + "/briefing.json"
	}
	briefing, err := loadBriefing(path)
	if err != nil {
		slog.Error("load briefing", "err", err)
		return 1
	}
	if err := orch.Synthesize(context.Background(), briefing); err != nil {
		slog.Error("synthesize", "err", err)
		return 1
	}
	return 0
}

func jtbdCommand(args []string) int {
	fs := flag.NewFlagSet("jtbd", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yaml", "pipeline config path")
	transcriptPath := fs.String("transcript", "", "interview transcript file")
	outPath := fs.String("out", "", "output json path")
	_ = fs.Parse(args)
	if *transcriptPath == "" {
		slog.Error("jtbd requires --transcript")
		return 1
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "err", err)
		return 1
	}
	transcript, err := os.ReadFile(*transcriptPath)
	if err != nil {
		slog.Error("read transcript", "err", err)
		return 1
	}
	llm := xai.NewClient(cfg.Qual.XAIBaseURL, os.Getenv("XAI_API_KEY"), cfg.Qual.XAIModel)
	analyzer := &jtbd.Analyzer{LLM: llm}
	result, err := analyzer.Analyze(context.Background(), string(transcript))
	if err != nil {
		slog.Error("jtbd", "err", err)
		return 1
	}
	if *outPath != "" {
		if err := jtbd.WriteResult(*outPath, result); err != nil {
			slog.Error("write result", "err", err)
			return 1
		}
	} else {
		fmt.Printf("Job: %s\nCircumstance: %s\nStruggle: %s\nDesired outcome: %s\n",
			result.Job, result.Circumstance, result.Struggle, result.DesiredOutcome)
	}
	return 0
}

func writeOnly(orch *pipeline.Orchestrator, briefing domain.Briefing, openData, issues []domain.Insight) int {
	if err := orch.WriteArtifacts(briefing, openData, issues); err != nil {
		slog.Error("write artifacts", "err", err)
		return 1
	}
	return 0
}

func loadInsights(explicit, fallback string) ([]domain.Insight, error) {
	path := explicit
	if path == "" {
		path = fallback
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []domain.Insight
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func loadBriefing(path string) (domain.Briefing, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return domain.Briefing{}, err
	}
	var briefing domain.Briefing
	if err := json.Unmarshal(b, &briefing); err != nil {
		return domain.Briefing{}, err
	}
	return briefing, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
