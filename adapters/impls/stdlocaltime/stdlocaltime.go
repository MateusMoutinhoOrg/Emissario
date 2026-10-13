package stdlocaltime

import (
	"time"

	localtimedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/localtimedeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// zone is the abbreviation and the offset from UTC, in seconds, of the
// process's local time zone at the instant seconds.
func zone(seconds int64) (string, int64) {
	name, offset := time.Unix(seconds, 0).In(time.Local).Zone()
	return name, int64(offset)
}

// Bind fills deps.Deps.LocaltimeDeps with the standard library's time.Local:
// the zone the TZ environment variable names, else the system's own.
func Bind(deps *deps.Deps) {
	deps.LocaltimeDeps = localtimedeps.Contract{
		Offset: func(seconds int64) int64 {
			_, offset := zone(seconds)
			return offset
		},
		Name: func(seconds int64) string {
			name, _ := zone(seconds)
			return name
		},
	}
}
