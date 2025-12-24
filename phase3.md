# Phase 3: AI Integration

**Timeline:** Week 3  
**Goal:** Implement AI provider clients with streaming support, connect to real LLM APIs

---

## Overview

This phase brings the AI to life. Users can send prompts and receive streaming responses from various providers (Ollama, OpenRouter, Groq, Together AI). The focus is on the OpenAI-compatible API format which most providers support.

---

## Features

| Feature | Description | Priority |
|---------|-------------|----------|
| AI client interface | Abstract provider interface | P0 |
| OpenAI-compatible client | Works with most providers | P0 |
| SSE streaming | Real-time response streaming | P0 |
| Streaming display | Show response as it arrives | P0 |
| Error handling | Graceful API error handling | P0 |
| Provider switching | Change providers via config | P1 |
| Retry logic | Automatic retry on transient failures | P2 |

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/ai/client.go` | Client interface definition |
| `internal/ai/openai.go` | OpenAI-compatible API client |
| `internal/ai/streaming.go` | SSE stream parsing |
| `internal/ai/types.go` | Request/response types |
| `internal/ai/errors.go` | Custom error types |

### Modified Files

| File | Changes |
|------|---------|
| `internal/ui/model.go` | Add AI client, streaming state |
| `internal/app/app.go` | Initialize AI client from config |

---

## Detailed File Specifications

### 1. `internal/ai/types.go`

```go
package ai

import "time"

// Message represents a chat message
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ChatRequest is the request body for chat completions
type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Stream      bool      `json:"stream"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
}

// ChatResponse is the non-streaming response
type ChatResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// StreamChunk is a single SSE chunk
type StreamChunk struct {
    ID      string         `json:"id"`
    Object  string         `json:"object"`
    Created int64          `json:"created"`
    Model   string         `json:"model"`
    Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
    Index        int         `json:"index"`
    Delta        DeltaContent `json:"delta"`
    FinishReason *string     `json:"finish_reason"`
}

type DeltaContent struct {
    Role    string `json:"role,omitempty"`
    Content string `json:"content,omitempty"`
}

// StreamEvent represents a streaming event
type StreamEvent struct {
    Content      string
    Done         bool
    Error        error
    FinishReason string
}
```

---

### 2. `internal/ai/client.go`

```go
package ai

import (
    "context"
)

// Client is the interface for AI providers
type Client interface {
    // Stream sends a prompt and returns a channel of streaming events
    Stream(ctx context.Context, messages []Message) <-chan StreamEvent
    
    // Complete sends a prompt and returns the full response
    Complete(ctx context.Context, messages []Message) (string, error)
    
    // GetModel returns the current model name
    GetModel() string
    
    // SetModel changes the active model
    SetModel(model string)
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
    Name    string
    APIKey  string
    BaseURL string
    Model   string
}

// NewClient creates a new AI client based on provider config
func NewClient(cfg ProviderConfig) (Client, error) {
    // All supported providers use OpenAI-compatible API
    return NewOpenAIClient(cfg)
}
```

---

### 3. `internal/ai/openai.go`

```go
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
```

---

### 4. `internal/ai/errors.go`

```go
package ai

import "fmt"

// APIError represents an error from the AI provider
type APIError struct {
    StatusCode int
    Message    string
    Provider   string
}

func (e APIError) Error() string {
    return fmt.Sprintf("%s API error (%d): %s", e.Provider, e.StatusCode, e.Message)
}

// RateLimitError indicates rate limiting
type RateLimitError struct {
    RetryAfter int // seconds
}

func (e RateLimitError) Error() string {
    return fmt.Sprintf("rate limited, retry after %d seconds", e.RetryAfter)
}

// IsRetryable returns true if the error is transient
func IsRetryable(err error) bool {
    if err == nil {
        return false
    }
    
    switch e := err.(type) {
    case APIError:
        // Retry on 5xx errors and 429 (rate limit)
        return e.StatusCode >= 500 || e.StatusCode == 429
    case RateLimitError:
        return true
    default:
        return false
    }
}
```

---

### 5. Updated `internal/ui/model.go` (AI Integration)

