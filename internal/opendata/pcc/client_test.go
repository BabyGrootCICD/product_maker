package pcc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/config"
)

const fixtureTender = `[
  {"category":"資訊服務","type":"招標公告","name":"資訊系統建置"},
  {"category":"資訊服務","type":"無法決標公告","name":"資訊系統建置流標"},
  {"category":"資訊服務","type":"無法決標公告","name":"資訊系統維護流標"},
  {"category":"勞務類","type":"招標公告","name":"勞務採購"}
]`

func TestRunAggregatesFailureRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixtureTender))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	cfg := config.PCCConfig{
		LookbackDays:         1,
		MinCases:             2,
		FailureRateThreshold: 0.4,
		Categories:           []string{"資訊服務", "勞務類", "財物類"},
	}

	insights, err := client.Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Metrics["failure_rate"] < 0.4 {
		t.Fatalf("unexpected failure_rate %v", insights[0].Metrics["failure_rate"])
	}
}
