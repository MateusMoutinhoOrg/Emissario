package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// Days are counted in the server's time zone: "today", a date of a filter and
// every bucket of the views over time start at a local midnight, or a local
// hour. An instant is Unix seconds; a local wall time is the same instant
// shifted by the zone's offset, so that timedeps — which spells UTC alone —
// spells it the way the server's clock reads it.

// DayLayout is how a day is spelled in a filter: 2026-10-13.
const DayLayout = "2006-01-02"

// TimeLayout is how an instant is spelled on a page and in an export.
const TimeLayout = "2006-01-02 15:04:05"

// daySeconds is the length of a day of wall time.
const daySeconds = 24 * 60 * 60

// hourSeconds is the length of an hour.
const hourSeconds = 60 * 60

// NowSeconds is the current instant, in Unix seconds.
func NowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1_000_000_000
}

// wallOf is the local wall time of the instant at.
func wallOf(sandbox *api.Sandbox, at int64) int64 {
	return at + sandbox.Deps.LocaltimeDeps.Offset(at)
}

// instantOf is the instant the local wall time wall is read at. The offset of
// the zone at that instant is not known before the instant is, so it is
// guessed from the wall time and checked once: that is exact everywhere but
// in the hour a daylight-saving change skips or repeats.
func instantOf(sandbox *api.Sandbox, wall int64) int64 {
	guess := wall - sandbox.Deps.LocaltimeDeps.Offset(wall)
	return wall - sandbox.Deps.LocaltimeDeps.Offset(guess)
}

// FormatLocal is the instant at spelled by layout in the server's time zone.
func FormatLocal(sandbox *api.Sandbox, at int64, layout string) string {
	return sandbox.Deps.TimeDeps.FormatUnix(wallOf(sandbox, at), layout)
}

// ZoneName is the abbreviation of the server's time zone now ("-03", "UTC").
func ZoneName(sandbox *api.Sandbox) string {
	return sandbox.Deps.LocaltimeDeps.Name(NowSeconds(sandbox))
}

// DayOf is the local day of the instant at, spelled DayLayout.
func DayOf(sandbox *api.Sandbox, at int64) string {
	return FormatLocal(sandbox, at, DayLayout)
}

// Today is the local day it is now.
func Today(sandbox *api.Sandbox) string {
	return DayOf(sandbox, NowSeconds(sandbox))
}

// wallDay reads day, spelled DayLayout exactly, as the wall time of its
// midnight.
func wallDay(sandbox *api.Sandbox, day string) (int64, bool) {
	wall, err := sandbox.Deps.TimeDeps.ParseUnix(DayLayout, day)
	if err != nil || sandbox.Deps.TimeDeps.FormatUnix(wall, DayLayout) != day {
		return 0, false
	}
	return wall, true
}

// ParseDay is the instant the local day spelled DayLayout starts at; false
// for anything else.
func ParseDay(sandbox *api.Sandbox, day string) (int64, bool) {
	wall, ok := wallDay(sandbox, day)
	if !ok {
		return 0, false
	}
	return instantOf(sandbox, wall), true
}

// AddDays is the day count days after day, both spelled DayLayout; day as it
// is when it does not read.
func AddDays(sandbox *api.Sandbox, day string, count int) string {
	wall, ok := wallDay(sandbox, day)
	if !ok {
		return day
	}
	return sandbox.Deps.TimeDeps.FormatUnix(wall+int64(count)*daySeconds, DayLayout)
}

// DaysBetween is how many days to is after from, both spelled DayLayout: 0
// for the same day, negative when to comes first.
func DaysBetween(sandbox *api.Sandbox, from string, to string) int {
	start, _ := wallDay(sandbox, from)
	end, _ := wallDay(sandbox, to)
	return int((end - start) / daySeconds)
}

// BucketStart is the instant the bucket holding at starts: its local hour,
// its local day or the first day of its local month.
func BucketStart(sandbox *api.Sandbox, at int64, bucket string) int64 {
	wall := wallOf(sandbox, at)
	switch bucket {
	case BucketHour:
		return instantOf(sandbox, wall-floorMod(wall, hourSeconds))
	case BucketMonth:
		start, _ := ParseDay(sandbox, sandbox.Deps.TimeDeps.FormatUnix(wall, "2006-01")+"-01")
		return start
	}
	return instantOf(sandbox, wall-floorMod(wall, daySeconds))
}

// NextBucket is the instant the bucket after the one starting at start
// starts.
func NextBucket(sandbox *api.Sandbox, start int64, bucket string) int64 {
	wall := wallOf(sandbox, start)
	next := start
	switch bucket {
	case BucketHour:
		next = instantOf(sandbox, wall-floorMod(wall, hourSeconds)+hourSeconds)
	case BucketMonth:
		year, month := yearMonth(sandbox, sandbox.Deps.TimeDeps.FormatUnix(wall, "2006-01"))
		if month == 12 {
			year, month = year+1, 1
		} else {
			month++
		}
		next, _ = ParseDay(sandbox, sandbox.Deps.StdDeps.Sprintf("%04d-%02d-01", year, month))
	default:
		next = instantOf(sandbox, wall-floorMod(wall, daySeconds)+daySeconds)
	}
	if next <= start {
		next = start + hourSeconds
	}
	return next
}

// yearMonth reads text, spelled 2006-01, as its year and month.
func yearMonth(sandbox *api.Sandbox, text string) (int64, int64) {
	parts := sandbox.Deps.StringsDeps.Split(text, "-")
	if len(parts) != 2 {
		return 1970, 1
	}
	year, _ := sandbox.Deps.StringsDeps.ParseInt(parts[0], 10, 64)
	month, _ := sandbox.Deps.StringsDeps.ParseInt(parts[1], 10, 64)
	return year, month
}

// BucketLabel is how the bucket starting at start is named on the chart.
func BucketLabel(sandbox *api.Sandbox, start int64, bucket string) string {
	switch bucket {
	case BucketHour:
		return FormatLocal(sandbox, start, "Jan 02 15:00")
	case BucketMonth:
		return FormatLocal(sandbox, start, "Jan 2006")
	}
	return FormatLocal(sandbox, start, "Mon Jan 02")
}

// floorMod is value modulo divisor, never negative.
func floorMod(value int64, divisor int64) int64 {
	rest := value % divisor
	if rest < 0 {
		rest += divisor
	}
	return rest
}
