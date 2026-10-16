package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
)

// The access log is the Marco Civil one (Lei 12.965/2014 art. 15): a company
// that runs an internet application for profit in Brazil keeps, for 6
// months, the date and time of every use of it from an IP (art. 5 VIII). Once
// a root turns it on, every page view the visualization records — before the
// ignored IPs, which filter statistics, not the log — and every call to the
// public APIs writes one access-log row: the time in UTC with milliseconds,
// the full IP, the source port of the visitor's connection — behind
// carrier-grade NAT many customers share one IPv4 address (STJ, 3rd Panel,
// 2019) — the kind of the request and the id of the record it wrote, so
// "who sent contact request 9f3a…" has an answer.
//
// It is apart from the records: it keeps the full IP whatever the IP mode,
// has its own retention, is read only by named readers with two-factor
// sign-in, and every read of it is inventoried (Decreto 8.771/2016 art. 13).
// A row is never written in the request's goroutine: it goes on a queue a
// writer drains, held while a backup is taken or restored.

// The kinds of an access-log row.
const (
	AccessView           = "view"
	AccessContactRequest = "contact-request"
	AccessAgent          = "agent"
)

// AccessKinds is every kind of an access-log row.
var AccessKinds = []string{AccessView, AccessContactRequest, AccessAgent}

// Access is one row of the access log, before it is written.
type Access struct {
	// AtMs is when the request came, in Unix milliseconds, UTC; 0 is now.
	AtMs int64
	// Ip is the full ip, IPv6 whole.
	Ip string
	// Port is the source port of the visitor's connection, "" when none
	// was read.
	Port string
	// Kind is AccessView, AccessContactRequest or AccessAgent.
	Kind string
	// Ref is the id of the record the request wrote, "" for none.
	Ref string
}

// accessQueueSize is how many rows wait for the writer at most.
const accessQueueSize = 10000

// accessQueue holds the rows the writer has not written yet.
var accessQueue = make(chan Access, accessQueueSize)

// accessStarted holds a token once StartAccessLog ran.
var accessStarted = make(chan struct{}, 1)

// accessDropped is how many rows found the queue full, guarded by
// accessDroppedLock.
var accessDropped int64

var accessDroppedLock = make(chan struct{}, 1)

// StartAccessLog runs the writer of the access log in a goroutine of its own,
// once per process.
func StartAccessLog(sandbox *api.Sandbox) {
	select {
	case accessStarted <- struct{}{}:
	default:
		return
	}
	go accessWriter(sandbox)
}

// AccessLogRunning tells whether the writer of the access log runs in this
// process: it runs in the one start-server started.
func AccessLogRunning(sandbox *api.Sandbox) bool {
	return len(accessStarted) > 0
}

// AccessDropped is how many rows found the queue full since the process
// started.
func AccessDropped(sandbox *api.Sandbox) int64 {
	accessDroppedLock <- struct{}{}
	defer func() { <-accessDroppedLock }()
	return accessDropped
}

// RecordAccess hands access to the writer when the access log is on. It
// never waits: with the queue full, the row is dropped, counted and warned
// of on stderr — never with its ip.
func RecordAccess(sandbox *api.Sandbox, access Access) {
	if !AccessLogRunning(sandbox) || access.Ip == "" {
		return
	}
	settings, err := CurrentSettings(sandbox)
	if err != nil || !settings.AccessLog {
		return
	}
	if access.AtMs == 0 {
		access.AtMs = NowMilliseconds(sandbox)
	}
	select {
	case accessQueue <- access:
	default:
		accessDroppedLock <- struct{}{}
		accessDropped++
		<-accessDroppedLock
		sandbox.Deps.StdDeps.Eprintf("warning: the access log queue is full: a %s row was dropped\n", access.Kind)
	}
}

// accessWriter is the body of the writer goroutine. It never returns.
func accessWriter(sandbox *api.Sandbox) {
	db := privacy_db.New(sandbox)
	for access := range accessQueue {
		HoldForBackups(sandbox)
		port := int64(0)
		if access.Port != "" {
			if parsed, err := sandbox.Deps.StringsDeps.ParseInt(access.Port, 10, 64); err == nil && parsed > 0 && parsed < 65536 {
				port = parsed
			}
		}
		if _, err := db.AddAccessLog(privacy_db.AccessLogInput{
			At:   access.AtMs,
			Ip:   access.Ip,
			Port: port,
			Kind: access.Kind,
			Ref:  access.Ref,
		}); err != nil {
			sandbox.Deps.StdDeps.Eprintf("warning: an access log row could not be recorded: %s\n", err.Error())
		}
	}
}