```go
package ui

import (
    "context"
    
    "github.com/yourusername/flux/internal/ai"
    "github.com/yourusername/flux/internal/ui/components"
    tea "github.com/charmbracelet/bubbletea"
)

// Custom message types for streaming
type streamChunkMsg string
type streamDoneMsg struct{}
type streamErrMsg struct{ err error }

type Model struct {
    // Components
    input    components.Input
    viewport components.Viewport
    messages components.Messages
    
    // AI
    aiClient  ai.Client
    streaming bool
    streamBuf string
    cancelFn  context.CancelFunc
    
    // State
    width    int
    height   int
    ready    bool
    quitting bool
    err      error
}

func NewModel(client ai.Client) Model {
    return Model{
        input:    components.NewInput(),
        messages: components.NewMessages(80),
        aiClient: client,
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            if m.streaming {
                // Cancel current stream
                if m.cancelFn != nil {
                    m.cancelFn()
                }
                m.streaming = false
                return m, nil
            }
            m.quitting = true
            return m, tea.Quit
            
        case "enter":
            if !m.streaming && m.input.Value() != "" {
                return m, m.sendMessage()
            }
        }
        
    case streamChunkMsg:
        m.streamBuf += string(msg)
        m.updateStreamingMessage()
        return m, nil
        
    case streamDoneMsg:
        m.finalizeStream()
        return m, nil
        
    case streamErrMsg:
        m.err = msg.err
        m.streaming = false
        m.messages.Add(components.RoleAssistant, 
            "Error: "+msg.err.Error())
        m.viewport.SetContent(m.messages.Render())
        return m, nil
        
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.handleResize()
        m.ready = true
    }
    
    // Update components (only if not streaming)
    if !m.streaming {
        var cmd tea.Cmd
        m.input, cmd = m.input.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    m.viewport, cmd := m.viewport.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m *Model) sendMessage() tea.Cmd {
    userMsg := m.input.Value()
    m.input.Reset()
    
    // Add user message
    m.messages.Add(components.RoleUser, userMsg)
    m.viewport.SetContent(m.messages.Render())
    m.viewport.GotoBottom()
    
    // Start streaming
    m.streaming = true
    m.streamBuf = ""
    
    // Build message history for API
    history := m.buildMessageHistory()
    
    return func() tea.Msg {
        ctx, cancel := context.WithCancel(context.Background())
        m.cancelFn = cancel
        
        events := m.aiClient.Stream(ctx, history)
        
        // Process stream in goroutine, send messages back
        go func() {
            for event := range events {
                if event.Error != nil {
                    // Send error through program
                    return
                }
                if event.Done {
                    return
                }
                // Send chunk (need to use program.Send)
            }
        }()
        
        return nil
    }
}

func (m *Model) streamCommand() tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithCancel(context.Background())
        m.cancelFn = cancel
        
        history := m.buildMessageHistory()
        events := m.aiClient.Stream(ctx, history)
        
        for event := range events {
            if event.Error != nil {
                return streamErrMsg{err: event.Error}
            }
            if event.Done {
                return streamDoneMsg{}
            }
            if event.Content != "" {
                return streamChunkMsg(event.Content)
            }
        }
        
        return streamDoneMsg{}
    }
}

func (m *Model) buildMessageHistory() []ai.Message {
    var history []ai.Message
    
    // Add system prompt if configured
    // history = append(history, ai.Message{Role: "system", Content: systemPrompt})
    
    for _, msg := range m.messages.Items() {
        history = append(history, ai.Message{
            Role:    string(msg.Role),
            Content: msg.Content,
        })
    }
    
    return history
}

func (m *Model) updateStreamingMessage() {
    // Render messages + current streaming buffer
    content := m.messages.Render()
    content += "\n" + m.renderStreamingIndicator() + m.streamBuf
    m.viewport.SetContent(content)
    m.viewport.GotoBottom()
}

func (m *Model) finalizeStream() {
    m.messages.Add(components.RoleAssistant, m.streamBuf)
    m.streamBuf = ""
    m.streaming = false
    m.viewport.SetContent(m.messages.Render())
    m.viewport.GotoBottom()
}

func (m Model) renderStreamingIndicator() string {
    return lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00D4AA")).
        Render("Assistant") + " (streaming...)\n"
}
```

---

### 6. Updated `internal/app/app.go`

