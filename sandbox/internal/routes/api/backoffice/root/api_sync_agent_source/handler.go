package api_sync_agent_source

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficesnapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// Handle answers POST /api/admin/root/sync-agent-source, the JSON twin of the
// Sync button of the agent source: its sync loop is asked for a sync now and
// {"status": "syncing"} is answered 202 at once, while it runs in the
// background. No agent source saved yet is answered 409, and no sync loop
// running in this process 503. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	outcome, err := websitesource.RequestSync(sandbox, websitesource.AgentName)
	if err != nil {
		return err
	}
	switch outcome {
	case websitesource.OutcomeNoSource:
		return sandbox.Deps.OpinionatedAgnosServer.WriteError(sandbox.Deps.SerializableDeps, *response, api.StatusConflict, "", websitesource.NoSourceMessage(sandbox, websitesource.AgentName))
	case websitesource.OutcomeNotRunning:
		return sandbox.Deps.OpinionatedAgnosServer.WriteError(sandbox.Deps.SerializableDeps, *response, api.StatusServiceUnavailable, "", websitesource.NotRunningMessage)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, backofficesnapshots.StatusAccepted, backofficeapi.JobJSON(sandbox, "syncing"))
}
