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
		var payload feishuMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Content.Post["zh_cn"].Title != "发布通知" {
			t.Fatalf("title = %s", payload.Content.Post["zh_cn"].Title)
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
		t.Fatalf("sent to ops = %d", sentToA)
	}
	if sentToB != 1 {
		t.Fatalf("sent to release = %d", sentToB)
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
		t.Fatalf("oldCount=%d newCount=%d", oldCount, newCount)
	}
}

func TestFactoryRequiresRobotName(t *testing.T) {
	_, err := NewFactory(RobotConfig{WebhookURL: "https://example.com/hook"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFactoryMissingRobot(t *testing.T) {
	factory, err := NewFactory()
	if err != nil {
		t.Fatal(err)
	}
	err = factory.Send(context.Background(), Robot("missing"), Message{Title: "服务异常"})
	if err == nil {
		t.Fatal("expected error")
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
		t.Fatalf("robots = %#v", robots)
	}
}
