package api_set_agent_url

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/set-agent-url, the JSON twin of the
// edit url form: the url is checked and written over the one of id — no
// headers keeping the ones saved, clear_headers removing them. An id no url
// holds is answered 404, a refused value 400 with its field. Only a root
// reaches it: the backoffice-api-root-guard answers anyone else under
// /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	url, refusal, err := agent.SetUrl(sandbox, int64(input.Body.Id), agent.UrlInput{
		Name:         input.Body.Name,
		Description:  input.Body.Description,
		Method:       input.Body.Method,
		Url:          input.Body.Url,
		Headers:      sandbox.Deps.StringsDeps.Join(input.Body.Headers, "\n"),
		ClearHeaders: input.Body.ClearHeaders,
	})
	if err != nil {
		return err
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.AgentUrlResponseJSON(sandbox, url))
}
