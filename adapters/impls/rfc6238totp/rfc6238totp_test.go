package rfc6238totp

import (
	"strings"
	"testing"
)

// The SHA-1 vectors of RFC 6238 appendix B, last 6 of their 8 digits.
func TestRfc6238Vectors(t *testing.T) {
	secret := encoding.EncodeToString([]byte("12345678901234567890"))
	for _, vector := range []struct {
		at   int64
		code string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	} {
		got, err := code(secret, vector.at)
		if err != nil || got != vector.code {
			t.Fatalf("at %d: %q %v, want %q", vector.at, got, err, vector.code)
		}
	}
}

func TestVerifySkewAndStep(t *testing.T) {
	secret, err := newSecret()
	if err != nil || len(secret) != 32 {
		t.Fatalf("secret %q %v", secret, err)
	}
	now := int64(1_800_000_000)
	previous, _ := code(secret, now-30)
	step, ok := verify(secret, previous, now, 1)
	if !ok || step != (now-30)/30 {
		t.Fatalf("previous step: %d %v", step, ok)
	}
	if _, ok := verify(secret, previous, now+60, 1); ok {
		t.Fatal("a code two steps old passed")
	}
	if _, ok := verify(secret, "12345", now, 1); ok {
		t.Fatal("five digits passed")
	}
	if _, ok := verify(strings.ToLower(secret), previous, now, 1); !ok {
		t.Fatal("a lower-case secret was refused")
	}
}

func TestQrPng(t *testing.T) {
	png, err := qrPng(uri("Emissario", "admin", "JBSWY3DPEHPK3PXP"))
	if err != nil || !strings.HasPrefix(string(png), "\x89PNG") {
		t.Fatalf("%.8q %v", png, err)
	}
}
