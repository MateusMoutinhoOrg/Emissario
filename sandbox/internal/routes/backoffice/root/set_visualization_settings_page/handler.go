package set_visualization_settings_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/root/set-visualization-settings with the form
// that POST /admin/root/set-visualization-settings reads, filled with the
// settings a root saved — or the defaults: tracking on, 90 days, no ip ignored
// — beside the country database in use, what is stored and the ip of the
// request. Only a root reaches it: the backoffice-root-guard answers anyone
// else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, err := visualization.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	form := backofficerender.VisualizationSettingsFormOf(sandbox, settings)
	return backofficerender.RenderVisualizationSettingsPage(sandbox, response, api.StatusOK, props.User, form, input.Notice, input.Removed, "", props.ClientIp)
}
