package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// retentionMilliseconds is how often the retention removes the old usage.
const retentionMilliseconds = 60 * 60 * 1000

// watchMilliseconds is how often the watcher looks for a restore.
const watchMilliseconds = 2000

// Start counts the tokens of today from agent-usage, then runs the writer, the
// retention and the watcher, each in a goroutine of its own, once per process:
// they live as long as the process does and stop with it.
func Start(sandbox *api.Sandbox) {
	select {
	case started <- struct{}{}:
	default:
		return
	}
	if err := LoadTally(sandbox); err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the agent usage of today could not be read: %s\n", err.Error())
	}
	go writer(sandbox)
	go retain(sandbox)
	go watch(sandbox)
}

// retain is the body of the retention goroutine: at the start and every hour,
// it removes the agent-usage older than UsageRetentionDays — they hold an ip.
// Summaries hold none, and stay until a root removes them. It never returns.
func retain(sandbox *api.Sandbox) {
	for {
		holdForBackups(sandbox)
		removed, err := RemoveExpiredUsage(sandbox)
		if err != nil {
			sandbox.Deps.StdDeps.Eprintf("warning: the old agent usage could not be removed: %s\n", err.Error())
		} else if removed > 0 {
			sandbox.Deps.StdDeps.Logf("agent: removed %d usage records older than %d days\n", removed, UsageRetentionDays)
		}
		sandbox.Deps.SleepDeps.Sleep(retentionMilliseconds)
	}
}

// RemoveExpiredUsage removes the agent-usage older than UsageRetentionDays,
// and answers how many.
func RemoveExpiredUsage(sandbox *api.Sandbox) (int, error) {
	cutoff := NowSeconds(sandbox) - UsageRetentionDays*24*60*60
	return agent_db.RemoveAgentUsagesBefore(sandbox, agent_db.New(sandbox), cutoff)
}

// watch is the body of the watcher goroutine: once a restore ends, it counts
// the tokens of today again from the database restored. It never returns.
func watch(sandbox *api.Sandbox) {
	for {
		if snapshots.Restoring(sandbox) {
			for snapshots.Restoring(sandbox) {
				sandbox.Deps.SleepDeps.Sleep(holdMilliseconds)
			}
			if err := LoadTally(sandbox); err != nil {
				sandbox.Deps.StdDeps.Eprintf("warning: the agent usage of today could not be read after the restore: %s\n", err.Error())
			}
		}
		sandbox.Deps.SleepDeps.Sleep(watchMilliseconds)
	}
}
