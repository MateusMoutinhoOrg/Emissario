package contact

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// A contact request never waits on the database: it is put on queue, and one
// goroutine, the writer, writes it to contact-db. While a backup is taken or
// restored the writer holds the queue, so no database is written under the
// snapshot, and writes it afterwards; the route keeps answering meanwhile. A
// full queue drops the request and counts it; a failed write goes to stderr.

// holdMilliseconds is how long the writer waits between two looks at a
// backup being taken or restored.
const holdMilliseconds = 500

// queue holds the contact requests the writer has not written yet.
var queue = make(chan contact_db.ContactRequestInput, QueueSize)

// started holds a token once Start ran, so a second call starts nothing.
var started = make(chan struct{}, 1)

// dropped is how many contact requests found the queue full since the
// process started, guarded by droppedLock.
var dropped int64

// droppedLock is held while dropped is read or written.
var droppedLock = make(chan struct{}, 1)

// Start runs the writer in a goroutine of its own, once per process: it lives
// as long as the process does and stops with it.
func Start(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	go writer(sandbox)
}

// Running tells whether the writer runs in this process: it runs in the one
// start-server started.
func Running(sandbox *api.Sandbox) bool {
	return len(started) > 0
}

// Dropped is how many contact requests found the queue full since the process
// started.
func Dropped(sandbox *api.Sandbox) int64 {
	droppedLock <- struct{}{}
	defer func() { <-droppedLock }()
	return dropped
}

// enqueue hands request to the writer, and tells whether it took it. It never
// waits: with the queue full, the request is dropped and counted in Dropped.
func enqueue(sandbox *api.Sandbox, request contact_db.ContactRequestInput) bool {
	select {
	case queue <- request:
		return true
	default:
		droppedLock <- struct{}{}
		dropped++
		<-droppedLock
		sandbox.Deps.StdDeps.Eprintf("warning: the contact queue is full: the contact request %s was dropped\n", request.Ref)
		return false
	}
}

// writer is the body of the writer goroutine. It never returns.
func writer(sandbox *api.Sandbox) {
	db := contact_db.New(sandbox)
	for request := range queue {
		holdForBackups(sandbox)
		if _, err := db.AddContactRequest(request); err != nil {
			sandbox.Deps.StdDeps.Eprintf("warning: the contact request %s could not be recorded: %s\n", request.Ref, err.Error())
		}
	}
}

// holdForBackups waits while a backup is taken or restored.
func holdForBackups(sandbox *api.Sandbox) {
	for snapshots.Creating(sandbox) || snapshots.Restoring(sandbox) {
		sandbox.Deps.SleepDeps.Sleep(holdMilliseconds)
	}
}
