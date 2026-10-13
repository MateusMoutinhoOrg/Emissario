package api_get_visualization_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/root/get-visualization-settings with the
// settings of the visualization beside what the page shows read only:
// {settings, geoip, stored, oldest_at, dropped}. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, err := visualization.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	stored, oldest, err := visualization.StoredViews(sandbox)
	if err != nil {
		return err
	}
	document := backofficeapi.VisualizationSettingsStateJSON(sandbox, settings, visualization.CurrentGeoip(sandbox), stored, oldest, visualization.Dropped(sandbox))
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
