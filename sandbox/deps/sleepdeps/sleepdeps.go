package sleepdeps

// This package is the sandbox's *copy* of the api a clock exposes to wait on —
// the same mechanic as stddeps, which hands out the clock's reading, for the
// same reason: pausing a goroutine needs `time`, which may not appear inside
// the sandbox. The contract is restated here, and the adapter — which lives
// outside the sandbox — is what fills it.
//
// A duration crosses this boundary as whole milliseconds, the way stddeps.Now
// hands an instant over as a plain integer.

// Contract is the waiting half of a clock, injected whole as the
// Deps.SleepDeps field. The git source sync loop is what reaches for it, to
// wait between two looks at the sources it keeps pulled.
type Contract struct {
	// Sleep pauses the calling goroutine for at least milliseconds. Zero or
	// a negative value returns at once.
	Sleep func(milliseconds int64)
}
