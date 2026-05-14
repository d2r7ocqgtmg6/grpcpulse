// Package alertmanager provides a client for sending alerts to an
// Alertmanager-compatible HTTP endpoint when service health transitions occur.
package alertmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config holds configuration for the Alertmanager client.
type Config struct {
	// URL is the base URL of the Alertmanager API, e.g. http://localhost:9093.
	URL string `yaml:"url"`
	// Timeout for each HTTP request. Defaults to 5s.
	Timeout time.Duration `yaml:"timeout"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Timeout: 5 * time.Second,
	}
}

// Alert represents a single Alertmanager alert payload.
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
}

// Client sends alerts to an Alertmanager endpoint.
type Client struct {
	cfg    Config
	httpCl *http.Client
}

// New creates a new Client. If cfg.Timeout is zero, the default is applied.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	return &Client{
		cfg: cfg,
		httpCl: &http.Client{Timeout: cfg.Timeout},
	}
}

// Send posts the given alerts to the Alertmanager /api/v2/alerts endpoint.
func (c *Client) Send(ctx context.Context, alerts []Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	body, err := json.Marshal(alerts)
	if err != nil {
		return fmt.Errorf("alertmanager: marshal alerts: %w", err)
	}

	url := c.cfg.URL + "/api/v2/alerts"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("alertmanager: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCl.Do(req)
	if err != nil {
		return fmt.Errorf("alertmanager: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alertmanager: unexpected status %d", resp.StatusCode)
	}
	return nil
}
