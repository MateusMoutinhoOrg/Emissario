package agent_start_server

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
)

// Handle runs in front of `start-server` and answers nothing, so start-server
// runs after it. It counts the tokens of today from agent-usage, for the
// limits of the public agent, and starts its goroutines: the writer, which
// turns the usage and the summaries /api/agent/* queues into records of
// agent-db (holding them while a backup is taken or restored), the retention,
// which removes every hour the usage older than 90 days, and the watcher,
// which counts the tokens of today again after a restore. They live as long
// as the process: when start-server stops, they stop.
//
// The agent source is synced by website-source-start-server, with the
// website's.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	agent.Start(sandbox)
	return nil
}
