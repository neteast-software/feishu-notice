package feishunotice

import "github.com/neteast-software/feishu-notice/internal/feishu"

// Secret signs Feishu custom robot requests.
type Secret string

// Sign returns a Feishu custom robot signature for the timestamp.
func (s Secret) Sign(timestamp string) string {
	return feishu.Sign(string(s), timestamp)
}
