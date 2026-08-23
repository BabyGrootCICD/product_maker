package xai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateOutreachMissingKey(t *testing.T) {
	c := NewClient("http://example", "", "grok-4-1-fast")
	_, err := c.GenerateOutreach(context.Background(), "title", "body")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestGenerateOutreachSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "Hey, saw your issue..."}},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", "grok-4-1-fast")
	out, err := c.GenerateOutreach(context.Background(), "Cache bug", "details")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected outreach draft")
	}
}

func TestRateTaskAxes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{
					"role":    "assistant",
					"content": `{"difficulty":4,"risk":2,"rationale":"需跨團隊協調"}`,
				}},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", "grok-4-1-fast")
	axes, err := c.RateTaskAxes(context.Background(), "t", "s", "issues", "https://x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if axes.Difficulty != 4 || axes.Risk != 2 {
		t.Fatalf("unexpected axes %+v", axes)
	}
}

func TestGenerateSolutionPaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "## Path 1\nApproach: ..."}},
			},
		})
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "key", "m")
	out, err := c.GenerateSolutionPaths(context.Background(), "t", "s", "issues", "u", 1.2)
	if err != nil || out == "" {
		t.Fatalf("err=%v out=%q", err, out)
	}
}
