package jtbd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
)

type Result struct {
	Job             string `json:"job"`
	Circumstance    string `json:"circumstance"`
	Struggle        string `json:"struggle"`
	DesiredOutcome  string `json:"desired_outcome"`
}

type Analyzer struct {
	LLM *xai.Client
}

func (a *Analyzer) Analyze(ctx context.Context, transcript string) (*Result, error) {
	if a.LLM == nil || a.LLM.APIKey == "" {
		return nil, fmt.Errorf("jtbd: missing XAI_API_KEY")
	}
	prompt := fmt.Sprintf(`分析以下訪談逐字稿，輸出 JSON（僅 JSON，不要 markdown）：
{
  "job": "...",
  "circumstance": "...",
  "struggle": "...",
  "desired_outcome": "..."
}

逐字稿：
%s`, transcript)

	raw, err := a.LLM.GenerateOutreach(ctx, "JTBD analysis", prompt)
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result Result
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse jtbd json: %w", err)
	}
	return &result, nil
}

func WriteResult(path string, result *Result) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