```go
package app

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yourusername/flux/internal/ai"
    "github.com/yourusername/flux/internal/config"
    "github.com/yourusername/flux/internal/ui"
)

func Run() error {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Get provider config
    provider, ok := cfg.Providers[cfg.Provider]
    if !ok {
        return fmt.Errorf("unknown provider: %s", cfg.Provider)
    }
    
    // Create AI client
    client, err := ai.NewClient(ai.ProviderConfig{
        Name:    cfg.Provider,
        APIKey:  provider.APIKey,
        BaseURL: provider.BaseURL,
        Model:   provider.Model,
    })
    if err != nil {
        return fmt.Errorf("failed to create AI client: %w", err)
    }
    
    // Create and run TUI
    model := ui.NewModel(client)
    p := tea.NewProgram(model, tea.WithAltScreen())
    
    _, err = p.Run()
    return err
}
```

---

## Testing

### Unit Tests

| Test File | Tests |
|-----------|-------|
| `internal/ai/openai_test.go` | Request building, response parsing |
| `internal/ai/streaming_test.go` | SSE parsing, chunk handling |
| `internal/ai/errors_test.go` | Error classification, retry logic |

### Integration Tests

| Test | Description |
|------|-------------|
| Mock server | Test against httptest server |
| Ollama local | Test with local Ollama if available |

### Mock Server for Testing

```go
func TestStreamingWithMockServer(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        flusher := w.(http.Flusher)
        
        chunks := []string{"Hello", " ", "World", "!"}
        for _, chunk := range chunks {
            fmt.Fprintf(w, "data: %s\n\n", makeChunkJSON(chunk))
            flusher.Flush()
            time.Sleep(10 * time.Millisecond)
        }
        fmt.Fprintf(w, "data: [DONE]\n\n")
    }))
    defer server.Close()
    
    client, _ := ai.NewOpenAIClient(ai.ProviderConfig{
        BaseURL: server.URL,
        Model:   "test",
    })
    
    // Test streaming
    events := client.Stream(context.Background(), []ai.Message{
        {Role: "user", Content: "test"},
    })
    
    var result string
    for event := range events {
        result += event.Content
    }
    
    assert.Equal(t, "Hello World!", result)
}
```

### Manual Testing Checklist

- [ ] Ollama: Start Ollama, verify streaming works
- [ ] OpenRouter: Set API key, verify connection
- [ ] Groq: Set API key, verify fast responses
- [ ] Cancel: Press Ctrl+C during stream to cancel
- [ ] Error: Test with invalid API key
- [ ] Network: Test behavior when offline
- [ ] Long response: Verify no timeout on lengthy responses

### Test Commands

```bash
# Run all AI tests
go test -v ./internal/ai/...

# Test with Ollama (requires running instance)
FLUX_PROVIDER=ollama go test -v ./internal/ai/... -run Integration

# Test specific provider
OPENROUTER_API_KEY=xxx go test -v ./internal/ai/... -run OpenRouter
```

---

## Provider-Specific Notes

### Ollama
- No API key required
- Default URL: `http://localhost:11434/v1`
- Supports OpenAI-compatible endpoint

### OpenRouter
- Requires `OPENROUTER_API_KEY`
- Base URL: `https://openrouter.ai/api/v1`
- Add `HTTP-Referer` header for attribution

### Groq
- Requires `GROQ_API_KEY`
- Base URL: `https://api.groq.com/openai/v1`
- Very fast inference

### Together AI
- Requires `TOGETHER_API_KEY`
- Base URL: `https://api.together.xyz/v1`

---

## Acceptance Criteria

1. **Streaming:** Responses stream character-by-character
2. **Multiple providers:** Can switch between providers via config
3. **Cancel:** Ctrl+C cancels ongoing stream
4. **Errors:** API errors display gracefully
5. **History:** Full conversation sent to API
6. **Model display:** Current model shown in status bar

---

## Definition of Done

- [ ] AI client interface implemented
- [ ] OpenAI-compatible client working
- [ ] SSE streaming parsing works
- [ ] Streaming display in TUI
- [ ] Error handling tested
- [ ] Multiple providers verified
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Committed to version control

---

## Notes

- Use `context.WithCancel` for stream cancellation
- Buffer size for channels should be reasonable (100+)
- Consider exponential backoff for retries
- Log API errors for debugging (optional verbose mode)
