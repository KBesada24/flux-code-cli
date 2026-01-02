package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	httpClient *http.Client
	config     ProviderConfig
}

func NewOpenAIClient(cfg ProviderConfig) (*OpenAIClient, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	return &OpenAIClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
		config: cfg,
	}, nil
}

func (c *OpenAIClient) Stream(ctx context.Context, messages []Message) <-chan StreamEvent {
	events := make(chan StreamEvent)

	go func() {
		defer close(events)

		reqBody := ChatRequest{
			Model:    c.config.Model,
			Messages: messages,
			Stream:   true,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			events <- StreamEvent{Error: fmt.Errorf("marshal error: %w", err)}
			return
		}

		req, err := http.NewRequestWithContext(
			ctx,
			"POST",
			c.config.BaseURL+"/chat/completions",
			bytes.NewReader(jsonBody),
		)
		if err != nil {
			events <- StreamEvent{Error: fmt.Errorf("request error: %w", err)}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			events <- StreamEvent{Error: fmt.Errorf("http error: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			events <- StreamEvent{
				Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body)),
			}
			return
		}

		// Parse SSE stream
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				events <- StreamEvent{Done: true}
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue // Skip malformed chunks
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					events <- StreamEvent{Content: delta.Content}
				}
				if chunk.Choices[0].FinishReason != nil {
					events <- StreamEvent{
						Done:         true,
						FinishReason: *chunk.Choices[0].FinishReason,
					}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			events <- StreamEvent{Error: fmt.Errorf("stream error: %w", err)}
		}
	}()

	return events
}

func (c *OpenAIClient) Complete(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.config.BaseURL+"/chat/completions",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) GetModel() string {
	return c.config.Model
}

func (c *OpenAIClient) SetModel(model string) {
	c.config.Model = model
}
