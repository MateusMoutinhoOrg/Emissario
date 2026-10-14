package list_agent_summaries_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/list-agent-summaries, open to every backoffice
// user: one page of the summaries of the period holding q, newest first. A
// period that does not read answers the page under a 400 with the reason.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	period, refusal := agent.ParsePeriod(sandbox, input.From, input.To)
	if refusal != nil {
		return backofficerender.RenderAgentSummariesPage(sandbox, response, refusal.Status, props.User, period.From, period.To, input.Q, refusal.Message, nil, visualization.Paging{}, "", 0)
	}
	summaries, paging, err := agent.ListSummaries(sandbox, period, input.Q, input.Page, input.Limit)
	if err != nil {
		return err
	}
	return backofficerender.RenderAgentSummariesPage(sandbox, response, api.StatusOK, props.User, period.From, period.To, input.Q, "", summaries, paging, input.Notice, input.Removed)
}
