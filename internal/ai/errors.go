package ai

import "net/http"

// IsRetryableHTTP returns true for status codes that should be retried.
func IsRetryableHTTP(status int) bool {
	return status == http.StatusTooManyRequests || (status >= 500 && status <= 599)
}
