package samgov

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/config"
)

const fixtureSAM = `{
  "opportunitiesData": [
    {"type":"Solicitation","naicsCode":"541512","noticeType":"o"},
    {"type":"Sources Sought","naicsCode":"541512","noticeType":"r"},
    {"type":"Special Notice","naicsCode":"541512","noticeType":"s"},
    {"type":"Award Notice","naicsCode":"541512","noticeType":"a"}
  ]
}`

func TestRunPainProxy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixtureSAM))
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"?", "test-key")
	cfg := config.SAMGovConfig{
		LookbackDays:         7,
		MaxRequests:          1,
		Limit:                100,
		MinCases:             3,
		PainProxyThreshold:   0.4,
		NAICS:                []string{"541512"},
	}

	insights, err := client.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Metrics["pain_proxy"] != 0.5 {
		t.Fatalf("unexpected pain_proxy %v", insights[0].Metrics["pain_proxy"])
	}
}
