package ai

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

// OpenAICompatibleProvider calls any OpenAI-compatible /chat/completions
// endpoint. Base URL, API key and model are all configurable.
type OpenAICompatibleProvider struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// NewOpenAICompatibleProvider constructs a provider using the supplied
// settings. baseURL should NOT include a trailing slash.
func NewOpenAICompatibleProvider(baseURL, apiKey, model string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenAICompatibleProvider) Name() string { return "openai-compatible" }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Analyze sends the prompt and returns the assistant content.
func (p *OpenAICompatibleProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	if p.BaseURL == "" {
		return "", fmt.Errorf("AI_BASE_URL is empty")
	}
	body := chatRequest{
		Model: p.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a careful assistant. Always respond with strict JSON only, no prose, no markdown fences."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		Stream:      false, // MiniMax-M3 / many proxies default to SSE; we need a single JSON object
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := p.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ai provider returned %d: %s", resp.StatusCode, string(rawResp))
	}

	// Some proxies ignore `stream:false` and still return SSE chunks
	// (e.g. MiniMax-M3 behind OmniRoute). If the body looks like SSE,
	// reassemble the streamed content into a single JSON envelope
	// before decoding.
	rawResp = reassembleSSE(rawResp)

	var cr chatResponse
	if err := json.Unmarshal(rawResp, &cr); err != nil {
		return "", fmt.Errorf("decode response: %w (raw=%s)", err, truncate(string(rawResp), 200))
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("ai provider returned no choices")
	}
	return cr.Choices[0].Message.Content, nil
}

// reassembleSSE walks Server-Sent Events lines ("data: {...}\n\n") and
// returns a single chat.completion-shaped JSON envelope with the final
// message content concatenated. Returns the input unchanged if it does
// not look like SSE.
func reassembleSSE(raw []byte) []byte {
	s := string(raw)
	if !strings.Contains(s, "data:") {
		return raw
	}
	var content strings.Builder
	// also capture reasoning_content in case the proxy puts the answer there
	var reasoning strings.Builder
	var id, model, obj string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var chunk struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Object  string `json:"object"`
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if id == "" {
			id = chunk.ID
			model = chunk.Model
			obj = chunk.Object
		}
		for _, ch := range chunk.Choices {
			content.WriteString(ch.Delta.Content)
			reasoning.WriteString(ch.Delta.ReasoningContent)
		}
	}
	if id == "" && content.Len() == 0 && reasoning.Len() == 0 {
		return raw
	}
	// Prefer real content; fall back to reasoning_content if the model
	// hid the answer there.
	final := content.String()
	if final == "" {
		final = reasoning.String()
	}
	envelope := map[string]any{
		"id":      id,
		"object":  obj,
		"model":   model,
		"choices": []map[string]any{{"index": 0, "message": map[string]any{"role": "assistant", "content": final}, "finish_reason": "stop"}},
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return raw
	}
	return b
}
