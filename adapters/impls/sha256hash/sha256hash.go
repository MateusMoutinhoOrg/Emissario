package sha256hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	hashdeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/hashdeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// sha256Hex fills hashdeps.Contract.Sha256Hex, returning the SHA-256 digest of
// content as lower-case hexadecimal.
func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// hmacSha256Hex fills hashdeps.Contract.HmacSha256Hex, returning the
// HMAC-SHA256 of content under key as lower-case hexadecimal.
func hmacSha256Hex(key []byte, content []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(content)
	return hex.EncodeToString(mac.Sum(nil))
}

// Bind fills deps.Deps.HashDeps with the standard library's crypto/sha256,
// crypto/hmac and encoding/hex.
func Bind(deps *deps.Deps) {
	deps.HashDeps = hashdeps.Contract{
		Sha256Hex: func(content []byte) string {
			return sha256Hex(content)
		},
		HmacSha256Hex: func(key []byte, content []byte) string {
			return hmacSha256Hex(key, content)
		},
		Equal: func(a []byte, b []byte) bool {
			return hmac.Equal(a, b)
		},
	}
}
