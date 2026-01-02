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
