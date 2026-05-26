package feishunotice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFactorySendsWithNamedRobot(t *testing.T) {
	var sentToA int
	var sentToB int
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sentToA++
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer serverA.Close()
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sentToB++
		var payload testMessagePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Content.Post["zh_cn"].Title != "发布通知" {
			t.Fatalf("标题 = %s", payload.Content.Post["zh_cn"].Title)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer serverB.Close()

	factory, err := NewFactory(
		RobotConfig{Name: Robot("ops"), WebhookURL: serverA.URL},
		RobotConfig{Name: Robot("release"), WebhookURL: serverB.URL},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := factory.Send(context.Background(), Robot("release"), Message{Title: "发布通知"}); err != nil {
		t.Fatal(err)
	}
	if sentToA != 0 {
		t.Fatalf("发送到 ops 次数 = %d", sentToA)
	}
	if sentToB != 1 {
		t.Fatalf("发送到 release 次数 = %d", sentToB)
	}
}

func TestFactoryRegisterReplacesRobot(t *testing.T) {
	var oldCount int
	oldServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		oldCount++
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer oldServer.Close()
	var newCount int
	newServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		newCount++
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer newServer.Close()

	factory, err := NewFactory(RobotConfig{Name: Robot("ops"), WebhookURL: oldServer.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := factory.Register(RobotConfig{Name: Robot("ops"), WebhookURL: newServer.URL}); err != nil {
		t.Fatal(err)
	}
	if err := factory.Send(context.Background(), Robot("ops"), Message{Title: "服务异常"}); err != nil {
		t.Fatal(err)
	}

	if oldCount != 0 || newCount != 1 {
		t.Fatalf("旧机器人次数=%d 新机器人次数=%d", oldCount, newCount)
	}
}

func TestFactorySendsCardWithNamedRobot(t *testing.T) {
	var sentToA int
	var sentToB int
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sentToA++
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer serverA.Close()
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sentToB++
		var payload testMessagePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.MsgType != "interactive" {
			t.Fatalf("消息类型 = %s", payload.MsgType)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	defer serverB.Close()

	factory, err := NewFactory(
		RobotConfig{Name: Robot("ops"), WebhookURL: serverA.URL},
		RobotConfig{Name: Robot("release"), WebhookURL: serverB.URL},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := factory.SendCard(context.Background(), Robot("release"), Card{"schema": "2.0"}); err != nil {
		t.Fatal(err)
	}
	if sentToA != 0 || sentToB != 1 {
		t.Fatalf("发送到A次数=%d 发送到B次数=%d", sentToA, sentToB)
	}
}

func TestFactoryRequiresRobotName(t *testing.T) {
	_, err := NewFactory(RobotConfig{WebhookURL: "https://example.com/hook"})
	if err == nil {
		t.Fatal("预期返回错误")
	}
}

func TestFactoryMissingRobot(t *testing.T) {
	factory, err := NewFactory()
	if err != nil {
		t.Fatal(err)
	}
	err = factory.Send(context.Background(), Robot("missing"), Message{Title: "服务异常"})
	if err == nil {
		t.Fatal("预期返回错误")
	}
}

func TestFactoryRobotsSorted(t *testing.T) {
	factory, err := NewFactory(
		RobotConfig{Name: Robot("release"), WebhookURL: "https://example.com/release"},
		RobotConfig{Name: Robot("ops"), WebhookURL: "https://example.com/ops"},
	)
	if err != nil {
		t.Fatal(err)
	}

	robots := factory.Robots()
	if len(robots) != 2 || robots[0] != "ops" || robots[1] != "release" {
		t.Fatalf("机器人列表 = %#v", robots)
	}
}
