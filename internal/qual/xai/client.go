package xai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	APIKey  string
	Model   string
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type TaskAxes struct {
	Difficulty int    `json:"difficulty"`
	Risk       int    `json:"risk"`
	Rationale  string `json:"rationale"`
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 120 * time.Second},
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
	}
}

func (c *Client) GenerateOutreach(ctx context.Context, title, body string) (string, error) {
	prompt := fmt.Sprintf(`你是一位技術型產品創辦人。你看到一個開發者在 GitHub 上抱怨：
Title: %s
Content: %s

請寫一段簡短、專業且帶有極客風格的 LinkedIn 破冰訊息。
目標：表達對他痛點的共鳴，並邀請他進行 15 分鐘的短暫視訊訪談，以協助設計新架構。
限制：不要像推銷，語氣要像是工程師之間的交流。`, title, truncate(body, 4000))
	return c.chat(ctx, prompt)
}

func (c *Client) RateTaskAxes(ctx context.Context, title, summary, pipeline, url string, metrics map[string]float64) (TaskAxes, error) {
	metricsJSON, _ := json.Marshal(metrics)
	prompt := fmt.Sprintf(`你是資深技術創辦人，評估一項探索任務的實作難度與風險。
只輸出 JSON（不要 markdown code fence），格式：
{"difficulty":1-5,"risk":1-5,"rationale":"<=200字中文"}

difficulty：1=可快速驗證，5=需大量工程／組織協調
risk：1=失敗成本低，5=方向錯了代價高或訊號嘈雜

Title: %s
Pipeline: %s
URL: %s
Summary: %s
Metrics: %s`, title, pipeline, url, truncate(summary, 1500), string(metricsJSON))

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return TaskAxes{}, err
	}
	raw = stripFence(raw)
	var axes TaskAxes
	if err := json.Unmarshal([]byte(raw), &axes); err != nil {
		return TaskAxes{}, fmt.Errorf("parse axes json: %w", err)
	}
	return axes, nil
}

func (c *Client) GenerateSolutionPaths(ctx context.Context, title, summary, pipeline, url string, priority float64) (string, error) {
	prompt := fmt.Sprintf(`你是資深軟體架構師。針對下列探索任務，產出 2–3 條可執行的 recommended solution paths。
用 markdown，每條包含：Approach / Why / Steps / Risks / Done-when。
語氣務實，禁止空泛口號；步驟要能對應 upstream issue。

Title: %s
Pipeline: %s
Priority: %.2f
URL: %s
Summary: %s`, title, pipeline, priority, url, truncate(summary, 2000))
	return c.chat(ctx, prompt)
}

func (c *Client) chat(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("xai: missing XAI_API_KEY")
	}
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 402 || resp.StatusCode == 429 {
		return "", fmt.Errorf("xai: rate limited or payment required (%d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("xai: status %d %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("xai: empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func stripFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
