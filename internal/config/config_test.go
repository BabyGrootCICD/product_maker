package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPipelineConfig(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "pipeline.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OpenData.PCC.FailureRateThreshold != 0.15 {
		t.Fatalf("unexpected pcc threshold %v", cfg.OpenData.PCC.FailureRateThreshold)
	}
	if cfg.Converge.TopN != 10 {
		t.Fatalf("unexpected converge top_n %d", cfg.Converge.TopN)
	}
}

func TestLoadReposConfig(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "repos.yaml")
	repos, err := LoadRepos(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos.Repos) < 5 {
		t.Fatalf("expected at least 5 repos, got %d", len(repos.Repos))
	}
}

func TestResolveReposPath(t *testing.T) {
	cfg := &Config{Issues: IssuesConfig{ReposFile: "configs/repos.yaml"}}
	if _, err := os.Stat("configs/repos.yaml"); err == nil {
		got := ResolveReposPath(cfg, "configs")
		if got != "configs/repos.yaml" {
			t.Fatalf("unexpected path %s", got)
		}
	}
}
