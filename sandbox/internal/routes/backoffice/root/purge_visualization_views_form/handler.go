package purge_visualization_views_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/purge-visualization-views, the action of the
// Delete views form: before=YYYY-MM-DD removes every view before that local
// day, all=1 every view, and the browser is sent back to the settings page
// with ?notice=purged and how many views went. A date that does not read
// answers the settings page under a 400 with the reason above it. Only a root
// reaches it: the backoffice-root-guard answers anyone else under /admin/root
// first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	removed, _, message, err := visualization.Purge(sandbox, input.Body.Before, input.Body.All == "1")
	if err != nil {
		return err
	}
	if message != "" {
		settings, err := visualization.ReadSettings(sandbox)
		if err != nil {
			return err
		}
		form := backofficerender.VisualizationSettingsFormOf(sandbox, settings)
		return backofficerender.RenderVisualizationSettingsPage(sandbox, response, api.StatusBadRequest, props.User, form, "", 0, message, props.ClientIp)
	}
	location := visualization.SettingsPath + "?notice=" + visualization.NoticePurged + "&removed=" + sandbox.Deps.StringsDeps.FormatInt(int64(removed), 10)
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, location)
}
