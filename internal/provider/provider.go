package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const maxRetries = 3

// doWithRetry executes an HTTP request with exponential backoff on retryable errors (429, 529, 5xx).
func doWithRetry(client *http.Client, newReq func() (*http.Request, error)) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		req, err := newReq()
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on rate limit or server errors
		if resp.StatusCode == 429 || resp.StatusCode == 529 || resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateStr(string(body), 200))
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Message represents a chat message for any provider.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response holds the result of a provider API call.
type Response struct {
	Content      string
	InputTokens  int
	OutputTokens int
	LatencyMs    int
}

// Provider is the interface for calling LLM APIs.
type Provider interface {
	Call(model string, system string, messages []Message) (*Response, error)
	Name() string
}

// Anthropic implements the Provider interface for Anthropic's API.
type Anthropic struct {
	apiKey string
	client *http.Client
}

func NewAnthropic() (*Anthropic, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	return &Anthropic{
		apiKey: key,
		client: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Call(model string, system string, messages []Message) (*Response, error) {
	start := time.Now()

	// Build request body
	anthropicMsgs := make([]map[string]string, len(messages))
	for i, m := range messages {
		anthropicMsgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body := map[string]any{
		"model":      model,
		"max_tokens": 2048,
		"messages":   anthropicMsgs,
	}
	if system != "" {
		body["system"] = system
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := doWithRetry(a.client, func() (*http.Request, error) {
		req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", a.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Anthropic response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, truncateStr(string(respBody), 300))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	text := ""
	if len(result.Content) > 0 {
		text = result.Content[0].Text
	}

	return &Response{
		Content:      text,
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
		LatencyMs:    int(time.Since(start).Milliseconds()),
	}, nil
}

// OpenAI implements the Provider interface for OpenAI's API.
type OpenAI struct {
	apiKey string
	client *http.Client
}

func NewOpenAI() (*OpenAI, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}
	return &OpenAI{
		apiKey: key,
		client: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (o *OpenAI) Name() string { return "openai" }

func (o *OpenAI) Call(model string, system string, messages []Message) (*Response, error) {
	start := time.Now()

	oaiMsgs := []map[string]string{}
	if system != "" {
		oaiMsgs = append(oaiMsgs, map[string]string{"role": "system", "content": system})
	}
	for _, m := range messages {
		oaiMsgs = append(oaiMsgs, map[string]string{"role": m.Role, "content": m.Content})
	}

	body := map[string]any{
		"model":      model,
		"max_tokens": 2048,
		"messages":   oaiMsgs,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := doWithRetry(o.client, func() (*http.Request, error) {
		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("calling OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading OpenAI response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, truncateStr(string(respBody), 300))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	text := ""
	if len(result.Choices) > 0 {
		text = result.Choices[0].Message.Content
	}

	return &Response{
		Content:      text,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		LatencyMs:    int(time.Since(start).Milliseconds()),
	}, nil
}

// ForModel returns the appropriate provider for a model string.
func ForModel(model string) (Provider, error) {
	switch {
	case isAnthropicModel(model):
		return NewAnthropic()
	case isOpenAIModel(model):
		return NewOpenAI()
	default:
		// Try Anthropic first, then OpenAI
		if p, err := NewAnthropic(); err == nil {
			return p, nil
		}
		if p, err := NewOpenAI(); err == nil {
			return p, nil
		}
		return nil, fmt.Errorf("no API key found — set ANTHROPIC_API_KEY or OPENAI_API_KEY")
	}
}

func isAnthropicModel(m string) bool {
	for _, prefix := range []string{"claude", "haiku"} {
		if len(m) >= len(prefix) && m[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func isOpenAIModel(m string) bool {
	for _, prefix := range []string{"gpt", "o1", "o3"} {
		if len(m) >= len(prefix) && m[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
