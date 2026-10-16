package rfc6238totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"

	"rsc.io/qr"

	totpdeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/totpdeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// secretBytes is how long a secret is: 160 bits, the size of HMAC-SHA1's
// output RFC 4226 recommends.
const secretBytes = 20

// stepSeconds is the time step of RFC 6238.
const stepSeconds = 30

// digits is how many digits a code has.
const digits = 6

// qrScale is how many image pixels one module of the QR code takes.
const qrScale = 6

// encoding is base32 without padding, the spelling authenticator apps read.
var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// decodeSecret reads secret as base32, ignoring spaces, case and padding.
func decodeSecret(secret string) ([]byte, error) {
	cleaned := strings.ToUpper(strings.TrimRight(strings.ReplaceAll(secret, " ", ""), "="))
	return encoding.DecodeString(cleaned)
}

// newSecret fills totpdeps.Contract.NewSecret.
func newSecret() (string, error) {
	key := make([]byte, secretBytes)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return encoding.EncodeToString(key), nil
}

// hotp is the code of key for counter (RFC 4226 §5.3).
func hotp(key []byte, counter uint64) string {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(message[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%0*d", digits, value%1000000)
}

// code fills totpdeps.Contract.Code.
func code(secret string, unixSeconds int64) (string, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	if unixSeconds < 0 {
		unixSeconds = 0
	}
	return hotp(key, uint64(unixSeconds/stepSeconds)), nil
}

// verify fills totpdeps.Contract.Verify.
func verify(secret string, given string, unixSeconds int64, skew int) (int64, bool) {
	given = strings.TrimSpace(strings.ReplaceAll(given, " ", ""))
	if len(given) != digits {
		return 0, false
	}
	key, err := decodeSecret(secret)
	if err != nil || len(key) == 0 {
		return 0, false
	}
	current := unixSeconds / stepSeconds
	matched, ok := int64(0), false
	for offset := -int64(skew); offset <= int64(skew); offset++ {
		step := current + offset
		if step < 0 {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(hotp(key, uint64(step))), []byte(given)) == 1 && !ok {
			matched, ok = step, true
		}
	}
	return matched, ok
}

// uri fills totpdeps.Contract.Uri.
func uri(issuer string, account string, secret string) string {
	label := url.PathEscape(issuer + ":" + account)
	query := url.Values{}
	query.Set("secret", secret)
	query.Set("issuer", issuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", fmt.Sprint(digits))
	query.Set("period", fmt.Sprint(stepSeconds))
	return "otpauth://totp/" + label + "?" + query.Encode()
}

// qrPng fills totpdeps.Contract.QrPng.
func qrPng(text string) ([]byte, error) {
	encoded, err := qr.Encode(text, qr.M)
	if err != nil {
		return nil, err
	}
	encoded.Scale = qrScale
	return encoded.PNG(), nil
}

// Bind fills deps.Deps.TotpDeps with crypto/hmac and crypto/sha1, and
// rsc.io/qr for the setup QR code.
func Bind(deps *deps.Deps) {
	deps.TotpDeps = totpdeps.Contract{
		NewSecret: newSecret,
		Code:      code,
		Verify:    verify,
		Uri:       uri,
		QrPng:     qrPng,
	}
}
