package set_agent_url_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/set-agent-url: the url is checked and
// written over the one of id — blank headers keeping the ones saved — and the
// browser is sent to the urls. A refused value answers the form again under
// its status with the reason above it. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	id, err := sandbox.Deps.StringsDeps.ParseInt(sandbox.Deps.StringsDeps.TrimSpace(input.Body.Id), 10, 64)
	if err != nil {
		return backofficerender.RenderAgentUrlsPage(sandbox, response, api.StatusNotFound, props.User, "", "No url has this id.")
	}
	sent := agent.UrlInput{
		Name:         input.Body.Name,
		Description:  input.Body.Description,
		Method:       input.Body.Method,
		Url:          input.Body.Url,
		Headers:      input.Body.Headers,
		ClearHeaders: input.Body.ClearHeaders == "1" || input.Body.ClearHeaders == "on",
	}
	url, refusal, err := agent.SetUrl(sandbox, id, sent)
	if err != nil {
		return err
	}
	if refusal != nil {
		if refusal.Status == api.StatusNotFound {
			return backofficerender.RenderAgentUrlsPage(sandbox, response, refusal.Status, props.User, "", refusal.Message)
		}
		form := backofficerender.AgentUrlForm{Id: input.Body.Id, Name: sent.Name, Description: sent.Description, Method: sent.Method, Url: sent.Url, Headers: sent.Headers}
		return backofficerender.RenderAgentUrlFormPage(sandbox, response, refusal.Status, props.User, true, form, sandbox.Deps.StringsDeps.Join(agent.HeaderNames(sandbox, url.Headers), ", "), refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, agent.UrlsPath+"?notice="+agent.NoticeSaved)
}
