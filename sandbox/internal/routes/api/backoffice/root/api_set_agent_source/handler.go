package api_set_agent_source

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
)

// Handle answers POST /api/admin/root/set-agent-source, the JSON twin of the
// agent source form: the repository url, branch and sync time are checked and
// written on the agent source, which its loop syncs on its next tick. A
// refused value is answered 400. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, message := websitesource.Validate(sandbox, websitesource.Settings{
		RepoUrl:  input.Body.RepoUrl,
		Branch:   input.Body.Branch,
		SyncTime: input.Body.SyncTime,
	})
	if message == "" {
		var err error
		message, err = websitesource.Save(sandbox, websitesource.AgentName, settings)
		if err != nil {
			return err
		}
	}
	if message != "" {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "", message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.AgentSourceJSON(sandbox, settings))
}
