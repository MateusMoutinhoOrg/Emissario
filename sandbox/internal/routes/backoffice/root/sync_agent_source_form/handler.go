package sync_agent_source_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// Handle answers POST /admin/root/sync-agent-source, the Sync button of the
// agent source page: its sync loop is asked for a sync now — whatever the
// sync time, and even when the checkout is already at the head of its branch
// — and the browser is sent back to the page at once, which refreshes until
// the sync ends. Only a root reaches it: the backoffice-root-guard answers
// anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	outcome, err := websitesource.RequestSync(sandbox, websitesource.AgentName)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, agent.SourcePath+"?notice="+outcome)
}
