package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Two limits hold the route, both per process and per ip — an IPv6 ip counts
// by its /64, which one visitor holds whole:
//
//   - rate-per-minute: requests in a minute, in a fixed window of
//     ratelimitdeps;
//   - ip-daily-limit: requests in the day of the server (TZ).
//
// Every request admitted counts in them, whether its body holds or not. The
// counters live in memory, so a restart forgets them.

// rateWindowSeconds is the window of rate-per-minute.
const rateWindowSeconds = 60

// dayWindowSeconds is the window of ip-daily-limit: its key holds the day, so
// it starts over at midnight.
const dayWindowSeconds = 24 * 60 * 60

// Admit is why a contact request from ip is refused before its body is looked
// at, nil when it goes on, with the settings it goes by: 503 while a restore
// writes the databases, with no writer in this process, or with the route
// off; 403 for a request another site's page sent — there is no CORS —; 429
// past the rate per minute or the daily limit of ip.
func Admit(sandbox *api.Sandbox, ip string, origin string, host string) (Settings, *Refusal) {
	if snapshots.Restoring(sandbox) {
		return Settings{}, &Refusal{Status: api.StatusServiceUnavailable, Message: CodeRestoring, RetryAfter: retryAfterSeconds}
	}
	if !Running(sandbox) {
		return Settings{}, &Refusal{Status: api.StatusServiceUnavailable, Message: CodeNotRunning}
	}
	settings, err := currentSettings(sandbox)
	if err != nil {
		return settings, failed(sandbox, "the contact settings could not be read", err)
	}
	if !settings.Enabled {
		return settings, &Refusal{Status: api.StatusServiceUnavailable, Message: CodeDisabled}
	}
	if !backofficehttp.SameOrigin(sandbox, origin, host) {
		return settings, &Refusal{Status: api.StatusForbidden, Field: "origin", Message: CodeOrigin}
	}
	key := ipKey(sandbox, ip)
	if hits := sandbox.Deps.RatelimitDeps.Hit("contact:"+key, rateWindowSeconds); int64(hits) > settings.RatePerMinute {
		return settings, &Refusal{Status: api.StatusTooManyRequests, Message: CodeRateLimit, RetryAfter: rateWindowSeconds}
	}
	if settings.IpDailyLimit > 0 {
		day := "contact-day:" + visualization.Today(sandbox) + ":" + key
		if hits := sandbox.Deps.RatelimitDeps.Hit(day, dayWindowSeconds); int64(hits) > settings.IpDailyLimit {
			return settings, &Refusal{Status: api.StatusTooManyRequests, Message: CodeIpDailyLimit, RetryAfter: secondsToMidnight(sandbox)}
		}
	}
	return settings, nil
}

// Receive checks body — at most settings.MaxBytes bytes of any JSON value —
// and hands it to the writer as a contact request from ip, exactly as it was
// sent, with the country of ip; the request keeps ip as the IP mode says,
// hashed by default. It answers the id of the contact request:
// 413 for a body too large, 400 for one that is not JSON, 503 with the queue
// full.
func Receive(sandbox *api.Sandbox, settings Settings, ip string, body []byte) (string, *Refusal) {
	if int64(len(body)) > settings.MaxBytes {
		return "", &Refusal{Status: api.StatusPayloadTooLarge, Message: CodeTooLarge}
	}
	if _, err := sandbox.Deps.SerializableDeps.ParseJson(string(body)); err != nil {
		return "", refuse("", CodeInvalidJson)
	}
	ref, err := sandbox.Deps.RandDeps.Hex(refBytes)
	if err != nil {
		return "", failed(sandbox, "the id of a contact request could not be drawn", err)
	}
	stored, err := privacy.Store(sandbox, ip)
	if err != nil {
		return "", failed(sandbox, "the ip of a contact request could not be hashed", err)
	}
	queued := enqueue(sandbox, contact_db.ContactRequestInput{
		Ref:     ref,
		At:      NowSeconds(sandbox),
		Ip:      stored,
		Country: visualization.Country(sandbox, ip),
		Body:    string(body),
		Size:    int64(len(body)),
	})
	if !queued {
		return "", &Refusal{Status: api.StatusServiceUnavailable, Message: CodeBusy, RetryAfter: retryAfterSeconds}
	}
	return ref, nil
}

// CreatedJSON is {"id": ref}: what /api/contact-requests answers.
func CreatedJSON(sandbox *api.Sandbox, ref string) *serializabledeps.SerializableObject {
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("id", ref)
	return document
}

// ipKey is the key ip counts under: an IPv4 ip whole, an IPv6 one its /64
// (2001:db8:1:2::/64).
func ipKey(sandbox *api.Sandbox, ip string) string {
	return privacy.GroupKey(sandbox, ip)
}

// secondsToMidnight is how many seconds the day of the server has left.
func secondsToMidnight(sandbox *api.Sandbox) int64 {
	next, ok := visualization.ParseDay(sandbox, visualization.AddDays(sandbox, visualization.Today(sandbox), 1))
	if !ok {
		return rateWindowSeconds
	}
	return max(1, next-NowSeconds(sandbox))
}
