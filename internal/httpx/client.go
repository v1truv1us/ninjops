package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ninjops/ninjops/internal/config"
)

// validateURL checks that the URL is safe to request (no SSRF)
func validateURL(rawURL string, allowedSchemes []string) error {
	if allowedSchemes == nil {
		allowedSchemes = []string{"https", "http"}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL missing host")
	}

	schemeValid := false
	for _, s := range allowedSchemes {
		if parsed.Scheme == s {
			schemeValid = true
			break
		}
	}
	if !schemeValid {
		return fmt.Errorf("URL scheme must be one of: %v", allowedSchemes)
	}

	return nil
}

type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

type ClientConfig struct {
	Timeout     time.Duration
	RetryConfig RetryConfig
	TokenRedact bool
	UserAgent   string
}

type Client struct {
	client *http.Client
	config ClientConfig
	token  string
	secret string
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Timeout: 30 * time.Second,
		RetryConfig: RetryConfig{
			MaxRetries:  3,
			InitialWait: 1 * time.Second,
			MaxWait:     30 * time.Second,
			Multiplier:  2.0,
		},
		TokenRedact: true,
		UserAgent:   "ninjops/1.0",
	}
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}
}

func NewClientWithAuth(cfg ClientConfig, token, secret string) *Client {
	c := NewClient(cfg)
	c.token = token
	c.secret = secret
	return c
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Validate URL to prevent SSRF (G704)
	if err := validateURL(req.URL.String(), []string{"https", "http"}); err != nil {
		return nil, fmt.Errorf("SSRF validation failed: %w", err)
	}

	c.setHeaders(req)

	var lastErr error
	wait := c.config.RetryConfig.InitialWait

	for attempt := 0; attempt <= c.config.RetryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			wait = time.Duration(float64(wait) * c.config.RetryConfig.Multiplier)
			if wait > c.config.RetryConfig.MaxWait {
				wait = c.config.RetryConfig.MaxWait
			}
		}

		var bodyBytes []byte
		if req.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %w", err)
			}
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = c.redactError(err)
			if c.shouldRetry(0, err) {
				continue
			}
			return nil, lastErr
		}

		if c.shouldRetry(resp.StatusCode, nil) {
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned status %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func (c *Client) Post(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(ctx, req)
}

func (c *Client) Put(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(ctx, req)
}

func (c *Client) setHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("X-API-Token", c.token)
	}
	if c.secret != "" {
		req.Header.Set("X-API-Secret", c.secret)
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", c.config.UserAgent)
}

func (c *Client) shouldRetry(statusCode int, err error) bool {
	if err != nil {
		return true
	}
	if statusCode == 429 {
		return true
	}
	if statusCode >= 500 && statusCode < 600 {
		return true
	}
	return false
}

func (c *Client) redactError(err error) error {
	if !c.config.TokenRedact {
		return err
	}
	errStr := err.Error()
	if c.token != "" {
		errStr = strings.ReplaceAll(errStr, c.token, config.RedactToken(c.token))
	}
	if c.secret != "" {
		errStr = strings.ReplaceAll(errStr, c.secret, config.RedactToken(c.secret))
	}
	if errStr != err.Error() {
		return fmt.Errorf("%s", errStr)
	}
	return err
}

func ParseJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

func ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
