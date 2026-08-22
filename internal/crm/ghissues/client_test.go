package ghissues

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpsertSignalCreates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/search/issues"):
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/issues"):
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 1, "html_url": "https://github.com/o/r/issues/1", "title": "t", "body": "b"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &Client{HTTP: srv.Client(), Token: "tok", Owner: "o", Repo: "r", BaseURL: srv.URL}
	_, err := client.UpsertSignal(context.Background(), "Signal", "body", "issues:actions/runner#1", []string{"signal"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpsertSignalUpdatesExisting(t *testing.T) {
	fp := "issues:actions/runner#1"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/search/issues"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{{"number": 1, "html_url": "https://github.com/o/r/issues/1", "title": "t", "body": FingerprintComment(fp)}},
			})
		case r.Method == http.MethodPatch:
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 1, "html_url": "https://github.com/o/r/issues/1", "title": "t2", "body": "b2"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &Client{HTTP: srv.Client(), Token: "tok", Owner: "o", Repo: "r", BaseURL: srv.URL}
	_, err := client.UpsertSignal(context.Background(), "Signal updated", "new body", fp, []string{"signal"})
	if err != nil {
		t.Fatal(err)
	}
}
