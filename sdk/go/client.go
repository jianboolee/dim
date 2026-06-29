package dim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type Config struct {
	BaseURL    string
	APIKey     string
	Token      string
	HTTPClient *http.Client
}

type Option func(*Config)

type Client struct {
	client *client
}

type client struct {
	baseURL    string
	apiKey     string
	token      string
	httpClient *http.Client
}

type apiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func newClient(cfg Config) *client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return &client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
		token:      cfg.Token,
		httpClient: httpClient,
	}
}

func New(options ...Option) *Client {
	var cfg Config
	for _, option := range options {
		option(&cfg)
	}
	return &Client{client: newClient(cfg)}
}

func WithBaseURL(baseURL string) Option {
	return func(cfg *Config) {
		cfg.BaseURL = baseURL
	}
}

func WithAPIKey(apiKey string) Option {
	return func(cfg *Config) {
		cfg.APIKey = apiKey
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *Config) {
		cfg.HTTPClient = httpClient
	}
}

func (c *client) withToken(token string) *client {
	return &client{
		baseURL:    c.baseURL,
		apiKey:     c.apiKey,
		token:      token,
		httpClient: c.httpClient,
	}
}

func (c *client) do(ctx context.Context, method string, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("X-Integration-Key", c.apiKey)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("im api status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out == nil {
		return nil
	}

	wrapped := apiResponse[json.RawMessage]{}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Data != nil {
		if wrapped.Code != 0 && wrapped.Code != http.StatusOK {
			return fmt.Errorf("im api code %d: %s", wrapped.Code, wrapped.Message)
		}
		if err := json.Unmarshal(wrapped.Data, out); err != nil {
			return fmt.Errorf("decode im data: %w", err)
		}
		return nil
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode im response: %w", err)
	}
	return nil
}
