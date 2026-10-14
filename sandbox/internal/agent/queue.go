package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// A request to the agent never waits on the database: the usage of every
// provider call and every summary are put on queue, and one goroutine, the
// writer, writes them to agent-db. While a backup is taken or restored the
// writer holds the queue, so no database is written under the snapshot, and
// writes it afterwards; the chat keeps answering meanwhile. A full queue
// drops the write and counts it; a failed write goes to stderr.

// holdMilliseconds is how long the writer and the watcher wait between two
// looks at a backup being taken or restored.
const holdMilliseconds = 500

// pending is one write the writer has not done yet: a usage or a summary.
type pending struct {
	Usage   *agent_db.AgentUsageInput
	Summary *agent_db.AgentSummaryInput
}

// queue holds the writes the writer has not done yet.
var queue = make(chan pending, QueueSize)

// started holds a token once Start ran, so a second call starts nothing.
var started = make(chan struct{}, 1)

// dropped is how many writes found the queue full since the process started,
// guarded by droppedLock.
var dropped int64

// droppedLock is held while dropped is read or written.
var droppedLock = make(chan struct{}, 1)

// Running tells whether the writer runs in this process: it runs in the one
// start-server started.
func Running(sandbox *api.Sandbox) bool {
	return len(started) > 0
}

// Dropped is how many writes found the queue full since the process started.
func Dropped(sandbox *api.Sandbox) int64 {
	droppedLock <- struct{}{}
	defer func() { <-droppedLock }()
	return dropped
}

// enqueue hands write to the writer. It never waits: with the queue full, the
// write is dropped and counted in Dropped.
func enqueue(sandbox *api.Sandbox, write pending) {
	select {
	case queue <- write:
	default:
		droppedLock <- struct{}{}
		dropped++
		<-droppedLock
		sandbox.Deps.StdDeps.Eprintf("warning: the agent queue is full: a write was dropped\n")
	}
}

// recordUsage queues the agent-usage of one provider call.
func recordUsage(sandbox *api.Sandbox, ip string, kind string, completion Completion) {
	enqueue(sandbox, pending{Usage: &agent_db.AgentUsageInput{
		At:           NowSeconds(sandbox),
		Ip:           ip,
		Kind:         kind,
		InputTokens:  completion.InputTokens,
		OutputTokens: completion.OutputTokens,
	}})
}

// writer is the body of the writer goroutine. It never returns.
func writer(sandbox *api.Sandbox) {
	db := agent_db.New(sandbox)
	for write := range queue {
		holdForBackups(sandbox)
		if write.Usage != nil {
			if _, err := db.AddAgentUsage(*write.Usage); err != nil {
				sandbox.Deps.StdDeps.Eprintf("warning: an agent usage could not be recorded: %s\n", err.Error())
			}
		}
		if write.Summary != nil {
			if _, err := db.AddAgentSummary(*write.Summary); err != nil {
				sandbox.Deps.StdDeps.Eprintf("warning: the agent summary %s could not be recorded: %s\n", write.Summary.Ref, err.Error())
			}
		}
	}
}

// holdForBackups waits while a backup is taken or restored.
func holdForBackups(sandbox *api.Sandbox) {
	for snapshots.Creating(sandbox) || snapshots.Restoring(sandbox) {
		sandbox.Deps.SleepDeps.Sleep(holdMilliseconds)
	}
}
