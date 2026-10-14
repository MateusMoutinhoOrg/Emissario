package set_agent_url_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/set-agent-url?id= with the form POST
// /admin/root/set-agent-url reads, filled with the url saved — its headers by
// name alone. An id no url holds answers the urls under a 404. Only a root
// reaches it: the backoffice-root-guard answers anyone else under /admin/root
// first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	url, ok := agent.FindUrl(sandbox, int64(input.Id))
	if !ok {
		return backofficerender.RenderAgentUrlsPage(sandbox, response, api.StatusNotFound, props.User, "", "No url has this id.")
	}
	saved := sandbox.Deps.StringsDeps.Join(agent.HeaderNames(sandbox, url.Headers), ", ")
	return backofficerender.RenderAgentUrlFormPage(sandbox, response, api.StatusOK, props.User, true, backofficerender.AgentUrlFormOf(sandbox, url), saved, "")
}
