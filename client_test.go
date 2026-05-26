package feishunotice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSendPostMessage(t *testing.T) {
	var payload feishuMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Fatalf("content-type = %s", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret", WithClock(func() time.Time {
		return time.Unix(1700000000, 0)
	}))
	if err != nil {
		t.Fatal(err)
	}

	err = client.Send(context.Background(), Message{
		Title: "服务异常",
		Lines: []string{"站点: Example", "状态: down"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if payload.MsgType != "post" {
		t.Fatalf("msg_type = %s", payload.MsgType)
	}
	if payload.Timestamp != "1700000000" {
		t.Fatalf("timestamp = %s", payload.Timestamp)
	}
	if payload.Sign != Secret("secret").Sign("1700000000") {
		t.Fatalf("sign = %s", payload.Sign)
	}
	post := payload.Content.Post["zh_cn"]
	if post.Title != "服务异常" {
		t.Fatalf("title = %s", post.Title)
	}
	if got := post.Content[1][0].Text; got != "状态: down" {
		t.Fatalf("line = %s", got)
	}
}

func TestClientSendWithoutSecret(t *testing.T) {
	var payload feishuMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Send(context.Background(), Message{Title: "服务恢复"}); err != nil {
		t.Fatal(err)
	}
	if payload.Timestamp != "" || payload.Sign != "" {
		t.Fatalf("unexpected signing fields: %+v", payload)
	}
}

func TestClientSendUsesLocaleAndTruncatesTitle(t *testing.T) {
	var payload feishuMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Send(context.Background(), Message{
		Title:  strings.Repeat("长", maxTitleLength+1),
		Locale: Locale("en_us"),
	}); err != nil {
		t.Fatal(err)
	}

	post := payload.Content.Post["en_us"]
	if len([]rune(post.Title)) != maxTitleLength {
		t.Fatalf("title length = %d", len([]rune(post.Title)))
	}
}

func TestClientSendRichParagraphs(t *testing.T) {
	var payload feishuMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{
		Title: "发布通知",
		Paragraphs: []Paragraph{
			{Text("详情: "), Link("查看", "https://example.com")},
			{At("all", "所有人")},
			{Image("img_ecffc3b9-8f14-400f-a014-05eca1a4310g")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	post := payload.Content.Post["zh_cn"]
	if post.Content[0][0].Tag != TagText {
		t.Fatalf("tag = %s", post.Content[0][0].Tag)
	}
	if post.Content[0][1].Tag != TagLink || post.Content[0][1].Href != "https://example.com" {
		t.Fatalf("link segment = %+v", post.Content[0][1])
	}
	if post.Content[1][0].Tag != TagAt || post.Content[1][0].UserID != "all" {
		t.Fatalf("at segment = %+v", post.Content[1][0])
	}
	if post.Content[2][0].Tag != TagImage || post.Content[2][0].ImageKey == "" {
		t.Fatalf("image segment = %+v", post.Content[2][0])
	}
}

func TestClientSendHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{Title: "服务异常"})

	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error = %v", err)
	}
	if httpErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d", httpErr.StatusCode)
	}
}

func TestClientSendResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":9499,"msg":"sign invalid"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{Title: "服务异常"})

	var responseErr ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("error = %v", err)
	}
	if responseErr.Code != 9499 {
		t.Fatalf("code = %d", responseErr.Code)
	}
}

func TestNewClientRequiresWebhookURL(t *testing.T) {
	_, err := NewClient("", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMessageRequiresTitle(t *testing.T) {
	client, err := NewClient("https://example.com/hook", "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSecretSign(t *testing.T) {
	sign := Secret("secret").Sign("1700000000")
	if sign == "" {
		t.Fatal("sign is empty")
	}
	if sign != "fiWS2+gh28DOydAv7hzONH/mDn9+b1Y4Y5ivXWXy8vA=" {
		t.Fatalf("sign = %s", sign)
	}
}
