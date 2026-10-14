package remove_agent_summaries_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/remove-agent-summaries, the Delete buttons
// of the summaries: one summary by its id, or every summary before a date. A
// deletion answers 303 to the summaries with how many it removed; a refusal
// answers the summaries under its status. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	removed, refusal, err := agent.RemoveSummaries(sandbox, input.Body.Id, input.Body.Before)
	if err != nil {
		return err
	}
	if refusal != nil {
		period, _ := agent.ParsePeriod(sandbox, "", "")
		return backofficerender.RenderAgentSummariesPage(sandbox, response, refusal.Status, props.User, period.From, period.To, "", refusal.Message, nil, visualization.Paging{}, "", 0)
	}
	location := agent.SummariesPath + "?notice=" + agent.NoticeRemoved + "&removed=" + sandbox.Deps.StringsDeps.FormatInt(int64(removed), 10)
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, location)
}
