# feishu-notice

Go package for Feishu custom robot notifications.

## 能力

- 单机器人：`NewClient(webhookURL, secret, options...)`
- 多机器人：`NewFactory(RobotConfig...)`
- 富文本消息：`Message` + `Send`
- 卡片消息：`Card` + `SendCard`
- 可选配置：`WithTimeout`、`WithHTTPClient`
- 错误类型：`HTTPError`、`ResponseError`

## 安装

```bash
go get github.com/neteast-software/feishu-notice
```

## 单机器人

```go
import (
	"context"
	"time"

	feishunotice "github.com/neteast-software/feishu-notice"
)

client, err := feishunotice.NewClient(webhookURL, secret, feishunotice.WithTimeout(10*time.Second))
if err != nil {
	return err
}

err = client.Send(ctx, feishunotice.Message{
	Title: "服务异常: Example",
	Lines: []string{
		"站点: Example",
		"状态: HTTP 503",
	},
})
```

`secret` 可为空；为空时不签名。

## 富文本

`Lines` 是纯文本快捷入口；混排用 `Paragraphs`。

```go
err = client.Send(context.Background(), feishunotice.Message{
	Title: "发布通知",
	Paragraphs: []feishunotice.Paragraph{
		{
			feishunotice.Text("详情: "),
			feishunotice.Link("查看", "https://example.com"),
		},
		{
			feishunotice.At("all", "所有人"),
		},
		{
			feishunotice.Image("img_xxx"),
		},
	},
})
```

`SegmentTag`:

| 常量 | 飞书 tag | 含义 |
|---|---|---|
| `TagText` | `text` | 文本 |
| `TagLink` | `a` | 链接 |
| `TagAt` | `at` | @ |
| `TagImage` | `img` | 图片 |

## 卡片

```go
err = client.SendCard(ctx, feishunotice.Card{
	"schema": "2.0",
	"header": map[string]any{
		"title":    map[string]any{"tag": "plain_text", "content": "服务状态"},
		"template": "green",
	},
	"body": map[string]any{
		"elements": []any{
			map[string]any{"tag": "markdown", "content": "**Example** 服务正常"},
		},
	},
})
```

## 多机器人

```go
factory, err := feishunotice.NewFactory(
	feishunotice.RobotConfig{Name: "ops", WebhookURL: opsWebhookURL, Secret: opsSecret},
	feishunotice.RobotConfig{Name: "release", WebhookURL: releaseWebhookURL, Secret: releaseSecret},
)
if err != nil {
	return err
}

err = factory.Send(ctx, "ops", feishunotice.Message{
	Title: "服务异常",
	Lines: []string{"状态: HTTP 503"},
})

err = factory.SendCard(ctx, "release", feishunotice.Card{
	"schema": "2.0",
	"body": map[string]any{
		"elements": []any{
			map[string]any{"tag": "markdown", "content": "发布完成"},
		},
	},
})
```

## 错误处理

```go
var httpErr feishunotice.HTTPError
if errors.As(err, &httpErr) {
	// httpErr.StatusCode / httpErr.Body
}

var responseErr feishunotice.ResponseError
if errors.As(err, &responseErr) {
	// responseErr.Code / responseErr.Message
}
```
