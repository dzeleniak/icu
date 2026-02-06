package satellite

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles API requests to spacebook.com
type Client struct {
	httpClient *http.Client
	tleURL     string
	satcatURL  string
}

// NewClient creates a new API client with a configured HTTP client
func NewClient(tleURL, satcatURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		tleURL:    tleURL,
		satcatURL: satcatURL,
	}
}

// FetchTLEs retrieves all TLE entries from the API.
// TLEs are returned as plain text with two lines per entry.
func (c *Client) FetchTLEs() ([]TLE, error) {
	resp, err := c.httpClient.Get(c.tleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TLEs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse TLE data (each TLE is 2 lines)
	var tles []TLE
	scanner := bufio.NewScanner(bytes.NewReader(body))
	var line1 string
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if lineNum%2 == 0 {
			line1 = line
		} else {
			tles = append(tles, TLE{
				Line1: line1,
				Line2: line,
			})
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading TLE data: %w", err)
	}

	return tles, nil
}

// FetchSATCATs retrieves all SATCAT entries from the API.
// SATCAT data is returned as JSON.
func (c *Client) FetchSATCATs() ([]SATCAT, error) {
	resp, err := c.httpClient.Get(c.satcatURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SATCATs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var satcats []SATCAT
	if err := json.Unmarshal(body, &satcats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SATCAT response: %w", err)
	}

	return satcats, nil
}
