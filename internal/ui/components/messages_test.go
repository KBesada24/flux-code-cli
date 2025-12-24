package components

import (
	"strings"
	"testing"
)

func TestNewMessages(t *testing.T) {
	msgs := NewMessages(80)

	if msgs.Count() != 0 {
		t.Errorf("New messages should be empty, got %d", msgs.Count())
	}
}

func TestMessagesAdd(t *testing.T) {
	msgs := NewMessages(80)

	msgs.Add(RoleUser, "Hello")
	if msgs.Count() != 1 {
		t.Errorf("Expected 1 message, got %d", msgs.Count())
	}

	msgs.Add(RoleAssistant, "Hi there!")
	if msgs.Count() != 2 {
		t.Errorf("Expected 2 messages, got %d", msgs.Count())
	}
}

func TestMessagesClear(t *testing.T) {
	msgs := NewMessages(80)

	msgs.Add(RoleUser, "Hello")
	msgs.Add(RoleAssistant, "Hi")
	msgs.Clear()

	if msgs.Count() != 0 {
		t.Errorf("After clear, expected 0 messages, got %d", msgs.Count())
	}
}

func TestMessagesRender(t *testing.T) {
	msgs := NewMessages(80)

	msgs.Add(RoleUser, "Hello")
	rendered := msgs.Render()

	if !strings.Contains(rendered, "You") {
		t.Error("Rendered user message should contain 'You'")
	}
	if !strings.Contains(rendered, "Hello") {
		t.Error("Rendered message should contain content")
	}
}

func TestMessagesRenderAssistant(t *testing.T) {
	msgs := NewMessages(80)

	msgs.Add(RoleAssistant, "# Heading\n\nSome text")
	rendered := msgs.Render()

	if !strings.Contains(rendered, "Assistant") {
		t.Error("Rendered assistant message should contain 'Assistant'")
	}
}

func TestMessagesSetWidth(t *testing.T) {
	msgs := NewMessages(80)

	// Should not panic
	msgs.SetWidth(120)
}
