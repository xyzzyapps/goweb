package fill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	defaultOpenAIBase = "https://api.openai.com/v1"
	defaultXAIBase    = "https://api.x.ai/v1"
)

type Endpoint struct {
	BaseURL string
	APIKey  string
	Model   string
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func ResolveEndpoint(modelFlag string) (Endpoint, error) {
	key := firstEnv("GOWEB_API_KEY", "OPENAI_API_KEY", "XAI_API_KEY")
	if key == "" {
		return Endpoint{}, fmt.Errorf("set OPENAI_API_KEY (or GOWEB_API_KEY / XAI_API_KEY)")
	}
	base := firstEnv("GOWEB_BASE_URL", "OPENAI_BASE_URL", "XAI_BASE_URL")
	if base == "" {
		if os.Getenv("XAI_API_KEY") != "" && os.Getenv("OPENAI_API_KEY") == "" {
			base = defaultXAIBase
		} else {
			base = defaultOpenAIBase
		}
	}
	base = strings.TrimRight(base, "/")
	model := modelFlag
	if model == "" {
		model = firstEnv("GOWEB_MODEL", "OPENAI_MODEL")
	}
	if model == "" {
		model = "gpt-4o"
	}
	return Endpoint{BaseURL: base, APIKey: key, Model: model}, nil
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []chatMsg `json:"messages"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func complete(ep Endpoint, user string) (string, error) {
	payload, err := json.Marshal(chatReq{
		Model: ep.Model,
		Messages: []chatMsg{
			{Role: "system", Content: "You fill literate-programming holes. Return only source text."},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", ep.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+ep.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm HTTP %d: %s", resp.StatusCode, truncate(string(raw), 400))
	}
	var out chatResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("llm: no choices")
	}
	return contentString(out.Choices[0].Message.Content)
}

func contentString(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", fmt.Errorf("llm content: %w", err)
	}
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.Text)
	}
	return b.String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
