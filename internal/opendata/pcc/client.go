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

	// Community mirrors can lag months behind wall-clock date.
	// Coarse-then-refine search avoids hundreds of empty daily fetches.
	anchor, err := c.findLatestDataDay(ctx, time.Now(), 400)
	if err != nil {
		return nil, err
	}
	if anchor.IsZero() {
		return nil, nil
	}

	for d := 0; d < cfg.LookbackDays; d++ {
		date := anchor.AddDate(0, 0, -d).Format("2006-01-02")
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

// findLatestDataDay returns the most recent day (from start, walking back)
// with at least one tender record. Returns zero time if none found within maxScanDays.
//
// Strategy: probe recent days daily, then sample monthly to tolerate mirror lag,
// then walk forward within the month to recover the true latest day.
func (c *Client) findLatestDataDay(ctx context.Context, start time.Time, maxScanDays int) (time.Time, error) {
	if maxScanDays <= 0 {
		maxScanDays = 400
	}

	for d := 0; d < 7 && d < maxScanDays; d++ {
		day := start.AddDate(0, 0, -d)
		ok, err := c.hasData(ctx, day)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			return day, nil
		}
	}

	var hit time.Time
	for d := 7; d < maxScanDays; d += 30 {
		day := start.AddDate(0, 0, -d)
		ok, err := c.hasData(ctx, day)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			hit = day
			break
		}
	}
	if hit.IsZero() {
		return time.Time{}, nil
	}

	// Walk forward from the monthly sample; stop after a short empty streak so we
	// do not burn requests past the mirror's latest ingest day.
	latest := hit
	emptyStreak := 0
	for i := 1; i <= 29; i++ {
		day := hit.AddDate(0, 0, i)
		if day.After(start) {
			break
		}
		ok, err := c.hasData(ctx, day)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			latest = day
			emptyStreak = 0
			continue
		}
		emptyStreak++
		if emptyStreak >= 3 {
			break
		}
	}
	return latest, nil
}

func (c *Client) hasData(ctx context.Context, day time.Time) (bool, error) {
	records, err := c.fetchDate(ctx, day.Format("2006-01-02"))
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
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
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	// Prefer array responses (including empty []). Empty days are valid — do not
	// fall through to object parsing, which fails with "cannot unmarshal array".
	if strings.HasPrefix(trimmed, "[") {
		var direct []Record
		if err := json.Unmarshal(body, &direct); err != nil {
			return nil, fmt.Errorf("parse pcc records: %w", err)
		}
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
