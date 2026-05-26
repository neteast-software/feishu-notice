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
	var payload testMessagePayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("请求方法 = %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Fatalf("内容类型 = %s", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret", withClock(func() time.Time {
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
		t.Fatalf("消息类型 = %s", payload.MsgType)
	}
	if payload.Timestamp != "1700000000" {
		t.Fatalf("时间戳 = %s", payload.Timestamp)
	}
	if payload.Sign != "fiWS2+gh28DOydAv7hzONH/mDn9+b1Y4Y5ivXWXy8vA=" {
		t.Fatalf("签名 = %s", payload.Sign)
	}
	post := payload.Content.Post["zh_cn"]
	if post.Title != "服务异常" {
		t.Fatalf("标题 = %s", post.Title)
	}
	if got := post.Content[1][0].Text; got != "状态: down" {
		t.Fatalf("正文行 = %s", got)
	}
}

func TestClientSendWithoutSecret(t *testing.T) {
	var payload testMessagePayload
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
		t.Fatalf("不应存在签名字段: %+v", payload)
	}
}

func TestClientSendUsesLocaleAndTruncatesTitle(t *testing.T) {
	var payload testMessagePayload
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
		t.Fatalf("标题长度 = %d", len([]rune(post.Title)))
	}
}

func TestClientSendRichParagraphs(t *testing.T) {
	var payload testMessagePayload
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
		t.Fatalf("节点类型 = %s", post.Content[0][0].Tag)
	}
	if post.Content[0][1].Tag != TagLink || post.Content[0][1].Href != "https://example.com" {
		t.Fatalf("链接节点 = %+v", post.Content[0][1])
	}
	if post.Content[1][0].Tag != TagAt || post.Content[1][0].UserID != "all" {
		t.Fatalf("@节点 = %+v", post.Content[1][0])
	}
	if post.Content[2][0].Tag != TagImage || post.Content[2][0].ImageKey == "" {
		t.Fatalf("图片节点 = %+v", post.Content[2][0])
	}
}

func TestClientSendCard(t *testing.T) {
	var payload testMessagePayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret", withClock(func() time.Time {
		return time.Unix(1700000000, 0)
	}))
	if err != nil {
		t.Fatal(err)
	}
	err = client.SendCard(context.Background(), Card{
		"schema": "2.0",
		"header": map[string]any{
			"title": map[string]any{"tag": "plain_text", "content": "卡片测试"},
		},
		"body": map[string]any{
			"elements": []any{
				map[string]any{"tag": "markdown", "content": "服务状态 **正常**"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if payload.MsgType != "interactive" {
		t.Fatalf("消息类型 = %s", payload.MsgType)
	}
	if payload.Sign != "fiWS2+gh28DOydAv7hzONH/mDn9+b1Y4Y5ivXWXy8vA=" {
		t.Fatalf("签名 = %s", payload.Sign)
	}
	if payload.Card["schema"] != "2.0" {
		t.Fatalf("卡片 = %#v", payload.Card)
	}
}

func TestCardRequiresContent(t *testing.T) {
	client, err := NewClient("https://example.com/hook", "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.SendCard(context.Background(), Card{})
	if err == nil {
		t.Fatal("预期返回错误")
	}
}

func TestClientSendHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "网关错误", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{Title: "服务异常"})

	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("错误 = %v", err)
	}
	if httpErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("状态码 = %d", httpErr.StatusCode)
	}
}

func TestClientSendResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":9499,"msg":"签名无效"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{Title: "服务异常"})

	var responseErr ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("错误 = %v", err)
	}
	if responseErr.Code != 9499 {
		t.Fatalf("响应码 = %d", responseErr.Code)
	}
}

func TestNewClientRequiresWebhookURL(t *testing.T) {
	_, err := NewClient("", "")
	if err == nil {
		t.Fatal("预期返回错误")
	}
}

func TestMessageRequiresTitle(t *testing.T) {
	client, err := NewClient("https://example.com/hook", "")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(context.Background(), Message{})
	if err == nil {
		t.Fatal("预期返回错误")
	}
}

type testMessagePayload struct {
	MsgType   string             `json:"msg_type"`
	Content   testContentPayload `json:"content"`
	Card      map[string]any     `json:"card"`
	Timestamp string             `json:"timestamp,omitempty"`
	Sign      string             `json:"sign,omitempty"`
}

type testContentPayload struct {
	Post map[string]testPostPayload `json:"post"`
}

type testPostPayload struct {
	Title   string      `json:"title"`
	Content []Paragraph `json:"content"`
}
