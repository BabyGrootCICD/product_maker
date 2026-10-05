package naics541519

import (
	"testing"
	"time"
)

func TestBuildQuery(t *testing.T) {
	searcher := NewSearcher()
	filters := map[string]string{
		"agency":   "DOD",
		"set_aside": "Small Business",
	}
	query := searcher.BuildQuery("software development", filters)
	if query.NAICSCode != "541519" {
		t.Fatalf("expected 541519, got %s", query.NAICSCode)
	}
	if query.Keyword != "software development" {
		t.Fatalf("expected 'software development', got %s", query.Keyword)
	}
	if query.Agency != "DOD" {
		t.Fatalf("expected 'DOD', got %s", query.Agency)
	}
}

func TestFormatQuery(t *testing.T) {
	searcher := NewSearcher()
	query := SearchQuery{
		NAICSCode: "541519",
		Keyword:   "cloud",
		Agency:    "DOD",
	}
	result := searcher.FormatQuery(query)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestParseOpportunity(t *testing.T) {
	searcher := NewSearcher()
	data := map[string]interface{}{
		"title":        "Software Development",
		"description":  "Web application development",
		"agency":       "DOD",
		"naics_code":   "541519",
		"set_aside":    "Small Business",
		"award_amount": "$500,000",
		"link":         "https://sam.gov/opp/123",
	}
	opp := searcher.ParseOpportunity(data)
	if opp.Title != "Software Development" {
		t.Fatalf("expected 'Software Development', got %s", opp.Title)
	}
	if opp.Agency != "DOD" {
		t.Fatalf("expected 'DOD', got %s", opp.Agency)
	}
	if opp.NAICSCode != "541519" {
		t.Fatalf("expected '541519', got %s", opp.NAICSCode)
	}
}

func TestFormatOpportunity(t *testing.T) {
	searcher := NewSearcher()
	opp := Opportunity{
		Title:    "Software Development",
		Agency:   "DOD",
		NAICSCode: "541519",
		Link:     "https://sam.gov/opp/123",
	}
	result := searcher.FormatOpportunity(opp)
	if result == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestSortByDate(t *testing.T) {
	opps := []Opportunity{
		{Title: "Old", PostedDate: time.Now().Add(-24 * time.Hour)},
		{Title: "New", PostedDate: time.Now()},
	}
	sorted := SortByDate(opps)
	if sorted[0].Title != "New" {
		t.Fatalf("expected 'New' first, got %s", sorted[0].Title)
	}
}

func TestFilterByAgency(t *testing.T) {
	opps := []Opportunity{
		{Title: "DOD", Agency: "DOD"},
		{Title: "NASA", Agency: "NASA"},
	}
	filtered := FilterByAgency(opps, "DOD")
	if len(filtered) != 1 {
		t.Fatalf("expected 1, got %d", len(filtered))
	}
}

func TestFormatResults(t *testing.T) {
	result := SearchResult{
		Opportunities: []Opportunity{{Title: "Test"}},
		TotalCount:    1,
		Page:          1,
		PageSize:      25,
	}
	json := FormatResults(result)
	if json == "" {
		t.Fatal("expected non-empty string")
	}
}

func BenchmarkBuildQuery(b *testing.B) {
	searcher := NewSearcher()
	filters := map[string]string{"agency": "DOD"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.BuildQuery("software", filters)
	}
}
