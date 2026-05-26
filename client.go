package feishunotice

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/neteast-software/feishu-notice/internal/feishu"
)

const defaultHTTPTimeout = 10 * time.Second

// Client sends messages to a Feishu custom robot webhook.
type Client struct {
	sender *feishu.Client
}

type clientConfig struct {
	webhookURL string
	secret     string
	httpClient *http.Client
	now        func() time.Time
}

// Option customizes Client construction.
type Option func(*clientConfig)

// WithHTTPClient sets the HTTP client used by Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *clientConfig) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout sets the timeout on the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		if timeout > 0 {
			c.httpClient = &http.Client{Timeout: timeout}
		}
	}
}

// WithClock sets the clock used for request signing.
func WithClock(now func() time.Time) Option {
	return func(c *clientConfig) {
		if now != nil {
			c.now = now
		}
	}
}

// NewClient creates a Feishu webhook client.
func NewClient(webhookURL string, secret string, options ...Option) (*Client, error) {
	config := clientConfig{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     strings.TrimSpace(secret),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		now:        time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	sender, err := feishu.NewClient(config.webhookURL, config.secret, config.httpClient, config.now)
	if err != nil {
		return nil, err
	}
	return &Client{sender: sender}, nil
}

// Send sends a post message to the configured webhook.
func (c *Client) Send(ctx context.Context, message Message) error {
	if c == nil || c.sender == nil {
		return errors.New("client is nil")
	}
	if err := message.Validate(); err != nil {
		return err
	}
	return c.sender.SendPost(ctx, feishu.PostMessage{
		Title:   message.safeTitle(),
		Locale:  message.Locale.String(),
		Content: message.feishuContent(),
	})
}
