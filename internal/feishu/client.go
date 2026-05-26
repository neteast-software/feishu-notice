package feishu

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

// Client adapts Feishu custom robot webhook protocol.
type Client struct {
	webhookURL string
	secret     string
	httpClient *http.Client
	now        func() time.Time
}

// NewClient creates a Feishu protocol client.
func NewClient(webhookURL string, secret string, httpClient *http.Client, now func() time.Time) (*Client, error) {
	if strings.TrimSpace(webhookURL) == "" {
		return nil, errors.New("webhook url is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if now == nil {
		now = time.Now
	}
	return &Client{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     strings.TrimSpace(secret),
		httpClient: httpClient,
		now:        now,
	}, nil
}

// PostMessage is the Feishu post message payload body.
type PostMessage struct {
	Title   string
	Locale  string
	Content [][]Segment
}

// Segment is one Feishu post rich-text node.
type Segment struct {
	Tag      string `json:"tag"`
	Text     string `json:"text,omitempty"`
	Href     string `json:"href,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
	ImageKey string `json:"image_key,omitempty"`
	UnEscape *bool  `json:"un_escape,omitempty"`
}

// SendPost sends a post message.
func (c *Client) SendPost(ctx context.Context, message PostMessage) error {
	payload := messagePayload{
		MsgType: "post",
		Content: contentPayload{
			Post: map[string]postPayload{
				message.Locale: {
					Title:   message.Title,
					Content: message.Content,
				},
			},
		},
	}
	if c.secret != "" {
		timestamp := strconv.FormatInt(c.now().Unix(), 10)
		payload.Timestamp = timestamp
		payload.Sign = Sign(c.secret, timestamp)
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

	var result responsePayload
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return ResponseError{Code: result.Code, Message: result.Message}
	}
	return nil
}

type messagePayload struct {
	MsgType   string         `json:"msg_type"`
	Content   contentPayload `json:"content"`
	Timestamp string         `json:"timestamp,omitempty"`
	Sign      string         `json:"sign,omitempty"`
}

type contentPayload struct {
	Post map[string]postPayload `json:"post"`
}

type postPayload struct {
	Title   string      `json:"title"`
	Content [][]Segment `json:"content"`
}

type responsePayload struct {
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
