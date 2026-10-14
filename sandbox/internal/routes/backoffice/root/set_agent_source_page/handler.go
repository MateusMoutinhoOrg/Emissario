package set_agent_source_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// Handle answers GET /admin/root/set-agent-source with the form POST
// /admin/root/set-agent-source reads, filled with the repository url, branch
// and sync time of the agent source — or the defaults, before a root saved
// any — beside what the last sync did. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, ok, err := websitesource.Read(sandbox, websitesource.AgentName)
	if err != nil {
		return err
	}
	if !ok {
		settings = websitesource.Defaults(sandbox)
	}
	return backofficerender.RenderAgentSourcePage(sandbox, response, api.StatusOK, props.User, settings, input.Notice, "")
}
