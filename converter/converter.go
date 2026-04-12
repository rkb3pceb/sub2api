// Package converter provides functionality to convert subscription links
// into various proxy client configuration formats.
package converter

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SupportedFormats lists all output formats this converter supports.
var SupportedFormats = []string{
	"clash",
	"singbox",
	"surge",
	"raw",
}

// FetchResult holds the raw content fetched from a subscription URL.
type FetchResult struct {
	Content  string
	Encoding string // "base64" or "plain"
}

// Client is an HTTP client wrapper for fetching subscription content.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new converter Client with sensible defaults.
func NewClient(timeout time.Duration, userAgent string) *Client {
	if timeout == 0 {
		// Increased default timeout from 15s to 30s; some subscription
		// providers are slow to respond and 15s caused frequent timeouts.
		timeout = 30 * time.Second
	}
	if userAgent == "" {
		userAgent = "sub2api/1.0 (https://github.com/your-org/sub2api)"
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

// Fetch retrieves the subscription content from the given URL.
// It automatically detects and decodes base64-encoded responses.
func (c *Client) Fetch(url string) (*FetchResult, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10 MB limit
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	raw := strings.TrimSpace(string(body))

	// Attempt base64 decode; treat as plain text on failure.
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(raw)
	}
	if err == nil && looksLikeProxyList(string(decoded)) {
		return &FetchResult{Content: string(decoded), Encoding: "base64"}, nil
	}

	return &FetchResult{Content: raw, Encoding: "plain"}, nil
}

// ParseProxyLines splits subscription content into individual proxy lines,
// filtering out blank lines and comments.
func ParseProxyLines(content string) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, line)
	}
	return result
}

// IsValidFormat reports whether the given format string is supported.
func IsValidFormat(format string) bool {
	for _, f := range SupportedFormats {
		if strings.EqualFold(f, format) {
			return true
		}
	}
	return false
}

// looksLikeProxyList is a heuristic to decide whether decoded content
// resembles a list of proxy URI