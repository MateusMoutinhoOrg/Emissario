package hashdeps

// This package is the sandbox's *copy* of the api a hashing library exposes —
// the same mechanic as argvdeps, embeddeps, iodeps, rundeps, stddeps, stringsdeps,
// sortdeps and templatedeps, for the same reason: the sandbox may import
// nothing but the sandbox, so `crypto/sha256`, `crypto/hmac` and
// `encoding/hex` may not appear inside it. The contract is restated here, and the adapter — which
// lives outside the sandbox — is what fills it.

// Contract is the hashing library injected whole as the Deps.HashDeps field.
type Contract struct {
	// Sha256Hex returns the SHA-256 digest of content, lower-case
	// hexadecimal. It is what every recorded example tree is compared by, so
	// the encoding is part of the golden and may not change.
	Sha256Hex func(content []byte) string

	// HmacSha256Hex returns the HMAC-SHA256 of content under key, lower-case
	// hexadecimal.
	HmacSha256Hex func(key []byte, content []byte) string

	// Equal tells whether a and b hold the same bytes, in a time that does
	// not depend on where they differ — how a signature is checked.
	Equal func(a []byte, b []byte) bool
}
