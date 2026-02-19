package nexus

import (
	"fmt"
)

type APIError struct {
	StatusCode int
	URL        string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("nexus api error: status=%d url=%s", e.StatusCode, e.URL)
	}
	// Keep body short-ish in error strings; callers can log the full body if needed.
	const max = 400
	b := e.Body
	if len(b) > max {
		b = b[:max] + "…"
	}
	return fmt.Sprintf("nexus api error: status=%d url=%s body=%q", e.StatusCode, e.URL, b)
}
