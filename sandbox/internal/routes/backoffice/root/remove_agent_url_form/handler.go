package remove_agent_url_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/remove-agent-url, the Remove button of a
// url: the url of id is removed and the browser sent back to the urls. An id
// no url holds answers the urls under a 404. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	id, err := sandbox.Deps.StringsDeps.ParseInt(sandbox.Deps.StringsDeps.TrimSpace(input.Body.Id), 10, 64)
	if err != nil {
		return backofficerender.RenderAgentUrlsPage(sandbox, response, api.StatusNotFound, props.User, "", "No url has this id.")
	}
	refusal, err := agent.RemoveUrl(sandbox, id)
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficerender.RenderAgentUrlsPage(sandbox, response, refusal.Status, props.User, "", refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, agent.UrlsPath+"?notice="+agent.NoticeRemoved)
}
