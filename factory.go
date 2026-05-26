package feishunotice

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Robot identifies a Feishu custom robot in a Factory.
type Robot string

// String returns the normalized robot name.
func (r Robot) String() string {
	return strings.TrimSpace(string(r))
}

// RobotConfig describes one Feishu custom robot.
type RobotConfig struct {
	Name       Robot
	WebhookURL string
	Secret     string
	Options    []Option
}

// Factory manages multiple Feishu robot clients.
type Factory struct {
	mu      sync.RWMutex
	clients map[Robot]*Client
}

// NewFactory creates a factory and registers the provided robots.
func NewFactory(configs ...RobotConfig) (*Factory, error) {
	f := &Factory{clients: make(map[Robot]*Client, len(configs))}
	for _, config := range configs {
		if err := f.Register(config); err != nil {
			return nil, err
		}
	}
	return f, nil
}

// Register adds or replaces a robot client.
func (f *Factory) Register(config RobotConfig) error {
	if f == nil {
		return errors.New("飞书机器人工厂不能为空")
	}
	name := Robot(config.Name.String())
	if name == "" {
		return errors.New("机器人名称不能为空")
	}
	client, err := NewClient(config.WebhookURL, config.Secret, config.Options...)
	if err != nil {
		return fmt.Errorf("机器人 %s: %w", name, err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.clients == nil {
		f.clients = make(map[Robot]*Client)
	}
	f.clients[name] = client
	return nil
}

func (f *Factory) client(name Robot) (*Client, bool) {
	if f == nil {
		return nil, false
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	client, exists := f.clients[Robot(name.String())]
	return client, exists
}

func (f *Factory) mustClient(name Robot) (*Client, error) {
	client, exists := f.client(name)
	if !exists {
		return nil, fmt.Errorf("机器人 %s 未注册", name.String())
	}
	return client, nil
}

// Send sends a message with the named robot.
func (f *Factory) Send(ctx context.Context, name Robot, message Message) error {
	client, err := f.mustClient(name)
	if err != nil {
		return err
	}
	return client.Send(ctx, message)
}

// SendCard sends an interactive card message with the named robot.
func (f *Factory) SendCard(ctx context.Context, name Robot, card Card) error {
	client, err := f.mustClient(name)
	if err != nil {
		return err
	}
	return client.SendCard(ctx, card)
}

// Robots returns registered robot names in stable order.
func (f *Factory) Robots() []Robot {
	if f == nil {
		return nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	robots := make([]Robot, 0, len(f.clients))
	for name := range f.clients {
		robots = append(robots, name)
	}
	sort.Slice(robots, func(i, j int) bool {
		return robots[i].String() < robots[j].String()
	})
	return robots
}
