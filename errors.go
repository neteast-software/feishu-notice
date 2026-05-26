package feishunotice

import "github.com/neteast-software/feishu-notice/internal/feishu"

// HTTPError describes a non-2xx Feishu HTTP response.
type HTTPError = feishu.HTTPError

// ResponseError describes a Feishu JSON response whose code is not zero.
type ResponseError = feishu.ResponseError
