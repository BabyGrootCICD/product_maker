package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/config"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
	"github.com/BabyGrootCICD/product_maker/internal/issues/score"
)

const graphQLURL = "https://api.github.com/graphql"

var graphQLEndpoint = graphQLURL

type Client struct {
	HTTP  *http.Client
	Token string
}

func NewClient(token string) *Client {
	return &Client{
		HTTP:  &http.Client{Timeout: 120 * time.Second},
		Token: token,
	}
}

type listQuery struct {
	Repository struct {
		Issues struct {
			Nodes []struct {
				Number    int    `json:"number"`
				Title     string `json:"title"`
				URL       string `json:"url"`
				Comments  struct {
					TotalCount int `json:"totalCount"`
				} `json:"comments"`
				Reactions struct {
					TotalCount int `json:"totalCount"`
				} `json:"reactions"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
			} `json:"nodes"`
		} `json:"issues"`
	} `json:"repository"`
}

type bodyQuery struct {
	Repository struct {
		Issue struct {
			Body string `json:"body"`
		} `json:"issue"`
	} `json:"repository"`
}

type issueNode struct {
	Number   int
	Title    string
	URL      string
	Comments int
	ThumbsUp int
	Labels   []string
}

func (c *Client) Run(ctx context.Context, cfg *config.Config, repos *config.ReposConfig) ([]domain.Insight, error) {
	scoreCfg := score.Config{
		MinComments: cfg.Issues.MinComments,
		MinThumbsUp: cfg.Issues.MinThumbsUp,
		MaxComments: cfg.Issues.MaxComments,
		LabelBonus:  cfg.Issues.LabelBonus,
	}

	var candidates []score.IssueCandidate
	for _, repo := range repos.Repos {
		nodes, err := c.listIssues(ctx, repo.Owner, repo.Name)
		if err != nil {
			return nil, fmt.Errorf("list issues %s/%s: %w", repo.Owner, repo.Name, err)
		}
		for _, n := range nodes {
			candidates = append(candidates, score.IssueCandidate{
				Owner:    repo.Owner,
				Name:     repo.Name,
				Theme:    repo.Theme,
				Number:   n.Number,
				Title:    n.Title,
				URL:      n.URL,
				Comments: n.Comments,
				ThumbsUp: n.ThumbsUp,
				Labels:   n.Labels,
			})
		}
	}

	insights := score.FilterAndScore(candidates, scoreCfg)
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].Score > insights[j].Score
	})

	topK := cfg.Issues.TopKBodies
	if topK > len(insights) {
		topK = len(insights)
	}
	for i := 0; i < topK; i++ {
		parts := strings.Split(strings.TrimPrefix(insights[i].Fingerprint, "issues:"), "#")
		if len(parts) != 2 {
			continue
		}
		repoParts := strings.SplitN(parts[0], "/", 2)
		if len(repoParts) != 2 {
			continue
		}
		var num int
		_, _ = fmt.Sscanf(parts[1], "%d", &num)
		body, err := c.fetchIssueBody(ctx, repoParts[0], repoParts[1], num)
		if err == nil {
			insights[i].Body = body
		}
	}

	return insights, nil
}

func (c *Client) listIssues(ctx context.Context, owner, name string) ([]issueNode, error) {
	query := `
query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    issues(states: OPEN, first: 50, orderBy: {field: COMMENTS, direction: DESC}) {
      nodes {
        number
        title
        url
        comments { totalCount }
        reactions(content: THUMBS_UP) { totalCount }
        labels(first: 10) { nodes { name } }
      }
    }
  }
}`
	var resp listQuery
	if err := c.graphQL(ctx, query, map[string]any{"owner": owner, "name": name}, &resp); err != nil {
		return nil, err
	}
	var nodes []issueNode
	for _, n := range resp.Repository.Issues.Nodes {
		var labels []string
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		nodes = append(nodes, issueNode{
			Number:   n.Number,
			Title:    n.Title,
			URL:      n.URL,
			Comments: n.Comments.TotalCount,
			ThumbsUp: n.Reactions.TotalCount,
			Labels:   labels,
		})
	}
	return nodes, nil
}

func (c *Client) fetchIssueBody(ctx context.Context, owner, name string, number int) (string, error) {
	query := `
query($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    issue(number: $number) {
      body
    }
  }
}`
	var resp bodyQuery
	if err := c.graphQL(ctx, query, map[string]any{"owner": owner, "name": name, "number": number}, &resp); err != nil {
		return "", err
	}
	return resp.Repository.Issue.Body, nil
}

func (c *Client) graphQL(ctx context.Context, query string, variables map[string]any, dest any) error {
	payload := map[string]any{"query": query, "variables": variables}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql status %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("graphql: %s", envelope.Errors[0].Message)
	}
	return json.Unmarshal(envelope.Data, dest)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
