package list_agent_urls_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/list-agent-urls: every url the public agent
// can call, by name, with the names of its headers — never their values — and
// the warning that whatever the agent can call, any visitor can. Only a root
// reaches it: the backoffice-root-guard answers anyone else under /admin/root
// first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	return backofficerender.RenderAgentUrlsPage(sandbox, response, api.StatusOK, props.User, input.Notice, "")
}
