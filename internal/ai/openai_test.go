package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kbesada/flux-code-cli/internal/ai"
)

func makeChunkJSON(content string) string {
	chunk := ai.StreamChunk{
		Choices: []ai.StreamChoice{
			{
				Delta: ai.DeltaContent{
					Content: content,
				},
			},
		},
	}
	b, _ := json.Marshal(chunk)
	return string(b)
}

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

	client, err := ai.NewOpenAIClient(ai.ProviderConfig{
		BaseURL: server.URL,
		Model:   "test",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test streaming
	events := client.Stream(context.Background(), []ai.Message{
		{Role: "user", Content: "test"},
	})

	var result string
	for event := range events {
		result += event.Content
	}

	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestCompleteWithMockServer(t *testing.T) {
	expected := "Hello World!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp := ai.ChatResponse{
			Choices: []ai.Choice{
				{
					Message: ai.Message{
						Role:    "assistant",
						Content: expected,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := ai.NewOpenAIClient(ai.ProviderConfig{
		BaseURL: server.URL,
		Model:   "test",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test completion
	result, err := client.Complete(context.Background(), []ai.Message{
		{Role: "user", Content: "test"},
	})
	if err != nil {
		t.Fatalf("Failed to complete: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
