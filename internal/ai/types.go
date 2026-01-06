package ai

import "context"

// ChatMessage represents a single message in a chat request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest defines a model-agnostic chat completion request.
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	Temperature float32
	MaxTokens   int
	Stream      bool
}

// ChatResponse is returned for non-streaming completions.
type ChatResponse struct {
	Content string
}

// Client defines the provider-agnostic AI client interface.
type Client interface {
	Complete(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)
	Model() string
	Provider() string
}

// StreamEventType enumerates streaming events.
type StreamEventType string

const (
	StreamEventChunk StreamEventType = "chunk"
	StreamEventDone  StreamEventType = "done"
	StreamEventError StreamEventType = "error"
)

// StreamEvent is emitted during a streaming completion.
type StreamEvent struct {
	Type    StreamEventType
	Content string
	Err     error
}
