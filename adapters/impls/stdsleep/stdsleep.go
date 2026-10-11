package stdsleep

import (
	"time"

	sleepdeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/sleepdeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// sleep fills sleepdeps.Contract.Sleep: the calling goroutine paused for
// milliseconds, at once for zero or less.
func sleep(milliseconds int64) {
	if milliseconds <= 0 {
		return
	}
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// Bind fills deps.Deps.SleepDeps with the standard library's time.Sleep.
func Bind(deps *deps.Deps) {
	deps.SleepDeps = sleepdeps.Contract{
		Sleep: func(milliseconds int64) {
			sleep(milliseconds)
		},
	}
}
