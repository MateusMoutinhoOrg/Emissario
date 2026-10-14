package set_agent_settings_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/set-agent-settings with the form POST
// /admin/root/set-agent-settings reads, filled with the settings saved — the
// api key never — beside the state of the agent. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, err := agent.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	return backofficerender.RenderAgentSettingsPage(sandbox, response, api.StatusOK, props.User, backofficerender.AgentSettingsFormOf(sandbox, settings), input.Notice, "")
}
