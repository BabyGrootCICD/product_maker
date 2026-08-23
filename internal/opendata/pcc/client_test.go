package pcc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/config"
)

const fixtureTender = `[
  {"category":"工程類","type":"招標公告","name":"道路改善工程"},
  {"category":"工程類","type":"無法決標公告","name":"道路改善工程流標"},
  {"category":"工程類","type":"無法決標公告","name":"橋樑維護流標"},
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
		Categories:           []string{"工程類", "勞務類", "財物類"},
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

func TestParseRecordsEmptyArray(t *testing.T) {
	records, err := parseRecords([]byte("[]"))
	if err != nil {
		t.Fatalf("empty array should not error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestParseRecordsWrappedObject(t *testing.T) {
	body := `{"records":[{"category":"勞務類","type":"招標公告","name":"x"}]}`
	records, err := parseRecords([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestFindLatestDataDaySkipsEmpty(t *testing.T) {
	emptyUntil := time.Now().AddDate(0, 0, -40).Format("2006-01-02")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		parts := strings.Split(r.URL.Path, "/")
		date := parts[len(parts)-1]
		if date > emptyUntil {
			_, _ = w.Write([]byte("[]"))
			return
		}
		_, _ = w.Write([]byte(fixtureTender))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	day, err := client.findLatestDataDay(context.Background(), time.Now(), 120)
	if err != nil {
		t.Fatal(err)
	}
	if day.IsZero() {
		t.Fatal("expected a non-zero anchor day")
	}
}
