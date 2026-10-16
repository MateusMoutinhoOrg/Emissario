package get_agent_summary_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/get-agent-summary?id=, open to every backoffice
// user: the summary whole, with the button that deletes it for a root. An id
// no summary holds answers the summaries under a 404.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	summary, ok := agent.FindSummary(sandbox, input.Id)
	if !ok {
		period, _ := agent.ParsePeriod(sandbox, "", "")
		return backofficerender.RenderAgentSummariesPage(sandbox, response, api.StatusNotFound, props.User, period.From, period.To, "", "No summary has this id.", nil, visualization.Paging{}, "", 0)
	}
	return backofficerender.RenderAgentSummaryPage(sandbox, response, props.User, summary, input.Notice, input.Reason)
}
