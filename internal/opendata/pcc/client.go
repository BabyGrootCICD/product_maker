package pcc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/opendata/detect"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

type Record struct {
	Category string `json:"category"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

func recordTitle(r Record) string {
	if r.Title != "" {
		return r.Title
	}
	return r.Name
}

func NewClient(baseURL string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 60 * time.Second},
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) Run(ctx context.Context, cfg config.PCCConfig) ([]domain.Insight, error) {
	stats := make(map[string]*detect.CategoryStats)
	now := time.Now()

	for d := 0; d < cfg.LookbackDays; d++ {
		date := now.AddDate(0, 0, -d).Format("2006-01-02")
		records, err := c.fetchDate(ctx, date)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			cat := r.Category
			if cat == "" {
				cat = inferCategory(recordTitle(r))
			}
			if !categoryAllowed(cat, cfg.Categories) {
				continue
			}
			s, ok := stats[cat]
			if !ok {
				s = &detect.CategoryStats{Category: cat}
				stats[cat] = s
			}
			s.TotalCases++
			if isFailed(r) {
				s.FailedCases++
			}
		}
	}

	var categoryStats []detect.CategoryStats
	for _, s := range stats {
		if s.TotalCases > 0 {
			s.FailureRate = float64(s.FailedCases) / float64(s.TotalCases)
		}
		categoryStats = append(categoryStats, *s)
	}

	return detect.FromCategoryStats(categoryStats, detect.DetectConfig{
		MinCases:             cfg.MinCases,
		FailureRateThreshold: cfg.FailureRateThreshold,
		Source:               "pcc",
		Theme:                "gov-procurement",
	}), nil
}

func (c *Client) fetchDate(ctx context.Context, date string) ([]Record, error) {
	url := fmt.Sprintf("%s/api/date/tender/%s", c.BaseURL, date)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("pcc api %s: status %d", date, resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pcc api %s: status %d body=%s", date, resp.StatusCode, truncate(string(body), 200))
	}
	return parseRecords(body)
}

func parseRecords(body []byte) ([]Record, error) {
	var direct []Record
	if err := json.Unmarshal(body, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}
	var wrapped struct {
		Records []Record `json:"records"`
		Data    []Record `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("parse pcc records: %w", err)
	}
	if len(wrapped.Records) > 0 {
		return wrapped.Records, nil
	}
	return wrapped.Data, nil
}

func isFailed(r Record) bool {
	text := strings.ToLower(r.Type + " " + r.Status + " " + recordTitle(r))
	return strings.Contains(text, "無法決標") ||
		strings.Contains(text, "流標") ||
		strings.Contains(text, "failed") ||
		strings.Contains(text, "廢標")
}

func inferCategory(title string) string {
	switch {
	case strings.Contains(title, "資訊"):
		return "資訊服務"
	case strings.Contains(title, "勞務"):
		return "勞務類"
	case strings.Contains(title, "財物"):
		return "財物類"
	default:
		return "其他"
	}
}

func categoryAllowed(cat string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if strings.Contains(cat, a) || strings.Contains(a, cat) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
