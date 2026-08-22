package samgov

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/opendata/detect"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	APIKey  string
}

type opportunity struct {
	Type       string `json:"type"`
	NAICSCode  string `json:"naicsCode"`
	NoticeType string `json:"noticeType"`
}

type searchResponse struct {
	OpportunitiesData []opportunity `json:"opportunitiesData"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 90 * time.Second},
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *Client) Run(ctx context.Context, cfg config.SAMGovConfig) ([]domain.Insight, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("samgov: missing SAM_API_KEY")
	}

	to := time.Now()
	from := to.AddDate(0, 0, -cfg.LookbackDays)
	stats := make(map[string]*detect.NAICSStats)
	requests := 0

	for _, naics := range cfg.NAICS {
		if requests >= cfg.MaxRequests {
			break
		}
		records, err := c.search(ctx, naics, from, to, cfg.Limit)
		requests++
		if err != nil {
			return nil, err
		}
		s, ok := stats[naics]
		if !ok {
			s = &detect.NAICSStats{NAICS: naics}
			stats[naics] = s
		}
		for _, o := range records {
			s.Total++
			if isSourcesSoughtOrSpecial(o) {
				s.SourcesSoughtOrSpecial++
			}
		}
	}

	var naicsStats []detect.NAICSStats
	for _, s := range stats {
		if s.Total > 0 {
			s.PainProxy = float64(s.SourcesSoughtOrSpecial) / float64(s.Total)
		}
		naicsStats = append(naicsStats, *s)
	}

	return detect.FromNAICSStats(naicsStats, detect.SAMDetectConfig{
		MinCases:           cfg.MinCases,
		PainProxyThreshold: cfg.PainProxyThreshold,
		Theme:              "gov-procurement",
	}), nil
}

func (c *Client) search(ctx context.Context, naics string, from, to time.Time, limit int) ([]opportunity, error) {
	params := url.Values{}
	params.Set("api_key", c.APIKey)
	params.Set("postedFrom", from.Format("01/02/2006"))
	params.Set("postedTo", to.Format("01/02/2006"))
	params.Set("ncode", naics)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", "0")

	reqURL := c.BaseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("samgov search naics=%s: status %d", naics, resp.StatusCode)
	}
	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse samgov response: %w", err)
	}
	return parsed.OpportunitiesData, nil
}

func isSourcesSoughtOrSpecial(o opportunity) bool {
	text := strings.ToLower(o.Type)
	if strings.Contains(text, "sources sought") || strings.Contains(text, "special notice") {
	 return true
	}
	nt := strings.ToLower(strings.TrimSpace(o.NoticeType))
	return nt == "r" || nt == "s"
}
