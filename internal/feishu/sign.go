package feishu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Sign returns a Feishu custom robot signature for the timestamp.
func Sign(secret string, timestamp string) string {
	stringToSign := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
