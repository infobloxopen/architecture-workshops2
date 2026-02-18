package depclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client calls the dependency simulator service.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a dep client pointing at the given base URL.
// SOLUTION: Transport-level timeout ensures no request exceeds 2s.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Call invokes the /work endpoint on the dep service.
// SOLUTION: Uses context-aware request so caller's deadline is respected.
func Call(ctx context.Context, c *Client, sleep string, failRate string) (string, error) {
	url := fmt.Sprintf("%s/work?sleep=%s&fail=%s", c.BaseURL, sleep, failRate)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("dep call failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading dep response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dep returned %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
