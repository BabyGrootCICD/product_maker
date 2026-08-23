package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutputDir string         `yaml:"output_dir"`
	OpenData  OpenDataConfig `yaml:"opendata"`
	Issues    IssuesConfig   `yaml:"issues"`
	Converge  ConvergeConfig `yaml:"converge"`
	CRM       CRMConfig      `yaml:"crm"`
	Qual      QualConfig     `yaml:"qual"`
	Tasks     TasksConfig    `yaml:"tasks"`
}

type TasksConfig struct {
	Path        string `yaml:"path"`
	AxesTopN    int    `yaml:"axes_top_n"`
	BranchTopN  int    `yaml:"branch_top_n"`
}

type OpenDataConfig struct {
	PCC    PCCConfig    `yaml:"pcc"`
	SAMGov SAMGovConfig `yaml:"samgov"`
}

type PCCConfig struct {
	Enabled                bool     `yaml:"enabled"`
	BaseURL                string   `yaml:"base_url"`
	LookbackDays           int      `yaml:"lookback_days"`
	MinCases               int      `yaml:"min_cases"`
	FailureRateThreshold   float64  `yaml:"failure_rate_threshold"`
	Categories             []string `yaml:"categories"`
}

type SAMGovConfig struct {
	Enabled              bool     `yaml:"enabled"`
	BaseURL              string   `yaml:"base_url"`
	LookbackDays         int      `yaml:"lookback_days"`
	MaxRequests          int      `yaml:"max_requests"`
	Limit                int      `yaml:"limit"`
	MinCases             int      `yaml:"min_cases"`
	PainProxyThreshold   float64  `yaml:"pain_proxy_threshold"`
	NAICS                []string `yaml:"naics"`
}

type IssuesConfig struct {
	MinComments   int               `yaml:"min_comments"`
	MinThumbsUp   int               `yaml:"min_thumbs_up"`
	MaxComments   int               `yaml:"max_comments"`
	LabelBonus    map[string]int    `yaml:"label_bonus"`
	TopKBodies    int               `yaml:"top_k_bodies"`
	ReposFile     string            `yaml:"repos_file"`
}

type ConvergeConfig struct {
	TopN             int     `yaml:"top_n"`
	UnmatchedTopN    int     `yaml:"unmatched_top_n"`
	JaccardThreshold float64 `yaml:"jaccard_threshold"`
}

type CRMConfig struct {
	DigestLabel string `yaml:"digest_label"`
	ReviewLabel string `yaml:"review_label"`
	SignalLabel string `yaml:"signal_label"`
}

type QualConfig struct {
	XAIBaseURL          string `yaml:"xai_base_url"`
	XAIModel            string `yaml:"xai_model"`
	MaxOutreachPerRun   int    `yaml:"max_outreach_per_run"`
}

type RepoEntry struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
	Theme string `yaml:"theme"`
}

type ReposConfig struct {
	Repos []RepoEntry `yaml:"repos"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.OutputDir == "" {
		c.OutputDir = "out"
	}
	if c.OpenData.PCC.LookbackDays <= 0 {
		c.OpenData.PCC.LookbackDays = 7
	}
	if c.OpenData.SAMGov.LookbackDays <= 0 {
		c.OpenData.SAMGov.LookbackDays = 7
	}
	if c.Converge.TopN <= 0 {
		c.Converge.TopN = 10
	}
	if c.Converge.UnmatchedTopN <= 0 {
		c.Converge.UnmatchedTopN = 3
	}
	if c.Issues.TopKBodies <= 0 {
		c.Issues.TopKBodies = 8
	}
	if c.Qual.MaxOutreachPerRun <= 0 {
		c.Qual.MaxOutreachPerRun = 10
	}
	if c.Tasks.Path == "" {
		c.Tasks.Path = "tasks.md"
	}
	if c.Tasks.AxesTopN <= 0 {
		c.Tasks.AxesTopN = 15
	}
	if c.Tasks.BranchTopN <= 0 {
		c.Tasks.BranchTopN = 3
	}
	return nil
}

func LoadRepos(path string) (*ReposConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read repos config: %w", err)
	}
	var cfg ReposConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse repos config: %w", err)
	}
	if len(cfg.Repos) == 0 {
		return nil, fmt.Errorf("repos config: no repos defined")
	}
	return &cfg, nil
}

func ResolveReposPath(cfg *Config, configDir string) string {
	if cfg.Issues.ReposFile == "" {
		return ""
	}
	if _, err := os.Stat(cfg.Issues.ReposFile); err == nil {
		return cfg.Issues.ReposFile
	}
	return configDir + "/" + cfg.Issues.ReposFile
}
