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
