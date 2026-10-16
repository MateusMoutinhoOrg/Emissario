package totpdeps

// This package is the sandbox's *copy* of the api a one-time password
// library exposes — the same mechanic as hashdeps and jwtdeps, for the same
// reason: HMAC-SHA1, base32 and a QR encoder are libraries, which may not
// appear inside the sandbox. The contract is restated here, and the adapter
// — which lives outside the sandbox — is what fills it.
//
// It is TOTP as RFC 6238 defines it and every authenticator app reads it:
// HMAC-SHA1, 6 digits, 30-second steps, the secret in base32. The backoffice
// asks for a code at sign-in once a user turned it on, and requires it of the
// readers of the access log. Only builtin types cross this boundary.

// Contract is the one-time password library injected whole as the
// Deps.TotpDeps field.
type Contract struct {
	// NewSecret answers a fresh secret: 20 random bytes in base32, upper
	// case, no padding.
	NewSecret func() (secret string, err error)

	// Code answers the 6-digit code of secret for the 30-second step the
	// Unix instant unixSeconds falls in. The error reports a secret that is
	// not base32.
	Code func(secret string, unixSeconds int64) (code string, err error)

	// Verify tells whether code is the one of secret for the step of
	// unixSeconds, or of one of the skew steps before or after it — the
	// clock of a phone drifts — and answers the step it matched, which the
	// caller keeps so the same code is never accepted twice. It compares in
	// constant time.
	Verify func(secret string, code string, unixSeconds int64, skew int) (step int64, ok bool)

	// Uri is the otpauth://totp/ URI an authenticator app reads secret
	// from, labelled issuer:account.
	Uri func(issuer string, account string, secret string) string

	// QrPng answers text drawn as a QR code, a PNG image: what the setup
	// page shows, from a route of its own the backoffice serves.
	QrPng func(text string) (png []byte, err error)
}
