package blocker

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// The IP blocker turns a request away before any route answers it, when the
// ip it came from is one a root blocked: one ip, a CIDR range, a hashed ip —
// as the records keep it in the hashed IP mode — or a whole country, looked
// up in the same country database as the page views. A blocked request is
// answered 403 {"error": "blocked"} by the ip-blocker middleware, which runs
// after the backoffice authentication: a request a session or an API token
// authenticated is never blocked, so a root who blocked their own network
// keeps the session they have, and a block that matches the ip the root adds
// it from is refused. remove-ip-blocks removes them from the command line.
//
// The blocks live in blocker-db and every request reads them from memory,
// read again from the store every cacheSeconds, after every change and after
// a restore. A block may end: the retention cleanup deletes it once it did.
// A hashed ip is matched under the current IP key, so a Rotate IP key leaves
// the hash blocks made before it matching no one.

const (
	// BlocksPath is the page of the blocks, where its forms send the browser
	// back to.
	BlocksPath = "/admin/root/ip-blocks"
	// AddPath is where the add form posts.
	AddPath = "/admin/root/add-ip-block"
	// RemovePath is where the remove form posts.
	RemovePath = "/admin/root/remove-ip-blocks"
)

// The kinds of a block: what its value is.
const (
	KindIp      = "ip"
	KindRange   = "range"
	KindHash    = "hash"
	KindCountry = "country"
)

const (
	// MaxBlocks is the most blocks there may be.
	MaxBlocks = 10000
	// MaxReasonLength is the longest reason of a block.
	MaxReasonLength = 200
	// MaxDays is the longest a block may last, in days.
	MaxDays = 36500
	// MaxIds is how many blocks one removal names.
	MaxIds = 100
	// MaxValueLength is the longest value a form or a JSON body sends.
	MaxValueLength = 64
	// cacheSeconds is how long the requests go by the blocks they read.
	cacheSeconds = 10
	// daySeconds is one day.
	daySeconds = 24 * 60 * 60
)

// CodeBlocked is the error a blocked request is answered with.
const CodeBlocked = "blocked"

// The notices the page shows after a form, from its ?notice=.
const (
	NoticeAdded   = "added"
	NoticeRemoved = "removed"
)

// Refusal is why a request is not answered as asked: its status, the field it
// blames ("" for none) and the message.
type Refusal struct {
	Status  int
	Field   string
	Message string
}

// refuse is a 400 blaming field.
func refuse(field string, message string) *Refusal {
	return &Refusal{Status: api.StatusBadRequest, Field: field, Message: message}
}

// NowSeconds is the current instant, in Unix seconds.
func NowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000000
}
