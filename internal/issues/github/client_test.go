package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/config"
)

func TestRunFiltersHighPainIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"issues": map[string]any{
						"nodes": []map[string]any{
							{
								"number": 42,
								"title":  "Runner cache corruption",
								"url":    "https://github.com/actions/runner/issues/42",
								"comments": map[string]any{"totalCount": 15},
								"reactions": map[string]any{"totalCount": 8},
								"labels": map[string]any{
									"nodes": []map[string]any{{"name": "wontfix"}},
								},
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	// Override graphQL URL by using a custom transport - simpler: test score path via Run with injected client
	// For unit test, we patch by making Client use test server - need to refactor graphQL URL
	// Use internal test hook: set graphQLEndpoint
	old := graphQLEndpoint
	graphQLEndpoint = srv.URL
	defer func() { graphQLEndpoint = old }()

	client := NewClient("test-token")
	cfg := &config.Config{
		Issues: config.IssuesConfig{
			MinComments: 10,
			MinThumbsUp: 5,
			MaxComments: 500,
			LabelBonus:  map[string]int{"wontfix": 5},
			TopKBodies:  0,
		},
	}
	repos := &config.ReposConfig{
		Repos: []config.RepoEntry{{Owner: "actions", Name: "runner", Theme: "devtools"}},
	}

	insights, err := client.Run(context.Background(), cfg, repos)
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Fingerprint != "issues:actions/runner#42" {
		t.Fatalf("unexpected fingerprint %s", insights[0].Fingerprint)
	}
}
