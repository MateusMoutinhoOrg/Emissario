package api_agent_usage

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
)

// Handle answers POST /api/admin/agent-usage, the JSON twin of the usage
// page, open to every backoffice user: the tokens of today against the
// limits, by day, by ip and the ips at their limit — every ip masked to a
// viewer. A period that does not read is answered 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	root := backofficeauth.Role(props.User.Role) == backofficeauth.RoleRoot
	period, refusal := agent.ParsePeriod(sandbox, input.Body.From, input.Body.To)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	report, err := agent.Usage(sandbox, period, root)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.AgentUsageJSON(sandbox, report))
}
