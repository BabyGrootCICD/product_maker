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
	HTTP     *http.Client
	BaseURL  string
	APIKey   string
	Model    string
}

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []chatMessage   `json:"messages"`
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

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 120 * time.Second},
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
	}
}

func (c *Client) GenerateOutreach(ctx context.Context, title, body string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("xai: missing XAI_API_KEY")
	}
	prompt := fmt.Sprintf(`你是一位技術型產品創辦人。你看到一個開發者在 GitHub 上抱怨：
Title: %s
Content: %s

請寫一段簡短、專業且帶有極客風格的 LinkedIn 破冰訊息。
目標：表達對他痛點的共鳴，並邀請他進行 15 分鐘的短暫視訊訪談，以協助設計新架構。
限制：不要像推銷，語氣要像是工程師之間的交流。`, title, truncate(body, 4000))

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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
