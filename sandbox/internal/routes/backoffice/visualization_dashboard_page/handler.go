package visualization_dashboard_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/visualization, open to every backoffice user,
// with backoffice/visualization_dashboard.html: the cards, the views over time
// and the top tables of the views the filters of the query string keep — the
// last 7 days when it names no period. A filter that does not read answers the
// page under a 400, the reason above the filter bar; an ip sent by a viewer
// answers the forbidden page.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	raw := visualization.RawFilter{
		From:    input.From,
		To:      input.To,
		Path:    input.Path,
		Prefix:  input.Prefix,
		Ip:      input.Ip,
		Country: input.Country,
		Device:  input.Device,
		Browser: input.Browser,
		Os:      input.Os,
		Status:  input.Status,
		Bots:    input.Bots,
	}
	root := backofficeauth.Role(props.User.Role) == backofficeauth.RoleRoot
	filter, refusal := visualization.ParseFilter(sandbox, raw, root)
	if refusal != nil && refusal.Status == api.StatusForbidden {
		return backofficerender.RenderForbiddenPage(sandbox, response, props.User)
	}
	if refusal != nil {
		return backofficerender.RenderVisualizationDashboardPage(sandbox, response, api.StatusBadRequest, props.User, raw, filter, refusal.Message, visualization.Metrics{})
	}
	selection, err := visualization.Select(sandbox, filter)
	if err != nil {
		return err
	}
	bucket, _ := visualization.ChooseBucket(sandbox, filter, "")
	metrics := visualization.Measure(sandbox, selection, bucket, visualization.DefaultTop, root)
	return backofficerender.RenderVisualizationDashboardPage(sandbox, response, api.StatusOK, props.User, raw, filter, "", metrics)
}
