package deepobservergo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

// Client is the main object to communicate with the Deep-Observer Control Plane
type Client struct {
	baseURL     string        // Server address (e.g., http://localhost:8090)
	httpClient  *http.Client  // Shared HTTP client to leverage connection pooling
	maxRetries  int           // Maximum number of retry attempts
	backoffBase time.Duration // Base duration for exponential backoff
}

// NewClient initializes a new Client with default configurations
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		maxRetries:  3,
		backoffBase: 500 * time.Millisecond,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Event represents the data structure of an operational event.
type Event struct {
	ID            string         `json:"id"`
	SchemaVersion string         `json:"schema_version"`
	EventType     string         `json:"event_type"` // incident, deployment, config_change, maintenance
	ServiceName   string         `json:"service_name"`
	Environment   string         `json:"environment"`
	Title         string         `json:"title"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Version       string         `json:"version,omitempty"`
	CommitHash    string         `json:"commit_hash,omitempty"`
	Actor         string         `json:"actor,omitempty"`
	Severity      string         `json:"severity,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// SendEvent sends an event to the Control Plane API with retry logic and lifecycle management
func (c *Client) SendEvent(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/events", c.baseURL)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(math.Pow(2, float64(attempt-1))) * c.backoffBase
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return fmt.Errorf("retry canceled: %w", ctx.Err())
			}
		}

		err = c.doRequest(ctx, url, payload)
		if err == nil {
			return nil
		}

		lastErr = err

		if !c.isRetryable(err) {
			break
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return &statusError{statusCode: resp.StatusCode}
}

// isRetryable determines if we should retry based on the error
func (c *Client) isRetryable(err error) bool {
	if se, ok := err.(*statusError); ok {
		return se.statusCode >= 500 || se.statusCode == 429
	}
	return true
}

type statusError struct {
	statusCode int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("bad status code: %d", e.statusCode)
}
