package feishu

import "testing"

func TestSign(t *testing.T) {
	sign := Sign("secret", "1700000000")
	if sign == "" {
		t.Fatal("签名为空")
	}
	if sign != "fiWS2+gh28DOydAv7hzONH/mDn9+b1Y4Y5ivXWXy8vA=" {
		t.Fatalf("签名 = %s", sign)
	}
}
