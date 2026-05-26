package feishunotice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	defaultLocale      = "zh_cn"
	maxTitleLength     = 120
)

// Client sends messages to a Feishu custom robot webhook.
type Client struct {
	webhookURL string
	secret     Secret
	httpClient *http.Client
	now        func() time.Time
}

// Option customizes Client construction.
type Option func(*Client)

// WithHTTPClient sets the HTTP client used by Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout sets the timeout on the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient = &http.Client{Timeout: timeout}
		}
	}
}

// WithClock sets the clock used for request signing.
func WithClock(now func() time.Time) Option {
	return func(c *Client) {
		if now != nil {
			c.now = now
		}
	}
}

// NewClient creates a Feishu webhook client.
func NewClient(webhookURL string, secret string, options ...Option) (*Client, error) {
	c := &Client{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     Secret(strings.TrimSpace(secret)),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		now:        time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(c)
		}
	}
	if c.webhookURL == "" {
		return nil, errors.New("webhook url is required")
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return c, nil
}

// Message is a rich-text post message for Feishu.
type Message struct {
	Title  string
	Lines  []string
	Locale Locale
}

// Locale is a Feishu post locale key, for example zh_cn or en_us.
type Locale string

// String returns the locale value.
func (l Locale) String() string {
	if strings.TrimSpace(string(l)) == "" {
		return defaultLocale
	}
	return strings.TrimSpace(string(l))
}

// Validate checks whether the message can be sent.
func (m Message) Validate() error {
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("message title is required")
	}
	return nil
}

// Send sends a post message to the configured webhook.
func (c *Client) Send(ctx context.Context, message Message) error {
	if c == nil {
		return errors.New("client is nil")
	}
	if err := message.Validate(); err != nil {
		return err
	}

	payload := feishuMessage{
		MsgType: "post",
		Content: feishuContent{
			Post: map[string]feishuPost{
				message.Locale.String(): {
					Title:   truncate(message.Title, maxTitleLength),
					Content: message.segments(),
				},
			},
		},
	}
	if c.secret != "" {
		timestamp := strconv.FormatInt(c.now().Unix(), 10)
		payload.Timestamp = timestamp
		payload.Sign = c.secret.Sign(timestamp)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return HTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}

	var result feishuResponse
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return ResponseError{Code: result.Code, Message: result.Message}
	}
	return nil
}

func (m Message) segments() [][]segment {
	content := make([][]segment, 0, len(m.Lines))
	for _, line := range m.Lines {
		content = append(content, []segment{{Tag: "text", Text: line}})
	}
	return content
}

type feishuMessage struct {
	MsgType   string        `json:"msg_type"`
	Content   feishuContent `json:"content"`
	Timestamp string        `json:"timestamp,omitempty"`
	Sign      string        `json:"sign,omitempty"`
}

type feishuContent struct {
	Post map[string]feishuPost `json:"post"`
}

type feishuPost struct {
	Title   string      `json:"title"`
	Content [][]segment `json:"content"`
}

type segment struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

type feishuResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// HTTPError describes a non-2xx Feishu HTTP response.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("feishu status %d", e.StatusCode)
	}
	return fmt.Sprintf("feishu status %d: %s", e.StatusCode, e.Body)
}

// ResponseError describes a Feishu JSON response whose code is not zero.
type ResponseError struct {
	Code    int
	Message string
}

func (e ResponseError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("feishu code %d", e.Code)
	}
	return fmt.Sprintf("feishu code %d: %s", e.Code, e.Message)
}

func truncate(value string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxLength {
		return value
	}
	return string(runes[:maxLength])
}
