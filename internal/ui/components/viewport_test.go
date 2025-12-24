package components

import (
	"testing"
)

func TestNewViewport(t *testing.T) {
	vp := NewViewport(80, 24)

	if !vp.Ready() {
		t.Error("New viewport should be ready")
	}
}

func TestViewportSetContent(t *testing.T) {
	vp := NewViewport(80, 24)

	// Should not panic
	vp.SetContent("Hello, World!")
}

func TestViewportSetSize(t *testing.T) {
	vp := NewViewport(80, 24)

	// Should not panic
	vp.SetSize(100, 30)
}

func TestViewportGotoBottom(t *testing.T) {
	vp := NewViewport(80, 24)
	vp.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")

	// Should not panic
	vp.GotoBottom()
}

func TestViewportScrollPercent(t *testing.T) {
	vp := NewViewport(80, 24)

	// Empty viewport should return a valid percent
	percent := vp.ScrollPercent()
	if percent < 0 || percent > 1 {
		t.Errorf("ScrollPercent should be between 0 and 1, got %f", percent)
	}
}
