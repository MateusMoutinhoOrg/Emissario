package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Three limits hold the chat, every one per process:
//
//   - rate-per-minute: requests of one ip in a minute, in a fixed window of
//     ratelimitdeps;
//   - daily-limit: tokens — input and output — of every visitor in the day;
//   - ip-daily-limit: tokens of one ip in the day.
//
// The day is the server's (TZ). An IPv6 ip counts by its /64, which one
// visitor holds whole. The tokens are counted here in memory — loaded from
// agent-usage at the start and after a restore — and checked before every
// provider call: a call reserves max-tokens until it ends, so requests served
// at the same time never pass a limit together.

// rateWindowSeconds is the window of rate-per-minute.
const rateWindowSeconds = 60

// tally is the tokens of one day, and the tokens the calls in flight reserved.
type tally struct {
	day          string
	total        int64
	byIp         map[string]int64
	reserved     int64
	reservedByIp map[string]int64
}

// today is the tally of the current day, guarded by tallyLock.
var today = tally{byIp: map[string]int64{}, reservedByIp: map[string]int64{}}

// tallyLock is held while today is read or written.
var tallyLock = make(chan struct{}, 1)

// IpKey is the key ip counts under in the rate limit of every minute, in
// memory: an IPv4 ip whole, an IPv6 one its /64 (2001:db8:1:2::/64); "" for
// no ip.
func IpKey(sandbox *api.Sandbox, ip string) string {
	return privacy.GroupKey(sandbox, ip)
}

// ShownKey is key — a hash, an ip or a /64 — as the reader reads it: a hash
// shortened for everyone, an ip whole to a root and masked to a viewer
// unless the settings show viewers full IPs.
func ShownKey(sandbox *api.Sandbox, key string, root bool) string {
	return privacy.ShowIp(sandbox, key, root)
}

// CheckRate counts one request of ip and refuses it — 429 rate-limit, retry
// after a minute — past rate-per-minute.
func CheckRate(sandbox *api.Sandbox, settings Settings, ip string) *Refusal {
	hits := sandbox.Deps.RatelimitDeps.Hit("agent:"+IpKey(sandbox, ip), rateWindowSeconds)
	if int64(hits) > settings.RatePerMinute {
		return &Refusal{Status: api.StatusTooManyRequests, Message: CodeRateLimit, RetryAfter: rateWindowSeconds}
	}
	return nil
}

// reserve holds settings.MaxTokens for one call of key, or refuses it — 429
// daily-limit or ip-daily-limit, retry after midnight — when the call could
// pass a limit. A call reserved must be released.
func reserve(sandbox *api.Sandbox, settings Settings, key string) *Refusal {
	tallyLock <- struct{}{}
	defer func() { <-tallyLock }()
	roll(sandbox)
	if settings.DailyLimit > 0 && today.total+today.reserved+settings.MaxTokens > settings.DailyLimit {
		return &Refusal{Status: api.StatusTooManyRequests, Message: CodeDailyLimit, RetryAfter: secondsToMidnight(sandbox)}
	}
	if key != "" && settings.IpDailyLimit > 0 && today.byIp[key]+today.reservedByIp[key]+settings.MaxTokens > settings.IpDailyLimit {
		return &Refusal{Status: api.StatusTooManyRequests, Message: CodeIpDailyLimit, RetryAfter: secondsToMidnight(sandbox)}
	}
	today.reserved += settings.MaxTokens
	if key != "" {
		today.reservedByIp[key] += settings.MaxTokens
	}
	return nil
}

// release gives back the reservation of one call of key and counts the
// tokens it used.
func release(sandbox *api.Sandbox, settings Settings, key string, used int64) {
	tallyLock <- struct{}{}
	defer func() { <-tallyLock }()
	roll(sandbox)
	today.reserved = max(0, today.reserved-settings.MaxTokens)
	today.total += used
	if key != "" {
		today.reservedByIp[key] = max(0, today.reservedByIp[key]-settings.MaxTokens)
		if today.reservedByIp[key] == 0 {
			delete(today.reservedByIp, key)
		}
		today.byIp[key] += used
	}
}

// roll starts the tally over when the day changed. The caller holds
// tallyLock.
func roll(sandbox *api.Sandbox) {
	day := visualization.Today(sandbox)
	if today.day == day {
		return
	}
	today.day = day
	today.total = 0
	today.byIp = map[string]int64{}
}

// secondsToMidnight is how long until the day of the server ends.
func secondsToMidnight(sandbox *api.Sandbox) int64 {
	next, ok := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, visualization.Today(sandbox), 1))
	if !ok {
		return 60
	}
	return max(1, next-NowSeconds(sandbox))
}

// Tally is the tokens of today so far — every visitor, and per ip key — as
// the limits count them.
type Tally struct {
	Day   string
	Total int64
	ByIp  map[string]int64
}

// TodayTally is a copy of the tally of today.
func TodayTally(sandbox *api.Sandbox) Tally {
	tallyLock <- struct{}{}
	defer func() { <-tallyLock }()
	roll(sandbox)
	copied := Tally{Day: today.day, Total: today.total, ByIp: map[string]int64{}}
	for key, tokens := range today.byIp {
		copied.ByIp[key] = tokens
	}
	return copied
}

// LoadTally counts the tokens of today again from agent-usage: what the
// limits hold to after the start and after a restore. The reservations of
// the calls in flight are kept.
func LoadTally(sandbox *api.Sandbox) error {
	day := visualization.Today(sandbox)
	start, _ := visualization.ParseDay(sandbox, day)
	rows, err := agent_db.ListAgentUsagesBetween(sandbox, agent_db.New(sandbox), start, NowSeconds(sandbox)+DeadlineMilliseconds/1000)
	if err != nil {
		return err
	}
	total := int64(0)
	byIp := map[string]int64{}
	for _, row := range rows {
		used := row.InputTokens + row.OutputTokens
		total += used
		if row.Ip != "" {
			byIp[privacy.StoredKey(sandbox, row.Ip)] += used
		}
	}
	tallyLock <- struct{}{}
	defer func() { <-tallyLock }()
	today.day = day
	today.total = total
	today.byIp = byIp
	return nil
}
