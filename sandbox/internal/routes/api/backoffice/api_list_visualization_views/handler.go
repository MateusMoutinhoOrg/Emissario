package api_list_visualization_views

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/list-visualization-views, open to every
// backoffice user, with one page of the views the filters of the body keep,
// newest first: {views, total, page, pages, limit}, each ip masked to a
// viewer.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	root := backofficeauth.Role(props.User.Role) == backofficeauth.RoleRoot
	filter, refusal := visualization.ParseFilter(sandbox, visualization.RawFilter{
		From:    input.Body.From,
		To:      input.Body.To,
		Path:    input.Body.Path,
		Prefix:  input.Body.Prefix,
		Ip:      input.Body.Ip,
		Country: input.Body.Country,
		Device:  input.Body.Device,
		Browser: input.Body.Browser,
		Os:      input.Body.Os,
		Status:  visualization.StatusParam(sandbox, input.Body.Status),
		Bots:    visualization.BotsParam(sandbox, input.Body.Bots),
	}, root)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	selection, err := visualization.Select(sandbox, filter)
	if err != nil {
		return err
	}
	views, paging := visualization.ListViews(sandbox, selection, input.Body.Page, input.Body.Limit, root)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.VisualizationViewsJSON(sandbox, views, paging))
}
