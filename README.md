# feishu-notice

统一封装飞书自定义机器人消息通知。

## 安装

```bash
go get github.com/neteast-software/feishu-notice
```

## 用法

```go
package main

import (
	"context"
	"time"

	feishunotice "github.com/neteast-software/feishu-notice"
)

func main() {
	client, err := feishunotice.NewClient(
		"https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
		"robot-secret",
		feishunotice.WithTimeout(10*time.Second),
	)
	if err != nil {
		panic(err)
	}

	err = client.Send(context.Background(), feishunotice.Message{
		Title: "服务异常: Example",
		Lines: []string{
			"站点: Example",
			"状态: HTTP 503",
		},
	})
	if err != nil {
		panic(err)
	}
}
```

`secret` 可为空；为空时不会生成 `timestamp` 和 `sign`。

## 多机器人

```go
factory, err := feishunotice.NewFactory(
	feishunotice.RobotConfig{
		Name:       feishunotice.Robot("ops"),
		WebhookURL: opsWebhookURL,
		Secret:     opsSecret,
		Options:    []feishunotice.Option{feishunotice.WithTimeout(10 * time.Second)},
	},
	feishunotice.RobotConfig{
		Name:       feishunotice.Robot("release"),
		WebhookURL: releaseWebhookURL,
		Secret:     releaseSecret,
	},
)
if err != nil {
	return err
}

err = factory.Send(context.Background(), feishunotice.Robot("ops"), feishunotice.Message{
	Title: "服务异常: Example",
	Lines: []string{"状态: HTTP 503"},
})
```

## service-health 接入

```go
factory, err := feishunotice.NewFactory(feishunotice.RobotConfig{
	Name:       feishunotice.Robot("default"),
	WebhookURL: config.FeishuWebhookURL,
	Secret:     config.FeishuSecret,
	Options:    []feishunotice.Option{feishunotice.WithTimeout(config.FeishuHTTPTimeout)},
})
if err != nil {
	return err
}
```
