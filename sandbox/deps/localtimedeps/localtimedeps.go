package localtimedeps

// This package is the sandbox's *copy* of the api a clock exposes about the
// time zone the process runs in — the same mechanic as timedeps, which spells
// instants in UTC, for the same reason: the zone lives in the `time` package
// and the operating system, neither of which may appear inside the sandbox.
// The contract is restated here, and the adapter — which lives outside the
// sandbox — is what fills it.
//
// An instant crosses this boundary as seconds since the Unix epoch, the way
// timedeps hands one over. Turning an instant into a local date is then
// timedeps.FormatUnix(seconds + Offset(seconds), layout).

// Contract is the local time zone of the process, injected whole as the
// Deps.LocaltimeDeps field. It is named after the capability, not after
// whoever reads it, so any part of the project that counts days the way the
// server's clock does reuses it.
type Contract struct {
	// Offset answers how many seconds the server's time zone is ahead of UTC
	// at the instant seconds — -10800 for America/Sao_Paulo, 3600 for
	// Europe/Paris in winter and 7200 in summer: a daylight-saving change is
	// in it.
	Offset func(seconds int64) int64
	// Name answers the abbreviation of the server's time zone at the instant
	// seconds ("-03", "CEST", "UTC").
	Name func(seconds int64) string
}
