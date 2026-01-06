package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StandardClientConfig defines the parameters for any OpenAI-compatible endpoint.
type StandardClientConfig struct {
	BaseURL    string
	APIKey     string
	AuthHeader string
	AuthPrefix string
	Model      string
	Provider   string
	HTTPClient *http.Client
}

// StandardClient implements a generic OpenAI-compatible chat client.
// It works with OpenAI, Ollama, Groq, OpenRouter, and others.
type StandardClient struct {
	baseURL    string
	apiKey     string
	authHeader string
	authPrefix string
	model      string
	provider   string
	httpClient *http.Client
}

// NewStandardClient creates a new generic AI client.
func NewStandardClient(cfg StandardClientConfig) (Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	// Note: API Key might be optional for some local providers like Ollama
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 60 * time.Second}
	}

	authHeader := cfg.AuthHeader
	if authHeader == "" {
		authHeader = "Authorization"
	}
	authPrefix := cfg.AuthPrefix
	if authPrefix == "" {
		authPrefix = "Bearer "
	}

	provider := cfg.Provider
	if provider == "" {
		provider = "custom"
	}

	return &StandardClient{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
		authHeader: authHeader,
		authPrefix: authPrefix,
		model:      cfg.Model,
		provider:   provider,
		httpClient: hc,
	}, nil
}

func (c *StandardClient) Model() string    { return c.model }
func (c *StandardClient) Provider() string { return c.provider }

func (c *StandardClient) Complete(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	payload := c.toPayload(req, false)
	body, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, err
	}
	c.applyHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChatResponse{}, c.httpError(resp)
	}

	var parsed standardResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ChatResponse{}, err
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, errors.New("no choices returned")
	}

	content := parsed.Choices[0].Message.Content
	return ChatResponse{Content: content}, nil
}

func (c *StandardClient) Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error) {
	payload := c.toPayload(req, true)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.applyHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	out := make(chan StreamEvent)
	go func() {
		defer close(out)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			out <- StreamEvent{Type: StreamEventError, Err: c.httpError(resp)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}

			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				out <- StreamEvent{Type: StreamEventDone}
				return
			}

			var chunk standardStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				out <- StreamEvent{Type: StreamEventError, Err: err}
				return
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					out <- StreamEvent{Type: StreamEventChunk, Content: choice.Delta.Content}
				}
			}
		}

		if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
			out <- StreamEvent{Type: StreamEventError, Err: err}
		}
	}()

	return out, nil
}

func (c *StandardClient) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set(c.authHeader, c.authPrefix+c.apiKey)
	}
}

func (c *StandardClient) httpError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("api error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
}

func (c *StandardClient) toPayload(req ChatRequest, stream bool) standardRequest {
	messages := make([]standardMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, standardMessage{Role: m.Role, Content: m.Content})
	}

	model := req.Model
	if model == "" {
		model = c.model
	}

	return standardRequest{
		Model:       model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      stream,
	}
}

type standardRequest struct {
	Model       string            `json:"model"`
	Messages    []standardMessage `json:"messages"`
	Temperature float32           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
}

type standardMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type standardResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type standardStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
