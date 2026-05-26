package feishunotice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Secret signs Feishu custom robot requests.
type Secret string

// Sign returns a Feishu custom robot signature for the timestamp.
func (s Secret) Sign(timestamp string) string {
	stringToSign := timestamp + "\n" + string(s)
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
