package agent_usage_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/agent-usage, open to every backoffice user: the
// tokens of today against the limits, by day, by ip and the ips at their
// limit — every ip masked to a viewer. A period that does not read answers
// the page under a 400 with the reason.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	root := backofficeauth.Role(props.User.Role) == backofficeauth.RoleRoot
	period, refusal := agent.ParsePeriod(sandbox, input.From, input.To)
	if refusal != nil {
		return backofficerender.RenderAgentUsagePage(sandbox, response, refusal.Status, props.User, period.From, period.To, refusal.Message, agent.UsageReport{})
	}
	report, err := agent.Usage(sandbox, period, root)
	if err != nil {
		return err
	}
	return backofficerender.RenderAgentUsagePage(sandbox, response, api.StatusOK, props.User, period.From, period.To, "", report)
}
