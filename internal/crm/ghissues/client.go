package ghissues

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	HTTP     *http.Client
	Token    string
	Owner    string
	Repo     string
	BaseURL  string
}

func NewClient(token, owner, repo string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 60 * time.Second},
		Token:   token,
		Owner:   owner,
		Repo:    repo,
		BaseURL: "https://api.github.com",
	}
}

type Issue struct {
	Number int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func (c *Client) CreateDigest(ctx context.Context, title, body string, labels []string) (*Issue, error) {
	return c.createIssue(ctx, title, body, labels)
}

func (c *Client) UpsertSignal(ctx context.Context, title, body string, fingerprint string, labels []string) (*Issue, error) {
	existing, err := c.findByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, err
	}
	body = ensureFingerprint(body, fingerprint)
	if existing != nil {
		return c.updateIssue(ctx, existing.Number, title, body)
	}
	return c.createIssue(ctx, title, body, labels)
}

func (c *Client) findByFingerprint(ctx context.Context, fingerprint string) (*Issue, error) {
	q := fmt.Sprintf(`repo:%s/%s in:body "%s"`, c.Owner, c.Repo, fingerprintMarker(fingerprint))
	reqURL := fmt.Sprintf("%s/search/issues?q=%s", c.BaseURL, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search issues: status %d %s", resp.StatusCode, string(raw))
	}
	var result struct {
		Items []Issue `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	return &result.Items[0], nil
}

func (c *Client) createIssue(ctx context.Context, title, body string, labels []string) (*Issue, error) {
	payload := map[string]any{
		"title":  title,
		"body":   body,
		"labels": labels,
	}
	var issue Issue
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("%s/repos/%s/%s/issues", c.BaseURL, c.Owner, c.Repo), payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) updateIssue(ctx context.Context, number int, title, body string) (*Issue, error) {
	payload := map[string]any{
		"title": title,
		"body":  body,
	}
	var issue Issue
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.BaseURL, c.Owner, c.Repo, number)
	if err := c.doJSON(ctx, http.MethodPatch, url, payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) doJSON(ctx context.Context, method, url string, payload any, dest any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: status %d %s", method, url, resp.StatusCode, string(raw))
	}
	if dest != nil && len(raw) > 0 {
		return json.Unmarshal(raw, dest)
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func FingerprintComment(fingerprint string) string {
	return fmt.Sprintf("<!-- fingerprint:%s -->", fingerprint)
}

func fingerprintMarker(fingerprint string) string {
	return "fingerprint:" + fingerprint
}

func ensureFingerprint(body, fingerprint string) string {
	marker := FingerprintComment(fingerprint)
	if strings.Contains(body, marker) {
		return body
	}
	return marker + "\n\n" + body
}

func AppendOutreachDraft(body, draft string) string {
	if strings.Contains(body, "## Outreach draft") {
		return body
	}
	return body + "\n\n## Outreach draft\n\n" + draft + "\n"
}
