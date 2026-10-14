package add_agent_url_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/add-agent-url: the url is checked and
// added, and the browser is sent to the urls. A refused value answers the
// form again, filled with what was sent, under a 400 with the reason above
// it. Only a root reaches it: the backoffice-root-guard answers anyone else
// under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	sent := agent.UrlInput{
		Name:        input.Body.Name,
		Description: input.Body.Description,
		Method:      input.Body.Method,
		Url:         input.Body.Url,
		Headers:     input.Body.Headers,
	}
	_, refusal, err := agent.AddUrl(sandbox, sent)
	if err != nil {
		return err
	}
	if refusal != nil {
		form := backofficerender.AgentUrlForm{Name: sent.Name, Description: sent.Description, Method: sent.Method, Url: sent.Url, Headers: sent.Headers}
		return backofficerender.RenderAgentUrlFormPage(sandbox, response, refusal.Status, props.User, false, form, "", refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, agent.UrlsPath+"?notice="+agent.NoticeAdded)
}
