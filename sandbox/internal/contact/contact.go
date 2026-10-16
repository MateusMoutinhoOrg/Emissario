package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// Contact requests are what the public website sends through POST
// /api/contact-requests: a JSON body of any kind — an object, an array, a
// string, a number — whose shape the website chooses, at most as large as the
// root set in the backoffice. Each one is stored as it was sent, with the
// instant it arrived, the ip it came from and the country of that ip
// (visualization.Country), in contact-db. Every backoffice user reads and
// searches them — by period, country and a text the JSON holds —; a root
// archives, unarchives and deletes them, and sets the limits.
//
// The route is limited per ip — requests per minute and per day, in fixed
// windows of ratelimitdeps (receive.go) — and never waits on the database:
// the record goes through a queue (queue.go), held while a backup is taken, so
// the website keeps sending meanwhile.

const (
	// RequestsPath lists the contact requests: the top-bar link.
	RequestsPath = "/admin/list-contact-requests"
	// RequestPath shows one contact request, ?id=<ref>.
	RequestPath = "/admin/get-contact-request"
	// ArchivePath archives or unarchives contact requests.
	ArchivePath = "/admin/root/archive-contact-requests"
	// RemovePath removes contact requests, by id or before a date.
	RemovePath = "/admin/root/remove-contact-requests"
	// SettingsPath is the settings page, and where its form posts.
	SettingsPath = "/admin/root/set-contact-settings"
	// PublicPath is the route the website sends a contact request to.
	PublicPath = "/api/contact-requests"
)

const (
	// QueueSize is how many contact requests wait for the writer at most.
	QueueSize = 10000
	// DefaultMaxBytes is the largest body accepted when no root set any.
	DefaultMaxBytes = 16384
	// MinMaxBytes is the smallest largest body a root may set: "{}".
	MinMaxBytes = 2
	// MaxMaxBytes is the largest body a root may let through: the max-bytes
	// of the contact-requests route, which refuses a larger one before it
	// is read.
	MaxMaxBytes = 1048576
	// DefaultRatePerMinute is the requests one ip may send a minute when no
	// root set any.
	DefaultRatePerMinute = 5
	// MaxRatePerMinute is the most requests a minute a root may let one ip
	// send.
	MaxRatePerMinute = 10000
	// MaxIpDailyLimit is the largest ip daily limit a root may set.
	MaxIpDailyLimit = 1000000
	// MaxContainsLength is the longest text a search looks for.
	MaxContainsLength = 200
	// MaxIds is how many contact requests one archive or removal names.
	MaxIds = 100
	// PreviewLength is how many characters of the body a listing shows.
	PreviewLength = 160
	// refBytes is how many random bytes the id of a contact request holds.
	refBytes = 8
	// retryAfterSeconds is the Retry-After of a 503 during a restore.
	retryAfterSeconds = 5
)

// The codes /api/contact-requests answers in the error of {"error","field"}.
const (
	CodeRateLimit    = "rate-limit"
	CodeIpDailyLimit = "ip-daily-limit"
	CodeDisabled     = "disabled"
	CodeNotRunning   = "not-running"
	CodeRestoring    = "restoring"
	CodeOrigin       = "origin"
	CodeTooLarge     = "too-large"
	CodeInvalidJson  = "invalid-json"
	CodeBusy         = "busy"
)

// The statuses a listing shows, from its ?status=.
const (
	StatusInbox    = "inbox"
	StatusArchived = "archived"
	StatusAll      = "all"
)

// Statuses is every status a listing accepts, the default first.
var Statuses = []string{StatusInbox, StatusArchived, StatusAll}

// The notices a page shows after a form, from its ?notice=.
const (
	NoticeSaved      = "saved"
	NoticeArchived   = "archived"
	NoticeUnarchived = "unarchived"
	NoticeRemoved    = "removed"
)

// Refusal is why a request is not answered as asked: its status, the field it
// blames ("" for none), the message — a code of /api/contact-requests or a
// sentence of the backoffice — and, for a 429 or a 503, the Retry-After in
// seconds. Notices, when it holds any, are the ids of the notices a save has
// to acknowledge first: it is then a 409.
type Refusal struct {
	Status     int
	Field      string
	Message    string
	RetryAfter int64
	Notices    []string
}

// refuse is a 400 blaming field.
func refuse(field string, message string) *Refusal {
	return &Refusal{Status: api.StatusBadRequest, Field: field, Message: message}
}

// Fail answers refusal: its Retry-After, then the failure the route returns.
func Fail(sandbox *api.Sandbox, setHeader func(key string, value string), refusal *Refusal) error {
	if refusal.RetryAfter > 0 {
		setHeader("Retry-After", sandbox.Deps.StringsDeps.FormatInt(refusal.RetryAfter, 10))
	}
	return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
}

// failed writes err to stderr and answers a 500 that tells nothing of it.
func failed(sandbox *api.Sandbox, what string, err error) *Refusal {
	sandbox.Deps.StdDeps.Eprintf("warning: %s: %s\n", what, err.Error())
	return &Refusal{Status: api.StatusInternalServerError, Message: "the contact request could not be received"}
}

// NowSeconds is the current instant, in Unix seconds.
func NowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000000
}
